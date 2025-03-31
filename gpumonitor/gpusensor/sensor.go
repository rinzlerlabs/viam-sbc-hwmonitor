package gpusensor

import "fmt"

type GPUReadingType string

const (
	GPUReadingTypeOther                            GPUReadingType = "other"
	GPUReadingTypeName                             GPUReadingType = "name"
	GPUReadingTypeUUID                             GPUReadingType = "uuid"
	GPUReadingTypePCIeAddress                      GPUReadingType = "pciAddress"
	GPUReadingTypeMemoryFree                       GPUReadingType = "memoryFree"
	GPUReadingTypeMemoryUsed                       GPUReadingType = "memoryUsed"
	GPUReadingTypeMemoryTotal                      GPUReadingType = "memoryTotal"
	GPUReadingTypeMemoryReserved                   GPUReadingType = "memoryReserved"
	GPUReadingTypeUtilizationGPU                   GPUReadingType = "utilizationGPU"
	GPUReadingTypeUtilizationMemory                GPUReadingType = "utilizationMemory"
	GPUReadingTypeUtilizationEncoder               GPUReadingType = "utilizationEncoder"
	GPUReadingTypeUtilizationDecoder               GPUReadingType = "utilizationDecoder"
	GPUReadingTypeUtilizationJPEG                  GPUReadingType = "utilizationJPEG"
	GPUReadingTypeUtilizationOFA                   GPUReadingType = "utilizationOFA"
	GPUReadingTypeTemperatureGPU                   GPUReadingType = "temperatureGPU"
	GPUReadingTypeTemperatureGPULimit              GPUReadingType = "temperatureGPULimit"
	GPUReadingTypeTemperatureMemory                GPUReadingType = "temperatureMemory"
	GPUReadingTypePowerDraw                        GPUReadingType = "powerDraw"
	GPUReadingTypePowerLimit                       GPUReadingType = "powerLimit"
	GPUReadingTypePowerState                       GPUReadingType = "powerState"
	GPUReadingTypeClocksGraphics                   GPUReadingType = "clocksGraphics"
	GPUReadingTypeClocksGraphicsMax                GPUReadingType = "clocksGraphicsMax"
	GPUReadingTypeClocksStreamingMultiProcessor    GPUReadingType = "clocksStreamingMultiProcessor"
	GPUReadingTypeClocksStreamingMultiProcessorMax GPUReadingType = "clocksStreamingMultiProcessorMax"
	GPUReadingTypeClocksMemory                     GPUReadingType = "clocksMemory"
	GPUReadingTypeCLocksMemoryMax                  GPUReadingType = "clocksMemoryMax"
	GPUReadingTypeClocksVideo                      GPUReadingType = "clocksVideo"
	GPUReadingTypeClocksVideoImageCompositor       GPUReadingType = "clocksVideoImageCompositor"
	GPUReadingTypeClocksJPEG                       GPUReadingType = "clocksJPEG"
	GPUReadingTypeClocksOFA                        GPUReadingType = "clocksOFA"
	GPUReadingTypeThrottleGPUIdle                  GPUReadingType = "throttleGPUIdle"
	GPUReadingTypeThrottleConfigSetting            GPUReadingType = "throttleConfigSetting"
	GPUReadingTypeThrottleConfigPowerLimit         GPUReadingType = "throttleConfigPowerLimit"
	GPUReadingTypeThrottleHardwareLimit            GPUReadingType = "throttleHardwareLimit"
	GPUReadingTypeThrottleHardwarePowerLimit       GPUReadingType = "throttleHardwarePowerLimit"
	GPUReadingTypeThrottleHardwareThermalLimit     GPUReadingType = "throttleHardwareThermalLimit"
	GPUReadingTypeThrottleSoftwareThermalLimit     GPUReadingType = "throttleSoftwareThermalLimit"
	GPUReadingTypeFanSpeed                         GPUReadingType = "fanSpeed"
	GPUReadingTypePCIeLinkGenGPUCurrent            GPUReadingType = "pcieLinkGenGPUCurrent"
	GPUReadingTypePCIeLinkGenGPUMax                GPUReadingType = "pcieLinkGenGPUMax"
	GPUReadingTypePCIeLinkGenHostMax               GPUReadingType = "pcieLinkGenHostMox"
	GPUReadingTypePCIeLinkGenMax                   GPUReadingType = "pcieLinkGenMax"
	GPUReadingTypePCIeWidthCurrent                 GPUReadingType = "pcieLinkWidthCurrent"
	GPUReadingTypePCIeWidthMax                     GPUReadingType = "pcieLinkWidthMax"
	GPUReadingTypeGPUModeCurrent                   GPUReadingType = "gpuModeCurrent"
	GPUReadingTypeGPUModePending                   GPUReadingType = "gpuModePending"
)

type GPUSensorReading struct {
	Type  GPUReadingType
	Value any
}

func (g GPUSensorReading) String(f fmt.State) string {
	if f.Flag('#') {
		return fmt.Sprintf("{ Type: %s, Current: %d }",
			g.Type, g.Value)
	} else {
		return fmt.Sprintf("{ %s %d }",
			g.Type, g.Value)
	}
}

func (g GPUSensorReading) Format(f fmt.State, c rune) {
	fmt.Fprintf(f, "%s", g.String(f))
}
