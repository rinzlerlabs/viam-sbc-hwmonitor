package utils

import "sync"

type CappedCollection[T interface{}] interface {
	Push(item T)
	Items() []T
}

type cappedCollection[T interface{}] struct {
	mu       sync.Mutex
	items    []T
	size     int
	position int
}

// NewCappedCollection creates a new CappedCollection with a specified maximum size.
// The collection will hold items of any type specified by the generic type parameter T.
//
// Parameters:
//   - size: The maximum number of items the collection can hold.
//
// Returns:
//
//	A pointer to a new CappedCollection instance with the specified size.
func NewCappedCollection[T interface{}](size int) CappedCollection[T] {
	return &cappedCollection[T]{
		items: make([]T, 0, size),
		size:  size,
	}
}

// Push adds an item to the CappedCollection. If the collection has not yet
// reached its maximum size, the item is appended to the end. If the collection
// is full, the item replaces the oldest item in the collection, maintaining
// the capped size. This method is thread-safe.
func (q *cappedCollection[T]) Push(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) < q.size {
		q.items = append(q.items, item)
		return
	}
	q.items[q.position] = item
	q.position = (q.position + 1) % q.size
}

// Items returns a slice containing all the items in the CappedCollection.
// The method locks the collection to ensure thread safety during the read operation.
// It creates a new slice with a capacity equal to the size of the collection and appends all items to it.
// The returned slice is a copy of the collection's items at the time of the call.
func (q *cappedCollection[T]) Items() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	ret := make([]T, 0, q.size)
	ret = append(ret, q.items...)

	return ret
}
