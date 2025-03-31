package voltages

import (
	"context"
	"testing"
	"time"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func TestRaspberryPiPowerSensors(t *testing.T) {
	Test().RequiresBoardType(boardtype.RaspberryPi).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	sensors, err := getRaspberryPiPowerSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, sensors)
	for _, s := range sensors {
		assert.NotNil(t, s)
	}
	waitForValues(t, sensors)
	for _, s := range sensors {
		assert.NotNil(t, s)
		m, err := s.GetReadingMap()
		require.NoError(t, err)
		assert.NotNil(t, m)
		for k, v := range m {
			logger.Infof("%s: %v", k, v)
			assert.NotEmpty(t, k)
			assert.NotZero(t, v)
		}
		defer s.Close()
	}
}

func waitForValues(t *testing.T, sensors []powerSensor) {
	timeout := time.Now().Add(10 * time.Second)
	for {
		if time.Now().After(timeout) {
			t.Fatal("Timed out waiting for sensors to have values")
		}
		allHaveValues := true
		for _, s := range sensors {
			m, err := s.GetReadingMap()
			require.NoError(t, err)
			if len(m) == 0 {
				allHaveValues = false
			}
			for _, v := range m {
				f, success := v.(float64)
				if !success {
					allHaveValues = false
				}
				if f == 0 {
					allHaveValues = false
				}
			}
		}
		if allHaveValues {
			break
		}
	}
}
