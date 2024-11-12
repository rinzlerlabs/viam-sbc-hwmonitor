package throttling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetThrottlingStates(t *testing.T) {
	res, err := parseThrottlingStates("throttled=0xe0008")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res[Undervolt])
	assert.False(t, res[ArmFrequencyCapped])
	assert.False(t, res[CurrentlyThrottled])
	assert.True(t, res[SoftTempLimitActive])
	assert.False(t, res[UnderVoltOccurred])
	assert.True(t, res[ArmFrequencyCapOccurred])
	assert.True(t, res[ThrottlingOccurred])
	assert.True(t, res[SoftTempLimitOccurred])

	res, err = parseThrottlingStates("throttled=0xe0006")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res[Undervolt])
	assert.True(t, res[ArmFrequencyCapped])
	assert.True(t, res[CurrentlyThrottled])
	assert.False(t, res[SoftTempLimitActive])
	assert.False(t, res[UnderVoltOccurred])
	assert.True(t, res[ArmFrequencyCapOccurred])
	assert.True(t, res[ThrottlingOccurred])
	assert.True(t, res[SoftTempLimitOccurred])
}
