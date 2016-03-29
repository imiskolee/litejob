package litejob

import (
	"time"
	"runtime"
	"fmt"
)

const (
	EngineRedis 	= "redis"
	EngineSqlite 	= "sqlite"
	EngineMemory 	= "memory"
)

type JobCallbackConfigure struct {

	After		JobCallback
	Before		JobCallback
}

//任务调度器的控制项
type DispatchConfigure struct {

	Engine 				string 								//队列引擎
	MaxConcurrency		uint32 								//最大并发数
	MaxReplyCount 		uint32 								//最大重试次数
	HeartInterval		time.Duration						//心跳时间
	DumpInterval 		time.Duration   					//数据持久化间隔时间
	Logfile 			string								//日志文件,默认为/std/stderr
	EngineConfigure 	map[string]interface{}				//引擎相关的配置项,localStorage支持dump_file参数,用于持久化内存信息
	Callback 			JobCallbackConfigure				//任务回调方法,每执行一次任务即调用一次该方法
}



//调度器状态
type DispatchStatus struct {
	Running 	uint32		//正在运行的任务数
	Len 		uint32		//当前任务长度

}


//任务调度器
type Dispatch struct{

	configure *DispatchConfigure
	storage   Storage
	running   uint32
	handlerList map[string]JobHandler
	log            *Log
}

func NewDispatch(configure *DispatchConfigure)(*Dispatch){

	dispatch := new(Dispatch)

	dispatch.configure = configure

	dispatch.handlerList = make(map[string]JobHandler,0)
	dispatch.log = NewLog(dispatch.configure.Logfile)

	switch dispatch.configure.Engine {

	case EngineRedis : {
		dispatch.storage = NewRedisStorage(configure)
	}
	}
	return dispatch
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
			runtime.Gosched()
			time.Sleep(5 * time.Millisecond)
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
