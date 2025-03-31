package clocks

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"go.viam.com/rdk/logging"
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]sensors.ClockSensor, error) {
	return genericwindows.GetClockSensors(ctx, logger)
}
