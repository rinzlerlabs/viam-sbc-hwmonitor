package windows

import (
	"errors"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

type WindowsConfig struct {
}

type windowsPowerManager struct {
	config *WindowsConfig
	logger logging.Logger
}

func NewPowerManager(config *WindowsConfig, logger logging.Logger) (*windowsPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return nil, utils.ErrPlatformNotSupported
}

func (pm *windowsPowerManager) ApplyPowerMode() (rebootRequired bool, err error) {
	return false, utils.ErrPlatformNotSupported
}

func (pm *windowsPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, utils.ErrPlatformNotSupported
}
