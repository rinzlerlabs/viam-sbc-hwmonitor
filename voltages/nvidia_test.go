package voltages

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestFoo(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	res, err := getJetsonPowerSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	time.Sleep(1 * time.Second)
	for _, s := range res {
		assert.NotNil(t, s)
		logger.Infof("s: %v", s.GetReadingMap())
		defer s.Close()
	}
}
