package voltages

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/raspberrypi"
)

// Jetson Orin Nano Jetpack 6 voltages
// cat /sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in*_label
// cat /sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in*_input

func getVoltages(ctx context.Context) (map[string]interface{}, error) {
	if sbcidentify.IsBoardType(raspberrypi.RaspberryPi) {
		return getRaspberryPiVoltages(ctx)
	}
	return make(map[string]interface{}), nil
}

func getRaspberryPiVoltages(ctx context.Context) (map[string]interface{}, error) {
	components := []string{"core", "sdram_c", "sdram_i", "sdram_p"}
	voltages := make(map[string]interface{})
	for _, component := range components {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			voltage, err := getRaspberryPiComponentVoltage(component)
			if err != nil {
				return nil, err
			}
			voltages[component] = voltage
		}
	}
	return voltages, nil
}

func getRaspberryPiComponentVoltage(component string) (Voltage float64, Err error) {
	proc := exec.Command("vcgencmd", "measure_volts", component)
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	output := string(outputBytes)
	return parseVcgencmdVoltage(output)
}

func parseVcgencmdVoltage(output string) (Voltage float64, Err error) {
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return 0, errors.New("unexpected output from vcgencmd")
	}
	return strconv.ParseFloat(strings.TrimSpace(strings.Replace(parts[1], "V", "", 1)), 64)
}
