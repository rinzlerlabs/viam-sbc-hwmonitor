package raspberrypi

import (
	"errors"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

func NewPowerManager(config *PowerManagerConfig, logger logging.Logger) (*raspiPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}

	return nil, utils.ErrPlatformNotSupported
}

func (pm *raspiPowerManager) ApplyPowerMode() (bool, error) {

	return false, utils.ErrPlatformNotSupported
}

func (pm *raspiPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return false, utils.ErrPlatformNotSupported
}
