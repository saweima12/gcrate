package delaywheel

import (
	"sync/atomic"

	"github.com/saweima12/gcrate/list"
)

type bucket struct {
	expiration int64
	tasks      *list.SyncList[*Task]
}

func newBucket() *bucket {
	return &bucket{
		expiration: -1,
		tasks:      list.NewSync[*Task](),
	}
}

func (bu *bucket) SetExpiration(d int64) bool {
	return atomic.SwapInt64(&bu.expiration, d) != d
}

func (bu *bucket) Expiration() int64 {
	return atomic.LoadInt64(&bu.expiration)
}

func (bu *bucket) AddTask(task *Task) {
	elm := bu.tasks.PushBack(task)
	task.elm = elm
	task.bucket = bu
}
