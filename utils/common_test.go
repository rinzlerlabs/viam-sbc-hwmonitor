package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemperatureParse(t *testing.T) {
	str := "temp=47.2'C\n"
	temp, err := parseTemperature(str)
	assert.NoError(t, err)
	assert.Equal(t, 47.2, temp)
}

func TestGetBoardType(t *testing.T) {
	board, e := GetBoardType()
	assert.NoError(t, e)
	assert.Equal(t, RaspberryPi5, board)
}
