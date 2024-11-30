package gpu_monitor

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestNvidiaGpuGetsFrequencies(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	nvidia, err := newNvidiaGpuMonitor(ctx, logger)
	assert.NoError(t, err)
	stats, err := nvidia.GetGPUStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Len(t, stats, 6)
	for _, stat := range stats {
		logger.Infof("GPU: %v", stat)
		assert.NotEmpty(t, stat.Name)
		assert.Greater(t, stat.CurrentFrequency, int64(0))
		assert.Greater(t, stat.MaxFrequency, int64(0))
		assert.Greater(t, stat.MinFrequency, int64(0))
		assert.NotEmpty(t, stat.Governor)
		assert.GreaterOrEqual(t, stat.Load, 0.0)
		assert.LessOrEqual(t, stat.Load, 100.0)
	}
}
