package delaywheel

import (
	"sync"
	"sync/atomic"
	"time"
)

type ctxPool struct {
	pool sync.Pool
	dw   *DelayWheel
}

func newCtxPool(dw *DelayWheel) *ctxPool {
	return &ctxPool{
		pool: sync.Pool{
			New: func() any {
				return new(TaskCtx)
			},
		},
		dw: dw,
	}
}

func (c *ctxPool) Get(t *Task) *TaskCtx {
	item := c.pool.Get().(*TaskCtx)
	item.t = t
	item.reSchedule = c.dw.addOrRun
	return item
}

func (c *ctxPool) Put(ctx *TaskCtx) {
	ctx.t = nil
	ctx.reSchedule = nil
	ctx.isSechuled = false
	c.pool.Put(ctx)
}

type TaskCtx struct {
	t          *Task
	reSchedule func(t *Task)

	mu         sync.Mutex
	isSechuled bool
}

func (ctx *TaskCtx) IsCancelled() bool {
	return ctx.t.isCancelled.Load()
}

func (ctx *TaskCtx) Cancel() {
	ctx.t.isCancelled.Store(true)
}

func (ctx *TaskCtx) TaskID() uint64 {
	return ctx.t.taskID
}

func (ctx *TaskCtx) Expiration() int64 {
	return ctx.t.expiration
}

func (ctx *TaskCtx) ExpireTime() time.Time {
	return msToTime(ctx.t.expiration)
}

func (ctx *TaskCtx) ReSchedule(d time.Duration) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.isSechuled {
		return
	}
	newExp := ctx.ExpireTime().Add(d)
	atomic.SwapInt64(&ctx.t.expiration, timeToMs(newExp))
	ctx.reSchedule(ctx.t)
	ctx.isSechuled = true

}
