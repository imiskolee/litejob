package litejob

type Storage interface{

	JobPush(job *Job) 	error
	JobPop()(*Job,error)
	JobFlush() 			error
	JobLen() 			uint32
}



