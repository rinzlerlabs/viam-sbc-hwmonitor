package jetson

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetNvidiaClockSensorsReturnsAllSensors(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	clocks, err := GetClockSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)

	// Expect one clock sensor per CPU exposed via sysfs cpufreq. Derive the
	// expectation from the same source the code uses rather than runtime.NumCPU,
	// since the clock sensors are built from these sysfs paths.
	cpus, err := sensors.GetSysFsCpuPaths()
	assert.NoError(t, err)
	requiredKeys := make([]string, 0, len(cpus))
	for _, cpu := range cpus {
		requiredKeys = append(requiredKeys, filepath.Base(cpu))
	}

	names := make([]string, 0, len(clocks))
	for i, clock := range clocks {
		t.Logf("Clock %d: %v", i, clock.Name())
		names = append(names, clock.Name())
	}

	// Every CPU must have a clock sensor.
	for _, key := range requiredKeys {
		assert.Contains(t, names, key)
	}

	// A GPU clock sensor (gpu0) is reported only when a GPU devfreq node is
	// present, so the total is either one-per-CPU or one-per-CPU plus gpu0.
	switch len(clocks) {
	case len(requiredKeys):
		assert.NotContains(t, names, "gpu0")
	case len(requiredKeys) + 1:
		assert.Contains(t, names, "gpu0")
	default:
		t.Fatalf("unexpected clock sensor count: got %d, want %d or %d",
			len(clocks), len(requiredKeys), len(requiredKeys)+1)
	}
}
