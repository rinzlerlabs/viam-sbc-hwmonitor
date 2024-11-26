package clocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseClockFrequency(t *testing.T) {
	s := []string{
		"frequency(48)=1800404352",
		"frequency(1)=500000992",
		"frequency(28)=0",
	}
	a := []int{
		1800404352,
		500000992,
		0,
	}

	for i, v := range s {
		freq, err := parseRasPiClockFrequency(v)
		assert.NoError(t, err)
		assert.Equal(t, a[i], freq)
	}
}
