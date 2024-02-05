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
		interval:     interval,
		taskMap:      make(map[TaskID]*Task),
		stopCh:       make(chan struct{}, 1),
		removeTaskCh: make(chan TaskID, 1),
		addTaskCh:    make(chan *Task, 1),
		execTaskCh:   make(chan Executor, bufSize),
	}
	result.currentTaskId.Store(0)
	return result
}

type timingWheel struct {
	interval      time.Duration
	wheel         *tWheel
	ticker        *time.Ticker
	currentTaskId atomic.Uint64
	mutex         sync.Mutex
	taskMap       map[TaskID]*Task

	stopCh       chan struct{}
	addTaskCh    chan *Task
	execTaskCh   chan Executor
	removeTaskCh chan TaskID
}

func (ti *timingWheel) AddTask(delay time.Duration, executor Executor) (TaskID, error) {
	if delay < 0 || delay < ti.interval {
		return 0, fmt.Errorf(ERR_DELAY_LESS_INTERVAL, delay, ti.interval)
	}

	if !ti.wheel.CheckDelay(delay) {
		return 0, fmt.Errorf(ERR_DELAY_EXCEEDS)
	}

	ti.mutex.Lock()
	ti.currentTaskId.Add(1)
	taskID := TaskID(ti.currentTaskId.Load())
	scheduleTime := time.Now().UTC().Add(delay)
	taskObj := &Task{
		taskId:       taskID,
		delay:        int64(delay),
		executor:     executor,
		scheduleTime: scheduleTime.UnixMilli(),
	}
	defer ti.mutex.Unlock()

	ti.addTaskCh <- taskObj
	return taskID, nil
}

func (ti *timingWheel) RemoveTask(id TaskID) {
	if id < 1 {
		return
	}
	if _, ok := ti.taskMap[id]; !ok {
		return
	}
	ti.removeTaskCh <- id
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

func (ti *timingWheel) ExecQueue() chan Executor {
	return ti.execTaskCh
}

func (ti *timingWheel) run() {
	for {
		select {
		case <-ti.ticker.C:
			ti.onTick()
		case task := <-ti.addTaskCh:
			ti.addTask(task)
		case tid := <-ti.removeTaskCh:
			ti.removeTask(tid)
		case <-ti.stopCh:
			return
		}
	}
}

func (ti *timingWheel) onTick() {

	tasks, nextTrigger := ti.wheel.Tick()
	// Add the tasks that need to be executed to the queue.
	for i := range tasks {
		ti.execTaskCh <- tasks[i].executor
		delete(ti.taskMap, tasks[i].taskId)
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
		ti.execTaskCh <- edgeTasks[i].executor
		delete(ti.taskMap, edgeTasks[i].taskId)
	}
}

func (ti *timingWheel) addTask(t *Task) {
	ti.wheel.AddTask(t)
	ti.taskMap[t.taskId] = t
}

func (ti *timingWheel) removeTask(tid TaskID) {
	if task, ok := ti.taskMap[tid]; ok {
		slot := task.slot
		slot.RemoveFirst(task)
		delete(ti.taskMap, task.taskId)
	}
}
