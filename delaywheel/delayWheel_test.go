package delaywheel_test

import (
	"fmt"
	"testing"
	"time"
	"unsafe"

	"github.com/saweima12/gcrate/delaywheel"
)

func TestAfterFuncAndAutoRun(t *testing.T) {
	// Create delaywheel and start it.
	dw, err := delaywheel.New(
		time.Millisecond, 20,
		delaywheel.WithAutoRun(),
	)
	if err != nil {
		t.Error(err)
		return
	}
	dw.Start()
	fmt.Println("delaywheel size:", unsafe.Sizeof(*dw))

	dw.AfterFunc(time.Second, func(ctx *delaywheel.TaskCtx) {
		fmt.Println("AutoRun: Hello")
	})

	<-time.After(time.Second * 5)
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

	task := <-dw.ExecTaskCh()
	task()
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
			item()
			fmt.Println("point")
		case <-limit.C:
			fmt.Println("After")
			return
		}
	}

}
