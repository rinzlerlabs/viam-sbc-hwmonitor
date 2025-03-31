package pwmfan

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/genericwindows"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

func GetTemperatureFunc() (func(ctx context.Context) (*sensors.SystemTemperatures, error), error) {
	return genericwindows.GetTemperatures, nil
}
