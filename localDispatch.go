package litejob

import (
	"sync"
	"time"
	"runtime"
	"fmt"
	"os"
	"encoding/gob"
)

const (
	EngineLocal = "local"
)

type LocalDispatch struct {
	sync.Mutex
	configure      *DispatchConfigure
	running        uint32
	handlerList    map[string]JobHandler
	JobList        []Job
	jobLen         uint32
	jobStoreageLen uint32
	log            *Log

	state 		  bool

}

func NewLocalDispatch(configure *DispatchConfigure) *LocalDispatch {

	dispatch := new(LocalDispatch)
	dispatch.configure = configure
	dispatch.handlerList = make(map[string]JobHandler, 32)
	dispatch.JobList = make([]Job, 1024)
	dispatch.jobStoreageLen = uint32(len(dispatch.JobList))
	dispatch.log = NewLog(configure.Logfile)

	dispatch.state = true


	return dispatch
}

func (this *LocalDispatch)scaleStorage() {

	if this.jobStoreageLen < 1024 {
		jobs := make([]Job, this.jobStoreageLen)
		this.JobList = append(this.JobList, jobs...)
		this.jobStoreageLen += this.jobStoreageLen
	}else {
		jobs := make([]Job, 1024)
		this.JobList = append(this.JobList, jobs...)
		this.jobStoreageLen += 1024
	}
}

func (this *LocalDispatch)RegisterHandler(name string, handler JobHandler) bool {
	this.handlerList[name] = handler
	return true
}

func (this *LocalDispatch)JobNew(name string,param interface{}) string {

	job := Job{
		Id:Guid(),
		Name:name,
		Param:param,
		CreateTime:time.Now(),
		Status : JobStatusWating,
	}

	this.JobPush(job)
	return job.Id
}


func (this *LocalDispatch)JobPush(job Job) bool {
	this.Lock()
	if this.jobStoreageLen - this.jobLen < 100 {
		this.scaleStorage()
	}else if(this.jobStoreageLen - this.jobLen > 1024){
		list := make([]Job,0)
		list = append(list,this.JobList[0:this.jobLen+1]...)
		this.JobList = list
		this.jobStoreageLen = uint32(len(this.JobList))
	}
	this.JobList[this.jobLen] = job
	this.jobLen++

	fmt.Fprintf(os.Stderr,"%d %d %d\n",this.jobLen,this.jobStoreageLen,len(this.JobList))

	this.Unlock()
	return true
}

func (this *LocalDispatch)JobLen() uint32 {
	return uint32(len(this.JobList))
}

func (this *LocalDispatch)FlushJob() bool {
	this.Lock()
	this.JobList = make([]Job,0)
	this.Unlock()
	return true
}

func (this *LocalDispatch)Status() DispatchStatus {
	return DispatchStatus{
		Len:this.jobLen,
		Running:this.running,
	}
}

func (this *LocalDispatch)loop() {

	lastDumpTime := time.Now()

	for {


		if ! this.state {
			break
		}

 		now := time.Now()

		if now.Sub(lastDumpTime) > (60 * time.Second) {
			this.Dump()
		}

		if this.running < this.configure.MaxConcurrency  && this.next() {
			time.Sleep(10 * time.Microsecond)
			runtime.GC()
			runtime.Gosched()
			continue;
		}

		time.Sleep(this.configure.HeartInterval)

	}
}

func (this *LocalDispatch)Start() bool {

	if ok := this.Load(); ok != true {
		return false
	}

	for{

		if ! this.state {
			break
		}
		this.loop()
		time.Sleep(10 * time.Second)
	}

	return true
}

func (this *LocalDispatch)Stop() bool{
	this.state = false
	this.Dump()
	return false
}

func (this *LocalDispatch)next() bool {

	//return false

	defer func() {
		if r := recover(); r != nil {
			this.log.Error("run job error")
		}
	}()

	if this.jobLen < 1 {
		return false
	}
	this.Lock()
	this.jobLen--
	job := this.JobList[0]
	this.JobList = this.JobList[1:]
	this.jobStoreageLen--
	go  this.do(job)
	this.Unlock()

	return true
}

func (this *LocalDispatch)do(job Job) {

	this.running++

	defer func(){this.running--}()

	handler, ok := this.handlerList[job.Name]

	if !ok {
		this.log.Error("unknow handler name:" + job.Name)
		return
	}

	start := time.Now()
	ret := handler(job.Param)
	end := time.Now()

	runTime := int(end.Sub(start).Seconds() * 1000)
	watingTime := int(end.Sub(job.CreateTime).Seconds())

	this.log.Normal("%40s %2d %8dms %8ds", job.Name, ret.Status, runTime, watingTime)

	if ret.Status == JobStatusAgain && job.replyCount < this.configure.MaxReplyCount {
		job.replyCount++
		this.JobPush(job)
	}else{
		if this.configure.JobCallback != nil {
			this.configure.JobCallback(job.Id,job.Name,int(ret.Status),ret.Msg)
		}
	}
}

//从内存dump到本地
func (this *LocalDispatch)Dump() bool {

	if this.jobLen < 1 {
		return true
	}

	file,ok := this.configure.EngineConfigure["dump_file"]

	if !ok {
		return false
	}

	this.Lock()

	f,err := os.Create(file.(string))

	if err != nil {
		return false
	}

	enc := gob.NewEncoder(f)
	err = enc.Encode(this.JobList[0:this.jobLen])

	if err != nil {
		fmt.Fprintln(os.Stderr,"[litejob] dump error:" + err.Error())
			return false
	}

	this.Unlock()
	return true
}



func (this *LocalDispatch)Load() bool {

	file,ok := this.configure.EngineConfigure["dump_file"]

	if !ok {
		return false
	}

	this.Lock()

	f,err := os.Open(file.(string))

	if err != nil {
		return false
	}

	enc := gob.NewDecoder(f)

	err = enc.Decode(&this.JobList)
	this.jobLen = uint32(len(this.JobList))
	this.jobStoreageLen = this.jobLen

	if err != nil {
		fmt.Fprintln(os.Stderr,"load file error:" + err.Error())
	}

	this.Unlock()
	return true
}

func init(){
	gob.Register(map[string]interface{}{})
}








