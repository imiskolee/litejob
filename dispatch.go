package litejob

import (
	"time"
	"runtime"
	"fmt"
)

type JobCallbackConfigure struct {

	After		JobCallback
	Before		JobCallback
}

type EngineConfigure map[string]interface{}

func (this *EngineConfigure)Get(name string,defaultVal interface{}) interface{} {
	if v,ok := (*this)[name];ok {
		return v
	}
	return defaultVal
}

// configure dispatch
type DispatchConfigure struct {

	Engine 				string 								// storage engine name.eg:(redis,sqlite,memory)
	MaxConcurrency		uint32 								// max concurrency goroutine
	MaxReplyCount 		uint32 								// max reply job to job list when job ask need again
	HeartInterval		time.Duration						// sleep some time when empty job list
	DumpInterval 		time.Duration   					// not used
	Logfile 			string								// log file path
	EngineConfigure 	EngineConfigure						// engine configure
	Callback 			JobCallbackConfigure				// job callbacks
}

//dispatch engine status for monitor
type DispatchStatus struct {
	Running 	uint32		//runing goroutine count
	Len 		uint32		//job list length
}

//
type Dispatch struct{

	configure *DispatchConfigure
	storage   Storage
	running   uint32
	handlerList map[string]JobHandler
	log            *Log
}

func NewDispatch(configure *DispatchConfigure)(*Dispatch,error){

	dispatch := new(Dispatch)

	dispatch.configure = configure

	dispatch.handlerList = make(map[string]JobHandler,0)
	dispatch.log = NewLog(dispatch.configure.Logfile)

	storage,err := GetStorage(dispatch.configure.Engine,dispatch.configure)

	if err != nil {
		return nil,err
	}

	dispatch.storage = storage
	return dispatch,nil
}


func (this *Dispatch)RegisterHandler(name string,handler JobHandler) {
	this.handlerList[name] = handler
}

func (this *Dispatch)JobNew(name string,param interface{}) (*Job,error) {

	job := &Job{
		Id:Guid(),
		Name:name,
		Param:param,
		Status:JobStatusWating,
		CreateTime:time.Now(),
	}

	err := this.storage.JobPush(job)

	if err != nil {
		this.log.Error("push job error:" + err.Error())
	}

	return job,err
}

func (this *Dispatch)Start() {

	this.Loop()
}

func (this *Dispatch)Loop() {

	for {

		if this.running < this.configure.MaxConcurrency && this.storage.JobLen() > 0 {
			this.running++
			go this.next()
			//remise CPU
			runtime.Gosched()

			//sleep 2ms when system is busy.
			if this.running > uint32(this.configure.MaxConcurrency / 2) {
				time.Sleep(5 * time.Millisecond)
			}

			continue
		}

		time.Sleep(this.configure.HeartInterval)
	}
}

func (this *Dispatch)next(){

	defer func(){
		this.running--
	}()

	defer func() {
		if r := recover(); r != nil {
			this.log.Error("run job error:" + fmt.Sprint(r))
		}
	}()

	job,err := this.storage.JobPop()

	if err != nil {
		return
	}

	if this.configure.Callback.Before != nil {
		this.configure.Callback.Before(job)
	}

	handler,ok := this.handlerList[job.Name]

	if ok == false {
		this.log.Error("unknow handler %s",job.Name)
		return
	}

	startTime := time.Now()

	ret := handler(job.Param)

	endTime := time.Now()

	job.Status = ret.Status

	if this.configure.Callback.After != nil {
		this.configure.Callback.After(job)
	}


	if ret.Status == JobStatusAgain && job.replyCount < this.configure.MaxReplyCount{
		job.replyCount++
		this.storage.JobPush(job)
	}

	this.log.Normal("%40s %2d %8dms %8ds %s", job.Name, ret.Status, int(endTime.Sub(startTime).Seconds()), int(endTime.Sub(job.CreateTime).Seconds()),ret.Msg)
}
