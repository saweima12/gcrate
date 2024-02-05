package core

import (
	"container/list"
	"sync"
)

type SafeList[T any] struct {
	mutex sync.Mutex
	list  *list.List
}

func NewSafeList[T any]() *SafeList[T] {
	return &SafeList[T]{
		list: list.New(),
	}
}

func (tl *SafeList[T]) PushBack(values ...T) {
	tl.mutex.Lock()
	for i := range values {
		tl.list.PushBack(values[i])
	}
	tl.mutex.Unlock()
}

func (tl *SafeList[T]) PopAll() []T {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	result := make([]T, 0, tl.list.Len())

	node := tl.popFront()
	for node != nil {
		item, ok := node.Value.(T)
		if !ok {
			continue
		}
		result = append(result, item)
		node = node.Next()
	}
	return result
}

func (tl *SafeList[T]) PopFront() (T, bool) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	var zero T
	node := tl.popFront()
	if node == nil {
		return zero, false
	}

	result, ok := node.Value.(T)
	if !ok {
		return zero, false
	}
	return result, true
}

func (tl *SafeList[T]) popFront() *list.Element {
	first := tl.list.Front()
	if first == nil {
		return nil
	}
	tl.list.Remove(first)
	return first
}
