package nvidia

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/gpumonitor/gpusensor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

const (
	nvidiaSmi = "nvidia-smi"
)

var (
	nvidiaSmiDefaultSensors        = utils.Values(nvidiaSensorsByGPUReadingTypes)
	nvidiaGPUReadingTypesBySensor  = utils.Invert(nvidiaSensorsByGPUReadingTypes)
	nvidiaSensorsByGPUReadingTypes = map[gpusensor.GPUReadingType]string{
		gpusensor.GPUReadingTypeName:                             "name",
		gpusensor.GPUReadingTypeUUID:                             "uuid",
		gpusensor.GPUReadingTypePCIeAddress:                      "gpu_bus_id",
		gpusensor.GPUReadingTypeMemoryFree:                       "memory.free",
		gpusensor.GPUReadingTypeMemoryUsed:                       "memory.used",
		gpusensor.GPUReadingTypeMemoryTotal:                      "memory.total",
		gpusensor.GPUReadingTypeMemoryReserved:                   "memory.reserved",
		gpusensor.GPUReadingTypeUtilizationGPU:                   "utilization.gpu",
		gpusensor.GPUReadingTypeUtilizationMemory:                "utilization.memory",
		gpusensor.GPUReadingTypeUtilizationEncoder:               "utilization.encoder",
		gpusensor.GPUReadingTypeUtilizationDecoder:               "utilization.decoder",
		gpusensor.GPUReadingTypeUtilizationJPEG:                  "utilization.jpeg",
		gpusensor.GPUReadingTypeUtilizationOFA:                   "utilization.ofa",
		gpusensor.GPUReadingTypeTemperatureGPU:                   "temperature.gpu",
		gpusensor.GPUReadingTypeTemperatureGPULimit:              "temperature.gpu.tlimit",
		gpusensor.GPUReadingTypeTemperatureMemory:                "temperature.memory",
		gpusensor.GPUReadingTypePowerDraw:                        "power.draw",
		gpusensor.GPUReadingTypePowerLimit:                       "power.limit",
		gpusensor.GPUReadingTypePowerState:                       "pstate",
		gpusensor.GPUReadingTypeClocksGraphics:                   "clocks.current.graphics",
		gpusensor.GPUReadingTypeClocksGraphicsMax:                "clocks.max.graphics",
		gpusensor.GPUReadingTypeClocksStreamingMultiProcessor:    "clocks.current.sm",
		gpusensor.GPUReadingTypeClocksStreamingMultiProcessorMax: "clocks.max.sm",
		gpusensor.GPUReadingTypeClocksMemory:                     "clocks.current.memory",
		gpusensor.GPUReadingTypeCLocksMemoryMax:                  "clocks.max.memory",
		gpusensor.GPUReadingTypeClocksVideo:                      "clocks.current.video",
		gpusensor.GPUReadingTypeThrottleGPUIdle:                  "clocks_event_reasons.gpu_idle",
		gpusensor.GPUReadingTypeThrottleConfigSetting:            "clocks_event_reasons.applications_clocks_setting",
		gpusensor.GPUReadingTypeThrottleConfigPowerLimit:         "clocks_event_reasons.sw_power_cap",
		gpusensor.GPUReadingTypeThrottleHardwareLimit:            "clocks_event_reasons.hw_slowdown",
		gpusensor.GPUReadingTypeThrottleHardwarePowerLimit:       "clocks_event_reasons.hw_power_brake_slowdown",
		gpusensor.GPUReadingTypeThrottleHardwareThermalLimit:     "clocks_event_reasons.hw_thermal_slowdown",
		gpusensor.GPUReadingTypeThrottleSoftwareThermalLimit:     "clocks_event_reasons.sw_thermal_slowdown",
		gpusensor.GPUReadingTypeFanSpeed:                         "fan.speed",
		gpusensor.GPUReadingTypePCIeLinkGenGPUCurrent:            "pcie.link.gen.gpucurrent",
		gpusensor.GPUReadingTypePCIeLinkGenGPUMax:                "pcie.link.gen.gpumax",
		gpusensor.GPUReadingTypePCIeLinkGenHostMax:               "pcie.link.gen.hostmax",
		gpusensor.GPUReadingTypePCIeLinkGenMax:                   "pcie.link.gen.max",
		gpusensor.GPUReadingTypePCIeWidthCurrent:                 "pcie.link.width.current",
		gpusensor.GPUReadingTypePCIeWidthMax:                     "pcie.link.width.max",
		gpusensor.GPUReadingTypeGPUModeCurrent:                   "gom.current",
		gpusensor.GPUReadingTypeGPUModePending:                   "gom.pending",
	}
)

type nvidiaGpuMonitor struct {
	logger         logging.Logger
	sensorsToQuery []string
}

func (n *nvidiaGpuMonitor) Close() error {
	return nil
}

func (n *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) (map[string][]gpusensor.GPUSensorReading, error) {

	output, err := getNvidiaSmiOutput()
	if err != nil {
		return nil, errors.Join(errors.New("error detecting gpus with nvidia-smi"), err)
	}
	n.logger.Debugf("nvidia-smi output: %s", output)
	return n.parseNvidiaSmiOutput(output)
}

func (n *nvidiaGpuMonitor) parseNvidiaSmiOutput(output []byte) (map[string][]gpusensor.GPUSensorReading, error) {
	stats := make(map[string][]gpusensor.GPUSensorReading, 0)
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
		readings := make([]gpusensor.GPUSensorReading, 0)
		if len(row) != len(headers) {
			return nil, errors.New("mismatched header and row lengths")
		}
		for i, header := range headers {
			normalizedHeader := getNormalizedHeader(header)
			row[i] = strings.TrimSpace(row[i])
			if nvidiaGPUReadingTypesBySensor[normalizedHeader] == gpusensor.GPUReadingTypeUUID {
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
				if normalizedHeader == string(gpusensor.GPUReadingTypeName) {
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
			readings = append(readings, gpusensor.GPUSensorReading{
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
