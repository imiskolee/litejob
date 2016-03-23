package litejob

import (
	"sync"
	"time"
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
}

func NewLocalDispatch(configure *DispatchConfigure) *LocalDispatch {

	dispatch := new(LocalDispatch)
	dispatch.configure = configure
	dispatch.handlerList = make(map[string]JobHandler, 32)
	dispatch.JobList = make([]Job, 1024)
	dispatch.jobStoreageLen = uint32(len(dispatch.JobList))
	dispatch.log = NewLog(configure.Logfile)
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

func (this *LocalDispatch)JobNew(name string,param interface{}) bool {

	job := Job{
		Name:name,
		Param:param,
		CreateTime:time.Now(),
		Status : JobStatusWating,
	}

	return this.JobPush(job)
}

func (this *LocalDispatch)JobPush(job Job) bool {
	this.Lock()
	if this.jobStoreageLen - this.jobLen < 100 {
		this.scaleStorage()
	}else{
		this.scaleStorage()
	}

	this.JobList[this.jobLen] = job
	this.jobLen++
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

	for {

		if this.running < this.configure.MaxConcurrency  && this.next() {
			continue;
		}
		time.Sleep(this.configure.HeartInterval)
	}
}

func (this *LocalDispatch)Start() bool {

	if ok := this.Load(); ok != true {
		return false
	}

	this.loop()

	return true
}


func (this *LocalDispatch)Stop() bool{

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
	job := this.JobList[0]
	this.JobList = this.JobList[1:]
	go  this.do(job)
	this.jobLen--
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
	}

}

//从内存dump到本地
func (this *LocalDispatch)Dump() bool {
	return true
}

func (this *LocalDispatch)Load() bool {

	return true
}










