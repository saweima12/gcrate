package timingwheel

import (
	"fmt"
	"time"
)

type Executor interface {
	Execute()
}

type TimingWheel interface {
	Start()
	Stop()
	AddWheel(slotNum uint16) TimingWheel
	AddTask(delay time.Duration, executor Executor) (*Task, error)
	ExecQueue() chan Executor
}

func New(interval time.Duration, bufSize int) *timingWheel {
	result := &timingWheel{
		interval:    interval,
		stopCh:      make(chan struct{}, 1),
		execQueueCh: make(chan Executor, bufSize),
	}

	return result
}

type timingWheel struct {
	interval time.Duration
	ticker   *time.Ticker
	wheel    *tWheel

	stopCh      chan struct{}
	execQueueCh chan Executor
}

func (ti *timingWheel) AddTask(delay time.Duration, executor Executor) (*Task, error) {
	scheduleTime := time.Now().UTC().Add(delay)
	taskObj := &Task{
		delay:        int64(delay),
		executor:     executor,
		scheduleTime: scheduleTime.UnixMilli(),
	}

	fmt.Println(taskObj)
	err := ti.wheel.AddTask(taskObj)
	if err != nil {
		return nil, err
	}

	return taskObj, nil
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
		ti.execQueueCh <- tasks[i].executor
	}
}
