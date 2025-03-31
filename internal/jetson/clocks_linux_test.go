package jetson

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.viam.com/rdk/logging"
)

func TestGetNvidiaClockSensorsReturnsAllSensors(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()
	clocks, err := GetClockSensors(ctx, logger)
	assert.NoError(t, err)
	assert.NotNil(t, clocks)
	assert.Equal(t, runtime.NumCPU()+1, len(clocks))
	requiredKeys := []string{"gpu0"}
	for i := 0; i < runtime.NumCPU(); i++ {
		requiredKeys = append(requiredKeys, "cpu"+strconv.Itoa(i))
	}
	for i, clock := range clocks {
		t.Logf("Clock %d: %v", i, clock.Name())
		for j, key := range requiredKeys {
			if clock.Name() == key {
				requiredKeys = append(requiredKeys[:j], requiredKeys[j+1:]...)
				break
			}
		}
	}
	assert.Empty(t, requiredKeys)
}
