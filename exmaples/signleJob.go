package main

import (
	"fmt"
	"time"

	"github.com/imiskolee/litejob"
	_ "github.com/imiskolee/litejob/queue"
)

func Handler(job *litejob.Job) litejob.Return {
	time.Sleep(100 * time.Millisecond)
	return litejob.Return{
		Status: litejob.JobStatusSuccess,
	}
}

func main() {

	configure := &litejob.Configure{
		QueueEngine:   "redis",
		HeartInterval: 1 * time.Second,
		QueueConfigure: &litejob.QueueConfigure{
			"host": "10.211.55.3",
		},
	}

	dispatch, err := litejob.NewDispatch(configure)
	if err != nil {
		fmt.Printf("error:%s\n", err)
		return
	}

	dispatch.Register("test", Handler, 20)

	for i := 0; i < 100; i++ {
		dispatch.Push("test", "test message")
		fmt.Println(i)
	}

	dispatch.Run()

}
