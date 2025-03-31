package jetson

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
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
