package temperatures

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/genericlinux"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/raspberrypi"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

func GetTemperatureFunc() (func(ctx context.Context) (*sensors.SystemTemperatures, error), error) {
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return raspberrypi.GetTemperatures, nil
	} else if sbcidentify.IsBoardType(boardtype.Jetson) {
		return jetson.GetTemperatures, nil
	} else {
		return genericlinux.GetTemperatures, nil
	}
}
