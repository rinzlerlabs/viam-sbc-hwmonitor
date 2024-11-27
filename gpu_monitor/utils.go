package gpu_monitor

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

func getGPUStats(ctx context.Context) (map[string]interface{}, error) {
	if sbcidentify.IsBoardType(nvidia.NVIDIA) {
		return getJetsonGPUStats(ctx)
	}
	return nil, ErrUnsupportedBoard
}

func getJetsonGPUStats(ctx context.Context) (map[string]interface{}, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	paths := []string{
		"/sys/devices/platform/bus@0/17000000.gpu/load",
		"/sys/devices/platform/17000000.ga10b/load",
	}
	resp := make(map[string]interface{})
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		loadStr, err := utils.ReadFileWithContext(ctxWithTimeout, path)
		if err != nil {
			return resp, err
		}
		load, err := strconv.Atoi(strings.TrimSpace(string(loadStr)))
		if err != nil {
			return resp, err
		}
		resp["gpu"] = utils.RoundValue(float64(load)/10, 1)
		return resp, nil
	}
	return resp, ErrUnsupportedBoard
}
