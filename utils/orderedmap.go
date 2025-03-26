package utils

import (
	"fmt"
	"iter"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

type OrderedMap[K comparable, V any] interface {
	Set(key K, value V) bool
	Get(key K) (V, bool)
	Delete(key K) bool
	Len() int
	Keys() iter.Seq[K]
	Values() iter.Seq[V]
	Front() *orderedmap.Element[K, V]
	Back() *orderedmap.Element[K, V]
	AllFromFront() iter.Seq2[K, V]
	AllFromBack() iter.Seq2[K, V]
	Has(key K) bool
}

type orderedMap[K comparable, V any] struct {
	*orderedmap.OrderedMap[K, V]
}

func NewOrderedMap[K comparable, V any]() OrderedMap[K, V] {
	return &orderedMap[K, V]{orderedmap.NewOrderedMap[K, V]()}
}

func (m *orderedMap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	i := 0
	for key, value := range m.AllFromFront() {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", key))
		sb.WriteString(": ")
		sb.WriteString(fmt.Sprintf("%v", value))
		i++
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *orderedMap[K, V]) Format(f fmt.State, c rune) {
	if f.Flag('#') {
		fmt.Fprintf(f, "%s", m.String())
	} else {
		fmt.Fprintf(f, "%v", m.String())
	}
}
