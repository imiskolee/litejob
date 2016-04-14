package main

import (
	"fmt"
	"os"
	"time"

	"github.com/imiskolee/litejob"
	_ "github.com/imiskolee/litejob/storage"
)

//start job server at before start

func Echo(job *litejob.Job) litejob.JobReturn {

	//	fmt.Fprint(os.Stderr,"echo")

	return litejob.JobReturn{
		Status: litejob.JobStatusAgain,
		Msg:    "",
	}
}

func main() {

	/**

	//默认redis配置
	var defaultRedisEngineConfigure litejob.EngineConfigure = litejob.EngineConfigure{

		"host"            	: "127.0.0.1",
		"port"           	 : 6379,
		"password"        	: "",
		"db"            	: 15,
		"job_key"        	: "litejob-job-list",
		"state_key"        	: "litejob-job-state",
		"max_try_count"    	: 3,
		"pool_timeout"    	: 15 * time.Second,
		"read_timeout"    	: 5 * time.Second,
		"write_timeout"    : 5 * time.Second,

	}
	*/

	configure := &litejob.DispatchConfigure{

		MaxConcurrency: 3,               //最大并发数,计算密集任务大于CPU个数没有意义,IO密集型的可以增大
		MaxReplyCount:  5,               //任务最大重试次数,当任务返回JobStatusAgain的时候将最多重试次数
		HeartInterval:  5 * time.Second, //心跳时间.在空闲时将Sleep该时间.可以理解为任务最大响应延迟
		Logfile:        "",              //日志文件地址,为空最输出到os.Stderr中
		Callback: litejob.JobCallbackConfigure{ //如不需要,可以不进行配置
			Before: func(job *litejob.Job) {}, // 任务处理之前
			After:  func(job *litejob.Job) {}, //  任务处理之后
		},
		Engine: "redis", //存储引擎
		EngineConfigure: litejob.EngineConfigure{
			"host":          "127.0.0.1", //redis 地址
			"port":          6379,
			"password":      "", // redis 密码
			"db":            15,
			"job_key":       "litejob-job-list",  // 用于存储任务队列的键
			"state_key":     "litejob-job-state", // 用于存储任务状态的键
			"max_try_count": 3,                   // redis内部重试次数
			"pool_timeout":  15 * time.Second,
			"read_timeout":  5 * time.Second,
			"write_timeout": 5 * time.Second,
		},
	}

	dispatch, err := litejob.NewDispatch(configure)

	if err != nil {
		fmt.Fprintln(os.Stderr, "init job server failed:"+err.Error())
	}

	dispatch.RegisterHandler("echo", Echo)

	jobIds := make([]string, 0)

	//启动1W个任务进程
	for i := 0; i < 10000; i++ {
		//第二个参数为interface{},但是最终会进行json序列化存储到引擎中,请保证其安全性
		job, err := dispatch.JobNew("echo", nil)
		if err == nil {
			jobIds = append(jobIds, job.Id)
		}
	}

	//以协程方式启动守护 go dispatch.Start()
	go dispatch.Start()

	/*
		time.Sleep(10 * time.Second)

		for i := len(jobIds)-1;i>0;i-- {
			state,err := dispatch.JobState(jobIds[i])
			fmt.Fprintln(os.Stdout,state,err)
		}
	*/

}
