package utils

import (
	"fmt"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
)

type OrderedMap[K comparable, V any] struct {
	*orderedmap.OrderedMap[K, V]
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{orderedmap.NewOrderedMap[K, V]()}
}

func (m *OrderedMap[K, V]) String() string {
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

func (m *OrderedMap[K, V]) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%v", m.String())
}
