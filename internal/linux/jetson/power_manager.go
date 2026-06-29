package jetson

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"
)

type PowerManagerConfig struct {
	PowerMode int    `json:"power_mode"`
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

type jetsonPowerManager struct {
	config *PowerManagerConfig
	logger logging.Logger
}

func NewPowerManager(config *PowerManagerConfig, logger logging.Logger) (*jetsonPowerManager, error) {
	if config == nil {
		return nil, errors.New("configuration cannot be nil")
	}
	return &jetsonPowerManager{
		config: config,
		logger: logger,
	}, nil
}

func (pm *jetsonPowerManager) ApplyPowerMode() (rebootRequired bool, err error) {
	currentPowerMode, err := pm.getCurrentPowerMode()
	if err != nil {
		return false, fmt.Errorf("failed to get current power mode: %v", err)
	}
	if currentPowerMode == pm.config.PowerMode {
		pm.logger.Debugf("Power mode is already set to %d", pm.config.PowerMode)
		return false, nil
	}

	cmd := exec.Command("nvpmodel", "-m", fmt.Sprintf("%d", pm.config.PowerMode))
	// Decline nvpmodel's interactive reboot prompt; we never reboot the device
	// automatically. A reboot-required change is reported back to the caller.
	cmd.Stdin = strings.NewReader("no\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// nvpmodel exits non-zero when the requested mode needs a reboot and our
		// "no" answer aborted it. Treat that as a non-fatal "reboot required"
		// result rather than failing the component build.
		if isRebootRequiredOutput(string(output)) {
			pm.logger.Warnf("Power mode %d requires a reboot to take effect. Run 'sudo nvpmodel -m %d' on the device (confirm the reboot prompt), or reboot after running it, to apply it.", pm.config.PowerMode, pm.config.PowerMode)
			return true, nil
		}
		return false, fmt.Errorf("failed to set power mode: %v, output: %s", err, string(output))
	}
	// nvpmodel applied the mode immediately; no reboot needed.
	return false, nil
}

// isRebootRequiredOutput reports whether nvpmodel declined the mode change
// because a reboot is required (our non-interactive "no" answer aborted it).
func isRebootRequiredOutput(output string) bool {
	return strings.Contains(output, "Reboot required")
}

func (pm *jetsonPowerManager) GetCurrentPowerMode() (interface{}, error) {
	return pm.getCurrentPowerMode()
}

// getCurrentPowerMode returns the active nvpmodel power mode as an integer.
func (pm *jetsonPowerManager) getCurrentPowerMode() (int, error) {
	cmd := exec.Command("nvpmodel", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get current power mode: %v, output: %s", err, string(output))
	}
	return parsePowerModeOutput(string(output))
}

// parsePowerModeOutput extracts the active power mode from `nvpmodel -q` output.
// The output contains a label line ("NV Power Mode: MAXN") and a separate line
// with the numeric mode, but the line ordering/blank lines vary across boards,
// so we scan for the first line that parses as an integer rather than assuming
// a fixed index.
func parsePowerModeOutput(output string) (int, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if mode, err := strconv.Atoi(line); err == nil {
			return mode, nil
		}
	}
	return 0, fmt.Errorf("could not find power mode in nvpmodel output: %q", output)
}
