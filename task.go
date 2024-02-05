package timingwheel

import (
	"timingwheel/core"
)

type TaskList struct {
	core.SafeList[*Task]
}

func NewTaskList() *TaskList {
	return &TaskList{
		SafeList: *core.NewSafeList[*Task](),
	}
}

type Task struct {
	delay        int64
	scheduleTime int64
	executor     Executor
}
