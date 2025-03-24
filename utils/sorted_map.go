package utils

import "sort"

// SortedMap maintains a map[string]any with keys sorted lexicographically.
type SortedMap[V any] struct {
	data map[string]V
	keys []string
}

// New creates a new empty SortedMap.
func New[V any]() *SortedMap[V] {
	return &SortedMap[V]{
		data: make(map[string]V),
		keys: []string{},
	}
}

// Set inserts or updates a key-value pair and keeps keys sorted.
func (sm *SortedMap[V]) Set(key string, value V) {
	_, exists := sm.data[key]
	sm.data[key] = value
	if !exists {
		sm.keys = append(sm.keys, key)
		sort.Strings(sm.keys)
	}
}

// Get retrieves the value and a boolean if it exists.
func (sm *SortedMap[V]) Get(key string) (V, bool) {
	val, ok := sm.data[key]
	return val, ok
}

// Delete removes a key if it exists.
func (sm *SortedMap[V]) Delete(key string) {
	if _, ok := sm.data[key]; ok {
		delete(sm.data, key)
		for i, k := range sm.keys {
			if k == key {
				sm.keys = append(sm.keys[:i], sm.keys[i+1:]...)
				break
			}
		}
	}
}

// Keys returns the sorted list of keys.
func (sm *SortedMap[V]) Keys() []string {
	return sm.keys
}

// Values returns values in key-sorted order.
func (sm *SortedMap[V]) Values() []V {
	values := make([]V, 0, len(sm.keys))
	for _, k := range sm.keys {
		values = append(values, sm.data[k])
	}
	return values
}

// Range calls the callback in sorted key order.
func (sm *SortedMap[V]) Range(f func(key string, value V)) {
	for _, k := range sm.keys {
		f(k, sm.data[k])
	}
}
