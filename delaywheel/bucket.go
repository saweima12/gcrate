package delaywheel

import (
	"sync/atomic"

	"github.com/saweima12/gcrate/pqueue"
)

type Bucket interface {
	pqueue.Delayer
}

type bucket struct {
	expiration int64
}

func NewBucket() Bucket {
	return &bucket{}
}

func (bu *bucket) SetExpiration(d int64) bool {
	return atomic.SwapInt64(&bu.expiration, d) != d
}

func (bu *bucket) Expiration() int64 {
	return atomic.LoadInt64(&bu.expiration)
}
