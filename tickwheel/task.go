package tickwheel

import (
	core "github.com/saweima12/gcrate/list"
)

type TaskID uint64

type Task struct {
	delay        int64
	scheduleTime int64
	executor     Executor
	taskId       TaskID
	slot         *TaskList
}

func (ta *Task) Equals(other *Task) bool {
	return ta.taskId == other.taskId
}

type TaskList struct {
	core.SyncList[*Task]
}

func NewTaskList() *TaskList {
	return &TaskList{
		SyncList: *core.NewSync[*Task](),
	}
}
