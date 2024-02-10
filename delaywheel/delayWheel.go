package delaywheel

import (
	"time"

	"github.com/saweima12/gcrate/pqueue"
)

type DelayWheel interface {
	Start()
	Stop()
}

func nowFunc() int64 {
	return timeToMs(time.Now().UTC())
}

func NewDelayedWheel(tickMs time.Duration, wheelSize int) DelayWheel {
	startMs := timeToMs(time.Now().UTC())

	return createDelayWheel(
		int64(tickMs),
		wheelSize,
		startMs,
	)
}

func createDelayWheel(tickMs int64, wheelSize int, startMs int64) DelayWheel {
	return &delayWheel{
		tickMs:      time.Duration(tickMs),
		wheelSize:   wheelSize,
		startMs:     startMs,
		currentTime: truncate(startMs, tickMs),
		maxInterval: tickMs * int64(wheelSize),

		queue: pqueue.NewDelayQueue[Bucket](nowFunc, wheelSize),
		wheel: newWheel(tickMs, wheelSize),

		stopCh: make(chan struct{}, 1),
	}
}

type delayWheel struct {
	tickMs      time.Duration
	wheelSize   int
	maxInterval int64
	startMs     int64
	currentTime int64

	wheel *tWheel
	queue pqueue.DelayQueue[Bucket]

	stopCh chan struct{}
}

func (de *delayWheel) Start() {
	go de.run()
}

func (de *delayWheel) Stop() {
	de.stopCh <- struct{}{}
}

func (de *delayWheel) run() {
	de.queue.Start()

	for {
		select {
		case <-de.stopCh:
			de.queue.Stop()
			return
		}
	}
}

func (de *delayWheel) addWheel() {
	target := de.wheel
	for target.next != nil {
		target = target.next
	}
}
