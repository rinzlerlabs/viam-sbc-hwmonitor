package gpu_monitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func TestNvidiaGPU(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	monitor, err := newNVIDIAGpuMonitor(logger)
	require.NoError(t, err)
	require.NotNil(t, monitor)
	stats, err := monitor.GetGPUStats(ctx)
	require.NoError(t, err)
	require.True(t, len(stats) > 0, "Expected GPU stats to be non-empty")
	logger.Infof("GPU Stats: %v", stats)
}
