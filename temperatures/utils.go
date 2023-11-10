package temperatures

import (
	"os/exec"
	"strconv"
	"strings"
)

func getPmicTemperature() (Temperature float64, Err error) {
	proc := exec.Command("vcgencmd", "measure_temp", "pmic")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return parseTemperature(string(outputBytes))
}

func getSoCTemperature() (Temperature float64, Err error) {
	proc := exec.Command("vcgencmd", "measure_temp")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return parseTemperature(string(outputBytes))
}

func parseTemperature(output string) (Temperature float64, Err error) {
	t := strings.Split(output, "=")
	t1 := strings.TrimSuffix(t[1], "'C\n")
	return strconv.ParseFloat(strings.TrimSpace(t1), 64)
}
