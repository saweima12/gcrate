package delaywheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/saweima12/gcrate/pqueue"
)

func DefaultNow() int64 {
	return timeToMs(time.Now().UTC())
}

func New(tick time.Duration, wheelSize int) (*DelayWheel, error) {
	startMs := timeToMs(time.Now().UTC())
	tickMs := int64(tick) / int64(time.Millisecond)

	if tickMs < 1 {
		return nil, fmt.Errorf("The tick must be greater than or equal to 1 Millisecond")
	}

	result := create(
		tickMs,
		wheelSize,
		startMs,
	)

	return result, nil
}

func create(tickMs int64, wheelSize int, startMs int64) *DelayWheel {
	initInterval := tickMs * int64(wheelSize)

	dw := &DelayWheel{
		tickMs:      tickMs,
		wheelSize:   wheelSize,
		startMs:     startMs,
		curInterval: initInterval,

		queue: pqueue.NewDelayQueue[*bucket](DefaultNow, wheelSize),
		wheel: newWheel(tickMs, wheelSize, startMs),

		execTaskCh: make(chan *Task, 1),
		addTaskCh:  make(chan *Task, 1),
		stopCh:     make(chan struct{}, 1),
	}
	dw.curTime.Store(startMs)
	dw.ctxPool = newCtxPool(dw)
	dw.taskPool = newTaskPool()
	return dw
}

type DelayWheel struct {
	wheelSize   int
	tickMs      int64
	curInterval int64
	startMs     int64
	curTime     atomic.Int64
	curTaskID   atomic.Uint64

	mu       sync.Mutex
	wheel    *tWheel
	taskPool *taskPool
	ctxPool  *ctxPool
	queue    pqueue.DelayQueue[*bucket]

	execTaskCh chan *Task
	addTaskCh  chan *Task
	stopCh     chan struct{}
}

func (de *DelayWheel) Start() {
	go de.run()
}

func (de *DelayWheel) Stop() {
	de.stopCh <- struct{}{}
}

func (de *DelayWheel) ExecTaskCh() <-chan *Task {
	return de.execTaskCh
}

func (de *DelayWheel) AfterFunc(d time.Duration, f func(task *TaskCtx)) {
	de.mu.Lock()
	defer de.mu.Unlock()

	t := de.createTask(d)
	t.executor = pureExec(f)
	de.addTaskCh <- t
}

func (de *DelayWheel) AfterExecute(d time.Duration, executor Executor) {
	de.mu.Lock()
	defer de.mu.Unlock()

	t := de.createTask(d)
	t.executor = executor

	de.addTaskCh <- t
}

func (de *DelayWheel) ScheduleFunc(d time.Duration, f func(ctx *TaskCtx)) {
	de.mu.Lock()
	defer de.mu.Unlock()

	t := de.createTask(d)

	sch := pureScheduler{d: d}
	t.executor = pureExec(func(ctx *TaskCtx) {
		// Calculate new expiration.
		expiration := sch.Next(msToTime(ctx.Expiration()))
		if !expiration.IsZero() {
			ctx.t.expiration = timeToMs(expiration)
			de.addOrRun(ctx.t)
			ctx.isSechuled = true
		}
		f(ctx)
	})

	de.addTaskCh <- t
}

func (de *DelayWheel) ScheduleExecute(d time.Duration, executor Executor) {
	de.mu.Lock()
	defer de.mu.Unlock()

	t := de.createTask(d)

	sch := pureScheduler{d: d}
	t.executor = pureExec(func(ctx *TaskCtx) {
		// Calculate new expiration.
		expiration := sch.Next(msToTime(ctx.t.expiration))
		if !expiration.IsZero() {
			ctx.t.expiration = timeToMs(expiration)
			de.addOrRun(ctx.t)
			ctx.isSechuled = true
		}
		executor.Execute(ctx)
	})

	de.addTaskCh <- t
}

func (de *DelayWheel) AdvanceClock(expiration int64) {
	if expiration < de.curTime.Load() {
		// if the expiration is less than current, ignore it.
		return
	}

	// update currentTime
	de.curTime.Store(expiration)
	de.wheel.advanceClock(expiration)
}

func (de *DelayWheel) run() {
	de.queue.Start()

	for {
		select {
		case task := <-de.addTaskCh:
			de.addOrRun(task)
		case bu := <-de.queue.ExpiredCh():
			de.handleExipredBucket(bu)
		case <-de.stopCh:
			de.queue.Stop()
			return
		}
	}
}

func (de *DelayWheel) handleExipredBucket(b *bucket) {
	// Advance wheel's currentTime.
	de.AdvanceClock(b.Expiration())

	for {
		// Extract all task
		task, ok := b.tasks.PopFront()
		if !ok {
			break
		}
		if task.isCancelled.Load() {
			continue
		}
		de.addOrRun(task)
	}

	b.SetExpiration(-1)
}

func (de *DelayWheel) add(task *Task) bool {
	curTime := de.curTime.Load()
	// When the expiration time less than one tick. execute directly.
	if task.expiration < curTime+de.tickMs {
		return false
	}

	// if the expiration is exceeds the curInterval, expand the wheel.
	if task.expiration > curTime+de.curInterval {
		// When the wheel need be extended, advanceClock to now.
		de.AdvanceClock(DefaultNow())
		for task.expiration > curTime+de.curInterval {
			de.expandWheel()
		}
	}

	// Choose a suitable wheel
	wheel := de.wheel
	for wheel.next != nil {
		if task.expiration <= wheel.curTime.Load()+wheel.interval {
			break
		}
		wheel = wheel.next
	}

	// Insert the task into bucket and setting expiration.
	bucket := wheel.addTask(task)
	if bucket.SetExpiration(task.expiration) {
		de.queue.Offer(bucket)
	}

	return true
}

func (de *DelayWheel) addOrRun(task *Task) {
	if !de.add(task) {
		de.execTaskCh <- task
	}
}

func (de *DelayWheel) expandWheel() {
	// Find last wheel
	target := de.wheel
	for target.next != nil {
		target = target.next
	}
	// Loading currnet time & create next layer wheel.
	curTime := de.curTime.Load()
	next := newWheel(target.interval, de.wheelSize, curTime)
	target.next = next

	atomic.StoreInt64(&de.curInterval, next.interval)
}

func (de *DelayWheel) createTask(d time.Duration) *Task {
	nId := de.curTaskID.Add(1)
	t := de.taskPool.Get()
	t.taskID = nId
	t.expiration = timeToMs(time.Now().UTC().Add(d))
	t.ctxPool = de.ctxPool
	t.taskPool = de.taskPool

	return t
}
