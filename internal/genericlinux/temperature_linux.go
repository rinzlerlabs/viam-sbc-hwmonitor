package genericlinux

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

func GetTemperatures(ctx context.Context) (*sensors.SystemTemperatures, error) {
	systemTemps := &sensors.SystemTemperatures{Extra: make(map[string]float64)}
	return systemTemps, nil
}
