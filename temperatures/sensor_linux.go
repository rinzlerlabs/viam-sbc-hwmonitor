package temperatures

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/raspberrypi"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

func GetTemperatureFunc(ctx context.Context) (func(ctx context.Context) (*sensors.SystemTemperatures, error), error) {
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return raspberrypi.GetTemperatures, nil
	} else if jetson.IsJetson() {
		return jetson.GetTemperatureFunc(ctx)
	} else {
		return linux.GetTemperatureFunc(ctx)
	}
}
