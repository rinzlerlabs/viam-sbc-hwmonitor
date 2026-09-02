package jetson

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/cpufrequtils"
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
	} else {
		cmd := exec.Command("nvpmodel", "-m", fmt.Sprintf("%d", pm.config.PowerMode))
		// Decline nvpmodel's interactive reboot prompt; we never reboot the
		// device automatically. A reboot-required change is reported back to
		// the caller.
		cmd.Stdin = strings.NewReader("no\n")
		output, cmdErr := cmd.CombinedOutput()
		// nvpmodel's exit code for a declined reboot prompt isn't documented
		// and isn't consistent to rely on, so check the output text
		// regardless of whether the command itself exited zero or non-zero.
		if isRebootRequiredOutput(string(output)) {
			pm.logger.Warnf("Power mode %d requires a reboot to take effect. Run 'sudo nvpmodel -m %d' on the device (confirm the reboot prompt), or reboot after running it, to apply it.", pm.config.PowerMode, pm.config.PowerMode)
			rebootRequired = true
		} else if cmdErr != nil {
			return false, fmt.Errorf("failed to set power mode: %v, output: %s", cmdErr, string(output))
		}
	}

	// Governor/frequency settings are independent of nvpmodel's power mode
	// and apply via the same generic cpufreq sysfs interface as the
	// Raspberry Pi, so continue applying them even when the power mode above
	// was already current (or requires a reboot to fully take effect).
	cfg := pm.config
	if cfg.Governor == "" && cfg.Frequency == 0 && cfg.Minimum == 0 && cfg.Maximum == 0 {
		return rebootRequired, nil
	}
	output, err := cpufrequtils.ApplyPolicy(cfg.Governor, cfg.Frequency, cfg.Minimum, cfg.Maximum)
	if err != nil {
		pm.logger.Errorf("Error configuring CPU: %s: %s", err, output)
		return rebootRequired, err
	}
	pm.logger.Infof("CPU configured: %s", output)
	return rebootRequired, nil
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

// String reports the power mode and CPU policy this manager applies. The sensor
// logs the PowerManager itself, and the default format would print this
// struct's two pointer fields as addresses. power_mode is always shown because
// 0 is a valid nvpmodel mode; the cpufreq settings are omitted when unset, to
// match ApplyPowerMode.
func (pm *jetsonPowerManager) String() string {
	settings := []string{fmt.Sprintf("power_mode=%d", pm.config.PowerMode)}
	if pm.config.Governor != "" {
		settings = append(settings, "governor="+pm.config.Governor)
	}
	if pm.config.Frequency != 0 {
		settings = append(settings, fmt.Sprintf("frequency=%d", pm.config.Frequency))
	}
	if pm.config.Minimum != 0 {
		settings = append(settings, fmt.Sprintf("minimum=%d", pm.config.Minimum))
	}
	if pm.config.Maximum != 0 {
		settings = append(settings, fmt.Sprintf("maximum=%d", pm.config.Maximum))
	}
	return "jetson(" + strings.Join(settings, ", ") + ")"
}
