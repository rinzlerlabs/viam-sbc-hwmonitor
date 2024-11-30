package cpu_monitor

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestCaptureCPUStats(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureCPUStats
	go sensor.task()

	for {
		if len(sensor.stats) > 0 {
			break
		}
	}
	cancel()
	sensor.wg.Wait()
	assert.Equal(t, runtime.NumCPU()+1, len(sensor.stats))
}

func TestCaptureCPUStatsExitsImmediately(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureCPUStats
	go sensor.task()
	sensor.cancelFunc()
	sensor.wg.Wait()
	assert.Equal(t, 0, len(sensor.stats))
}

func TestCaptureCPUStatsRespectsSleepTime(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      make(map[string]interface{}),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
		sleepTime:  100 * time.Millisecond,
	}

	sensor.task = sensor.captureCPUStats
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
	assert.Equal(t, runtime.NumCPU()+1, len(sensor.stats))
	testLength := end.Sub(now)
	logger.Infof("Test took %s", testLength)
	assert.True(t, testLength > 100*time.Millisecond)
	assert.True(t, testLength < 200*time.Millisecond)
}
