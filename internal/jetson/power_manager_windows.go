package jetson

import (
	"errors"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

func NewPowerManager(config *PowerManagerConfig, logger logging.Logger) (*jetsonPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return nil, utils.ErrPlatformNotSupported
}

func (pm *jetsonPowerManager) ApplyPowerMode() (rebootRequired bool, err error) {
	return false, utils.ErrPlatformNotSupported
}

func (pm *jetsonPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, utils.ErrPlatformNotSupported
}
