package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"strings"

	"github.com/imiskolee/litejob"
	"gopkg.in/redis.v3"
)

var defaultRedisQueueEngineConfigure litejob.QueueConfigure = litejob.QueueConfigure{

	"host":          "127.0.0.1",
	"port":          6379,
	"password":      "",
	"database":      13,
	"pool_timeout":  30000,
	"read_timeout":  5000,
	"write_timeout": 5000,
	"max_try_count": 5,
	"prefix":        "litejob",
}

type Redis struct {
	sync.Mutex
	client     *redis.Client
	configure  litejob.QueueConfigure
	queues     map[string]*litejob.JobQueueState
	lastUpdate time.Time
}

func NewRedis(configure *litejob.QueueConfigure) litejob.Queue {
	configure.Merge(defaultRedisQueueEngineConfigure)

	engine := new(Redis)

	engine.configure = *configure

	options := &redis.Options{

		Addr:         fmt.Sprintf("%s:%d", engine.configure["host"], engine.configure["port"]),
		DB:           int64(engine.configure["database"].(int)),
		PoolTimeout:  time.Duration(engine.configure["pool_timeout"].(int)) * time.Millisecond,
		ReadTimeout:  time.Duration(engine.configure["read_timeout"].(int)) * time.Millisecond,
		WriteTimeout: time.Duration(engine.configure["read_timeout"].(int)) * time.Millisecond,
		MaxRetries:   engine.configure["max_try_count"].(int),
		Password:     engine.configure["password"].(string),
	}

	engine.client = redis.NewClient(options)
	engine.queues = make(map[string]*litejob.JobQueueState, 0)
	return engine
}

func (queue *Redis) RegisterJob(name string, max int) {

	state := new(litejob.JobQueueState)
	state.Max = max
	state.Work = true
	queue.Lock()
	queue.queues[name] = state
	queue.Unlock()
}

func (queue *Redis) Push(job *litejob.Job) error {

	data, err := job.MarshalBinary()

	if err != nil {
		return errors.New("[litejob] MarshalBinary error:" + err.Error())
	}
	queue.Lock()
	cmd := queue.client.LPush(queue.getJobKey(job.Name), string(data))
	queue.Unlock()
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return err
}

func (queue *Redis) PushFront(job *litejob.Job) error {

	data, err := job.MarshalBinary()
	if err != nil {
		return errors.New("[litejob] MarshalBinary error:" + err.Error())
	}
	queue.Lock()
	cmd := queue.client.RPush(queue.getJobKey(job.Name), string(data))
	queue.Unlock()
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return err
}

func (queue *Redis) GetQueueSize(name string) (uint32, error) {

	jobKey := queue.getJobKey(name)

	cmd := queue.client.LLen(jobKey)

	if cmd.Err() != nil {
		return 0, cmd.Err()
	}

	val := cmd.Val()
	return uint32(val), nil
}

func (queue *Redis) PopN(max int) ([]litejob.Job, error) {

	now := time.Now()

	if now.Sub(queue.lastUpdate) > 60*time.Second {
		queue.syncState()
		queue.lastUpdate = now
	}

	waitings := []string{}
	for name, v := range queue.queues {
		//过滤
		if v.Len < 1 || v.Work == false || v.Max <= v.Running {
			continue
		}
		waitings = append(waitings, name)
	}

	var jobs []litejob.Job

	for _, name := range waitings {

		cmd := queue.client.RPop(queue.getJobKey(name))

		if cmd.Err() != nil {
			continue
		}

		var job litejob.Job

		data, err := cmd.Bytes()

		if err != nil {
			continue
		}

		err = job.UnmarshalBinary(data)

		if err != nil {
			continue
		}
		queue.Lock()
		s := queue.queues[name]
		s.Running++
		queue.Unlock()
		jobs = append(jobs, job)

	}

	return jobs, nil

}

func (queue *Redis) FlushJob(job *litejob.Job) {

	queue.Lock()
	s := (queue.queues[job.Name])
	s.Running--
	queue.Unlock()
}

func (queue *Redis) UpdateJobStatus(name string, status litejob.JobStatus, msg string) {

}

func (queue *Redis) syncState() {

	for name, v := range queue.queues {
		cmd := queue.client.LLen(queue.getJobKey(name))
		if cmd.Err() != nil {
			continue
		}
		v.Len = int(cmd.Val())
		queue.Lock()
		queue.queues[name] = v
		queue.Unlock()
	}
}

func (queue *Redis) Monitor() map[string]map[string]interface{} {

	cmd := queue.client.Keys(queue.getJobKey("*"))
	keys, _ := cmd.Result()
	ret := make(map[string]map[string]interface{}, 0)

	for _, key := range keys {
		name := strings.Replace(key, queue.getJobKey(""), "", 2)
		cmd := queue.client.LLen(key)
		ret[name] = map[string]interface{}{
			"Len": cmd.Val(),
		}
	}
	return ret
}

func (queue *Redis) getJobKey(name string) string {

	return fmt.Sprintf("%s:%s:%s", queue.configure["prefix"], "job", name)
}

func (queue *Redis) getStateKey(id string) string {

	return fmt.Sprintf("%s:%s:%s", queue.configure["prefix"], "state", id)
}

func (queue *Redis) GetAllQueue() map[string]*litejob.JobQueueState {
	return queue.queues
}

func init() {
	litejob.RegisterQueue("redis", NewRedis)
}
