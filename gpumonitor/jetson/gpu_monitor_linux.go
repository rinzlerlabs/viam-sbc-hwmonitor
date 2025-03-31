package jetson

import (
	"fmt"

	"go.viam.com/rdk/logging"
)

func NewJetsonGpuMonitor(logger logging.Logger) (*jetsonGpuMonitor, error) {
	gpuSensors, err := getJetsonGpuSensors()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU load sensors: %w", err)
	}

	return &jetsonGpuMonitor{logger: logger, sensors: gpuSensors}, nil
}
