package temperature

import (
	"os/exec"
	"strconv"
	"strings"
)

func getPmicTemperature() (float64, error) {
	proc := exec.Command("vcgencmd", "measure_temp", "pmic")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSuffix(strings.Split(string(outputBytes), "=")[1], "'C\n"), 64)
}

func getSoCTemperature() (float64, error) {
	proc := exec.Command("vcgencmd", "measure_temp")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSuffix(strings.Split(string(outputBytes), "=")[1], "'C\n"), 64)
}
