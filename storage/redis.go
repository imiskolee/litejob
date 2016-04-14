package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/imiskolee/litejob"
	"gopkg.in/redis.v3"
)

type Redis struct {
	redisClient          *redis.Client
	configure            *litejob.DispatchConfigure
	redisKey             string
	stateKey             string
	redisStatusKeyPrefix string
	recommitTryCount     int
}

var defaultRedisEngineConfigure litejob.EngineConfigure = litejob.EngineConfigure{

	"host":          "127.0.0.1",
	"port":          6379,
	"password":      "",
	"db":            15,
	"job_key":       "litejob-job-list",
	"state_key":     "litejob-job-state",
	"max_try_count": 3,
	"pool_timeout":  15 * time.Second,
	"read_timeout":  5 * time.Second,
	"write_timeout": 5 * time.Second,
	"state_expires": 24 * 7 * time.Hour, //a weekly

}

func NewRedis(configure *litejob.DispatchConfigure) litejob.Storage {

	storage := new(Redis)
	storage.configure = configure

	//重构这一大段代码
	storage.redisKey = configure.EngineConfigure.Get("job_key", defaultRedisEngineConfigure["job_key"]).(string)
	storage.stateKey = configure.EngineConfigure.Get("state_key", defaultRedisEngineConfigure["state_key"]).(string)

	storage.recommitTryCount = configure.EngineConfigure.Get("max_try_count", defaultRedisEngineConfigure["max_try_count"]).(int)

	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", configure.EngineConfigure.Get("host", defaultRedisEngineConfigure["host"]).(string),
			configure.EngineConfigure.Get("port", defaultRedisEngineConfigure["port"]).(int)),
		DB:           int64(configure.EngineConfigure.Get("db", defaultRedisEngineConfigure["db"]).(int)),
		PoolTimeout:  configure.EngineConfigure.Get("pool_timeout", defaultRedisEngineConfigure["pool_timeout"]).(time.Duration),
		ReadTimeout:  configure.EngineConfigure.Get("read_timeout", defaultRedisEngineConfigure["read_timeout"]).(time.Duration),
		WriteTimeout: configure.EngineConfigure.Get("write_timeout", defaultRedisEngineConfigure["write_timeout"]).(time.Duration),
		MaxRetries:   configure.EngineConfigure.Get("max_try_count", defaultRedisEngineConfigure["max_try_count"]).(int),
	}

	password := configure.EngineConfigure.Get("password", defaultRedisEngineConfigure["password"]).(string)

	if len(password) > 0 {
		options.Password = password
	}

	storage.redisClient = redis.NewClient(options)
	return storage
}

func (this *Redis) JobPush(job *litejob.Job) error {
	var err error = nil
	data, err := job.MarshalBinary()
	if err == nil {
		cmd := this.redisClient.LPush(this.redisKey, string(data))
		err = cmd.Err()
		if cmd.Err() == nil {
			return nil
		}
	}
	return errors.New("redis cmmand errror:" + err.Error())
}

func (this *Redis) JobPop() (*litejob.Job, error) {

	cmd := this.redisClient.RPop(this.redisKey)

	if cmd.Err() == nil {
		job := new(litejob.Job)
		err := cmd.Scan(job)
		return job, err
	}

	return nil, cmd.Err()
}

func (this *Redis) JobLen() uint32 {

	cmd := this.redisClient.LLen(this.redisKey)

	if cmd.Err() == nil {
		return uint32(cmd.Val())
	}

	return 0
}

func (this *Redis) JobFlush() error {

	if cmd := this.redisClient.Del(this.redisKey); cmd.Err() == nil {
		return nil
	}

	return errors.New("redis cmmand errror")
}

func (this *Redis) JobStateUpdate(state *litejob.JobState) error {

	if len(this.stateKey) > 0 && len(state.JobId) > 0 {
		var err error = nil
		data, err := state.MarshalBinary()
		if err == nil {
			key := fmt.Sprintf("%s:%s", this.stateKey, state.JobId)
			cmd := this.redisClient.Set(key, data, 0)
			return cmd.Err()
		}
	}

	return nil
}

func (this *Redis) JobState(jobId string) (*litejob.JobState, error) {

	state := new(litejob.JobState)
	state.Status = litejob.JobStatusUnknow.String()
	state.Msg = "unknow error"

	if len(this.stateKey) > 0 {
		key := fmt.Sprintf("%s:%s", this.stateKey, jobId)
		cmd := this.redisClient.Get(key)
		if cmd.Err() == nil {
			err := cmd.Scan(state)
			return state, err
		}
		return state, cmd.Err()
	}
	return state, nil
}

func init() {
	litejob.RegisterStorage("redis", NewRedis)
}
