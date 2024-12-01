package clocks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

var (
	ErrNoGpuClockFound = errors.New("no valid GPU clock path found")
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]clockSensor, error) {
	if sbcidentify.IsRaspberryPi() {
		return getRaspberryPiClockSensors(ctx, logger)
	} else if sbcidentify.IsNvidia() {
		return getNvidiaClockSensors(ctx, logger)
	}
	boardtype, err := sbcidentify.GetBoardType()
	if err != nil {
		logger.Warnf("Failed to get board type: %v", err)
	}
	logger.Warnf("No clock sensors found for %s", boardtype)
	return nil, nil
}

func getSysFsClock(ctx context.Context, path string) (current, min, max int64, err error) {
	current, err = readIntFromFile(ctx, filepath.Join(path, "cpufreq/cpuinfo_cur_freq"))
	if err != nil {
		return 0, 0, 0, err
	}
	min, err = readIntFromFile(ctx, filepath.Join(path, "cpufreq/cpuinfo_min_freq"))
	if err != nil {
		return 0, 0, 0, err
	}
	max, err = readIntFromFile(ctx, filepath.Join(path, "cpufreq/cpuinfo_max_freq"))
	if err != nil {
		return 0, 0, 0, err
	}
	return current, min, max, nil
}

func readIntFromFile(ctx context.Context, path string) (int64, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()
	file, err := utils.ReadFileWithContext(ctxWithTimeout, path)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(string(file)), 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func getSysFsCpuPaths() ([]string, error) {
	paths, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if err != nil {
		return nil, err
	}
	validPaths := make([]string, 0)
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		validPaths = append(validPaths, path)
	}

	return validPaths, nil
}
