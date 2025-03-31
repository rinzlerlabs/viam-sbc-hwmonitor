package raspberrypi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemperatureParse(t *testing.T) {
	str := "temp=47.2'C\n"
	temp, err := parseTemperature(str)
	require.NoError(t, err)
	require.Equal(t, 47.2, temp)
}
