//go:build linux
// +build linux

package clocks

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.NotNil(t, clock)
	}
	assert.NoError(t, err)
	sensor := &Config{
		sensors:    clocks,
		logger:     logger,
		cancelCtx:  ctx,
		cancelFunc: cancelFunc,
	}
	time.Sleep(2 * time.Second)
	sensor.Close(ctx)
	readings, err := sensor.Readings(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(clocks), len(readings))
	for i, reading := range readings {
		t.Logf("Reading %s: %v", i, reading)
	}
}
