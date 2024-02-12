package delaywheel

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type Executor interface {
	Execute(task *TaskCtx)
}

type Scheduler interface {
	Next(d time.Time) time.Time
}

type taskPool struct {
	pool sync.Pool
}

func newTaskPool() *taskPool {
	return &taskPool{
		pool: sync.Pool{
			New: func() any {
				return new(Task)
			},
		},
	}
}

func (tp *taskPool) Get() *Task {
	return tp.pool.Get().(*Task)
}

func (tp *taskPool) Put(t *Task) {
	t.taskID = 0
	t.expiration = 0
	t.executor = nil
	t.isCancelled.Store(false)
	t.taskPool = nil
	t.ctxPool = nil
	t.elm = nil
	t.bucket = nil
	tp.pool.Put(t)
}

type Task struct {
	taskID      uint64
	expiration  int64
	executor    Executor
	isCancelled atomic.Bool

	taskPool *taskPool
	ctxPool  *ctxPool
	elm      *list.Element
	bucket   *bucket
}

func (dt *Task) TaskID() uint64 {
	return dt.taskID
}

func (dt *Task) Expiration() int64 {
	return dt.expiration
}

func (dt *Task) Execute() {
	ctx := dt.ctxPool.Get(dt)
	dt.executor.Execute(ctx)

	isSchedule := ctx.isSechuled

	dt.ctxPool.Put(ctx)
	if !isSchedule {
		dt.taskPool.Put(dt)
	}

}

func pureExec(f func(task *TaskCtx)) *pureExecutor {
	return &pureExecutor{
		f: f,
	}
}

type pureExecutor struct {
	f func(task *TaskCtx)
}

func (we *pureExecutor) Execute(task *TaskCtx) {
	we.f(task)
}

type pureScheduler struct {
	d time.Duration
}

func (pu *pureScheduler) Next(cur time.Time) time.Time {
	return cur.Add(pu.d)
}
