package voltages

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestRaspberryPiPowerSensors(t *testing.T) {
	Test().RequiresBoardType(boardtype.RaspberryPi).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	res, err := getRaspberryPiPowerSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	for _, s := range res {
		assert.NotNil(t, s)
		assert.NoError(t, s.StartUpdating())
	}
	time.Sleep(2 * time.Second)
	for _, s := range res {
		assert.NotNil(t, s)
		logger.Infof("s: %v", s.GetReadingMap())
		defer s.Close()
	}
}
