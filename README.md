### LiteJob

#### 为什么需要LiteJob?

如果你有以下需求,可以尝试使用litejob.

* 如果系统中已经有成熟的队列,消息系统.可以使用litejob做本地重试队列,尽可能保证任务分发成功.
* 直接使用litejob做为一种瞬时,超负载的流量控制手段.


### Document
----
# litejob
--
    import "litejob"


## Usage

```go
const (
	EngineLocal = "local"
)
```


#### type Dispatch

```go
type Dispatch interface {
	JobNew(name string, param interface{}) string         //推送一个任务
	JobPush(job Job) bool                                 //重入一条任务
	JobLen() uint32                                       //当前队列长度
	FlushJob() bool                                       //清空所有任务
	Status() DispatchStatus                               //状态
	Start() bool                                          // 开始服务,如果设置了dump_file,则会从硬盘预加载任务队列
	Stop() bool                                           // 停止服务,如果设置了dump_file,则会持久化一次到硬盘.
	RegisterHandler(name string, handler JobHandler) bool //注册一个任务
}
```

任务调度器

#### func  NewDispatch

```go
func NewDispatch(configure *DispatchConfigure) Dispatch
```
获取一个调度器

#### type DispatchConfigure

```go
type DispatchConfigure struct {
	Engine          string                 //队列引擎
	MaxConcurrency  uint32                 //最大并发数
	MaxReplyCount   uint32                 //最大重试次数
	HeartInterval   time.Duration          //心跳时间
	DumpInterval    time.Duration          //数据持久化间隔时间
	Logfile         string                 //日志文件,默认为/std/stderr
	EngineConfigure map[string]interface{} //引擎相关的配置项,localStorage支持dump_file参数,用于持久化内存信息
	JobCallback     JobCallback            //任务回调方法,每执行一次任务即调用一次该方法
}
```

任务调度器的控制项

#### type DispatchStatus

```go
type DispatchStatus struct {
	Running uint32 //正在运行的任务数
	Len     uint32 //当前任务长度
}
```

调度器状态

#### type Job

```go
type Job struct {
	Id         string      //任务id,使用/dev/urandom 获取
	Name       string      //任务名称,调度器注册的任务名称
	Param      interface{} //任务参数
	CreateTime time.Time   //任务创建时间
	FlushTime  time.Time   //任务结束时间
	Status     JobStatus   //任务状态
}
```

任务

#### type JobCallback

```go
type JobCallback func(id string, name string, status int, msg string)
```

任务回调函数定义

#### type JobHandler

```go
type JobHandler func(raw interface{}) JobReturn
```

任务函数定义

#### type JobReturn

```go
type JobReturn struct {
	Status JobStatus //任务状态
	Msg    string    //任务执行结果
}
```

任务函数返回信息

#### type JobStatus

```go
type JobStatus uint
```

任务状态

```go
const (
	JobStatusWating  JobStatus = 1 //任务已经添加成功,等待执行
	JobStatusDoing   JobStatus = 2 //任务正在执行
	JobStatusFailed  JobStatus = 3 //任务执行失败
	JobStatusSuccess JobStatus = 4 //任务执行成功
	JobStatusAgain   JobStatus = 5 //任务重新执行,受限于max_reply
)
```









