package gpumonitor

import (
	"context"
	"testing"

	"github.com/rinzlerlabs/sbcidentify/boardtype"
	. "github.com/rinzlerlabs/sbcidentify/test"
	"github.com/stretchr/testify/require"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func TestGetReadings_Jetson(t *testing.T) {
	Test().RequiresBoardType(boardtype.Jetson).ShouldSkip(t)
}

func TestGetReadings_NvidiaGPU(t *testing.T) {
	skipIfNoNvidiaDriver(t)

	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	sensor := &Config{logger: logger}
	config := resource.NewEmptyConfig(resource.NewName(API, "gpu-monitor"), Model)
	config.ConvertedAttributes = &ComponentConfig{}
	err := sensor.Reconfigure(ctx, nil, config)
	require.NoError(t, err)
	require.NotNil(t, sensor.gpuMonitor)
	stats, err := sensor.Readings(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, stats)
	logger.Infof("Readings: %#v", stats)
}
