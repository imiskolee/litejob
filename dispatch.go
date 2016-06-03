package litejob

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type Dispatch struct {
	sync.Mutex
	ID        string
	queue     Queue
	configure *Configure
	handlers  map[string]JobHandler
	running   int
	count     int
}

func NewDispatch(configure *Configure) (*Dispatch, error) {

	dispatch := new(Dispatch)

	dispatch.configure = configure

	queue, err := GetQueue(dispatch.configure.QueueEngine, dispatch.configure.QueueConfigure)

	if err != nil {
		return nil, err
	}
	dispatch.queue = queue
	dispatch.handlers = make(map[string]JobHandler, 0)
	return dispatch, nil
}

func (dispatch *Dispatch) Register(name string, handler JobHandler, max int) {
	dispatch.Lock()
	dispatch.handlers[name] = handler
	dispatch.queue.RegisterJob(name, max)
	dispatch.Unlock()
}

func (dispatch *Dispatch) Push(name, message string) (*Job, error) {

	job := NewJob(name, message)
	err := dispatch.queue.Push(job)
	return job, err
}

func (dispatch *Dispatch) Run() {
	dispatch.loop()
}

func (dispatch *Dispatch) GetAllQueue() map[string]*JobQueueState {
	return dispatch.queue.GetAllQueue()
}

func (dispatch *Dispatch) Monitor() map[string]map[string]interface{} {
	return dispatch.queue.Monitor()
}

func (dispatch *Dispatch) loop() {

	for {
		jobs, err := dispatch.queue.PopN(10)

		if err != nil {
			goto next
		}

		if len(jobs) < 1 {
			goto next
		}

		dispatch.Lock()
		for _, job := range jobs {
			dispatch.running++
			go dispatch.runJob(job)
		}
		dispatch.Unlock()
		runtime.Gosched()
		continue

	next:
		runtime.Gosched()
		time.Sleep(dispatch.configure.HeartInterval)
	}
}

func (dispatch *Dispatch) runJob(job Job) {

	defer func() {
		dispatch.running--
		dispatch.queue.FlushJob(&job)
	}()

	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	handler, ok := dispatch.handlers[job.Name]

	if !ok {
		fmt.Fprintf(os.Stdout, "[litejob] unregisterd job:"+job.Name)
		return
	}

	if dispatch.configure.Before != nil {
		dispatch.configure.Before(&job)
	}

	startTime := time.Now()
	ret := handler(&job)
	endTime := time.Now()
	job.Status = ret.Status

	if dispatch.configure.After != nil {
		dispatch.configure.After(&job)
	}

	runTime := endTime.Sub(startTime)
	waitTime := endTime.Sub(job.Created)

	if ret.Status == JobStatusAgain {
		if job.ReplyCount < dispatch.configure.MaxReplyCount {
			job.Status = ret.Status
			job.ReplyCount++
			dispatch.queue.Push(&job)
		} else {
			ret.Status = JobStatusFailed
		}
	}

	dispatch.Lock()
	dispatch.count++
	fmt.Fprintf(os.Stdout, "[%s] %6d %32s %32s %s %s %s %s\n", endTime.Format(time.ANSIC), dispatch.count, job.Name, job.ID, ret.Status, runTime, waitTime, ret.Msg)
	dispatch.Unlock()

	dispatch.queue.UpdateJobStatus(job.Name, ret.Status, ret.Msg)

}
