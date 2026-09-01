package gpumonitor

import (
	"context"
	"fmt"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

// ErrUnsupportedBoard wraps the shared utils.ErrBoardNotSupported sentinel so
// callers can match either this GPU-specific message or any other unsupported-
// board error (e.g. the Jetson GPU monitor's own "no GPU sensors found") via a
// single errors.Is(err, utils.ErrBoardNotSupported) check.
var ErrUnsupportedBoard = fmt.Errorf("gpu stats not supported on this board: %w", utils.ErrBoardNotSupported)

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
	// Prefer the Jetson sysfs monitor for Tegra integrated GPUs; nvidia-smi
	// does not work on Tegra.
	if jetson.IsJetson() {
		return jetson.NewJetsonGpuMonitor(logger)
	} else if sensors.HasNvidiaSmiCommand(logger) {
		return sensors.NewNVIDIAGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
