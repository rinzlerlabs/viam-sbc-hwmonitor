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
	jetson, err := newJetsonGpuMonitor(ctx, logger)
	assert.NoError(t, err)
	stats, err := jetson.GetGPUStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Len(t, stats, 6)
	for _, stat := range stats {
		logger.Infof("GPU: %v", stat)
		assert.NotEmpty(t, stat.Name)
		assert.Greater(t, stat.CurrentValue, int64(0))
		assert.Greater(t, stat.MaxValue, int64(0))
		assert.Greater(t, stat.MinValue, int64(0))
	}
}
