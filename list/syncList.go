package list

import (
	"container/list"
	"sync"
)

type SyncList[T comparable] struct {
	gl    *GenericList[T]
	mutex sync.Mutex
}

func NewSync[T comparable]() *SyncList[T] {
	return &SyncList[T]{
		gl: NewGeneric[T](),
	}
}

func (tl *SyncList[T]) PushFront(value T) *list.Element {
	tl.mutex.Lock()
	rtn := tl.gl.PushFront(value)
	tl.mutex.Unlock()
	return rtn
}

func (tl *SyncList[T]) PushBack(value T) *list.Element {
	tl.mutex.Lock()
	rtn := tl.gl.PushBack(value)
	tl.mutex.Unlock()
	return rtn
}

func (tl *SyncList[T]) PopAll() []T {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	return tl.gl.PopAll()
}

func (tl *SyncList[T]) PopFront() (T, bool) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	return tl.gl.PopFront()
}

func (tl *SyncList[T]) PopBack() (T, bool) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	return tl.gl.PopBack()
}

func (tl *SyncList[T]) RemoveFirst(t T) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.gl.RemoveFirst(t)
}
