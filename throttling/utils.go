package throttling

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

func getThrottlingStates() (Undervolt, ArmFrequencyCapped, CurrentlyThrottled, SoftTempLimitActive, UnderVoltOccurred, ArmFrequencyCapOccurred, ThrottlingOccurred, SoftTempLimitOccurred bool, Err error) {
	proc := exec.Command("vcgencmd", "get_throttled")
	outputBytes, err := proc.Output()
	if err != nil {
		return false, false, false, false, false, false, false, false, err
	}
	output := string(outputBytes)
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return false, false, false, false, false, false, false, false, errors.New("unexpected output from vcgencmd")
	}
	hex := strings.Replace(parts[1], "0x", "", 1)
	throttlingStates, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return false, false, false, false, false, false, false, false, err
	}
	return throttlingStates&0x1 == 1, throttlingStates&0x2 == 1, throttlingStates&0x4 == 1, throttlingStates&0x8 == 1, throttlingStates&0x10000 == 1, throttlingStates&0x20000 == 1, throttlingStates&0x40000 == 1, throttlingStates&0x80000 == 1, nil
}
