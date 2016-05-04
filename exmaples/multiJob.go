package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/imiskolee/litejob"
	_ "github.com/imiskolee/litejob/queue"
)

func Handler(job *litejob.Job) litejob.Return {
	//	fmt.Print(job)
	//	time.Sleep(100 * time.Millisecond)
	//	runtime.Gosched()
	return litejob.Return{
		Status: litejob.JobStatusFailed,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	configure := &litejob.Configure{
		QueueEngine:   "redis",
		HeartInterval: 1 * time.Second,
		QueueConfigure: &litejob.QueueConfigure{
			"host": "10.211.55.3",
		},
		MaxReplyCount: 5,
	}

	dispatch, err := litejob.NewDispatch(configure)
	if err != nil {
		fmt.Printf("error:%s\n", err)
		return
	}

	totalJobs := 100
	//模拟100个任务
	for i := 0; i < totalJobs; i++ {
		dispatch.Register(fmt.Sprintf("test_%d", i), Handler, 2)
	}

	for i := 0; i < 100000; i++ {

		job := fmt.Sprintf("test_%d", int(rand.Int31())%totalJobs)
		dispatch.Push(job, "test message")
		fmt.Println(i)
	}

	dispatch.Run()

}
