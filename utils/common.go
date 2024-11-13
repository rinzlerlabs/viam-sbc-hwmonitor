package utils

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type BoardType string

var (
	Unknown      BoardType = "Unknown"
	RaspberryPi4 BoardType = "Raspberry Pi 4"
	RaspberryPi5 BoardType = "Raspberry Pi 5"
)

func GetPmicTemperature() (Temperature float64, Err error) {
	proc := exec.Command("vcgencmd", "measure_temp", "pmic")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return parseTemperature(string(outputBytes))
}

func GetSoCTemperature() (Temperature float64, Err error) {
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

func GetBoardType() (BoardType, error) {
	c, err := os.ReadFile("/sys/firmware/devicetree/base/model")
	if err != nil {
		return Unknown, err
	}
	str := string(c)
	if strings.HasPrefix(str, string(RaspberryPi4)) {
		return RaspberryPi4, nil
	}

	if strings.HasPrefix(str, string(RaspberryPi5)) {
		return RaspberryPi5, nil
	}

	return Unknown, nil
}
