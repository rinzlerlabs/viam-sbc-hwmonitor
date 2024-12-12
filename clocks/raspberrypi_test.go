package clocks

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetRaspPiClockSensorsReturnsAllSensors(t *testing.T) {
	Test().RequiresBoardType(boardtype.RaspberryPi4).ShouldSkip(t)
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	clocks, err := getRaspberryPiClockSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	assert.Equal(t, runtime.NumCPU()+12, len(clocks))
	requiredKeys := []string{"arm", "core", "h264", "isp", "v3d", "uart", "pwm", "emmc", "pixel", "vec", "hdmi", "dpi"}
	for i := 0; i < runtime.NumCPU(); i++ {
		requiredKeys = append(requiredKeys, "cpu"+strconv.Itoa(i))
	}
	for i, clock := range clocks {
		t.Logf("Clock %d: %v", i, clock.GetName())
		for j, key := range requiredKeys {
			if clock.GetName() == key {
				requiredKeys = append(requiredKeys[:j], requiredKeys[j+1:]...)
				break
			}
		}
	}

	assert.Empty(t, requiredKeys)
}
