package gpumonitor

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestCaptureCPUStats(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      utils.NewCappedCollection[sample](5),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureGPUStats
	go sensor.task()

	for {
		if len(sensor.stats.Items()) > 0 {
			break
		}
	}
	cancel()
	sensor.wg.Wait()
	assert.Equal(t, 1, len(sensor.stats.Items()))
}

func TestCaptureCPUStatsExitsImmediately(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      utils.NewCappedCollection[sample](5),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
	}

	sensor.task = sensor.captureGPUStats
	go sensor.task()
	sensor.cancelFunc()
	sensor.wg.Wait()
	assert.Equal(t, 0, len(sensor.stats.Items()))
}

func TestCaptureCPUStatsRespectsSleepTime(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	sensor := &Config{
		stats:      utils.NewCappedCollection[sample](5),
		cancelCtx:  ctx,
		cancelFunc: cancel,
		logger:     logger,
		sleepTime:  100 * time.Millisecond,
	}

	sensor.task = sensor.captureGPUStats
	now := time.Now()
	go sensor.task()

	for {
		if len(sensor.stats.Items()) > 0 {
			break
		}
	}
	cancel()
	sensor.wg.Wait()
	end := time.Now()
	assert.Equal(t, 1, len(sensor.stats.Items()))
	testLength := end.Sub(now)
	logger.Infof("Test took %s", testLength)
	assert.True(t, testLength > 100*time.Millisecond)
	assert.True(t, testLength < 200*time.Millisecond)
}

func TestGetReadings(t *testing.T) {
	t.Skip("Skipping for now until i redo this implementation")
	// Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	// logger := logging.NewTestLogger(t)
	// ctx, cancel := context.WithCancel(context.Background())
	// sensor := &Config{
	// 	stats:      utils.NewCappedCollection[sample](5),
	// 	cancelCtx:  ctx,
	// 	cancelFunc: cancel,
	// 	logger:     logger,
	// }
	// sensor.stats.Push(sample{
	// 	DeviceStats: []gpuComponentStats{
	// 		{Name: "gpu0", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 50},
	// 		{Name: "gpu1", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 50},
	// 	},
	// })
	// sensor.stats.Push(sample{
	// 	DeviceStats: []gpuComponentStats{
	// 		{Name: "gpu0", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 60},
	// 		{Name: "gpu1", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 40},
	// 	},
	// })
	// sensor.stats.Push(sample{
	// 	DeviceStats: []gpuComponentStats{
	// 		{Name: "gpu0", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 70},
	// 		{Name: "gpu1", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 30},
	// 	},
	// })
	// sensor.stats.Push(sample{
	// 	DeviceStats: []gpuComponentStats{
	// 		{Name: "gpu0", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 80},
	// 		{Name: "gpu1", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 20},
	// 	},
	// })
	// sensor.stats.Push(sample{
	// 	DeviceStats: []gpuComponentStats{
	// 		{Name: "gpu0", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 90},
	// 		{Name: "gpu1", CurrentValue: 100, MaxValue: 200, MinValue: 50, Governor: "test", Load: 10},
	// 	},
	// })

	// readings, err := sensor.Readings(ctx, nil)
	// assert.Nil(t, err)
	// assert.Equal(t, 8, len(readings))
	// logger.Infof("Readings: %v", readings)
	// assert.Equal(t, 70.0, readings["gpu0-load"])
	// assert.Equal(t, 30.0, readings["gpu1-load"])
}

func TestCollectionSizeCalculation(t *testing.T) {
	sleepTime := 100 * time.Millisecond
	assert.Equal(t, 10, calculateCollectionSize(sleepTime))

	sleepTime = 1 * time.Second
	assert.Equal(t, 1, calculateCollectionSize(sleepTime))

	sleepTime = 1 * time.Millisecond
	assert.Equal(t, 1000, calculateCollectionSize(sleepTime))

	sleepTime = 10 * time.Second
	assert.Equal(t, 1, calculateCollectionSize(sleepTime))
}
