package jetson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPowerModeOnlyAppliesOnce(t *testing.T) {
}

func TestParsePowerModeOutput(t *testing.T) {
	output := `NV Power Mode: 15W
0`
	powerMode, err := parsePowerModeOutput(output)
	require.NoError(t, err)
	require.Equal(t, 0, powerMode)

	desiredPowerMode := 0

	if powerMode != desiredPowerMode { // This is just used to confirm that testing an int returned as interface{} can still be compared to an int reliably
		t.Fatalf("Power mode should be equal to desired power mode %d, but got %d", desiredPowerMode, powerMode)
	}

	output = `NV Power Mode: 7W
1`
	powerMode, err = parsePowerModeOutput(output)
	require.NoError(t, err)
	require.Equal(t, 1, powerMode)

	if powerMode == desiredPowerMode {
		t.Fatalf("Power mode should be equal to desired power mode %d, but got %d", desiredPowerMode, powerMode)
	}
}
