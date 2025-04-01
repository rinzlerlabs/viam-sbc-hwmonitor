package raspberrypi

import (
	"errors"
	"os/exec"
	"strconv"

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
	args := make([]string, 0)
	if pm.config.Governor != "" {
		args = append(args, "--governor", pm.config.Governor)
	}
	if pm.config.Frequency != 0 {
		args = append(args, "--freq", strconv.Itoa(pm.config.Frequency))
	}
	if pm.config.Minimum != 0 {
		args = append(args, "--min", strconv.Itoa(pm.config.Minimum))
	}
	if pm.config.Maximum != 0 {
		args = append(args, "--max", strconv.Itoa(pm.config.Maximum))
	}

	if len(args) > 0 {
		proc := exec.Command("cpufreq-set", args...)

		outputBytes, err := proc.Output()
		if err != nil {
			pm.logger.Errorf("Error configuring CPU: %s", err)
		}
		pm.logger.Infof("CPU configured: %s", string(outputBytes))
	} else {
		pm.logger.Info("No configuration changes made")
	}
	return false, nil
}

func (pm *raspiPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return nil, nil
}
