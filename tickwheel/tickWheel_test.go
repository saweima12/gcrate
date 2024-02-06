package tickwheel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/saweima12/gcrate/tickwheel"
)

type TestExecute struct {
	Name string
}

func (te *TestExecute) Execute() {
	fmt.Println(te.Name)
}

func TestTimingWheel(t *testing.T) {
	r := tickwheel.New(time.Second/2, 100).
		AddWheel(10).
		AddWheel(10)

	r.Start()

	r.AddTask(time.Second*5+1000, &TestExecute{Name: "John"})
	r.AddTask(time.Second*6, &TestExecute{Name: "Denny"})

	for {
		select {
		case executor := <-r.ExecCh():
			executor.Execute()

		case <-time.After(time.Second * 10):
			fmt.Printf("%#v\n", r)
			return
		}
	}
}
