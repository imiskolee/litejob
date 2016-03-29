package litejob

import (
	"time"

	"encoding/json"
)

//任务状态
type JobStatus uint

const (

	JobStatusWating 	JobStatus = 1 	//任务已经添加成功,等待执行
	JobStatusDoing 		JobStatus = 2 	//任务正在执行
	JobStatusFailed		JobStatus = 3  	//任务执行失败
	JobStatusSuccess 	JobStatus = 4  	//任务执行成功
	JobStatusAgain 		JobStatus = 5  	//任务重新执行,受限于max_reply
	JobStatusKill 		JobStatus = 100
)



//任务函数返回信息
type JobReturn struct {

	Status  	JobStatus	//任务状态
	Msg 		string		//任务执行结果
}

//任务函数定义
type JobHandler func(raw interface{}) JobReturn

//任务回调函数定义
type JobCallback func(job *Job)

//任务
type Job struct {

	Id 			string	//任务id,使用/dev/urandom 获取
	Name 		string	//任务名称,调度器注册的任务名称
	Param 		interface{}		//任务参数
	CreateTime  time.Time		//任务创建时间
	FlushTime 	time.Time		//任务结束时间
	Status 		JobStatus		//任务状态
	replyCount 	uint32			//重试次数
}

func (this *Job)MarshalBinary() ([]byte,error){

	return json.Marshal(this)
}


func (this *Job)UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data,this)
}










