package gpumonitor

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func skipIfNoNvidiaDriver(t *testing.T) {
	logger := logging.NewTestLogger(t)
	if hasNvidiaSmiCommand(logger) == false {
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
	for key, stat := range stats {
		require.NotEmpty(t, stat)
		require.NotEmpty(t, key)
		for _, reading := range stat {
			require.NotEmpty(t, reading)
			require.NotEmpty(t, reading.Type)
			require.NotNil(t, reading.Value)
			logger.Infof("Sensor: %s, Name: %s, Value: %v", key, reading.Type, reading.Value)
		}
	}
}

func TestNvidiaSmiParsing(t *testing.T) {
	b, err := os.ReadFile("testdata/nvidia-smi.out")
	require.NoError(t, err)
	gpuMonitor := &nvidiaGpuMonitor{
		logger: logging.NewTestLogger(t),
	}
	stats, err := gpuMonitor.parseNvidiaSmiOutput(b)
	require.NoError(t, err)
	require.NotNil(t, stats)
	for gpu, stat := range stats {
		require.NotEmpty(t, gpu)
		for _, reading := range stat {
			require.NotEmpty(t, reading)
			require.NotEmpty(t, reading.Type)
			require.NotNil(t, reading.Value)
		}
	}
}

// func TestNvidiaGpu_Readings(t *testing.T) {
// 	skipIfNoNvidiaDriver(t)
// 	ctx := context.Background()
// 	logger := logging.NewTestLogger(t)
// 	jetson, err := newNVIDIAGpuMonitor(logger)
// 	require.NoError(t, err)
// 	stats, err := jetson.GetGPUStats(ctx)
// 	require.NoError(t, err)
// 	require.NotNil(t, stats)
// 	require.Len(t, stats, 12)
// 	sensor := &Config{
// 		logger: logger,
// 		stats:  utils.NewCappedCollection[sample](10),
// 	}
// 	sensor.stats.Push(sample{DeviceStats: stats})
// 	res, err := sensor.Readings(ctx, nil)
// 	require.NoError(t, err)
// 	require.NotNil(t, res)
// 	logger.Infof("Readings: %#v", res)
// }
