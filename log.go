package litejob

import (
	"log"
	"os"
)

type Log struct {

	log *log.Logger

}

func NewLog(path string) *Log{

	file,err:= os.Create(path)

	if err != nil {
		file = os.Stderr
	}else{

	}


	log := log.New(file,"",log.LstdFlags|log.Lshortfile)
	m := new(Log)
	m.log = log

	return m
}

func (this *Log)Normal(format string,v ...interface{}) {
	this.log.Printf(format,v...)
}

func (this *Log)Error(format string,v ...interface{}){
	this.log.Printf(format,v...)
}

