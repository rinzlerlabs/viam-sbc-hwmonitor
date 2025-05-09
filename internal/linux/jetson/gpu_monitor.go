package jetson

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	ErrDevicePathNotFound = errors.New("device path not found")
	ErrStatsNotAvailable  = errors.New("stats not available for this device")

	jetpack5Sensors = []jetsonGpuSensor{
		{sensorType: sensors.GPUReadingTypeClocksGraphics, currentValuePath: "/sys/class/devfreq/17000000.ga10b/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksVideo, currentValuePath: "/sys/class/devfreq/15480000.nvdec/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksMemory, currentValuePath: "/sys/kernel/debug/clk/emc/clk_rate"},
		{sensorType: sensors.GPUReadingTypeClocksVideoImageCompositor, currentValuePath: "/sys/class/devfreq/15340000.vic/cur_freq"},
		{sensorType: sensors.GPUReadingTypeUtilizationGPU, currentValuePath: "/sys/devices/platform/gpu.0/load", multiplier: 0.1},
		{sensorType: sensors.GPUReadingTypeMemoryFree, currentValuePath: "/sys/kernel/debug/nvmap/iovmm/free_size", regex: regexp.MustCompile(`([0-9]+)\s+bytes\s*$`)},
		{sensorType: sensors.GPUReadingTypeMemoryUsed, currentValuePath: "/sys/kernel/debug/nvmap/stats/total_memory"},
	}
	jetpack6Sensors = []jetsonGpuSensor{
		{sensorType: sensors.GPUReadingTypeClocksGraphics, currentValuePath: "/sys/class/devfreq/17000000.gpu/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksVideo, currentValuePath: "/sys/class/devfreq/15480000.nvdec/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksMemory, currentValuePath: "/sys/kernel/debug/clk/emc/clk_rate"},
		{sensorType: sensors.GPUReadingTypeClocksJPEG, currentValuePath: "/sys/class/devfreq/15380000.nvjpg/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksJPEG, currentValuePath: "/sys/class/devfreq/15540000.nvjpg/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksVideoImageCompositor, currentValuePath: "/sys/class/devfreq/15340000.vic/cur_freq"},
		{sensorType: sensors.GPUReadingTypeClocksOFA, currentValuePath: "/sys/class/devfreq/15a50000.ofa/cur_freq"},
		{sensorType: sensors.GPUReadingTypeUtilizationGPU, currentValuePath: "/sys/devices/platform/bus@0/gpu.0/load", multiplier: 0.1},
		{sensorType: sensors.GPUReadingTypeMemoryFree, currentValuePath: "/sys/kernel/debug/nvmap/iovmm/free_size", regex: regexp.MustCompile(`([0-9]+)\s+bytes\s*$`)},
		{sensorType: sensors.GPUReadingTypeMemoryUsed, currentValuePath: "/sys/kernel/debug/nvmap/stats/total_memory"},
	}
)

func NewJetsonGpuMonitor(logger logging.Logger) (*jetsonGpuMonitor, error) {
	gpuSensors, err := getJetsonGpuSensors()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU load sensors: %w", err)
	}

	return &jetsonGpuMonitor{logger: logger, sensors: gpuSensors}, nil
}

type jetsonGpuMonitor struct {
	logger  logging.Logger
	sensors []jetsonGpuSensor
}

type jetsonGpuSensor struct {
	sensorType       sensors.GPUReadingType
	currentValuePath string
	multiplier       float64
	regex            *regexp.Regexp
}

func (d *jetsonGpuSensor) CurrentValue(ctx context.Context) (float64, error) {
	b, err := utils.ReadFileWithContext(ctx, d.currentValuePath)
	if err != nil {
		return 0, err
	}
	var value float64
	if d.regex != nil {
		matches := d.regex.FindStringSubmatch(string(b))
		if len(matches) > 1 {
			value, err = utils.ParseFloat64(matches[1])
			if err != nil {
				return 0, fmt.Errorf("failed to parse value: %w", err)
			}
		} else {
			return 0, fmt.Errorf("failed to match regex: %s", d.regex.String())
		}
	} else {
		value, err = utils.ParseFloat64(string(b))
		if err != nil {
			return 0, fmt.Errorf("failed to parse value: %w", err)
		}
	}
	if d.multiplier != 0 {
		return value * d.multiplier, nil
	}
	return value, nil
}

func (s *jetsonGpuSensor) GetSensorReading(ctx context.Context) (*sensors.GPUSensorReading, error) {
	currentValue, err := s.CurrentValue(ctx)
	if err != nil {
		return nil, err
	}
	return &sensors.GPUSensorReading{
		Type:  s.sensorType,
		Value: currentValue,
	}, nil
}

func getJetsonGpuSensors() ([]jetsonGpuSensor, error) {
	if _, err := os.Stat(jetpack5Sensors[0].currentValuePath); !os.IsNotExist(err) {
		return jetpack5Sensors, nil
	} else if _, err := os.Stat(jetpack6Sensors[0].currentValuePath); !os.IsNotExist(err) {
		return jetpack6Sensors, nil
	}

	return nil, errors.New("no load sensors found")
}

func (m *jetsonGpuMonitor) GetGPUStats(ctx context.Context) (map[string][]sensors.GPUSensorReading, error) {
	stats := make([]sensors.GPUSensorReading, 0)
	for _, sensor := range m.sensors {
		m.logger.Debugf("Getting stats for %s", sensor.sensorType)
		stat, err := sensor.GetSensorReading(ctx)
		if err != nil {
			m.logger.Errorf("Failed to get sensor reading for %s: %v", sensor.sensorType, err)
			continue
		}
		stats = append(stats, *stat)
	}
	return map[string][]sensors.GPUSensorReading{
		"gpu0": stats,
	}, nil
}

func (m *jetsonGpuMonitor) Close() error {
	// No resources to clean up for Jetson GPU monitor
	return nil
}
