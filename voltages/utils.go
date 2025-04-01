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
	} else if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return jetson.GetPowerSensors(ctx, logger)
	}
	return make([]sensors.PowerSensor, 0), nil
}
