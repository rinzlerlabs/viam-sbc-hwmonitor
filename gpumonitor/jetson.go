package gpumonitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

var (
	ErrDevicePathNotFound = errors.New("device path not found")
	ErrStatsNotAvailable  = errors.New("stats not available for this device")

	frequencyBasePath  = "/sys/class/devfreq/"
	jetsonClockSensors = map[string]string{
		"nvenc": "encoder",
		"nvdec": "decoder",
		"nvjpg": "jpg",
		"ofa":   "ofa",
		"gpu":   "graphic",
		"vic":   "vic",
	}
	jetpack5LoadSensors = map[string]string{
		"gpu": "/sys/devices/platform/gpu.0/load",
	}
	jetpack6LoadSensors = map[string]string{
		"gpu": "/sys/devices/platform/bus@0/gpu.0/load",
	}
)

type jetsonGpuMonitor struct {
	logger  logging.Logger
	sensors []gpuSensor
}

func newJetsonGpuFrequencySensor(name string, path string) *jetsonGpuSensor {
	return &jetsonGpuSensor{
		name:             name,
		sensorType:       GPUSensorTypeFrequency,
		currentValuePath: filepath.Join(path, "cur_freq"),
		minValuePath:     filepath.Join(path, "min_freq"),
		maxValuePath:     filepath.Join(path, "max_freq"),
	}
}

func newJetsonGpuLoadSensor(name, path string) *jetsonGpuSensor {
	minValue := int64(0)
	maxValue := int64(100)
	return &jetsonGpuSensor{
		name:             name,
		sensorType:       GPUSensorTypeLoad,
		currentValuePath: path,
		minValue:         &minValue,
		maxValue:         &maxValue,
	}
}

type jetsonGpuSensor struct {
	name             string
	sensorType       gpuSensorType
	currentValuePath string
	minValuePath     string
	maxValuePath     string
	minValue         *int64
	maxValue         *int64
}

func (d *jetsonGpuSensor) CurrentValue(ctx context.Context) (int64, error) {
	return utils.ReadInt64FromFileWithContext(ctx, d.currentValuePath)
}

func (d *jetsonGpuSensor) MaxValue(ctx context.Context) (int64, error) {
	if d.maxValue != nil {
		return *d.maxValue, nil
	}
	freq, err := utils.ReadInt64FromFileWithContext(ctx, d.maxValuePath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.maxValue = &freq
	return freq, nil
}

func (d *jetsonGpuSensor) MinValue(ctx context.Context) (int64, error) {
	if d.minValue != nil {
		return *d.minValue, nil
	}
	freq, err := utils.ReadInt64FromFileWithContext(ctx, d.minValuePath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.minValue = &freq
	return freq, nil
}

func (s *jetsonGpuSensor) Name() string {
	return s.name
}

func (s *jetsonGpuSensor) GetSensorReading(ctx context.Context) (*gpuSensorReading, error) {
	currentValue, err := s.CurrentValue(ctx)
	if err != nil {
		return nil, err
	}
	maxValue, err := s.MaxValue(ctx)
	if err != nil {
		return nil, err
	}
	minValue, err := s.MinValue(ctx)
	if err != nil {
		return nil, err
	}

	return &gpuSensorReading{
		Name:         s.name,
		Type:         s.sensorType,
		CurrentValue: currentValue,
		MaxValue:     maxValue,
		MinValue:     minValue,
	}, nil
}

func getJetsonGpuClockSensors() ([]gpuSensor, error) {
	clockSensors := make([]gpuSensor, 0)
	dirs, err := os.ReadDir(frequencyBasePath)
	if err != nil {
		return nil, err
	}
	dupeNames := make(map[string]int)
	for _, dir := range dirs {
		name := dir.Name()
		isValidDevice := false
		for postfix := range jetsonClockSensors {
			if strings.Contains(name, postfix) {
				isValidDevice = true
			}
		}
		if !isValidDevice {
			continue
		}

		realPath, err := filepath.EvalSymlinks(filepath.Join(frequencyBasePath, name))
		if err != nil {
			return nil, err
		}
		dirInfo, err := os.Stat(realPath)
		if err != nil {
			return nil, err
		}
		if !dirInfo.IsDir() {
			continue
		}
		fmt.Printf("Found GPU Sensor %s\n", name)
		prettyName := strings.Split(name, ".")[1]
		if _, ok := dupeNames[prettyName]; ok {
			prettyName = fmt.Sprintf("%s_%d", prettyName, dupeNames[prettyName])
		}
		dupeNames[prettyName]++
		sensor := newJetsonGpuFrequencySensor(prettyName, realPath)
		clockSensors = append(clockSensors, sensor)
	}
	return clockSensors, nil
}

func getJetsonGpuLoadSensors() ([]gpuSensor, error) {
	if _, err := os.Stat(jetpack5LoadSensors["gpu"]); !os.IsNotExist(err) {
		return []gpuSensor{newJetsonGpuLoadSensor("gpu", jetpack5LoadSensors["gpu"])}, nil
	} else if _, err := os.Stat(jetpack6LoadSensors["gpu"]); !os.IsNotExist(err) {
		return []gpuSensor{newJetsonGpuLoadSensor("gpu", jetpack6LoadSensors["gpu"])}, nil
	}

	return nil, errors.New("no load sensors found")
}

func newJetsonGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	devices := make([]gpuSensor, 0)
	clocks, err := getJetsonGpuClockSensors()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU clock sensors: %w", err)
	}
	devices = append(devices, clocks...)
	loadSensors, err := getJetsonGpuLoadSensors()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU load sensors: %w", err)
	}
	devices = append(devices, loadSensors...)

	return &jetsonGpuMonitor{logger: logger, sensors: devices}, nil
}

func (m *jetsonGpuMonitor) GetGPUStats(ctx context.Context) ([]gpuSensorReading, error) {
	stats := make([]gpuSensorReading, 0)
	for _, device := range m.sensors {
		m.logger.Debugf("Getting stats for %s", device.Name())
		stat, err := device.GetSensorReading(ctx)
		if err != nil {
			m.logger.Errorf("Failed to get sensor reading for %s: %v", device.Name(), err)
			continue
		}
		stats = append(stats, *stat)
	}
	return stats, nil
}

func (m *jetsonGpuMonitor) Close() error {
	// No resources to clean up for Jetson GPU monitor
	return nil
}
