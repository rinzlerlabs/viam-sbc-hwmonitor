package gpu_monitor

import (
	"context"
	"errors"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"go.viam.com/rdk/logging"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
	ErrInitializingNvml = errors.New("failed to initialize NVML")
	ErrNvmlFailure      = errors.New("NVML failure")
)

type GpuMonitor interface {
	GetGPUStats(ctx context.Context) (map[string]interface{}, error)
	Close()
}

type nvidiaGpuMonitor struct {
	logger logging.Logger
}

func newNvidiaGpuMonitor(logger logging.Logger) (GpuMonitor, error) {
	if res := nvml.Init(); res != nvml.SUCCESS {
		return nil, ErrInitializingNvml
	}
	return &nvidiaGpuMonitor{logger: logger}, nil
}

func (m *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) (map[string]interface{}, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		m.logger.Info("Failed to get device count %v", ret)
		return nil, ErrNvmlFailure
	}
	resp := make(map[string]interface{})
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			m.logger.Info("Failed to get device handle %v", ret)
			return nil, ErrNvmlFailure
		}
		name, _ := nvml.DeviceGetName(device)
		utilization, ret := nvml.DeviceGetUtilizationRates(device)
		if ret != nvml.SUCCESS {
			m.logger.Info("Failed to get device utilization %v", ret)
			return nil, ErrNvmlFailure
		}
		resp[name+"_gpu"] = float64(utilization.Gpu) / 10
		resp[name+"_memory"] = float64(utilization.Memory) / 10
	}
	return resp, nil
}

func (m *nvidiaGpuMonitor) Close() {
	nvml.Shutdown()
}

func newGpuMonitor(logger logging.Logger) (GpuMonitor, error) {
	if sbcidentify.IsBoardType(nvidia.NVIDIA) {
		return newNvidiaGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
