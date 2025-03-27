package gpumonitor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func skipIfNoNvidiaDriver(t *testing.T) {
	if hasNvidiaSmiCommand() == false {
		t.Skipf("This test requires an NVIDIA gpu and the nvidia-smi command to be present")
	}
}

func TestNvidiaGPU(t *testing.T) {
	skipIfNoNvidiaDriver(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	monitor, err := newNVIDIAGpuMonitor(logger)
	require.NoError(t, err)
	require.NotNil(t, monitor)
	stats, err := monitor.GetGPUStats(ctx)
	require.NoError(t, err)
	require.Len(t, stats, 12)
	logger.Infof("GPU Stats: %#v", stats)
}

func TestNvidiaGpu_Readings(t *testing.T) {
	skipIfNoNvidiaDriver(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	jetson, err := newNVIDIAGpuMonitor(logger)
	require.NoError(t, err)
	stats, err := jetson.GetGPUStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Len(t, stats, 12)
	sensor := &Config{
		logger: logger,
		stats:  utils.NewCappedCollection[sample](10),
	}
	sensor.stats.Push(sample{DeviceStats: stats})
	res, err := sensor.Readings(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	logger.Infof("Readings: %#v", res)
}
