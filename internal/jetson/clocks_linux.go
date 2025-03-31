package jetson

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"go.viam.com/rdk/logging"
)

type jetsonClockSensor struct {
	logger     logging.Logger
	mu         sync.RWMutex
	name       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	sensorType string
	path       string
}

func (s *jetsonClockSensor) Close() error {
	s.cancelFunc()
	return nil
}

func (s *jetsonClockSensor) Name() string {
	return s.name
}

func (s *jetsonClockSensor) GetReadingMap() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var frequency int64
	var err error
	switch s.sensorType {
	case "sysfs":
		frequency, err = s.readSysfsClock()
	default:
		return nil, errors.New("unknown sensor type")
	}
	return map[string]interface{}{
		s.name: frequency,
	}, err
}

func (s *jetsonClockSensor) readSysfsClock() (int64, error) {
	current, err := sensors.GetSysFsClock(s.cancelCtx, s.path)
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

func GetClockSensors(ctx context.Context, logger logging.Logger) ([]*jetsonClockSensor, error) {
	s := make([]*jetsonClockSensor, 0)
	sysFsCpus, err := sensors.GetSysFsCpuPaths()
	if err != nil {
		return nil, err
	}
	for _, cpu := range sysFsCpus {
		sensor := newNvidiaJetsonCpuClockSensor(ctx, logger, cpu)
		s = append(s, sensor)
	}
	s = append(s, newNvidiaJetsonGpuClockSensor(ctx, logger))
	return s, nil
}
