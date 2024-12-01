package clocks

import (
	"context"
	"runtime"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"

	. "github.com/rinzlerlabs/sbcidentify/test"
)

func TestGetNvidiaClockSensors(t *testing.T) {
	logger := logging.NewTestLogger(t)
	Test().RequiresBoardType(boardtype.NVIDIA).ShouldSkip(t)
	ctx := context.Background()
	clocks, err := getClockSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	assert.Equal(t, runtime.NumCPU()+1, len(clocks))
	for i, clock := range clocks {
		t.Logf("Clock %d: %v", i, clock.GetName())
	}
}
