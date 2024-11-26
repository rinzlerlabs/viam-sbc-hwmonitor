package throttling

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"github.com/rinzlerlabs/sbcidentify/raspberrypi"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
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

func getThrottlingStates(ctx context.Context) (map[string]interface{}, error) {
	if sbcidentify.IsBoardType(raspberrypi.RaspberryPi) {
		return getRasPiThrottlingStates(ctx)
	} else if sbcidentify.IsBoardType(nvidia.NVIDIA) {
		return getJetsonThrottlingStates(ctx)
	}
	return nil, fmt.Errorf("board not supported")
}

func getRasPiThrottlingStates(ctx context.Context) (map[string]interface{}, error) {
	proc := exec.CommandContext(ctx, "vcgencmd", "get_throttled")
	outputBytes, err := proc.Output()
	if err != nil {
		return nil, err
	}
	output := string(outputBytes)
	return parseRasPiThrottlingStates(output)
}

func parseRasPiThrottlingStates(output string) (map[string]interface{}, error) {
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected output from vcgencmd %s", output)
	}
	hex := strings.TrimSpace(strings.Replace(parts[1], "0x", "", 1))
	throttlingStates, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
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

func getJetsonThrottlingStates(ctx context.Context) (map[string]interface{}, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	dirs, err := filepath.Glob("/sys/class/thermal/cooling_device*")
	if err != nil {
		return nil, err
	}

	throttlingStates := make(map[string]interface{})
	for _, dir := range dirs {
		deviceType, err := utils.ReadFileWithContext(ctxWithTimeout, filepath.Join(dir, "type"))
		if err != nil {
			return nil, err
		}

		curState, err := utils.ReadFileWithContext(ctxWithTimeout, filepath.Join(dir, "cur_state"))
		if err != nil {
			return nil, err
		}

		maxState, err := utils.ReadFileWithContext(ctxWithTimeout, filepath.Join(dir, "max_state"))
		if err != nil {
			return nil, err
		}

		deviceTypeStr := strings.TrimSpace(string(deviceType))
		curStateInt, err := strconv.Atoi(strings.TrimSpace(string(curState)))
		if err != nil {
			return nil, err
		}

		maxStateInt, err := strconv.Atoi(strings.TrimSpace(string(maxState)))
		if err != nil {
			return nil, err
		}

		throttlingStates[deviceTypeStr] = curStateInt > 0 && curStateInt == maxStateInt
	}

	return throttlingStates, nil
}
