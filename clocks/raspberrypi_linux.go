//go:build linux
// +build linux

package clocks

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"go.viam.com/rdk/logging"
)

var (
	raspi4Clocks = []string{
		"arm",
		"core",
		"h264",
		"isp",
		"v3d",
		"uart",
		"pwm",
		"emmc",
		"pixel",
		"vec",
		"hdmi",
		"dpi",
	}

	raspi5Clocks = []string{
		"arm",
		"core",
		"emmc",
		"hdmi",
		"isp",
		"uart",
		"v3d",
	}
)

type raspberryPiClockSensor struct {
	logger     logging.Logger
	mu         sync.RWMutex
	wg         sync.WaitGroup
	name       string
	sensorType string
	path       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	updateTask func()
	frequency  int64
}

func (s *raspberryPiClockSensor) StartUpdating() error {
	updateInterval := 1 * time.Second
	s.updateTask = func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for {
			select {
			case <-s.cancelCtx.Done():
				return
			case <-time.After(updateInterval):
				var frequency int64
				var err error
				switch s.sensorType {
				case "vcgencmd":
					frequency, err = s.readVcgencmdClock()
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
			}
		}
	}
	go s.updateTask()
	return nil
}

func (s *raspberryPiClockSensor) readVcgencmdClock() (int64, error) {
	cmd := exec.CommandContext(s.cancelCtx, "vcgencmd", "measure_clock", s.name)
	output, err := cmd.Output()
	if err != nil {
		s.logger.Errorw("failed to measure clock", "sensor", s.name, "error", err)
		return 0, err
	}
	outputStr := string(output)
	parts := strings.Split(outputStr, "=")
	if len(parts) != 2 {
		s.logger.Errorw("unexpected output format", "sensor", s.name, "output", outputStr)
		return 0, err
	}
	frequency, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		s.logger.Errorw("failed to parse frequency", "sensor", s.name, "output", outputStr, "error", err)
		return 0, err
	}
	s.logger.Debugf("measured clock frequency", "sensor", s.name, "frequency", frequency)
	return frequency, nil
}

func (s *raspberryPiClockSensor) readSysfsClock() (int64, error) {
	current, err := getSysFsClock(s.cancelCtx, s.path)
	if err != nil {
		s.logger.Errorw("failed to read sysfs clock", "sensor", s.name, "error", err)
		return 0, err
	}
	s.logger.Debugw("measured clock frequency", "sensor", s.name, "current", current)
	return current, nil
}

func (s *raspberryPiClockSensor) Close() {
	s.cancelFunc()
	s.wg.Wait()
}

func (s *raspberryPiClockSensor) GetReadingMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		s.name: s.frequency,
	}
}

func (s *raspberryPiClockSensor) Name() string {
	return s.name
}

func getRaspberryPi4ClockSensors(ctx context.Context, logger logging.Logger) []clockSensor {
	sensors := make([]clockSensor, 0)
	for _, name := range raspi4Clocks {
		sensor := newRaspberryPiVcgencmdSensor(ctx, logger, name)
		sensors = append(sensors, sensor)
	}
	return sensors
}

func getRaspberryPi5ClockSensors(ctx context.Context, logger logging.Logger) []clockSensor {
	sensors := make([]clockSensor, 0)
	for _, name := range raspi5Clocks {
		sensor := newRaspberryPiVcgencmdSensor(ctx, logger, name)
		sensors = append(sensors, sensor)
	}
	return sensors
}

func newRaspberryPiVcgencmdSensor(ctx context.Context, logger logging.Logger, name string) *raspberryPiClockSensor {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &raspberryPiClockSensor{
		logger:     logger,
		name:       name,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		sensorType: "vcgencmd",
	}
}

func newRaspberryPiSysFsSensor(ctx context.Context, logger logging.Logger, path string) *raspberryPiClockSensor {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	parts := strings.Split(path, "/")
	sensorName := parts[len(parts)-1]
	return &raspberryPiClockSensor{
		logger:     logger,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		name:       sensorName,
		sensorType: "sysfs",
		path:       filepath.Join(path, "cpufreq/cpuinfo_cur_freq"),
	}
}

func getRaspberryPiClockSensors(ctx context.Context, logger logging.Logger) ([]clockSensor, error) {
	sensors := make([]clockSensor, 0)
	if sbcidentify.IsBoardType(boardtype.RaspberryPi5) {
		sensors = append(sensors, getRaspberryPi5ClockSensors(ctx, logger)...)
	} else if sbcidentify.IsBoardType(boardtype.RaspberryPi4) {
		sensors = append(sensors, getRaspberryPi4ClockSensors(ctx, logger)...)
	} else {
		b, e := sbcidentify.GetBoardType()
		logger.Warnf("No vcgencmd clock sensors found for %s %s", b, e)
	}
	sysFsCpus, err := getSysFsCpuPaths()
	if err != nil {
		return nil, err
	}
	for _, cpu := range sysFsCpus {
		sensor := newRaspberryPiSysFsSensor(ctx, logger, cpu)
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}
