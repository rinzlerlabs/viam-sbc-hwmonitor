package pwm_fan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestGetInternalFanSpeed(t *testing.T) {
	ctx := context.Background()
	fan, err := newFan(nil, "", "", true)
	assert.NoError(t, err)
	assert.NotNil(t, fan)
	assert.NotNil(t, fan.internalFan)
	speed, err := fan.GetSpeed(ctx)
	assert.NoError(t, err)
	assert.NotZero(t, speed)
}

func TestSetInternalFanSpeed(t *testing.T) {
	ctx := context.Background()
	fan, err := newFan(nil, "", "", true)
	assert.NoError(t, err)
	assert.NotNil(t, fan)
	assert.NotNil(t, fan.internalFan)

	err = fan.SetSpeed(ctx, 1)
	assert.NoError(t, err)

	speed, err := fan.GetSpeed(ctx)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), speed)

	err = fan.SetSpeed(ctx, 0)
	assert.NoError(t, err)

	speed, err = fan.GetSpeed(ctx)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), speed)
}
