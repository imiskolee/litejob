package litejob

import (
	"gopkg.in/redis.v3"
	"fmt"
	"time"
	"errors"
	"github.com/astaxie/beego"
)

type RedisStorage struct {

	redisClient 			*redis.Client
	configure 				*DispatchConfigure
	redisKey  			 	string
	redisStatusKeyPrefix 	string
	recommitTryCount 	int
}



func NewRedisStorage(configure *DispatchConfigure) *RedisStorage {

	storage := new(RedisStorage)
	storage.configure = configure

	var host string = "127.0.0.1"
	var port int 	= 6379
	var db   int    = 15
	var redisKey  string = "litejob-auto-list"
	var tryCount int = 3
	var password string = ""

	if _,ok := configure.EngineConfigure["redis_host"];ok {
		host = configure.EngineConfigure["redis_host"].(string)
	}
	if _,ok := configure.EngineConfigure["redis_port"];ok {
		port = configure.EngineConfigure["redis_port"].(int)
	}

	if _,ok := configure.EngineConfigure["redis_db"];ok {
		db = configure.EngineConfigure["redis_db"].(int)
	}

	if _,ok := configure.EngineConfigure["redis_key"];ok {
		redisKey = configure.EngineConfigure["redis_key"].(string)
	}

	if _,ok := configure.EngineConfigure["redis_password"];ok {
		password = configure.EngineConfigure["redis_password"].(string)
	}

	if _,ok := configure.EngineConfigure["redis_command_try_count"];ok {
		tryCount = configure.EngineConfigure["redis_command_try_count"].(int)
	}

	storage.redisKey = redisKey
	storage.recommitTryCount = tryCount

	options := &redis.Options{
		Addr : fmt.Sprintf("%s:%d",host,port),
		DB 	 : int64(db),
		PoolTimeout: 5 * time.Second,
	}

	if len(password) > 0 {
		options.Password = password
	}

	storage.redisClient = redis.NewClient(options)
	return storage
}

func (this *RedisStorage)JobPush(job *Job) error {

	var err error = nil
	for i := 0; i < this.recommitTryCount ;i++ {
		data,err := job.MarshalBinary()
		if err == nil {
			cmd := this.redisClient.LPush(this.redisKey,string(data))
			err = cmd.Err()
			if cmd.Err() == nil {
				return nil
			}
		}
		return err
	}

	return errors.New("redis cmmand errror:" + err.Error())
}

func (this *RedisStorage)JobPop()(*Job,error){

	cmd := this.redisClient.LPop(this.redisKey)

	if cmd.Err() == nil {
		job := new(Job)
		err := cmd.Scan(job)
		return job,err
	}

	return nil,cmd.Err()
}

func (this *RedisStorage)JobLen() uint32 {

	cmd := this.redisClient.LLen(this.redisKey)

	beego.Debug(cmd.Result())

	if cmd.Err() == nil {
		return uint32(cmd.Val())
	}

	return 0
}

func (this *RedisStorage)JobFlush() error {

	for i := 0; i < this.recommitTryCount ;i++ {
		if cmd := this.redisClient.Del(this.redisKey);cmd.Err() == nil {
			return nil
		}
	}
	return errors.New("redis cmmand errror")
}

