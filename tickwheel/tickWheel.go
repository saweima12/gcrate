package tickwheel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Executor interface {
	Execute()
}

type TickWheel interface {
	AddWheel(slotNum uint16) TickWheel
	AddTask(delay time.Duration, executor Executor) (TaskID, error)
	RemoveTask(id TaskID) error

	Start()
	Stop()
	ExecCh() chan Executor
	TaskCount() int
}

func New(interval time.Duration, bufSize int) TickWheel {
	result := &tickWheel{
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

type tickWheel struct {
	interval      time.Duration
	wheel         *tWheel
	ticker        *time.Ticker
	currentTaskId atomic.Uint64
	mutex         sync.Mutex

	taskMap      map[TaskID]*Task
	stopCh       chan struct{}
	addTaskCh    chan *Task
	execTaskCh   chan Executor
	removeTaskCh chan TaskID
}

func (ti *tickWheel) AddWheel(slotNum uint16) TickWheel {
	if ti.wheel == nil {
		ti.wheel = newWheel(slotNum, int64(ti.interval))
		return ti
	}

	target := ti.wheel
	for target.next != nil {
		target = target.next
	}
	target.next = newWheel(slotNum, target.tickDur*int64(target.slotNum))
	return ti
}

func (ti *tickWheel) AddTask(delay time.Duration, executor Executor) (TaskID, error) {
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

func (ti *tickWheel) RemoveTask(id TaskID) error {
	if id < 1 {
		return fmt.Errorf(ERR_ID_NOT_AVALIABLE, id)
	}
	if _, ok := ti.taskMap[id]; !ok {
		return fmt.Errorf(ERR_NOT_FOUND, id)
	}
	ti.removeTaskCh <- id
	return nil
}

// Create a new wheel to attach to the tail of the wheel.

func (ti *tickWheel) Start() {
	if ti.wheel == nil {
		return
	}

	ti.ticker = time.NewTicker(ti.interval)
	go ti.run()
}

func (ti *tickWheel) Stop() {
	ti.stopCh <- struct{}{}
}

func (ti *tickWheel) ExecCh() chan Executor {
	return ti.execTaskCh
}

func (ti *tickWheel) TaskCount() int {
	return len(ti.taskMap)
}

func (ti *tickWheel) run() {
	for {
		select {
		case <-ti.ticker.C:
			ti.onTick()
		case task := <-ti.addTaskCh:
			ti.addTask(task)
		case tid := <-ti.removeTaskCh:
			ti.removeTask(tid)
		case <-ti.stopCh:
			ti.ticker.Stop()
			return
		}
	}
}

func (ti *tickWheel) onTick() {
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

func (ti *tickWheel) addTask(t *Task) {
	ti.wheel.AddTask(t)
	ti.taskMap[t.taskId] = t
}

func (ti *tickWheel) removeTask(tid TaskID) {
	if task, ok := ti.taskMap[tid]; ok {
		slot := task.slot
		slot.RemoveFirst(task)
		delete(ti.taskMap, task.taskId)
	}
}
