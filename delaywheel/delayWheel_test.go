package delaywheel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/saweima12/gcrate/delaywheel"
)

func TestAfterFunc(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(time.Millisecond, 20)
	if err != nil {
		t.Error(err)
		return
	}
	dw.Start()

	dw.AfterFunc(time.Second, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("Hello")
	})

	item := <-dw.ExecTaskCh()
	fmt.Println(item)
	item.Execute()
}

type TestExecutor struct {
	count int
}

func (te *TestExecutor) Execute(task *delaywheel.TaskCtx) {
	te.count += 1
	fmt.Println("Test count", te.count)
}

func TestAfterExec(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(time.Millisecond, 20)
	if err != nil {
		t.Error(err)
		return
	}
	dw.Start()

	dw.AfterExecute(time.Second, &TestExecutor{})

	item := <-dw.ExecTaskCh()
	fmt.Println(item)
	item.Execute()
}

func TestScheduleFunc(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(time.Millisecond, 20)
	if err != nil {
		t.Error(err)
		return
	}

	dw.Start()
	dw.ScheduleFunc(time.Second*3, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("Hello")
	})

	limit := time.NewTimer(time.Second * 10)
	for {
		select {
		case item := <-dw.ExecTaskCh():
			item.Execute()
			fmt.Println("point")
		case <-limit.C:
			fmt.Println("After")
			return
		}
	}

}
