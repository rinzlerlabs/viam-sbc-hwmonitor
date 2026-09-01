package cpumonitor

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
	viamutils "go.viam.com/utils"
)

func TestCaptureCPUStats(t *testing.T) {
	logger := logging.NewTestLogger(t)
	sensor := &Config{
		logger:    logger,
		sleepTime: 1 * time.Second,
	}

	sensor.workers = viamutils.NewBackgroundStoppableWorkers(sensor.startUpdating)

	var reading map[string]any
	for {
		r, err := sensor.Readings(context.Background(), nil)
		require.NoError(t, err)
		if len(r) > 0 {
			reading = r
			break
		}
	}
	sensor.Close(context.Background())
	require.Equal(t, runtime.NumCPU()+1, len(reading))
	for k, v := range reading {
		logger.Infof("%v: %v", k, v)
	}
}

func TestCaptureCPUStatsExitsImmediately(t *testing.T) {
	logger := logging.NewTestLogger(t)
	sensor := &Config{
		logger:    logger,
		sleepTime: 1 * time.Second,
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

	var reading map[string]any
	for {
		r, err := sensor.Readings(ctx, nil)
		require.NoError(t, err)
		if len(r) > 0 {
			reading = r
			break
		}
	}
	sensor.Close(ctx)
	end := time.Now()
	assert.Equal(t, runtime.NumCPU()+1, len(reading))
	testLength := end.Sub(now)
	logger.Infof("Test took %s", testLength)
	assert.True(t, testLength > 100*time.Millisecond)
	assert.True(t, testLength < 200*time.Millisecond)
}
