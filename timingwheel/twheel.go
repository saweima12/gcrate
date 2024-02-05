package timingwheel

import (
	"fmt"
	"sync/atomic"
	"time"
)

type tWheel struct {
	slots      []*TaskList
	interval   int64
	maxDelay   int64
	slotNum    uint16
	currentPos atomic.Uint32
	next       *tWheel
}

type TriggerNextFunc func() ([]*Task, TriggerNextFunc)

func newWheel(slotNum uint16, interval int64) *tWheel {
	result := &tWheel{
		slotNum:  slotNum,
		interval: interval,
		maxDelay: interval * int64(slotNum),
		slots:    make([]*TaskList, slotNum),
	}

	result.currentPos.Store(0)
	// Initialize each slots.
	for i := range result.slots {
		result.slots[i] = NewTaskList()
	}
	return result
}

func (t *tWheel) Tick() (taskList []*Task, nextTrigger TriggerNextFunc) {
	t.currentPos.Add(1)
	if uint16(t.currentPos.Load()) >= t.slotNum {
		t.currentPos.Store(0)

		list := t.slots[t.currentPos.Load()]
		return list.PopAll(), t.triggerNext
	}

	list := t.slots[t.currentPos.Load()]
	return list.PopAll(), nil
}

func (t *tWheel) AddTask(task *Task) error {
	if task.delay <= t.maxDelay {
		pos := t.calculatePos(task.delay)
		fmt.Printf("AddTask to %d, %d\n", pos, t.interval)
		t.slots[pos].PushBack(task)
		return nil
	}

	if t.next == nil {
		return fmt.Errorf(ERR_DELAY_EXCEEDS)
	}

	return t.next.AddTask(task)
}

func (t *tWheel) triggerNext() (extTasks []*Task, nextTrigger TriggerNextFunc) {

	if t.next == nil {
		return
	}

	result := make([]*Task, 0)
	taskList, next := t.next.Tick()
	fmt.Println("triggerNext")

	now := time.Now().UTC().UnixMilli()

	for i := range taskList {
		offset := (taskList[i].scheduleTime - now) * int64(time.Millisecond)
		fmt.Println(offset)
		if offset <= t.interval {
			result = append(result, taskList[i])
			continue
		}
		taskList[i].delay = offset
		t.AddTask(taskList[i])
	}

	return result, next
}

func (t *tWheel) calculatePos(delay int64) int {
	blockNum := delay / t.interval
	return int(t.currentPos.Load()+uint32(blockNum)) % int(t.slotNum)
}
