package processmonitor

import (
	"context"
	"testing"
	"time"

	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetProcessInfo(t *testing.T) {
	Test().RequiresRoot().ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensor := &Config{
		logger: logger,
		info:   &procInfo{Name: "viam-agent", IncludeEnv: true, IncludeCmdline: true, IncludeCwd: true, IncludeOpenFiles: true, IncludeUlimits: true},
	}
	now := time.Now()
	readings, err := sensor.Readings(ctx, nil)
	elapsed := time.Since(now)
	logger.Infof("Elapsed time: %v", elapsed)
	assert.NoError(t, err)
	assert.NotNil(t, readings)
	assert.NotEmpty(t, readings)
	logger.Infof("Proc readings: %v", readings)
}
