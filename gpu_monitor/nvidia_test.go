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
	require.Len(t, stats, 11)
	logger.Infof("GPU Stats: %#v", stats)
}
