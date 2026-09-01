package temperatures

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/windows"
)

func GetTemperatureFunc(ctx context.Context) (func(ctx context.Context) (*sensors.SystemTemperatures, error), error) {
	return windows.GetTemperatures, nil
}
