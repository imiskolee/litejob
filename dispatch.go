package litejob

import "time"

//任务调度器的控制项
type DispatchConfigure struct {

	Engine 				string 			//队列引擎
	MaxConcurrency		uint32 			//最大并发数
	MaxReplyCount 		uint32 			//最大重试次数
	HeartInterval		time.Duration	//心跳时间
	DumpInterval 		time.Duration   //数据持久化间隔时间
	Logfile 			string			//日志文件,默认为/std/stderr
	EngineConfigure 	map[string]interface{}	//引擎相关的配置项,localStorage支持dump_file参数,用于持久化内存信息
	JobCallback 		JobCallback				//任务回调方法,每执行一次任务即调用一次该方法
}

//调度器状态
type DispatchStatus struct {
	Running 	uint32		//正在运行的任务数
	Len 		uint32		//当前任务长度
}


//任务调度器
type Dispatch interface{

	JobNew(name string,param interface{}) string	//推送一个任务
	JobPush(job Job) 	bool			//重入一条任务
	JobLen() 			uint32			//当前队列长度
	FlushJob() 			bool			//清空所有任务
	Status()   			DispatchStatus	//状态
	Start()				bool			// 开始服务,如果设置了dump_file,则会从硬盘预加载任务队列
	Stop()				bool			// 停止服务,如果设置了dump_file,则会持久化一次到硬盘.
	RegisterHandler(name string,handler JobHandler) bool	//注册一个任务
}

//获取一个调度器
func NewDispatch(configure *DispatchConfigure)(Dispatch) {

	if configure.Engine == EngineLocal {
		return NewLocalDispatch(configure)
	}
	return nil
}
