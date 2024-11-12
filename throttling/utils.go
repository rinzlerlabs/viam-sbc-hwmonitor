package throttling

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	Undervolt               = "undervolt"
	ArmFrequencyCapped      = "armFrequencyCapped"
	CurrentlyThrottled      = "currentlyThrottled"
	SoftTempLimitActive     = "softTempLimitActive"
	UnderVoltOccurred       = "undervoltOccurred"
	ArmFrequencyCapOccurred = "armFrequencyCapOccurred"
	ThrottlingOccurred      = "throttlingOccurred"
	SoftTempLimitOccurred   = "softTempLimitOccurred"
)

func getThrottlingStates() (map[string]bool, error) {
	proc := exec.Command("vcgencmd", "get_throttled")
	outputBytes, err := proc.Output()
	if err != nil {
		return nil, err
	}
	output := string(outputBytes)
	return parseThrottlingStates(output)
}

func parseThrottlingStates(output string) (map[string]bool, error) {
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected output from vcgencmd %s", output)
	}
	hex := strings.TrimSpace(strings.Replace(parts[1], "0x", "", 1))
	throttlingStates, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return nil, err
	}
	return map[string]bool{
		Undervolt:               throttlingStates&0x1 != 0,
		ArmFrequencyCapped:      throttlingStates&0x2 != 0,
		CurrentlyThrottled:      throttlingStates&0x04 != 0,
		SoftTempLimitActive:     throttlingStates&0x8 != 0,
		UnderVoltOccurred:       throttlingStates&0x10000 != 0,
		ArmFrequencyCapOccurred: throttlingStates&0x20000 != 0,
		ThrottlingOccurred:      throttlingStates&0x40000 != 0,
		SoftTempLimitOccurred:   throttlingStates&0x80000 != 0,
	}, nil
}
