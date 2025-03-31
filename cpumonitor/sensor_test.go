package cpumonitor

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
	viamutils "go.viam.com/utils"
)

func TestCaptureCPUStats(t *testing.T) {
	logger := logging.NewTestLogger(t)
	sensor := &Config{
		logger: logger,
	}

	sensor.workers = viamutils.NewBackgroundStoppableWorkers(sensor.startUpdating)

	for {
		if len(sensor.reading) > 0 {
			break
		}
	}
	sensor.Close(context.Background())
	assert.Equal(t, runtime.NumCPU()+1, len(sensor.reading))
}

func TestCaptureCPUStatsExitsImmediately(t *testing.T) {
	logger := logging.NewTestLogger(t)
	sensor := &Config{
		logger: logger,
	}

	sensor.workers = viamutils.NewBackgroundStoppableWorkers(sensor.startUpdating)
	start := time.Now()
	sensor.Close(context.Background())
	end := time.Now()
	assert.Less(t, end.Sub(start), 100*time.Millisecond)
}

func TestCaptureCPUStatsRespectsSleepTime(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensor := &Config{
		logger:    logger,
		sleepTime: 100 * time.Millisecond,
	}

	now := time.Now()
	sensor.workers = viamutils.NewBackgroundStoppableWorkers(sensor.startUpdating)

	for {
		if len(sensor.reading) > 0 {
			break
		}
	}
	sensor.Close(ctx)
	end := time.Now()
	assert.Equal(t, runtime.NumCPU()+1, len(sensor.reading))
	testLength := end.Sub(now)
	logger.Infof("Test took %s", testLength)
	assert.True(t, testLength > 100*time.Millisecond)
	assert.True(t, testLength < 200*time.Millisecond)
}
