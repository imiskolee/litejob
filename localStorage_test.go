package litejob

import (
	"testing"
)


func SmallHandler(v interface{}) JobReturn{
	return JobReturn{}
}

func TestDumpLoad(t *testing.T){


	configure := &DispatchConfigure{
		Engine:"local",
		MaxConcurrency:10,
		EngineConfigure:map[string]interface{} {
			"dump_file" : "/tmp/test.dat",
		},
	}

	dispatch := NewDispatch(configure)

	dispatch.RegisterHandler("test",SmallHandler)

	for i := 0; i< 200000;i++ {
		dispatch.JobNew("test",map[string]interface{}{"a":make([]byte,1024*2)})
	}

	dispatch.Stop()
	dis2 := NewDispatch(configure)
	dis2.RegisterHandler("test",SmallHandler)
	dis2.Start()

}

