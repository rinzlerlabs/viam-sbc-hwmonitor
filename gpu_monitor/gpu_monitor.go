package gpu_monitor

import (
	"context"
	"errors"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"go.viam.com/rdk/logging"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type GpuMonitor interface {
	GetGPUStats(ctx context.Context) (map[string]interface{}, error)
	Close()
}

func newGpuMonitor(logger logging.Logger) (GpuMonitor, error) {
	if sbcidentify.IsBoardType(nvidia.NVIDIA) {
		return newNvidiaGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
