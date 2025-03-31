package clocks

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]ClockSensor, error) {
	return nil, utils.ErrPlatformNotSupported
}
