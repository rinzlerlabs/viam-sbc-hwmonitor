package gpumonitor

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/gpumonitor/gpusensor"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
)

func TestReadings(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	gpuMonitor := &mockGpuMonitor{}
	require.NotNil(t, gpuMonitor)
	sensor := &Config{
		logger:     logger,
		gpuMonitor: gpuMonitor,
	}
	res, err := sensor.Readings(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 1)
	logger.Infof("Readings: %#v", res)
}

type mockGpuMonitor struct{}

func (m *mockGpuMonitor) Close() error { return nil }
func (m *mockGpuMonitor) GetGPUStats(context.Context) (map[string][]gpusensor.GPUSensorReading, error) {
	return map[string][]gpusensor.GPUSensorReading{
		"gpu0": {
			{Type: gpusensor.GPUReadingTypeClocksGraphics, Value: 1000},
		},
	}, nil
}
