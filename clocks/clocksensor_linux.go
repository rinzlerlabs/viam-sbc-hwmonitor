package clocks

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/raspberrypi"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"go.viam.com/rdk/logging"
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]sensors.ClockSensor, error) {
	if sbcidentify.IsRaspberryPi() {
		return raspberrypi.GetClockSensors(ctx, logger)
	} else if sbcidentify.IsNvidia() {
		return jetson.GetClockSensors(ctx, logger)
	}
	boardtype, err := sbcidentify.GetBoardType()
	if err != nil {
		logger.Warnf("Failed to get board type: %v", err)
	}
	logger.Debugf("No SBC clock sensors found for %s, assuming genericlinux", boardtype)

	return linux.GetClockSensors(ctx, logger)
}
