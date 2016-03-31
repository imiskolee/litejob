package litejob

import "errors"

type StorageInitFunc func(configure *DispatchConfigure) Storage

var storages map[string]StorageInitFunc

type Storage interface{

	JobPush(job *Job) 	error
	JobPop()(*Job,error)
	JobFlush() 			error
	JobLen() 			uint32
}


func RegisterStorage(name string,storage StorageInitFunc) {
	storages[name] = storage
}

func GetStorage(name string,configure *DispatchConfigure) (Storage,error){
	if s,ok := storages[name];ok {
		return s(configure),nil
	}
	return nil,errors.New("[LiteJob] unregister storage :" + name)
}


func init(){
	storages = make(map[string]StorageInitFunc,0)
}





