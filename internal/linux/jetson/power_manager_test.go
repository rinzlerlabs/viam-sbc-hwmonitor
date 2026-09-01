package jetson

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPowerModeOnlyAppliesOnce(t *testing.T) {
}

func TestIsRebootRequiredOutput(t *testing.T) {
	// Captured from a real device: `nvpmodel -m 1` declined via the
	// interactive prompt.
	output := `NVPM WARN: Golden image context is already created
NVPM WARN: Reboot required for changing to this power mode: 1
NVPM WARN: DO YOU WANT TO REBOOT NOW? enter YES/yes to confirm:
NVPM ERROR: optMask is 1, no request for power mode`
	require.True(t, isRebootRequiredOutput(output))

	require.False(t, isRebootRequiredOutput("NV Power Mode: MAXN\n0"))
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
