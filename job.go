package litejob

import "time"

type JobStatus uint

const (

	//任务已经添加成功,等待执行
	JobStatusWating 	JobStatus = 1
	//任务正在执行
	JobStatusDoing 		JobStatus = 2
	//任务执行失败
	JobStatusFailed		JobStatus = 3
	//任务执行成功
	JobStatusSuccess 	JobStatus = 4
	//任务重新执行,受限于max_reply
	JobStatusAgain 		JobStatus = 5
)

type JobReturn struct {

	Status  	JobStatus
	Msg 		string
}

type JobHandler func(raw interface{}) JobReturn

type JobCallback func(id string,name string,status int,msg string)

type Job struct {

	Id 			string
	Name 		string
	Param 		interface{}
	CreateTime  time.Time
	FlushTime 	time.Time
	Status 		JobStatus
	replyCount 	uint32
}

/*
func (this *Job)Encode()([]byte,error) {

}

func (this *Job)Decode([]byte,error) {

}
*/












