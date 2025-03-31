package jetson

import (
	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func NewJetsonGpuMonitor(logger logging.Logger) (*jetsonGpuMonitor, error) {
	return nil, utils.ErrPlatformNotSupported
}
