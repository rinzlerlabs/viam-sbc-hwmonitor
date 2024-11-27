package utils

import "sync"

type CappedCollection[T interface{}] struct {
	mu       sync.Mutex
	items    []T
	size     int
	position int
}

func NewCappedCollection[T interface{}](size int) *CappedCollection[T] {
	return &CappedCollection[T]{
		items: make([]T, 0, size),
		size:  size,
	}
}

func (q *CappedCollection[T]) Push(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) < q.size {
		q.items = append(q.items, item)
		return
	}
	q.items[q.position] = item
	q.position = (q.position + 1) % q.size
}

func (q *CappedCollection[T]) Items() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	ret := make([]T, 0, q.size)
	for _, item := range q.items {
		ret = append(ret, item)
	}

	return ret
}
