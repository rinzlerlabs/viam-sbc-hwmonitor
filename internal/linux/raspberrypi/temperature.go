package raspberrypi

import (
	"context"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
)

var raspberryPiTemperatureSensors = []sensors.TemperatureReader{
	NewVcgencmdSensor("CPU", ""),
	NewVcgencmdSensor("PMIC", "pmic"),
}

func GetTemperatures(ctx context.Context) (*sensors.SystemTemperatures, error) {
	systemTemps := &sensors.SystemTemperatures{Extra: make(map[string]float64)}
	for _, sensor := range raspberryPiTemperatureSensors {
		temp, err := sensor.Read(ctx)
		if err != nil {
			continue
		}
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

func NewVcgencmdSensor(name, subcommand string) sensors.TemperatureReader {
	return &VcgencmdSensor{name: name, subcommand: subcommand}
}

type VcgencmdSensor struct {
	name       string
	subcommand string
}

func (t *VcgencmdSensor) Read(ctx context.Context) (float64, error) {
	proc := exec.CommandContext(ctx, "vcgencmd", "measure_temp", t.subcommand)
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return parseTemperature(string(outputBytes))
}

func (t *VcgencmdSensor) Name() string {
	return t.name
}

func parseTemperature(output string) (Temperature float64, Err error) {
	t := strings.Split(output, "=")
	t1 := strings.TrimSuffix(t[1], "'C\n")
	return strconv.ParseFloat(strings.TrimSpace(t1), 64)
}
