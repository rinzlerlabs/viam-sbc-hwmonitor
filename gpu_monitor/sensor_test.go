package gpu_monitor

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"

	. "github.com/thegreatco/gotestutils"
)

func TestCaptureCPUStats(t *testing.T) {
	Test().RequiresBoardType(nvidia.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureGPUStats
	go sensor.task()

	for {
		if len(sensor.stats) > 0 {
			break
		}
	}
	cancel()
	sensor.wg.Wait()
	assert.Equal(t, 5, len(sensor.stats))
}

func TestCaptureCPUStatsExitsImmediately(t *testing.T) {
	Test().RequiresBoardType(nvidia.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureGPUStats
	go sensor.task()
	sensor.cancelFunc()
	sensor.wg.Wait()
	assert.Equal(t, 0, len(sensor.stats))
}

func TestCaptureCPUStatsRespectsSleepTime(t *testing.T) {
	Test().RequiresBoardType(nvidia.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
		sleepTime:  100 * time.Millisecond,
	}

	sensor.task = sensor.captureGPUStats
	now := time.Now()
	go sensor.task()

	for {
		if len(sensor.stats) > 0 {
			break
		}
	}
	cancel()
	sensor.wg.Wait()
	end := time.Now()
	assert.Equal(t, 5, len(sensor.stats))
	testLength := end.Sub(now)
	logger.Infof("Test took %s", testLength)
	assert.True(t, testLength > 100*time.Millisecond)
	assert.True(t, testLength < 200*time.Millisecond)
}
