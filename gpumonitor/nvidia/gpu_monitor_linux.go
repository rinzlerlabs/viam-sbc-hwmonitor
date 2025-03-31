package nvidia

import (
	"os/exec"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func HasNvidiaSmiCommand(logger logging.Logger) bool {
	cmd := exec.Command("which", "nvidia-smi")
	stdOut, stdErr := cmd.CombinedOutput()
	logger.Debugf("which nvidia-smi command output: %s", stdOut)
	if stdErr != nil {
		logger.Debugf("nvidia-smi command not found: %v", stdErr)
		return false
	}
	return true
}

func NewNVIDIAGpuMonitor(logger logging.Logger, extraSensors ...string) (*nvidiaGpuMonitor, error) {
	sensorsToQuery := make(map[string]bool)
	for _, sensor := range nvidiaSmiDefaultSensors {
		sensorsToQuery[sensor] = true
	}
	for _, sensor := range extraSensors {
		sensorsToQuery[sensor] = true
	}

	monitor := &nvidiaGpuMonitor{
		logger:         logger,
		sensorsToQuery: utils.Keys(sensorsToQuery),
	}

	return monitor, nil
}
