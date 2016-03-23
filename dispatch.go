package litejob

import "time"

type DispatchConfigure struct {

	Engine 				string 			//队列引擎
	MaxConcurrency		uint32 			//最大并发数
	MaxReplyCount 		uint32 			//最大重试次数
	HeartInterval		time.Duration	//心跳时间
	Logfile 			string
	EngineConfigure 	map[string]interface{}
}


type DispatchStatus struct {

	Running 	uint32
	Len 		uint32
	Jobs 		uint32
}

type Dispatch interface{

	JobNew(name string,param interface{}) bool
	JobPush(job Job) 	bool			//推入一条任务
	JobLen() 			uint32			//当前队列长度
	FlushJob() 			bool			//清空所有任务
	Status()   			DispatchStatus	//状态
	Start()				bool			// 开始服务
	Stop()				bool			// 停止服务

	RegisterHandler(name string,handler JobHandler) bool
}


func NewDispatch(configure *DispatchConfigure)(Dispatch) {

	if configure.Engine == "local" {
		return NewLocalDispatch(configure)
	}
	return nil
}


