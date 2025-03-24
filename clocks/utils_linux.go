//go:build linux
// +build linux

package clocks

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/rinzlerlabs/sbcidentify"
	"go.viam.com/rdk/logging"
)

var (
	ErrNoGpuClockFound = errors.New("no valid GPU clock path found")
)

func getClockSensors(ctx context.Context, logger logging.Logger) ([]clockSensor, error) {
	if sbcidentify.IsRaspberryPi() {
		return getRaspberryPiClockSensors(ctx, logger)
	} else if sbcidentify.IsNvidia() {
		return getNvidiaJetsonClockSensors(ctx, logger)
	}
	boardtype, err := sbcidentify.GetBoardType()
	if err != nil {
		logger.Warnf("Failed to get board type: %v", err)
	}
	logger.Warnf("No clock sensors found for %s", boardtype)
	return nil, nil
}

func getSysFsClock(ctx context.Context, path string) (int64, error) {
	return readIntFromFile(ctx, path)
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
