package litejob

import (
	"errors"
)

var queues map[string]QueueInitFunc

type QueueInitFunc func(configure *QueueConfigure) Queue

type JobQueueState struct {
	Name    string
	Len     int
	Work    bool
	Max     int
	Running int
}

//Queue 任务队列
type Queue interface {
	Push(job *Job) error
	PopN(max int) ([]Job, error)
	RegisterJob(name string, max int)
	UpdateJobStatus(name string, status JobStatus, msg string)
	GetAllQueue() map[string]*JobQueueState
	FlushJob(job *Job)
	Monitor() map[string]map[string]interface{}
}

func GetQueue(name string, configure *QueueConfigure) (Queue, error) {

	initFunc, ok := queues[name]

	if !ok {
		return nil, errors.New("unregisterd queue:" + name)
	}

	return initFunc(configure), nil
}

func RegisterQueue(name string, initFunc QueueInitFunc) {
	queues[name] = initFunc
}

func init() {
	queues = make(map[string]QueueInitFunc, 0)
}
