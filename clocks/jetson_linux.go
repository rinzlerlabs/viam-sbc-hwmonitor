//go:build linux
// +build linux

package clocks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.viam.com/rdk/logging"
)

type jetsonClockSensor struct {
	logger     logging.Logger
	mu         sync.RWMutex
	wg         sync.WaitGroup
	name       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	updateTask func()
	frequency  int64
	sensorType string
	path       string
}

func (s *jetsonClockSensor) Close() {
	s.cancelFunc()
	s.wg.Wait()
}

func (s *jetsonClockSensor) Name() string {
	return s.name
}

func (s *jetsonClockSensor) GetReadingMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		s.name: s.frequency,
	}
}

func (s *jetsonClockSensor) StartUpdating() error {
	updateInterval := 1 * time.Second
	s.updateTask = func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for {
			select {
			case <-s.cancelCtx.Done():
				return
			case <-time.After(updateInterval):
				s.logger.Debug("Updating clock frequency")
				var frequency int64
				var err error
				switch s.sensorType {
				case "sysfs":
					frequency, err = s.readSysfsClock()
				}
				if err != nil {
					s.logger.Errorf("failed to read clock frequency: %v", err)
					continue
				}
				s.mu.Lock()
				s.frequency = frequency
				s.mu.Unlock()
				s.logger.Debugf("Updated clock frequency: %d", frequency)
			}
		}
	}
	go s.updateTask()
	return nil
}

func (s *jetsonClockSensor) readSysfsClock() (int64, error) {
	current, err := getSysFsClock(s.cancelCtx, s.path)
	if err != nil {
		s.logger.Errorw("failed to read sysfs clock", "sensor", s.name, "error", err)
		return 0, err
	}
	s.logger.Debugw("measured clock frequency", "sensor", s.name, "current", current)
	return current, nil
}

func newNvidiaJetsonCpuClockSensor(ctx context.Context, logger logging.Logger, path string) *jetsonClockSensor {
	logger.Debugf("Initializing NVIDIA Jetson CPU clock sensor: %v", path)
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	parts := strings.Split(path, "/")
	sensorName := parts[len(parts)-1]
	s := &jetsonClockSensor{
		logger:     logger.Sublogger(sensorName),
		name:       sensorName,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		path:       filepath.Join(path, "cpufreq/cpuinfo_cur_freq"),
		sensorType: "sysfs",
	}
	return s
}

func newNvidiaJetsonGpuClockSensor(ctx context.Context, logger logging.Logger) *jetsonClockSensor {
	logger.Debug("Initializing NVIDIA Jetson GPU clock sensor")
	paths := []string{
		"/sys/devices/platform/bus@0/17000000.gpu/devfreq/17000000.gpu/cur_freq",
		"/sys/devices/platform/17000000.ga10b/devfreq/17000000.ga10b/cur_freq",
	}
	var gpuPath string
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		gpuPath = path
		break
	}
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	name := "gpu0"
	return &jetsonClockSensor{
		logger:     logger.Sublogger(name),
		name:       name,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		path:       gpuPath,
		sensorType: "sysfs",
	}
}

func getNvidiaJetsonClockSensors(ctx context.Context, logger logging.Logger) ([]clockSensor, error) {
	sensors := make([]clockSensor, 0)
	sysFsCpus, err := getSysFsCpuPaths()
	if err != nil {
		return nil, err
	}
	for _, cpu := range sysFsCpus {
		sensor := newNvidiaJetsonCpuClockSensor(ctx, logger, cpu)
		sensors = append(sensors, sensor)
	}
	sensors = append(sensors, newNvidiaJetsonGpuClockSensor(ctx, logger))
	return sensors, nil
}
