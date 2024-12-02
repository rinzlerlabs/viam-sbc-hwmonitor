package memory_monitor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetMemory(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensor := &Config{}
	now := time.Now()
	readings, err := sensor.Readings(ctx, nil)
	elapsed := time.Since(now)
	logger.Infof("Elapsed time: %v", elapsed)
	assert.NoError(t, err)
	assert.NotNil(t, readings)
	logger.Infof("Memory readings: %v", readings)
}
