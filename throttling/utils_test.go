package throttling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetThrottlingStatesForRasPi(t *testing.T) {
	res, err := parseRasPiThrottlingStates("throttled=0xe0008")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res[Undervolt].(bool))
	assert.False(t, res[CurrentlyThrottled].(bool))
	assert.True(t, res[SoftTempLimitActive].(bool))
	assert.False(t, res[UnderVoltOccurred].(bool))
	assert.False(t, res[ArmFrequencyCapped].(bool))
	assert.True(t, res[ArmFrequencyCapOccurred].(bool))
	assert.True(t, res[ThrottlingOccurred].(bool))
	assert.True(t, res[SoftTempLimitOccurred].(bool))

	res, err = parseRasPiThrottlingStates("throttled=0xe0006")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res[Undervolt].(bool))
	assert.True(t, res[ArmFrequencyCapped].(bool))
	assert.True(t, res[CurrentlyThrottled].(bool))
	assert.False(t, res[SoftTempLimitActive].(bool))
	assert.False(t, res[UnderVoltOccurred].(bool))
	assert.True(t, res[ArmFrequencyCapOccurred].(bool))
	assert.True(t, res[ThrottlingOccurred].(bool))
	assert.True(t, res[SoftTempLimitOccurred].(bool))
}
