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

const (
	GPUReadingTypeOther                            gpuReadingType = "other"
	GPUReadingTypeName                             gpuReadingType = "name"
	GPUReadingTypeUUID                             gpuReadingType = "uuid"
	GPUReadingTypePCIeAddress                      gpuReadingType = "pciAddress"
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
	GPUReadingTypeClocksVideoImageCompositor       gpuReadingType = "clocksVideoImageCompositor"
	GPUReadingTypeClocksJPEG                       gpuReadingType = "clocksJPEG"
	GPUReadingTypeClocksOFA                        gpuReadingType = "clocksOFA"
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

type gpuMonitor interface {
	// Close closes the GPU monitor.
	Close() error
	// GetGPUStats returns a map of GPU sensor readings.
	// The key is the identifier for the GPU in the system.
	// The value is a slice of gpuSensorReading.
	// The slice contains the readings for each sensor on the GPU.
	// The readings are in the order they were found.
	// The readings are not guaranteed to be in any particular order.
	// The readings are guaranteed to be unique.
	GetGPUStats(context.Context) (map[string][]gpuSensorReading, error)
}

type gpuSensorReading struct {
	Type  gpuReadingType
	Value any
}

func (g gpuSensorReading) String(f fmt.State) string {
	if f.Flag('#') {
		return fmt.Sprintf("{ Type: %s, Current: %d }",
			g.Type, g.Value)
	} else {
		return fmt.Sprintf("{ %s %d }",
			g.Type, g.Value)
	}
}

func (g gpuSensorReading) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%s", g.String(f))
}

func newGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return newJetsonGpuMonitor(logger)
	} else if hasNvidiaSmiCommand(logger) {
		return newNVIDIAGpuMonitor(logger)
	}
	return nil, ErrUnsupportedBoard
}
