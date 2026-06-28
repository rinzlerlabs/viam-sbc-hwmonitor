package linux

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

func GetTemperatures(ctx context.Context) (*sensors.SystemTemperatures, error) {
	return sensors.ReadSysfsThermalZones(ctx)
}
