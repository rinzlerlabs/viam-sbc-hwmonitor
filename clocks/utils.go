package clocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	ErrNoGpuClockFound = errors.New("no valid GPU clock path found")
)

func getSystemClocks(ctx context.Context) (map[string]interface{}, error) {
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return getRasPiSystemClocks(ctx)
	} else if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return getJetsonSystemClocks(ctx)
	}
	return nil, fmt.Errorf("board not supported")
}

func getRasPiSystemClocks(ctx context.Context) (map[string]interface{}, error) {
	cpuClocks, err := getSysDevicesSystemCPUClocks(ctx)
	if err != nil {
		return nil, err
	}

	return cpuClocks, nil
}

func getJetsonSystemClocks(ctx context.Context) (map[string]interface{}, error) {
	cpuClocks, err := getSysDevicesSystemCPUClocks(ctx)
	if err != nil {
		return nil, err
	}

	gpuClock, err := getJetsonGPUClocks(ctx)
	if err != nil {
		return nil, err
	}

	cpuClocks["gpu"] = gpuClock
	return cpuClocks, nil
}

func getSysDevicesSystemCPUClocks(ctx context.Context) (map[string]interface{}, error) {
	dirs, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if err != nil {
		return nil, err
	}

	clockFrequencies := make(map[string]interface{})
	for _, dir := range dirs {
		cpu := filepath.Base(dir)
		clockFrequencies[cpu], err = readIntFromFile(ctx, filepath.Join(dir, "cpufreq/cpuinfo_cur_freq"))
		if err != nil {
			return nil, err
		}
		clockFrequencies[cpu+"_min"], err = readIntFromFile(ctx, filepath.Join(dir, "cpufreq/cpuinfo_min_freq"))
		if err != nil {
			return nil, err
		}
		clockFrequencies[cpu+"_max"], err = readIntFromFile(ctx, filepath.Join(dir, "cpufreq/cpuinfo_max_freq"))
		if err != nil {
			return nil, err
		}
	}

	return clockFrequencies, nil
}

func readIntFromFile(ctx context.Context, path string) (int, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()
	file, err := utils.ReadFileWithContext(ctxWithTimeout, path)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(strings.TrimSpace(string(file)))
	if err != nil {
		return 0, err
	}
	return i, nil
}

func getJetsonGPUClocks(ctx context.Context) (int, error) {
	paths := []string{
		"/sys/devices/platform/bus@0/17000000.gpu/devfreq/17000000.gpu/cur_freq",
		"/sys/devices/platform/17000000.ga10b/devfreq/17000000.ga10b/cur_freq",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		clock, err := utils.ReadFileWithContext(ctx, path)
		if err != nil {
			return 0, err
		}
		clockInt, err := strconv.Atoi(strings.TrimSpace(string(clock)))
		if err != nil {
			return 0, err
		}
		return clockInt, nil
	}

	return 0, ErrNoGpuClockFound
}
