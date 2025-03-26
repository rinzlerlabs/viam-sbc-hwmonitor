package raspberrypi

import (
	"errors"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

func NewPowerManager(config *RaspiConfig, logger logging.Logger) (*raspiPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return &raspiPowerManager{
		config: config,
		logger: logger,
	}, nil
}

type raspiPowerManager struct {
	config *RaspiConfig
	logger logging.Logger
}

type RaspiConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

func (pm *raspiPowerManager) ApplyPowerMode() (bool, error) {
	err := utils.InstallPackage("cpufrequtils")
	if err != nil {
		return false, errors.Join(err, errors.New("error installing cpufrequtils"))
	}

	return false, nil
}

func (pm *raspiPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, nil
}
