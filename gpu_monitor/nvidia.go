package gpu_monitor

import (
	"context"
	"errors"
	"fmt"
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
	nvidiaDevicePostfixes = []string{"gpu", "ga10b", "gp10b", "vic", "nvjpg", "nvdec", "ofa"}
	statsEnabledPostfixes = []string{"gpu", "ga10b", "gp10b"}
	deviceCounts          = make(map[string]int)
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

func newNvidiaGpuDevice(ctx context.Context, fullname string) (*nvidiaGpuDevice, error) {
	prettyName := fullname
	parts := strings.Split(fullname, ".")
	if len(parts) > 1 {
		prettyName = parts[1]
	}
	statsPath, err := getStatsPath(fullname)
	if err != nil {
		return nil, err
	}
	device := &nvidiaGpuDevice{
		Name:          fullname,
		PrettyName:    prettyName,
		FrequencyPath: filepath.Join(frequencyBasePath, fullname),
		StatsPath:     statsPath,
	}
	return device, nil
}

type nvidiaGpuDevice struct {
	Name          string
	PrettyName    string
	FrequencyPath string
	StatsPath     string
	minFrequency  *int64
	maxFrequency  *int64
}

func (d *nvidiaGpuDevice) GetCurrentFrequency(ctx context.Context) (int64, error) {
	freqPath := filepath.Join(d.FrequencyPath, "cur_freq")
	return utils.ReadInt64FromFileWithContext(ctx, freqPath)
}

func (d *nvidiaGpuDevice) GetMaxFrequency(ctx context.Context) (int64, error) {
	if d.maxFrequency != nil {
		return *d.maxFrequency, nil
	}
	freqPath := filepath.Join(d.FrequencyPath, "max_freq")
	freq, err := utils.ReadInt64FromFileWithContext(ctx, freqPath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.maxFrequency = &freq
	return freq, nil
}

func (d *nvidiaGpuDevice) GetMinFrequency(ctx context.Context) (int64, error) {
	if d.minFrequency != nil {
		return *d.minFrequency, nil
	}
	freqPath := filepath.Join(d.FrequencyPath, "min_freq")
	freq, err := utils.ReadInt64FromFileWithContext(ctx, freqPath)
	if err != nil {
		return 0, err
	}
	// cache the value for the future
	d.minFrequency = &freq
	return freq, nil
}

func (d *nvidiaGpuDevice) GetGovernor(ctx context.Context) (string, error) {
	govPath := filepath.Join(d.FrequencyPath, "governor")
	return utils.ReadFileWithContext(ctx, govPath)
}

func (d *nvidiaGpuDevice) GetLoad(ctx context.Context) (float64, error) {
	if d.StatsPath == "" {
		return 0, ErrStatsNotAvailable
	}
	loadPath := filepath.Join(d.StatsPath, "load")
	load, err := utils.ReadInt64FromFileWithContext(ctx, loadPath)
	return float64(load), err
}

type nvidiaGpuMonitor struct {
	logger  logging.Logger
	devices []*nvidiaGpuDevice
}

func newNvidiaGpuMonitor(ctx context.Context, logger logging.Logger) (GpuMonitor, error) {
	devices := make([]*nvidiaGpuDevice, 0)
	dirs, err := os.ReadDir(frequencyBasePath)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		name := dir.Name()
		isValidDevice := false
		for _, postfix := range nvidiaDevicePostfixes {
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
		logger.Infof("Found GPU Device %s", name)
		device, err := newNvidiaGpuDevice(ctx, name)
		if err != nil {
			return nil, err
		}
		if _, ok := deviceCounts[device.PrettyName]; !ok {
			deviceCounts[device.PrettyName] = 0

		} else {
			deviceCounts[device.PrettyName]++
		}
		device.PrettyName = fmt.Sprintf("%s%d", device.PrettyName, deviceCounts[device.PrettyName])
		for _, dev := range devices {
			if dev.PrettyName == device.PrettyName {
				device.PrettyName = device.PrettyName + "2"
			}
		}
		devices = append(devices, device)
	}
	return &nvidiaGpuMonitor{logger: logger, devices: devices}, nil
}

func (m *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) ([]gpuDeviceStats, error) {
	stats := make([]gpuDeviceStats, 0)
	for _, device := range m.devices {
		m.logger.Debugf("Getting stats for %s", device.Name)
		curFreq, err := device.GetCurrentFrequency(ctx)
		if err != nil {
			return nil, err
		}
		maxFreq, err := device.GetMaxFrequency(ctx)
		if err != nil {
			return nil, err
		}
		minFreq, err := device.GetMinFrequency(ctx)
		if err != nil {
			return nil, err
		}
		gov, err := device.GetGovernor(ctx)
		if err != nil {
			return nil, err
		}
		load, err := device.GetLoad(ctx)
		if err != nil && err != ErrStatsNotAvailable {
			return nil, err
		}
		stats = append(stats, gpuDeviceStats{
			Name:             device.PrettyName,
			CurrentFrequency: curFreq,
			MaxFrequency:     maxFreq,
			MinFrequency:     minFreq,
			Governor:         gov,
			Load:             load,
		})
	}
	return stats, nil
}

func (m *nvidiaGpuMonitor) Close() {
}
