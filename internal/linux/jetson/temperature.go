package jetson

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

var jetsonTemperatureSensors = []sensors.TemperatureReader{
	sensors.NewFileTemperatureSensor("CPU", "/sys/devices/virtual/thermal/thermal_zone0/temp"),
	sensors.NewFileTemperatureSensor("GPU", "/sys/devices/virtual/thermal/thermal_zone1/temp"),
	sensors.NewFileTemperatureSensor("CV0", "/sys/devices/virtual/thermal/thermal_zone2/temp"),
	sensors.NewFileTemperatureSensor("CV1", "/sys/devices/virtual/thermal/thermal_zone3/temp"),
	sensors.NewFileTemperatureSensor("CV2", "/sys/devices/virtual/thermal/thermal_zone4/temp"),
	sensors.NewFileTemperatureSensor("SOC0", "/sys/devices/virtual/thermal/thermal_zone5/temp"),
	sensors.NewFileTemperatureSensor("SOC1", "/sys/devices/virtual/thermal/thermal_zone6/temp"),
	sensors.NewFileTemperatureSensor("SOC2", "/sys/devices/virtual/thermal/thermal_zone7/temp"),
	sensors.NewFileTemperatureSensor("TJ", "/sys/devices/virtual/thermal/thermal_zone8/temp"),
}

func GetTemperatures(ctx context.Context) (*sensors.SystemTemperatures, error) {
	systemTemps := &sensors.SystemTemperatures{Extra: make(map[string]float64)}
	for _, sensor := range jetsonTemperatureSensors {
		temp, err := sensor.Read(ctx)
		if err != nil {
			continue
		}
		temp = float64(int((temp/1000)*100)) / 100
		switch sensor.Name() {
		case "CPU":
			systemTemps.CPU = &temp
		case "GPU":
			systemTemps.GPU = &temp
		default:
			systemTemps.Extra[sensor.Name()] = temp
		}
	}

	return systemTemps, nil
}
