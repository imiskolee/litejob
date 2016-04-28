package litejob

type JobHandler func(job *Job) Return

type Callback func(job *Job)
