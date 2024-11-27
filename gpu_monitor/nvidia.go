package gpu_monitor

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

var (
	searchPaths = []string{
		"/sys/devices/platform/bus@0/17000000.gpu",
		"/sys/devices/platform/17000000.ga10b",
	}

	frequencyBasePath = "/sys/class/devfreq/"
	gpuStatsKeys      = []string{"load", "cur_freq", "max_freq", "min_freq", "governor"}
)

type nvidiaGpuMonitor struct {
	logger logging.Logger
}

func newNvidiaGpuMonitor(logger logging.Logger) (GpuMonitor, error) {
	return &nvidiaGpuMonitor{logger: logger}, nil
}

func (m *nvidiaGpuMonitor) GetGPUStats(ctx context.Context) (map[string]interface{}, error) {

}

func (m *nvidiaGpuMonitor) GetGPUFrequencies(ctx context.Context) (map[string]interface{}, error) {
	dirs, err := os.ReadDir(frequencyBasePath)
	if err != nil {
		return nil, err
	}

	frequencies := make(map[string]interface{})
	for _, dir := range dirs {
		if dir.IsDir() {
			frequencyPath := filepath.Join(frequencyBasePath, dir.Name(), "cur_freq")
			data, err := utils.ReadFileWithContext(ctx, frequencyPath)
			if err != nil {
				m.logger.Errorw("failed to read frequency", "path", frequencyPath, "error", err)
				continue
			}
			frequencies[dir.Name()] = string(data)
		}
	}
	return frequencies, nil
}

func (m *nvidiaGpuMonitor) Close() {
}
