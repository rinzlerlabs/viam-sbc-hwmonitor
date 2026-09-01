package linux

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

// GetTemperatureFunc discovers this board's kernel thermal zones once and
// returns a closure that only re-reads their current temperatures on each
// call, instead of re-globbing sysfs and re-reading every zone's static
// "type" file on every call — this matters for callers that poll on a fixed
// interval (e.g. pwm_fan).
func GetTemperatureFunc(ctx context.Context) (func(context.Context) (*sensors.SystemTemperatures, error), error) {
	reader, err := sensors.NewThermalZoneReader(ctx)
	if err != nil {
		return nil, err
	}
	return reader.Read, nil
}
