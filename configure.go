package litejob

import "time"

type QueueConfigure map[string]interface{}

func (configure *QueueConfigure) Merge(src QueueConfigure) {
	for k, v := range src {
		if _, ok := (*configure)[k]; !ok {
			(*configure)[k] = v
		}
	}
}

type Configure struct {
	QueueEngine    string
	LogFile        string
	MaxReplyCount  int
	HeartInterval  time.Duration
	QueueConfigure *QueueConfigure
	After          Callback
	Before         Callback
}
