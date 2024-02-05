package timingwheel

import (
	core "github.com/saweima12/gcrate/synclist"
)

type TaskID uint64

type Task struct {
	delay        int64
	scheduleTime int64
	executor     Executor
	taskId       TaskID
}

type TaskList struct {
	core.SafeList[*Task]
}

func NewTaskList() *TaskList {
	return &TaskList{
		SafeList: *core.NewGeneric[*Task](),
	}
}
