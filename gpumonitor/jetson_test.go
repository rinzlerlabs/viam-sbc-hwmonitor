package gpumonitor

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func TestJetsonGpuGetsFrequencies(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	jetson, err := newJetsonGpuMonitor(logger)
	require.NoError(t, err)
	stats, err := jetson.GetGPUStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Len(t, stats, 7)
	for _, stat := range stats {
		logger.Infof("GPU: %#v", stat)
		assert.NotEmpty(t, stat.Name)
	}
}

func TestJetsonGpu_Readings(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	jetson, err := newJetsonGpuMonitor(logger)
	require.NoError(t, err)
	stats, err := jetson.GetGPUStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Len(t, stats, 7)
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
