package gpumonitor

import (
	"context"
	"errors"
	"fmt"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"go.viam.com/rdk/logging"
)

var (
	ErrUnsupportedBoard = errors.New("gpu stats not supported on this board")
)

type gpuReadingType string
type gpuSensorType string

const (
	// GPU sensor types
	GPUSensorTypePower      gpuSensorType = "power"
	GPUSensorTypePowerState gpuSensorType = "power_state"
	GPUSensorTypeFrequency  gpuSensorType = "frequency"
	GPUSensorTypeLoad       gpuSensorType = "load"
	GPUSensorTypeMemory     gpuSensorType = "memory"

	GPUReadingTypeMemoryFree                       gpuReadingType = "memoryFree"
	GPUReadingTypeMemoryUsed                       gpuReadingType = "memoryUsed"
	GPUReadingTypeMemoryTotal                      gpuReadingType = "memoryTotal"
	GPUReadingTypeMemoryReserved                   gpuReadingType = "memoryReserved"
	GPUReadingTypeUtilizationGPU                   gpuReadingType = "utilizationGPU"
	GPUReadingTypeUtilizationMemory                gpuReadingType = "utilizationMemory"
	GPUReadingTypeUtilizationEncoder               gpuReadingType = "utilizationEncoder"
	GPUReadingTypeUtilizationDecoder               gpuReadingType = "utilizationDecoder"
	GPUReadingTypeUtilizationJPEG                  gpuReadingType = "utilizationJPEG"
	GPUReadingTypeUtilizationOFA                   gpuReadingType = "utilizationOFA"
	GPUReadingTypeTemperatureGPU                   gpuReadingType = "temperatureGPU"
	GPUReadingTypeTemperatureGPULimit              gpuReadingType = "temperatureGPULimit"
	GPUReadingTypeTemperatureMemory                gpuReadingType = "temperatureMemory"
	GPUReadingTypePowerDraw                        gpuReadingType = "powerDraw"
	GPUReadingTypePowerLimit                       gpuReadingType = "powerLimit"
	GPUReadingTypePowerState                       gpuReadingType = "powerState"
	GPUReadingTypeClocksGraphics                   gpuReadingType = "clocksGraphics"
	GPUReadingTypeClocksGraphicsMax                gpuReadingType = "clocksGraphicsMax"
	GPUReadingTypeClocksStreamingMultiProcessor    gpuReadingType = "clocksStreamingMultiProcessor"
	GPUReadingTypeClocksStreamingMultiProcessorMax gpuReadingType = "clocksStreamingMultiProcessorMax"
	GPUReadingTypeClocksMemory                     gpuReadingType = "clocksMemory"
	GPUReadingTypeCLocksMemoryMax                  gpuReadingType = "clocksMemoryMax"
	GPUReadingTypeClocksVideo                      gpuReadingType = "clocksVideo"
	GPUReadingTypeThrottleGPUIdle                  gpuReadingType = "throttleGPUIdle"
	GPUReadingTypeThrottleConfigSetting            gpuReadingType = "throttleConfigSetting"
	GPUReadingTypeThrottleConfigPowerLimit         gpuReadingType = "throttleConfigPowerLimit"
	GPUReadingTypeThrottleHardwareLimit            gpuReadingType = "throttleHardwareLimit"
	GPUReadingTypeThrottleHardwarePowerLimit       gpuReadingType = "throttleHardwarePowerLimit"
	GPUReadingTypeThrottleHardwareThermalLimit     gpuReadingType = "throttleHardwareThermalLimit"
	GPUReadingTypeThrottleSoftwareThermalLimit     gpuReadingType = "throttleSoftwareThermalLimit"
	GPUReadingTypeFanSpeed                         gpuReadingType = "fanSpeed"
	GPUReadingTypePCIeLinkGenGPUCurrent            gpuReadingType = "pcieLinkGenGPUCurrent"
	GPUReadingTypePCIeLinkGenGPUMax                gpuReadingType = "pcieLinkGenGPUMax"
	GPUReadingTypePCIeLinkGenHostMax               gpuReadingType = "pcieLinkGenHostMox"
	GPUReadingTypePCIeLinkGenMax                   gpuReadingType = "pcieLinkGenMax"
	GPUReadingTypePCIeWidthCurrent                 gpuReadingType = "pcieLinkWidthCurrent"
	GPUReadingTypePCIeWidthMax                     gpuReadingType = "pcieLinkWidthMax"
	GPUReadingTypeGPUModeCurrent                   gpuReadingType = "gpuModeCurrent"
	GPUReadingTypeGPUModePending                   gpuReadingType = "gpuModePending"
)

type gpuSensor interface {
	GetSensorReading(context.Context) (*gpuSensorReading, error)
	Name() string
}

type gpuMonitor interface {
	Close() error
	GetGPUStats(context.Context) ([][]gpuSensorReading, error)
}

type gpuSensorReading struct {
	Name  string
	Type  gpuReadingType
	Value int64
}

func (g gpuSensorReading) String(f fmt.State) string {
	if f.Flag('#') {
		return fmt.Sprintf("{ Name: %s, Type: %s, Current: %d }",
			g.Name, g.Type, g.Value)
	} else {
		return fmt.Sprintf("{ %s %s %d }",
			g.Name, g.Type, g.Value)
	}
}

func (g gpuSensorReading) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%s", g.String(f))
}

func newGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return newJetsonGpuMonitor(logger)
	}
	if hasNvidiaSmiCommand() {
		return newNVIDIAGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
