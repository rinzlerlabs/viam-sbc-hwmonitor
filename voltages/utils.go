package voltages

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/raspberrypi"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"go.viam.com/rdk/logging"
)

func getPowerSensors(ctx context.Context, logger logging.Logger) ([]sensors.PowerSensor, error) {
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return raspberrypi.GetPowerSensors(ctx, logger)
	}
	// All other Linux boards: discover INA3221 power monitors via sysfs (Jetson
	// and similar). Returns an empty set if no supported monitor is present.
	// Not gated on board identification, which can fail on some Jetsons (Orin).
	return jetson.GetPowerSensors(ctx, logger)
}
