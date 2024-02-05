package timingwheel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/saweima12/gcrate/timingwheel"
)

type TestExecute struct {
	Name string
}

func (te *TestExecute) Execute() {
	fmt.Println(te.Name)
}

func Test_TimingWheel(t *testing.T) {
	r := timingwheel.New(time.Second/2, 100).
		AddWheel(10).
		AddWheel(10)

	r.Start()

	for {
		select {
		case executor := <-r.ExecCh():
			executor.Execute()

		case <-time.After(time.Second * 60):
			fmt.Printf("%#v\n", r)
			return
		}
	}

}
