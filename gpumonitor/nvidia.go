package gpumonitor

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

var (
	errUnsupportedSensorType   = errors.New("unsupported sensor type")
	nvidia_smi                 = "nvidia-smi"
	nvidia_smi_default_columns = []string{
		"name",
		"count",
		"uuid",
		"gpu_serial",
		"pci.bus_id",
		"pcie.link.gen.gpucurrent",
		"pcie.link.gen.max",
		"pcie.link.gen.gpumax",
		"pcie.link.gen.hostmax",
		"pcie.link.width.current",
		"pcie.link.width.max",
		"display_mode",
		"display_active",
		"vbios_version",
		"gpu_operation_mode.current",
		"gpu_operation_mode.pending",
		"fan.speed",
		"pstate",
		"clocks_event_reasons.hw_slowdown",
		"clocks_event_reasons.sw_power_cap",
		"clocks_event_reasons.hw_thermal_slowdown",
		"clocks_event_reasons.active",
		"clocks_event_reasons.gpu_idle",
		"clocks_event_reasons.applications_clocks_setting",
		"clocks_event_reasons.hw_power_brake_slowdown",
		"clocks_event_reasons.sw_thermal_slowdown",
		"clocks_event_reasons.sync_boost",
		"memory.total",
		"memory.reserved",
		"memory.used",
		"memory.free",
		"compute_mode",
		"compute_cap",
		"utilization.gpu",
		"utilization.memory",
		"utilization.encoder",
		"utilization.decoder",
		"utilization.jpeg",
		"utilization.ofa",
		"encoder.stats.sessionCount",
		"encoder.stats.averageFps",
		"encoder.stats.averageLatency",
		"temperature.gpu",
		"temperature.gpu.tlimit",
		"temperature.memory",
		"power.management",
		"power.draw",
		"power.draw.average",
		"power.draw.instant",
		"power.limit",
		"enforced.power.limit",
		"power.default_limit",
		"power.min_limit",
		"power.max_limit",
		"clocks.current.graphics",
		"clocks.current.sm",
		"clocks.current.memory",
		"clocks.current.video",
		"clocks.max.graphics",
		"clocks.max.sm",
		"clocks.max.memory",
	}
	nvidia_smi_parameters = []string{
		"--query-gpu",
		strings.Join(utils.Values(nvidiaColumnsByGPUReadingTypes), ","),
		"--format=csv",
	}

	nvidiaColumnsByGPUReadingTypes = map[gpuReadingType]string{
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
		GPUReadingTypeTemperatureGPULimit:              "temperature.gpi.tlimit",
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
		GPUReadingTypeGPUModeCurrent:                   "gpu_operation_mode.current",
		GPUReadingTypeGPUModePending:                   "gpu_operation_mode.pending",
	}

	nvidiaGPUReadingTypesByColumns = utils.Invert(nvidiaColumnsByGPUReadingTypes)
)

func hasNvidiaSmiCommand() bool {
	cmd := exec.Command("which", "nvidia-smi")
	err := cmd.Run()
	return err == nil
}

func newNVIDIAGpuMonitor(logger logging.Logger) (gpuMonitor, error) {
	monitor := &nvidiaGpuMonitor{
		logger: logger,
	}

	return monitor, nil
}

type nvidiaGpuMonitor struct {
	logger logging.Logger
}

func (n *nvidiaGpuMonitor) Close() error {
	return nil
}

func (n *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) ([]gpuSensorReading, error) {
	stats := make([]gpuSensorReading, 0)

	cmd := exec.Command(nvidia_smi, nvidia_smi_parameters...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Join(errors.New("error detecting gpus with nvidia-smi"), err)
	}
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
		if len(row) != len(headers) {
			return nil, errors.New("mismatched header and row lengths")
		}
		for i, header := range headers {
			row[i] = strings.TrimSpace(row[i])
			val, err := strconv.Atoi(row[i])
			if err != nil {
				n.logger.Infof("error converting value to int: %s", row[i])
				continue
			}
			stats = append(stats, gpuSensorReading{
				Name:  header,
				Type:  nvidiaGPUReadingTypesByColumns[header],
				Value: int64(val),
			})
		}
	}

	return stats, nil
}

type nvidiaSmiSample struct {
}
