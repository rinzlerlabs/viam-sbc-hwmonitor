package temperatures

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/windows"
)

func GetTemperatureFunc() (func(ctx context.Context) (*sensors.SystemTemperatures, error), error) {
	return windows.GetTemperatures, nil
}
