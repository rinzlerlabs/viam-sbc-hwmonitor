package processmonitor

import (
	"context"
	"testing"
	"time"

	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func TestGetProcessInfo(t *testing.T) {
	Test().RequiresRoot().ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensor := &Config{
		logger: logger,
		info:   &procInfo{Name: "sh"},
	}
	now := time.Now()
	readings, err := sensor.getCPUStats(ctx)
	require.NoError(t, err)
	elapsed := time.Since(now)
	logger.Infof("Elapsed time: %v", elapsed)
	assert.NoError(t, err)
	assert.NotNil(t, readings)
	assert.NotEmpty(t, readings)
	logger.Infof("Proc readings: %v", readings)
}

func TestGetProcessInfo_ProcessDoesNotExist(t *testing.T) {
	Test().RequiresRoot().ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensor := &Config{
		logger: logger,
		info:   &procInfo{Name: "1234"},
	}
	readings, err := sensor.getCPUStats(ctx)
	require.Error(t, err)
	require.Nil(t, readings)
}
