package gpu_monitor

import (
	"context"
	"errors"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"go.viam.com/rdk/logging"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type GpuMonitor interface {
	GetGPUStats(ctx context.Context) ([]gpuDeviceStats, error)
	Close()
}

func newGpuMonitor(ctx context.Context, logger logging.Logger) (GpuMonitor, error) {
	if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return newNvidiaGpuMonitor(ctx, logger)
	}
	return nil, ErrUnsupportedBoard
}

type gpuDeviceStats struct {
	Name             string
	CurrentFrequency int64
	MaxFrequency     int64
	MinFrequency     int64
	Governor         string
	Load             float64
}

func gpuDeviceStatsToMap(stats []gpuDeviceStats) map[string]interface{} {
	result := make(map[string]interface{})
	for _, stat := range stats {
		result[stat.Name+"-current_frequency"] = stat.CurrentFrequency
		result[stat.Name+"-max_frequency"] = stat.MaxFrequency
		result[stat.Name+"-min_frequency"] = stat.MinFrequency
		result[stat.Name+"-governor"] = stat.Governor
		result[stat.Name+"-load"] = stat.Load
	}
	return result
}
