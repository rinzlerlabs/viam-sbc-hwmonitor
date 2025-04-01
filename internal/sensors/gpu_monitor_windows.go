package sensors

import (
	"os/exec"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func HasNvidiaSmiCommand(logger logging.Logger) bool {
	cmd := exec.Command("where", "nvidia-smi")
	stdOut, stdErr := cmd.CombinedOutput()
	logger.Debugf("where nvidia-smi command output: %s", stdOut)
	if stdErr != nil {
		logger.Debugf("nvidia-smi command not found: %v", stdErr)
		return false
	}
	return true
}

func NewNVIDIAGpuMonitor(logger logging.Logger, extraSensors ...string) (*nvidiaGpuMonitor, error) {
	return nil, utils.ErrPlatformNotSupported
}
