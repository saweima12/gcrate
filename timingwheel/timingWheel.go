package timingwheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Executor interface {
	Execute()
}

type TimingWheel interface {
	Start()
	Stop()
	AddWheel(slotNum uint16) TimingWheel
	AddTask(delay time.Duration, executor Executor) (TaskID, error)
	ExecQueue() chan Executor
}

func New(interval time.Duration, bufSize int) TimingWheel {
	result := &timingWheel{
		interval:    interval,
		stopCh:      make(chan struct{}, 1),
		execQueueCh: make(chan Executor, bufSize),
	}
	result.currentTaskId.Store(0)
	return result
}

type timingWheel struct {
	interval      time.Duration
	ticker        *time.Ticker
	currentTaskId atomic.Uint64

	wheel       *tWheel
	stopCh      chan struct{}
	execQueueCh chan Executor
	mutex       sync.Mutex
}

func (ti *timingWheel) AddTask(delay time.Duration, executor Executor) (TaskID, error) {
	if delay < 0 || delay < ti.interval {
		return 0, fmt.Errorf(ERR_DELAY_LESS_INTERVAL, delay, ti.interval)
	}
	ti.mutex.Lock()
	defer ti.mutex.Unlock()

	ti.currentTaskId.Add(1)
	taskID := TaskID(ti.currentTaskId.Load())
	scheduleTime := time.Now().UTC().Add(delay)
	taskObj := &Task{
		taskId:       taskID,
		delay:        int64(delay),
		executor:     executor,
		scheduleTime: scheduleTime.UnixMilli(),
	}

	err := ti.wheel.AddTask(taskObj)
	if err != nil {
		return 0, err
	}
	return taskID, nil
}

func (ti *timingWheel) RemoveTask(id TaskID) {

}

func (ti *timingWheel) ExecQueue() chan Executor {
	return ti.execQueueCh
}

// Create a new wheel to attach to the tail of the wheel.
func (ti *timingWheel) AddWheel(slotNum uint16) TimingWheel {
	if ti.wheel == nil {
		ti.wheel = newWheel(slotNum, int64(ti.interval))
		return ti
	}

	target := ti.wheel
	for target.next != nil {
		target = target.next
	}
	target.next = newWheel(slotNum, target.interval*int64(target.slotNum))
	return ti
}

func (ti *timingWheel) Start() {
	if ti.wheel == nil {
		return
	}

	ti.ticker = time.NewTicker(ti.interval)
	go ti.run()
}

func (ti *timingWheel) Stop() {
	ti.stopCh <- struct{}{}
}

func (ti *timingWheel) run() {
	for {
		select {
		case <-ti.ticker.C:
			ti.onTick()
		case <-ti.stopCh:
			return
		}
	}
}

func (ti *timingWheel) onTick() {

	tasks, nextTrigger := ti.wheel.Tick()
	// Add the tasks that need to be executed to the queue.
	for i := range tasks {
		ti.execQueueCh <- tasks[i].executor
	}

	var next TriggerNextFunc
	next = nextTrigger

	if next == nil {
		return
	}

	edgeTasks := []*Task{}
	for next != nil {
		tasks, nextTrigger := next()
		edgeTasks = append(edgeTasks, tasks...)
		next = nextTrigger
	}

	// Add the edge tasks that need to be executed to the queue.
	for i := range edgeTasks {
		ti.execQueueCh <- edgeTasks[i].executor
	}

}

func (ti *timingWheel) getTask(id TaskID) {

}
