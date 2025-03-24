package clocks

import (
	"context"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"go.viam.com/rdk/logging"
)

func newNVidiaGPUClockSensor(logger logging.Logger) clockSensor {
	// Initialize NVML
	ret := nvml.Init()
	// Check if we failed to initialize and not because it's already initialized
	// (which can happen if clocks and gpu_monitor components are both loaded)
	if ret != nvml.SUCCESS && ret != nvml.ERROR_ALREADY_INITIALIZED {
		logger.Errorf("Failed to initialize NVML: %v", ret)
		return nil
	}
	return &nvidiaClockSensor{
		logger: logger,
	}
}

type nvidiaClockSensor struct {
	logger     logging.Logger
	mu         sync.RWMutex
	wg         sync.WaitGroup
	name       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	updateTask func()
	frequency  int64
	sensorType string
}

func (s *nvidiaClockSensor) Close() {
	s.cancelFunc()
	s.wg.Wait()
	if err := nvml.Shutdown(); err != nvml.SUCCESS {
		s.logger.Errorf("Failed to shutdown NVML: %v", err)
	}
}

func (s *nvidiaClockSensor) Name() string {
	return s.name
}

func (s *nvidiaClockSensor) GetReadingMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		s.name: s.frequency,
	}
}

func (s *nvidiaClockSensor) StartUpdating() error {
	updateInterval := 1 * time.Second
	s.updateTask = func() {
		s.wg.Add(1)
		defer s.wg.Done()
		deviceCount, ret := nvml.DeviceGetCount()
		if ret != nvml.SUCCESS {
			s.logger.Errorf("Failed to get device count: %v", ret)
			s.logger.Errorf("Exiting update task")
			return
		}
		if deviceCount == 0 {
			s.logger.Errorf("No NVIDIA devices found")
			s.logger.Errorf("Exiting update task")
			return
		}
		s.logger.Infof("Found %d NVIDIA devices", deviceCount)
		for {
			select {
			case <-s.cancelCtx.Done():
				return
			case <-time.After(updateInterval):
				s.logger.Debug("Updating clock frequency")
				for i := 0; i < deviceCount; i++ {
					device, ret := nvml.DeviceGetHandleByIndex(i)
					if ret != nvml.SUCCESS {
						s.logger.Errorf("Failed to get device handle: %v", ret)
						continue
					}
					frequency, ret := device.GetClockInfo(nvml.CLOCK_GRAPHICS)
					if ret != nvml.SUCCESS {
						s.logger.Errorf("Failed to get clock info: %v", ret)
						continue
					}
					s.mu.Lock()
					s.frequency = int64(frequency)
					s.mu.Unlock()
					s.logger.Infof("Device %d clock frequency: %d MHz", i, frequency)
				}
			}
		}
	}
	go s.updateTask()
	return nil
}
