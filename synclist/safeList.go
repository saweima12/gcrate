package synclist

import (
	clist "container/list"
	"sync"
)

type Equatable[T any] interface {
	Equals(other T) bool
}

type SafeList[T Equatable[T]] struct {
	mutex sync.Mutex
	list  *clist.List
}

func New[T Equatable[T]]() *SafeList[T] {
	return &SafeList[T]{
		list: clist.New(),
	}
}

func (tl *SafeList[T]) PushFront(values ...T) {
	tl.mutex.Lock()
	for i := range values {
		tl.list.PushFront(values[i])
	}
	tl.mutex.Unlock()
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

	for {
		node := tl.remove(tl.list.Front())
		if node == nil {
			break
		}

		data, ok := node.Value.(T)
		if !ok {
			continue
		}
		result = append(result, data)
	}
	return result
}

func (tl *SafeList[T]) PopFront() (T, bool) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	var zero T
	node := tl.remove(tl.list.Front())
	if node == nil {
		return zero, false
	}

	result, ok := node.Value.(T)
	if !ok {
		return zero, false
	}
	return result, true
}

func (tl *SafeList[T]) PopBack() (T, bool) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	var zero T
	node := tl.remove(tl.list.Back())
	if node == nil {
		return zero, false
	}

	result, ok := node.Value.(T)
	if !ok {
		return zero, false
	}
	return result, true
}

func (tl *SafeList[T]) RemoveFirst(t T) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	node := tl.list.Front()
	for {
		if node == nil {
			return
		}

		val, ok := node.Value.(T)
		if !ok {
			continue
		}

		if val.Equals(t) {
			tl.list.Remove(node)
		}
		node = node.Next()
	}
}

func (tl *SafeList[T]) remove(node *clist.Element) *clist.Element {
	if node == nil {
		return nil
	}
	tl.list.Remove(node)
	return node
}
