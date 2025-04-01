package linux

import (
	"errors"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

type LinuxConfig struct {
}

type linuxPowerManager struct {
}

func NewPowerManager(config *LinuxConfig, logger logging.Logger) (*linuxPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return nil, utils.ErrPlatformNotSupported
}

func (pm *linuxPowerManager) ApplyPowerMode() (rebootRequired bool, err error) {
	return false, utils.ErrPlatformNotSupported
}

func (pm *linuxPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, utils.ErrPlatformNotSupported
}
