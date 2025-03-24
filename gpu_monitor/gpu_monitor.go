package gpu_monitor

import (
	"context"
	"errors"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type gpuSensorType string

const (
	// GPU sensor types
	GPUSensorTypePower     gpuSensorType = "power"
	GPUSensorTypeFrequency gpuSensorType = "frequency"
	GPUSensorTypeLoad      gpuSensorType = "load"
	GPUSensorTypeMemory    gpuSensorType = "memory"
)

type gpuSensor interface {
	GetSensorReading(context.Context) (*gpuSensorReading, error)
	Name() string
}

type gpuMonitor interface {
	Close() error
	GetGPUStats(context.Context) ([]*gpuSensorReading, error)
}

type gpuSensorReading struct {
	Name         string
	Type         gpuSensorType
	CurrentValue int64
	MaxValue     int64
	MinValue     int64
}

func gpuDeviceStatsToMap(stats []gpuSensorReading) map[string]interface{} {
	result := make(map[string]interface{})
	for _, stat := range stats {
		result[stat.Name+"-current_value"] = stat.CurrentValue
		result[stat.Name+"-max_value"] = stat.MaxValue
		result[stat.Name+"-min_value"] = stat.MinValue
	}
	return result
}
