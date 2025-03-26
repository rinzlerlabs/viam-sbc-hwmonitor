package gpumonitor

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type gpuSensorType string

const (
	// GPU sensor types
	GPUSensorTypePower      gpuSensorType = "power"
	GPUSensorTypePowerState gpuSensorType = "power_state"
	GPUSensorTypeFrequency  gpuSensorType = "frequency"
	GPUSensorTypeLoad       gpuSensorType = "load"
	GPUSensorTypeMemory     gpuSensorType = "memory"
)

type gpuSensor interface {
	GetSensorReading(context.Context) (*gpuSensorReading, error)
	Name() string
}

type gpuMonitor interface {
	Close() error
	GetGPUStats(context.Context) ([]gpuSensorReading, error)
}

type gpuSensorReading struct {
	Name            string
	Type            gpuSensorType
	HasMaxValue     bool
	HasMinValue     bool
	HasCurrentValue bool
	CurrentValue    int64
	MaxValue        int64
	MinValue        int64
}

func (g gpuSensorReading) String(f fmt.State) string {
	if f.Flag('#') {
		return fmt.Sprintf("{ Name: %s, Type: %s, Current: %d, Max: %d, Min: %d }",
			g.Name, g.Type, g.CurrentValue, g.MaxValue, g.MinValue)
	} else {
		return fmt.Sprintf("{ %s %s %d %d %d }",
			g.Name, g.Type, g.CurrentValue, g.MaxValue, g.MinValue)
	}
}

func (g gpuSensorReading) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%s", g.String(f))
}
