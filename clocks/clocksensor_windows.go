package clocks

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/windows"
	"go.viam.com/rdk/logging"
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]sensors.ClockSensor, error) {
	return windows.GetClockSensors(ctx, logger)
}
