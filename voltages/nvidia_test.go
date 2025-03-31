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

func TestJetsonPowerSensors(t *testing.T) {
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	res, err := getJetsonPowerSensors(ctx, logger)
	require.NoError(t, err)
	require.NotNil(t, res)
	time.Sleep(1 * time.Second)
	for _, s := range res {
		require.NotNil(t, s)
		readings, err := s.GetReadingMap()
		require.NoError(t, err)
		assert.NotNil(t, readings)
		logger.Infof("s: %v", readings)
		defer s.Close()
	}
}
