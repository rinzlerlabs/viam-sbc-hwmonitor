package jetson

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func TestJetsonGpuGetsFrequencies(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	jetson, err := NewJetsonGpuMonitor(logger)
	require.NoError(t, err)
	gpuStats, err := jetson.GetGPUStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, gpuStats)
	require.Len(t, gpuStats, 7)
	for _, gpuStat := range gpuStats {
		for _, stat := range gpuStat {
			logger.Infof("GPU: %#v", stat)
			assert.NotEmpty(t, stat.Type)
		}
	}
}

func TestJetsonGPUReadingTypeMemoryFree(t *testing.T) {
	var sensor jetsonGpuSensor
	for _, sensor = range jetpack5Sensors {
		if sensor.sensorType == sensors.GPUReadingTypeMemoryFree {
			break
		}
	}

	matches := sensor.regex.FindStringSubmatch("Max allocatable IOVMM memory: 4078006272 bytes")
	require.Len(t, matches, 2)
	assert.Equal(t, "4078006272", matches[1])
	value, err := utils.ParseFloat64(matches[1])
	require.NoError(t, err)
	assert.Equal(t, 4078006272.0, value)
}
