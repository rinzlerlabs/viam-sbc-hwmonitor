package gpumonitor

import (
	"context"
	"errors"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"go.viam.com/rdk/logging"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type gpuMonitor interface {
	// Close closes the GPU monitor.
	Close() error
	// GetGPUStats returns a map of GPU sensor readings.
	// The key is the identifier for the GPU in the system.
	// The value is a slice of gpuSensorReading.
	// The slice contains the readings for each sensor on the GPU.
	// The readings are in the order they were found.
	// The readings are not guaranteed to be in any particular order.
	// The readings are guaranteed to be unique.
	GetGPUStats(context.Context) (map[string][]sensors.GPUSensorReading, error)
}

func newGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	// Prefer the Jetson sysfs monitor for Tegra integrated GPUs. Check for the
	// GPU sysfs node directly (not just board identification), since detection
	// can fail on some Jetsons (e.g. Orin) and nvidia-smi does not work on Tegra.
	if sbcidentify.IsBoardType(boardtype.NVIDIA) || jetson.HasJetsonGpu() {
		return jetson.NewJetsonGpuMonitor(logger)
	} else if sensors.HasNvidiaSmiCommand(logger) {
		return sensors.NewNVIDIAGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
