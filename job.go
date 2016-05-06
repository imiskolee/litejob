package litejob

import (
	"encoding/json"
	"time"
)

//JobStatus 代表任务状态
type JobStatus string

const (
	//JobStatusWaiting 代表任务已经成功加入队列，正在等待执行
	JobStatusWaiting JobStatus = "WAITING"
	//JobStatusSuccess 表示任务已经执行成功
	JobStatusSuccess JobStatus = "SUCCESS"
	//JobStatusFailed 表示任务在逻辑上执行失败
	JobStatusFailed JobStatus = "FAILED"
	//JobStatusAgain 表示任务正在重试
	JobStatusAgain JobStatus = "AGAIN"
	//JobStatusUnknow 表示未知任务状态
	JobStatusUnknow JobStatus = "UNKNOW"
	//JobStatusError 任务失败（参数异常，网络异常等）
	JobStatusError JobStatus = "ERROR"
)

//JobState 代表任务完整状态
//JobState 不会被永久存储
type JobState struct {
	ID         string    //任务ID
	Status     JobStatus //任务状态
	Created    time.Time //任务创建时间
	LastUpdate time.Time //任务状态最后更新时间
	Msg        string    //任务返回的信息
}

func (state *JobState) MarshalBinary() ([]byte, error) {
	return json.Marshal(state)
}

func (state *JobState) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, state)
}

type Job struct {
	ID         string
	Name       string
	Message    string
	Created    time.Time
	Status     JobStatus
	ReplyCount int
}

func NewJob(name, message string) *Job {
	job := new(Job)
	job.ID = GUID()
	job.Name = name
	job.Message = message
	job.Created = time.Now()
	job.Status = JobStatusWaiting
	return job
}

func (state *Job) MarshalBinary() ([]byte, error) {
	return json.Marshal(state)
}

func (state *Job) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, state)
}

type Return struct {
	Status JobStatus
	Msg    string
}
