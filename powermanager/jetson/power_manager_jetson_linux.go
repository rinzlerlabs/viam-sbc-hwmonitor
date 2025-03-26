package jetson

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"
)

func NewPowerManager(config *JetsonConfig, logger logging.Logger) (*jetsonPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return &jetsonPowerManager{
		config: config,
		logger: logger,
	}, nil
}

type jetsonPowerManager struct {
	config *JetsonConfig
	logger logging.Logger
}

type JetsonConfig struct {
	PowerMode int    `json:"power_mode"`
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

func (pm *jetsonPowerManager) ApplyPowerMode() (rebootRequired bool, err error) {
	currentPowerMode, err := pm.GetCurrentPowerMode()
	if err != nil {
		return false, fmt.Errorf("failed to get current power mode: %v", err)
	}
	if currentPowerMode == nil {
		return false, errors.New("current power mode is nil")
	}
	if currentPowerMode == pm.config.PowerMode {
		pm.logger.Debugf("Power mode is already set to %d", pm.config.PowerMode)
		return false, nil
	}
	cmd := exec.Command("nvpmodel", "-m", fmt.Sprintf("%d", pm.config.PowerMode))
	cmd.Stdin = strings.NewReader("no\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to set power mode: %v, output: %s", err, string(output))
	}
	return true, nil
}

func (pm *jetsonPowerManager) GetCurrentPowerMode() (interface{}, error) {
	cmd := exec.Command("nvpmodel", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get current power mode: %v, output: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func parsePowerModeOutput(output string) (int, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return 0, errors.New("unexpected output format")
	}
	powerModeLine := lines[1]
	powerMode := strings.TrimSpace(powerModeLine)
	powerModeInt, err := strconv.Atoi(powerMode)
	if err != nil {
		return 0, fmt.Errorf("failed to parse power mode: %v", err)
	}
	return powerModeInt, nil
}
