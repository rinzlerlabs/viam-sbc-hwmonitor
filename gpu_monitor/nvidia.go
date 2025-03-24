//go:build full_nvidia_support
// +build full_nvidia_support

package gpu_monitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

var (
	errUnsupportedSensorType = errors.New("unsupported sensor type")
	clockTypes               = map[string]nvml.ClockType{
		"graphics": nvml.CLOCK_GRAPHICS,
		"sm":       nvml.CLOCK_SM,
		"mem":      nvml.CLOCK_MEM,
		"video":    nvml.CLOCK_VIDEO,
		"count":    nvml.CLOCK_COUNT,
	}

	nvidiaLoadSensors   = []string{"encoder", "decoder", "jpg", "ofa"}                     // Sensors that report load (utilization) values
	nvidiaPowerSensors  = []string{"power", "power_limit", "default_limit", "power_state"} // Sensors that report power values
	nvidiaMemorySensors = []string{"mem"}                                                  // Sensors that report memory values
)

func hasNVIDIAGPU() bool {
	err := utils.NVMLManager.Acquire()
	defer utils.NVMLManager.Release()
	return err == nil
}

func newNVIDIAGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	ret := utils.NVMLManager.Acquire()
	if ret != nil {
		logger.Errorf("Failed to acquire NVML: %v", ret)
		return nil, ret
	}
	gpus := make([]nvidiaGpu, 0)
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		logger.Errorf("Failed to get device count: %v", ret)
		return nil, ret
	}
	for i := range deviceCount {
		gpu, err := newNvidiaGpu(logger, i)
		if err != nil {
			logger.Errorf("Failed to create GPU %d: %v", i, err)
			continue
		}
		gpus = append(gpus, *gpu)
	}
	return &nvidiaGpuMonitor{logger: logger, gpus: gpus}, nil
}

type nvidiaGpuMonitor struct {
	logger logging.Logger
	gpus   []nvidiaGpu
}

func (n *nvidiaGpuMonitor) Close() error {
	return utils.NVMLManager.Release()
}

func (n *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) ([]gpuSensorReading, error) {
	stats := make([]gpuSensorReading, 0)
	for _, gpu := range n.gpus {
		for _, sensor := range gpu.sensors {
			stat, err := sensor.GetSensorReading(ctx)
			if err != nil {
				n.logger.Errorf("Failed to get sensor reading for %s: %v", sensor.Name(), err)
				continue
			}
			stats = append(stats, *stat)
		}
	}
	return stats, nil
}

func newNvidiaGpu(logger logging.Logger, index int) (*nvidiaGpu, error) {
	nvmlDevice, ret := nvml.DeviceGetHandleByIndex(index)
	if ret != nvml.SUCCESS {
		logger.Errorf("Failed to get device handle for index %d: %v", index, ret)
		return nil, errors.Join(ret, fmt.Errorf("failed to get device handle for GPU %d", index))
	}
	name, ret := nvmlDevice.GetName()
	if ret != nvml.SUCCESS {
		logger.Errorf("Failed to get device name for index %d: %v", index, ret)
		return nil, errors.Join(ret, fmt.Errorf("failed to get device name for GPU %d", index))
	}
	sensors := make([]gpuSensor, 0)
	for name, clockType := range clockTypes { // Iterate over default clocks, not all clocks are supported so we check each one
		_, ret := nvmlDevice.GetClockInfo(clockType) // Current clocks
		if ret != nvml.SUCCESS {
			logger.Warnf("Failed to get clock info for device %s: %v", name, ret)
			continue
		}
		s, err := newNvidiaGpuClockSensor(name, nvmlDevice, clockType)
		if err != nil {
			logger.Errorf("Failed to create clock sensor for device %s: %v", name, err)
			continue
		}
		sensors = append(sensors, s)
	}

	for _, loadSensor := range nvidiaLoadSensors {
		if s, ret := newNvidiaGpuSensor(loadSensor, nvmlDevice, GPUSensorTypeLoad); ret == nvml.SUCCESS || ret == nil {
			sensors = append(sensors, s)
		} else {
			logger.Errorf("Failed to create %s sensor for device %s: %v", loadSensor, name, ret)
		}
	}
	for _, powerSensor := range nvidiaPowerSensors {
		if s, ret := newNvidiaGpuSensor(powerSensor, nvmlDevice, GPUSensorTypePower); ret == nvml.SUCCESS || ret == nil {
			sensors = append(sensors, s)
		} else {
			logger.Errorf("Failed to create %s sensor for device %s: %v", powerSensor, name, ret)
		}
	}
	for _, memorySensor := range nvidiaMemorySensors {
		if s, ret := newNvidiaGpuSensor(memorySensor, nvmlDevice, GPUSensorTypeMemory); ret == nvml.SUCCESS || ret == nil {
			sensors = append(sensors, s)
		} else {
			logger.Errorf("Failed to create %s sensor for device %s: %v", memorySensor, name, ret)
		}
	}
	return &nvidiaGpu{
		logger:  logger,
		name:    name,
		index:   index,
		sensors: sensors,
	}, nil
}

type nvidiaGpu struct {
	logger  logging.Logger
	name    string
	index   int
	sensors []gpuSensor
}

func (d *nvidiaGpu) GetName() string {
	return d.name
}

func (d *nvidiaGpu) Close() error {
	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		return ret
	}
	return nil
}

func newNvidiaGpuClockSensor(name string, nvmlDevice nvml.Device, clockType nvml.ClockType) (*nvidiaGpuClock, error) {
	if nvmlDevice == nil {
		return nil, fmt.Errorf("nvmlDevice cannot be nil")
	}
	maxFreq, ret := nvmlDevice.GetMaxClockInfo(clockType)
	if ret != nvml.SUCCESS {
		return nil, errors.Join(ret, fmt.Errorf("failed to get max clock info for %s", name))
	}
	return &nvidiaGpuClock{
		name:         name,
		clockType:    clockType,
		nvmlDevice:   nvmlDevice,
		maxFrequency: maxFreq,
	}, nil
}

type nvidiaGpuClock struct {
	name         string
	clockType    nvml.ClockType
	nvmlDevice   nvml.Device
	maxFrequency uint32
}

func (s *nvidiaGpuClock) Name() string {
	return s.name
}

func (s *nvidiaGpuClock) GetSensorReading(context.Context) (*gpuSensorReading, error) {
	reading := &gpuSensorReading{
		Name: s.name,
		Type: GPUSensorTypeFrequency,
	}
	if s.HasMinValue() {
		minFreq, err := s.MinValue()
		if err != nil {
			return nil, err
		}
		reading.MinValue = int64(minFreq)
	}
	if s.HasMaxValue() {
		maxFreq, err := s.MaxValue()
		if err != nil {
			return nil, err
		}
		reading.MaxValue = int64(maxFreq)
	}
	if s.HasCurrentValue() {
		curFreq, err := s.CurrentValue()
		if err != nil {
			return nil, err
		}
		reading.CurrentValue = int64(curFreq)
	}
	return reading, nil
}

func (s *nvidiaGpuClock) HasMinValue() bool {
	return false
}

func (s *nvidiaGpuClock) MinValue() (float64, error) {
	return 0, nil
}

func (s *nvidiaGpuClock) HasMaxValue() bool {
	return true
}

func (s *nvidiaGpuClock) MaxValue() (float64, error) {
	return float64(s.maxFrequency), nil
}

func (s *nvidiaGpuClock) HasCurrentValue() bool {
	return true
}

func (s *nvidiaGpuClock) CurrentValue() (float64, error) {
	if !s.HasCurrentValue() {
		return 0, errors.New("current value not supported for this sensor")
	}
	frequency, ret := s.nvmlDevice.GetClockInfo(s.clockType)
	if ret != nvml.SUCCESS {
		return 0, errors.Join(ret, fmt.Errorf("failed to get current clock frequency for %s", s.name))
	}
	return float64(frequency), nil
}

func newNvidiaGpuSensor(name string, nvmlDevice nvml.Device, sensorType gpuSensorType) (*nvidiaGpuSensor, error) {
	if nvmlDevice == nil {
		return nil, fmt.Errorf("nvmlDevice cannot be nil")
	}
	if _, err := getReading(sensorType, name, nvmlDevice); err != nvml.SUCCESS {
		return nil, err
	}
	return &nvidiaGpuSensor{
		name:       name,
		sensorType: sensorType,
		nvmlDevice: nvmlDevice,
	}, nil
}

type nvidiaGpuSensor struct {
	name       string
	sensorType gpuSensorType
	nvmlDevice nvml.Device
}

func (s *nvidiaGpuSensor) Name() string {
	return s.name
}

func (s *nvidiaGpuSensor) GetSensorReading(context.Context) (*gpuSensorReading, error) {
	reading := &gpuSensorReading{
		Name: s.name,
		Type: GPUSensorTypeFrequency,
	}
	if s.HasMinValue() {
		minFreq, err := s.MinValue()
		if err != nil {
			return nil, err
		}
		reading.MinValue = int64(minFreq)
	}
	if s.HasMaxValue() {
		maxFreq, err := s.MaxValue()
		if err != nil {
			return nil, err
		}
		reading.MaxValue = int64(maxFreq)
	}
	if s.HasCurrentValue() {
		curFreq, err := s.CurrentValue()
		if err != nil {
			return nil, err
		}
		reading.CurrentValue = int64(curFreq)
	}
	return reading, nil
}

func (s *nvidiaGpuSensor) HasMinValue() bool {
	return false
}

func (s *nvidiaGpuSensor) MinValue() (float64, error) {
	if !s.HasMinValue() {
		return 0, errors.New("min value not supported for this sensor")
	}
	return 0, nil
}

func (s *nvidiaGpuSensor) HasMaxValue() bool {
	return false
}

func (s *nvidiaGpuSensor) MaxValue() (float64, error) {
	if !s.HasMaxValue() {
		return 0, errors.New("max value not supported for this sensor")
	}
	return 0, nil
}

func (s *nvidiaGpuSensor) HasCurrentValue() bool {
	return true
}

func (s *nvidiaGpuSensor) CurrentValue() (float64, error) {
	if !s.HasCurrentValue() {
		return 0, errors.New("current value not supported for this sensor")
	}
	vals, err := getReading(s.sensorType, s.name, s.nvmlDevice)
	if err != nvml.SUCCESS && err != nil {
		return 0, err
	}
	if len(vals) == 0 {
		return 0, errors.New("no values returned from sensor")
	}
	return float64(vals[0]), nil
}

func getReading(sensorType gpuSensorType, sensor string, nvmlDevice nvml.Device) ([]uint32, error) {
	switch sensorType {
	case GPUSensorTypePower:
		switch sensor {
		case "power":
			val, ret := nvmlDevice.GetPowerUsage()
			return []uint32{val}, ret
		case "power_limit":
			val, ret := nvmlDevice.GetPowerManagementLimit()
			return []uint32{val}, ret
		case "default_limit":
			val, ret := nvmlDevice.GetPowerManagementDefaultLimit()
			return []uint32{val}, ret
		case "power_state":
			val, ret := nvmlDevice.GetPowerState()
			return []uint32{uint32(val)}, ret
		default:
			return nil, errUnsupportedSensorType
		}
	case GPUSensorTypeLoad:
		switch sensor {
		case "encoder":
			val1, val2, ret := nvmlDevice.GetEncoderUtilization()
			return []uint32{val1, val2}, ret
		case "decoder":
			val1, val2, ret := nvmlDevice.GetDecoderUtilization()
			return []uint32{val1, val2}, ret
		case "jpg":
			val1, val2, ret := nvmlDevice.GetJpgUtilization()
			return []uint32{val1, val2}, ret
		case "ofa":
			val1, val2, ret := nvmlDevice.GetOfaUtilization()
			return []uint32{val1, val2}, ret
		default:
			return nil, errUnsupportedSensorType
		}
	case GPUSensorTypeFrequency:
		panic("foo")
	case GPUSensorTypeMemory:
		switch sensor {
		case "mem":
			mem, ret := nvmlDevice.GetMemoryInfo()
			return []uint32{uint32(mem.Used), uint32(mem.Total)}, ret
		default:
			return nil, errUnsupportedSensorType
		}
	default:
		return nil, fmt.Errorf("unsupported sensor type: %s", sensorType)
	}
}
