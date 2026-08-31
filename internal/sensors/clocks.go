package sensors

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

type ClockSensor interface {
	Close() error
	GetReadingMap() (map[string]interface{}, error)
	Name() string
}

// gpuClockMarkers are substrings of the devfreq device node names used by the
// GPU across SBC generations (ex: Jetson Nano "gpu", Orin "gpu"/"ga10b").
var gpuClockMarkers = []string{"gpu", "ga10b", "gv11b", "gp10b", "gm20b", "gk20a"}

// GetSysFsClockSensors builds clock sensors from generic sysfs interfaces: one
// per CPU (from cpufreq) and one for the GPU (from devfreq). It works on any
// Linux SBC exposing these paths, independent of board identification.
func GetSysFsClockSensors(ctx context.Context, logger logging.Logger) ([]ClockSensor, error) {
	clockSensors := make([]ClockSensor, 0)

	cpus, err := GetSysFsCpuPaths()
	if err != nil {
		return nil, err
	}
	for _, cpu := range cpus {
		clockSensors = append(clockSensors, newFileClockSensor(ctx, logger, filepath.Base(cpu), cpuClockPath(cpu)))
	}

	if gpuPath := findGpuClockPath(); gpuPath != "" {
		clockSensors = append(clockSensors, newFileClockSensor(ctx, logger, "gpu0", gpuPath))
	} else {
		logger.Debug("no GPU devfreq clock path found; skipping GPU clock sensor")
	}

	return clockSensors, nil
}

// cpuClockPath prefers cpuinfo_cur_freq (the actual hardware frequency) but
// falls back to scaling_cur_freq, which is more widely readable.
func cpuClockPath(cpu string) string {
	cpuinfo := filepath.Join(cpu, "cpufreq", "cpuinfo_cur_freq")
	if _, err := os.Stat(cpuinfo); err == nil {
		return cpuinfo
	}
	return filepath.Join(cpu, "cpufreq", "scaling_cur_freq")
}

// findGpuClockPath locates the GPU devfreq cur_freq file via the generic
// /sys/class/devfreq interface. Returns "" if no GPU devfreq device is found.
func findGpuClockPath() string {
	matches, err := filepath.Glob("/sys/class/devfreq/*/cur_freq")
	if err != nil {
		return ""
	}
	for _, match := range matches {
		lower := strings.ToLower(match)
		for _, marker := range gpuClockMarkers {
			if strings.Contains(lower, marker) {
				return match
			}
		}
	}
	return ""
}

type fileClockSensor struct {
	logger     logging.Logger
	name       string
	path       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
}

func newFileClockSensor(ctx context.Context, logger logging.Logger, name, path string) *fileClockSensor {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return &fileClockSensor{
		logger:     logger.Sublogger(name),
		name:       name,
		path:       path,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
}

func (s *fileClockSensor) Close() error {
	s.cancelFunc()
	return nil
}

func (s *fileClockSensor) Name() string {
	return s.name
}

func (s *fileClockSensor) GetReadingMap() (map[string]interface{}, error) {
	frequency, err := GetSysFsClock(s.cancelCtx, s.path)
	if err != nil {
		s.logger.Errorw("failed to read sysfs clock", "sensor", s.name, "path", s.path, "error", err)
		return nil, err
	}
	return map[string]interface{}{s.name: frequency}, nil
}

func GetSysFsClock(ctx context.Context, path string) (int64, error) {
	return readIntFromFile(ctx, path)
}

func GetSysFsCpuPaths() ([]string, error) {
	paths, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*")
	if err != nil {
		return nil, err
	}
	validPaths := make([]string, 0)
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		validPaths = append(validPaths, path)
	}

	return validPaths, nil
}

func readIntFromFile(ctx context.Context, path string) (int64, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	file, err := utils.ReadFileWithContext(ctxWithTimeout, path)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(string(file)), 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
