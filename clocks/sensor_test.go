package clocks

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"

	. "github.com/rinzlerlabs/sbcidentify/test"
)

func TestRaspberryPiGetReadings(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx, cancelFunc := context.WithCancel(context.Background())
	Test().RequiresRoot().RequiresBoardType(boardtype.RaspberryPi).ShouldSkip(t)
	clocks, err := getClockSensors(ctx, logger)
	assert.NoError(t, err)
	sensor := &Config{
		sensors:    clocks,
		logger:     logger,
		cancelCtx:  ctx,
		cancelFunc: cancelFunc,
	}
	readings, err := sensor.Readings(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(clocks), len(readings))
}

func TestNvidiaGetReadings(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx, cancelFunc := context.WithCancel(context.Background())
	Test().RequiresRoot().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	clocks, err := getClockSensors(ctx, logger)
	for _, clock := range clocks {
		err := clock.StartUpdating()
		assert.NoError(t, err)
	}
	assert.NoError(t, err)
	sensor := &Config{
		sensors:    clocks,
		logger:     logger,
		cancelCtx:  ctx,
		cancelFunc: cancelFunc,
	}
	time.Sleep(1 * time.Second)
	cancelFunc()
	readings, err := sensor.Readings(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(clocks), len(readings))
	for i, reading := range readings {
		t.Logf("Reading %s: %v", i, reading)
	}
}
