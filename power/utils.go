package power

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

func getVoltages() (Core, SDRAM_c, SDRAM_i, SDRAM_p float64, Err error) {
	core, err := getComponentVoltage("core")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	sdram_c, err := getComponentVoltage("sdram_c")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	sdram_i, err := getComponentVoltage("sdram_i")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	sdram_p, err := getComponentVoltage("sdram_p")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return core, sdram_c, sdram_i, sdram_p, nil
}

func getComponentVoltage(component string) (float64, error) {
	proc := exec.Command("vcgencmd", "measure_volts", component)
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	output := string(outputBytes)
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return 0, errors.New("unexpected output from vcgencmd")
	}
	voltage, err := strconv.ParseFloat(strings.Replace(parts[1], "V", "", 1), 64)
	if err != nil {
		return 0, err
	}
	return voltage, nil
}
