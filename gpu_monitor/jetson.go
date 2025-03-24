package gpu_monitor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

var (
	ErrDevicePathNotFound = errors.New("device path not found")
	ErrStatsNotAvailable  = errors.New("stats not available for this device")
	searchPaths           = []string{
		"/sys/devices/platform/bus@0/",
		"/sys/devices/platform/",
	}

	frequencyBasePath     = "/sys/class/devfreq/"
	statsEnabledPostfixes = []string{"gpu", "ga10b", "gp10b"}
	deviceCounts          = make(map[string]int)
	jetsonClockSensors    = map[string]string{
		"nvenc": "encoder",
		"nvdec": "decoder",
		"nvjpg": "jpg",
		"ofa":   "ofa",
		"gpu":   "graphis",
		"vic":   "vic",
	}
	jetpack5LoadSensors = map[string]string{
		"gpu": "/sys/devices/platform/gpu.0/load",
	}
	jetpack6LoadSensors = map[string]string{
		"gpu": "/sys/devices/platform/bus@0/gpu.0/load",
	}
)

func getStatsPath(name string) (string, error) {
	isStatsEnabled := false
	for _, postfix := range statsEnabledPostfixes {
		if strings.Contains(name, postfix) {
			isStatsEnabled = true
		}
	}
	if !isStatsEnabled {
		return "", nil
	}
	for _, path := range searchPaths {
		devicePath := filepath.Join(path, name)
		if _, err := os.Stat(devicePath); err == nil {
			return devicePath, nil
		}
	}
	return "", ErrDevicePathNotFound
}

func newJetsonGpuComponent(fullname string) (*jetsonGpuComponent, error) {
	prettyName := fullname
	parts := strings.Split(fullname, ".")
	if len(parts) > 1 {
		prettyName = parts[1]
	}
	statsPath, err := getStatsPath(fullname)
	if err != nil {
		return nil, err
	}
	device := &jetsonGpuComponent{
		Name:          fullname,
		PrettyName:    prettyName,
		FrequencyPath: filepath.Join(frequencyBasePath, fullname),
		StatsPath:     statsPath,
	}
	return device, nil
}

type jetsonGpuComponent struct {
	Name          string
	PrettyName    string
	FrequencyPath string
	StatsPath     string
	minFrequency  *int64
	maxFrequency  *int64
}

func (d *jetsonGpuComponent) GetGovernor(ctx context.Context) (string, error) {
	govPath := filepath.Join(d.FrequencyPath, "governor")
	return utils.ReadFileWithContext(ctx, govPath)
}

func (d *jetsonGpuComponent) GetLoad(ctx context.Context) (float64, error) {
	if d.StatsPath == "" {
		return 0, ErrStatsNotAvailable
	}
	loadPath := filepath.Join(d.StatsPath, "load")
	load, err := utils.ReadInt64FromFileWithContext(ctx, loadPath)
	return float64(load), err
}

type jetsonGpuMonitor struct {
	logger  logging.Logger
	sensors []gpuSensor
}

type jetsonGpuClock struct {
	name         string
	path         string
	maxFrequency *int64
	minFrequency *int64
}

func (d *jetsonGpuClock) getCurrentFrequency(ctx context.Context) (int64, error) {
	freqPath := filepath.Join(d.path, "cur_freq")
	return utils.ReadInt64FromFileWithContext(ctx, freqPath)
}

func (d *jetsonGpuClock) getMaxFrequency(ctx context.Context) (int64, error) {
	if d.maxFrequency != nil {
		return *d.maxFrequency, nil
	}
	freqPath := filepath.Join(d.path, "max_freq")
	freq, err := utils.ReadInt64FromFileWithContext(ctx, freqPath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.maxFrequency = &freq
	return freq, nil
}

func (d *jetsonGpuClock) getMinFrequency(ctx context.Context) (int64, error) {
	if d.minFrequency != nil {
		return *d.minFrequency, nil
	}
	freqPath := filepath.Join(d.path, "min_freq")
	freq, err := utils.ReadInt64FromFileWithContext(ctx, freqPath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.minFrequency = &freq
	return freq, nil
}

func (s *jetsonGpuClock) Name() string {
	return s.name
}

func (s *jetsonGpuClock) GetSensorReading(ctx context.Context) (*gpuSensorReading, error) {
	curFreq, err := s.getCurrentFrequency(ctx)
	if err != nil {
		return nil, err
	}
	maxFreq, err := s.getMaxFrequency(ctx)
	if err != nil {
		return nil, err
	}
	minFreq, err := s.getMinFrequency(ctx)
	if err != nil {
		return nil, err
	}

	return &gpuSensorReading{
		Name:         s.name,
		Type:         GPUSensorTypeFrequency,
		CurrentValue: curFreq,
		MaxValue:     maxFreq,
		MinValue:     minFreq,
	}, nil
}

func newJetsonGpuMonitor(ctx context.Context, logger logging.Logger) (gpuMonitor, error) {
	devices := make([]gpuSensor, 0)
	dirs, err := os.ReadDir(frequencyBasePath)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		name := dir.Name()
		isValidDevice := false
		for _, postfix := range jetsonClockSensors {
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
		logger.Infof("Found GPU Sensor %s", name)
		// device, err := newJetsonGpuComponent(ctx, name)
		// if err != nil {
		// 	return nil, err
		// }
		// if _, ok := deviceCounts[device.PrettyName]; !ok {
		// 	deviceCounts[device.PrettyName] = 0

		// } else {
		// 	deviceCounts[device.PrettyName]++
		// }
		// device.PrettyName = fmt.Sprintf("%s%d", device.PrettyName, deviceCounts[device.PrettyName])
		// for _, dev := range devices {
		// 	if dev.PrettyName == device.PrettyName {
		// 		device.PrettyName = device.PrettyName + "2"
		// 	}
		// }
		// devices = append(devices, device)
	}
	return &jetsonGpuMonitor{logger: logger, sensors: devices}, nil
}

func (m *jetsonGpuMonitor) GetGPUStats(ctx context.Context) ([]*gpuSensorReading, error) {
	stats := make([]*gpuSensorReading, 0)
	for _, device := range m.sensors {
		m.logger.Debugf("Getting stats for %s", device.Name)
		stat, err := device.GetSensorReading(ctx)
		if err != nil {
			m.logger.Errorf("Failed to get sensor reading for %s: %v", device.Name, err)
			continue
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func (m *jetsonGpuMonitor) Close() error {
	// No resources to clean up for Jetson GPU monitor
	return nil
}
