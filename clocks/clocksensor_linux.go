package clocks

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/raspberrypi"
	"go.viam.com/rdk/logging"
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]ClockSensor, error) {
	if sbcidentify.IsRaspberryPi() {
		sensors, err := raspberrypi.GetClockSensors(ctx, logger)
		if err != nil {
			return nil, err
		}
		ret := make([]ClockSensor, 0, len(sensors))
		for _, sensor := range sensors {
			ret = append(ret, sensor)
		}
		return ret, nil
	} else if sbcidentify.IsNvidia() {
		sensors, err := jetson.GetClockSensors(ctx, logger)
		if err != nil {
			return nil, err
		}
		ret := make([]ClockSensor, 0, len(sensors))
		for _, sensor := range sensors {
			ret = append(ret, sensor)
		}
		return ret, nil
	}
	boardtype, err := sbcidentify.GetBoardType()
	if err != nil {
		logger.Warnf("Failed to get board type: %v", err)
	}
	logger.Warnf("No clock sensors found for %s", boardtype)
	return nil, nil
}
