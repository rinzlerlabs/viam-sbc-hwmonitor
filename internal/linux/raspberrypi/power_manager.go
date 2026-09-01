package raspberrypi

import (
	"errors"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/cpufrequtils"
	"go.viam.com/rdk/logging"
)

type PowerManagerConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

type raspiPowerManager struct {
	config *PowerManagerConfig
	logger logging.Logger
}

func NewPowerManager(config *PowerManagerConfig, logger logging.Logger) (*raspiPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}

	return &raspiPowerManager{
		config: config,
		logger: logger,
	}, nil
}

func (pm *raspiPowerManager) ApplyPowerMode() (bool, error) {
	cfg := pm.config
	if cfg.Governor == "" && cfg.Frequency == 0 && cfg.Minimum == 0 && cfg.Maximum == 0 {
		pm.logger.Info("No configuration changes made")
		return false, nil
	}

	output, err := cpufrequtils.ApplyPolicy(cfg.Governor, cfg.Frequency, cfg.Minimum, cfg.Maximum)
	if err != nil {
		pm.logger.Errorf("Error configuring CPU: %s: %s", err, output)
		return false, err
	}
	pm.logger.Infof("CPU configured: %s", output)
	return false, nil
}

func (pm *raspiPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, nil
}
