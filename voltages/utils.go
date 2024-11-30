package voltages

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"go.viam.com/rdk/logging"
)

// Jetson Orin Nano Jetpack 6 voltages
// cat /sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in*_label
// cat /sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in*_input

func getPowerSensors(ctx context.Context, logger logging.Logger) ([]powerSensor, error) {
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return getRaspberryPiPowerSensors(ctx, logger)
	} else if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return getJetsonPowerSensors(ctx, logger)
	}
	return make([]powerSensor, 0), nil
}
