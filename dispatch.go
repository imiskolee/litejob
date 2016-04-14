package litejob

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type JobCallbackConfigure struct {
	After  JobCallback
	Before JobCallback
}

type EngineConfigure map[string]interface{}

func (this *EngineConfigure) Get(name string, defaultVal interface{}) interface{} {
	if v, ok := (*this)[name]; ok {
		return v
	}
	return defaultVal
}

// configure dispatch
type DispatchConfigure struct {
	Engine          string               // storage engine name.eg:(redis,sqlite,memory)
	MaxConcurrency  uint32               // max concurrency goroutine
	MaxReplyCount   uint32               // max reply job to job list when job ask need again
	HeartInterval   time.Duration        // sleep some time when empty job list
	DumpInterval    time.Duration        // not used
	Logfile         string               // log file path
	EngineConfigure EngineConfigure      // engine configure
	Callback        JobCallbackConfigure // job callbacks
}

//dispatch engine status for monitor
type DispatchStatus struct {
	Running uint32 //runing goroutine count
	Len     uint32 //job list length
}

//
type Dispatch struct {
	sync.Mutex
	configure   *DispatchConfigure
	storage     Storage
	running     uint32
	handlerList map[string]JobHandler
	log         *Log
	Count       uint32
}

func NewDispatch(configure *DispatchConfigure) (*Dispatch, error) {

	dispatch := new(Dispatch)

	dispatch.configure = configure

	dispatch.handlerList = make(map[string]JobHandler, 0)
	dispatch.log = NewLog(dispatch.configure.Logfile)

	storage, err := GetStorage(dispatch.configure.Engine, dispatch.configure)

	if err != nil {
		return nil, err
	}

	dispatch.storage = storage
	return dispatch, nil
}

func (this *Dispatch) RegisterHandler(name string, handler JobHandler) {
	this.handlerList[name] = handler
}

func (this *Dispatch) JobNew(name string, param interface{}) (*Job, error) {

	job := &Job{
		Id:         Guid(),
		Name:       name,
		Param:      param,
		Status:     JobStatusWating,
		CreateTime: time.Now(),
	}

	err := this.storage.JobPush(job)

	if err != nil {
		this.log.Error("push job error:" + err.Error())
	}

	return job, err
}

func (this *Dispatch) Start() {

	this.Loop()
}

func (this *Dispatch) Loop() {

	for {
		//队列非空
		if this.storage.JobLen() > 0 {
			if this.running < this.configure.MaxConcurrency {
				this.Lock()
				this.running++
				this.Unlock()
				go this.next()
				//remise CPU
			}
			runtime.Gosched()
			continue

		}

		time.Sleep(this.configure.HeartInterval)
		runtime.Gosched()
	}
}

func (this *Dispatch) JobState(jobId string) (*JobState, error) {

	return this.storage.JobState(jobId)
}

func (this *Dispatch) next() {

	defer func() {
		this.Lock()
		this.running--
		this.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			this.log.Error("run job error:" + fmt.Sprint(r))
			debug.PrintStack()
		}
	}()

	job, err := this.storage.JobPop()

	if err != nil {
		return
	}

	job.Status = JobStatusDoing

	if this.configure.Callback.Before != nil {
		this.configure.Callback.Before(job)
	}

	handler, ok := this.handlerList[job.Name]

	if ok == false {
		this.log.Error("unknow handler %s", job.Name)
		return
	}

	startTime := time.Now()

	ret := handler(job)

	endTime := time.Now()

	job.Status = ret.Status

	if this.configure.Callback.After != nil {
		this.configure.Callback.After(job)
	}

	if ret.Status == JobStatusAgain {
		if job.ReplyCount < this.configure.MaxReplyCount {
			job.ReplyCount++
			this.storage.JobPush(job)
		} else {
			ret.Status = JobStatusFailed
			ret.Msg = "over max reply times."
		}
	}

	state := &JobState{
		JobId:    job.Id,
		Status:   ret.Status.String(),
		Msg:      ret.Msg,
		WaitTime: int(endTime.Sub(job.CreateTime).Seconds()),
		RunTime:  int(endTime.Sub(startTime).Nanoseconds() / 1000000),
	}

	this.storage.JobStateUpdate(state)

	this.Count++

	this.log.Normal("%6d %3d %32s %32s %10s %8dms %8ds %s", this.Count, this.running, job.Name, job.Id, ret.Status, state.RunTime, state.WaitTime, ret.Msg)
}
