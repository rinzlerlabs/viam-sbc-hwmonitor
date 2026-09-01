package gpumonitor

import (
	"context"
	"fmt"
	"testing"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
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

// TestErrUnsupportedBoardWrapsSharedSentinel locks in that gpumonitor's own
// ErrUnsupportedBoard participates in the shared utils.ErrBoardNotSupported
// chain, which is what lets Reconfigure's errors.Is(err, utils.ErrBoardNotSupported)
// check match both this package's own fallback and the Jetson GPU monitor's
// "no GPU sensors found" error.
func TestErrUnsupportedBoardWrapsSharedSentinel(t *testing.T) {
	require.ErrorIs(t, ErrUnsupportedBoard, utils.ErrBoardNotSupported)
}

// TestReadingsUnsupportedBoardReportsSpecificReason regression-tests the
// gpumonitor graceful-degradation fix: when Reconfigure degrades because the
// underlying GPU monitor construction failed with an unsupported-board error
// (e.g. the Jetson GPU monitor's "no GPU sensors found"), Readings should
// surface that specific reason rather than a generic one.
func TestReadingsUnsupportedBoardReportsSpecificReason(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	specific := fmt.Errorf("no GPU sensors found: %w", utils.ErrBoardNotSupported)
	sensor := &Config{
		logger:         logger,
		gpuMonitor:     nil,
		unsupportedErr: specific,
	}
	res, err := sensor.Readings(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"error": specific.Error()}, res)
}

// TestReadingsUnsupportedBoardFallsBackWhenReasonMissing covers the
// defensive fallback for a gpuMonitor==nil Config with no unsupportedErr set
// (shouldn't happen via Reconfigure, but Readings must not panic).
func TestReadingsUnsupportedBoardFallsBackWhenReasonMissing(t *testing.T) {
	ctx := context.Background()
	logger := logging.NewTestLogger(t)
	sensor := &Config{
		logger:     logger,
		gpuMonitor: nil,
	}
	res, err := sensor.Readings(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"error": ErrUnsupportedBoard.Error()}, res)
}

type mockGpuMonitor struct{}

func (m *mockGpuMonitor) Close() error { return nil }
func (m *mockGpuMonitor) GetGPUStats(context.Context) (map[string][]sensors.GPUSensorReading, error) {
	return map[string][]sensors.GPUSensorReading{
		"gpu0": {
			{Type: sensors.GPUReadingTypeClocksGraphics, Value: 1000},
		},
	}, nil
}
