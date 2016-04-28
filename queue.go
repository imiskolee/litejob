package litejob

import "errors"

var queues map[string]QueueInitFunc

type QueueInitFunc func(configure *QueueConfigure) Queue

//Queue 任务队列
type Queue interface {
	Push(job *Job) error
	PopN(max int) ([]Job, error)
	RegisterJob(name string, max int)
	UpdateJobStatus(name string, status JobStatus, msg string)
	FlushJob(job *Job)
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
