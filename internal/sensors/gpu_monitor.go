package sensors

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

const (
	nvidiaSmi = "nvidia-smi"
)

var (
	nvidiaSmiDefaultSensors        = utils.Values(nvidiaSensorsByGPUReadingTypes)
	nvidiaGPUReadingTypesBySensor  = utils.Invert(nvidiaSensorsByGPUReadingTypes)
	nvidiaSensorsByGPUReadingTypes = map[GPUReadingType]string{
		GPUReadingTypeName:                             "name",
		GPUReadingTypeUUID:                             "uuid",
		GPUReadingTypePCIeAddress:                      "gpu_bus_id",
		GPUReadingTypeMemoryFree:                       "memory.free",
		GPUReadingTypeMemoryUsed:                       "memory.used",
		GPUReadingTypeMemoryTotal:                      "memory.total",
		GPUReadingTypeMemoryReserved:                   "memory.reserved",
		GPUReadingTypeUtilizationGPU:                   "utilization.gpu",
		GPUReadingTypeUtilizationMemory:                "utilization.memory",
		GPUReadingTypeUtilizationEncoder:               "utilization.encoder",
		GPUReadingTypeUtilizationDecoder:               "utilization.decoder",
		GPUReadingTypeUtilizationJPEG:                  "utilization.jpeg",
		GPUReadingTypeUtilizationOFA:                   "utilization.ofa",
		GPUReadingTypeTemperatureGPU:                   "temperature.gpu",
		GPUReadingTypeTemperatureGPULimit:              "temperature.gpu.tlimit",
		GPUReadingTypeTemperatureMemory:                "temperature.memory",
		GPUReadingTypePowerDraw:                        "power.draw",
		GPUReadingTypePowerLimit:                       "power.limit",
		GPUReadingTypePowerState:                       "pstate",
		GPUReadingTypeClocksGraphics:                   "clocks.current.graphics",
		GPUReadingTypeClocksGraphicsMax:                "clocks.max.graphics",
		GPUReadingTypeClocksStreamingMultiProcessor:    "clocks.current.sm",
		GPUReadingTypeClocksStreamingMultiProcessorMax: "clocks.max.sm",
		GPUReadingTypeClocksMemory:                     "clocks.current.memory",
		GPUReadingTypeCLocksMemoryMax:                  "clocks.max.memory",
		GPUReadingTypeClocksVideo:                      "clocks.current.video",
		GPUReadingTypeThrottleGPUIdle:                  "clocks_event_reasons.gpu_idle",
		GPUReadingTypeThrottleConfigSetting:            "clocks_event_reasons.applications_clocks_setting",
		GPUReadingTypeThrottleConfigPowerLimit:         "clocks_event_reasons.sw_power_cap",
		GPUReadingTypeThrottleHardwareLimit:            "clocks_event_reasons.hw_slowdown",
		GPUReadingTypeThrottleHardwarePowerLimit:       "clocks_event_reasons.hw_power_brake_slowdown",
		GPUReadingTypeThrottleHardwareThermalLimit:     "clocks_event_reasons.hw_thermal_slowdown",
		GPUReadingTypeThrottleSoftwareThermalLimit:     "clocks_event_reasons.sw_thermal_slowdown",
		GPUReadingTypeFanSpeed:                         "fan.speed",
		GPUReadingTypePCIeLinkGenGPUCurrent:            "pcie.link.gen.gpucurrent",
		GPUReadingTypePCIeLinkGenGPUMax:                "pcie.link.gen.gpumax",
		GPUReadingTypePCIeLinkGenHostMax:               "pcie.link.gen.hostmax",
		GPUReadingTypePCIeLinkGenMax:                   "pcie.link.gen.max",
		GPUReadingTypePCIeWidthCurrent:                 "pcie.link.width.current",
		GPUReadingTypePCIeWidthMax:                     "pcie.link.width.max",
		GPUReadingTypeGPUModeCurrent:                   "gom.current",
		GPUReadingTypeGPUModePending:                   "gom.pending",
	}
)

type nvidiaGpuMonitor struct {
	logger         logging.Logger
	sensorsToQuery []string
}

func (n *nvidiaGpuMonitor) Close() error {
	return nil
}

func (n *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) (map[string][]GPUSensorReading, error) {

	output, err := getNvidiaSmiOutput()
	if err != nil {
		return nil, errors.Join(errors.New("error detecting gpus with nvidia-smi"), err)
	}
	n.logger.Debugf("nvidia-smi output: %s", output)
	return n.parseNvidiaSmiOutput(output)
}

func (n *nvidiaGpuMonitor) parseNvidiaSmiOutput(output []byte) (map[string][]GPUSensorReading, error) {
	stats := make(map[string][]GPUSensorReading, 0)
	reader := csv.NewReader(bytes.NewReader(output))
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, errors.Join(errors.New("error reading headers from nvidia-smi output"), err)
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Join(errors.New("error reading rows from nvidia-smi output"), err)
	}

	for _, row := range rows {
		var gpuID string
		readings := make([]GPUSensorReading, 0)
		if len(row) != len(headers) {
			return nil, errors.New("mismatched header and row lengths")
		}
		for i, header := range headers {
			normalizedHeader := getNormalizedHeader(header)
			row[i] = strings.TrimSpace(row[i])
			if nvidiaGPUReadingTypesBySensor[normalizedHeader] == GPUReadingTypeUUID {
				gpuID = row[i]
				continue
			}
			var val any
			switch row[i] {
			case "[N/A]":
				fallthrough
			case "N/A":
				val = row[i]
			case "Active":
				val = true
			case "Not Active":
				val = false
			default:
				if strings.Contains(row[i], "-") || strings.Contains(row[i], ":") || strings.HasPrefix(row[i], "P") {
					val = row[i]
					continue
				}
				val, err = strconv.ParseInt(row[i], 10, 64)
				if err == nil {
					break
				}
				val, err = strconv.ParseFloat(row[i], 64)
				if err == nil {
					break
				}
				if normalizedHeader == string(GPUReadingTypeName) {
					val = row[i]
					break
				}
				n.logger.Debugf("error parsing value %s for sensor %s, treating as a string: %v", row[i], normalizedHeader, err)
				val = row[i]
			}
			if _, ok := nvidiaGPUReadingTypesBySensor[normalizedHeader]; !ok {
				n.logger.Debugf("skipping sensor %s, it is not in the list", normalizedHeader)
				continue
			}
			readings = append(readings, GPUSensorReading{
				Type:  nvidiaGPUReadingTypesBySensor[normalizedHeader],
				Value: val,
			})
		}
		stats[gpuID] = readings
	}
	return stats, nil
}

func getNvidiaSmiOutput() ([]byte, error) {
	cmd := exec.Command(nvidiaSmi, "--query-gpu", strings.Join(nvidiaSmiDefaultSensors, ","), "--format=csv,nounits")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Join(errors.New("error detecting gpus with nvidia-smi"), err)
	}
	return output, nil
}

func getNormalizedHeader(header string) string {
	for _, sensor := range nvidiaSmiDefaultSensors {
		if strings.Contains(header, sensor) {
			return sensor
		}
	}
	return header
}
