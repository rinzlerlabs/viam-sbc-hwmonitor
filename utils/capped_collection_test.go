package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CappedCollection(t *testing.T) {
	collection := NewCappedCollection[int](5)
	assert.Len(t, collection.Items(), 0)
	assert.Equal(t, []int{}, collection.Items())

	collection.Push(1)
	assert.Len(t, collection.Items(), 1)
	assert.Equal(t, []int{1}, collection.Items())

	collection.Push(2)
	assert.Len(t, collection.Items(), 2)
	assert.Equal(t, []int{1, 2}, collection.Items())

	collection.Push(3)
	assert.Len(t, collection.Items(), 3)
	assert.Equal(t, []int{1, 2, 3}, collection.Items())

	collection.Push(4)
	assert.Len(t, collection.Items(), 4)
	assert.Equal(t, []int{1, 2, 3, 4}, collection.Items())

	collection.Push(5)
	assert.Len(t, collection.Items(), 5)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, collection.Items())

	collection.Push(6)
	assert.Len(t, collection.Items(), 5)
	assert.Equal(t, []int{6, 2, 3, 4, 5}, collection.Items())
}
