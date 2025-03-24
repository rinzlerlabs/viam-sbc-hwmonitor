//go:build !full_nvidia_support
// +build !full_nvidia_support

package gpu_monitor

import (
	"context"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"go.viam.com/rdk/logging"
)

func newGpuMonitor(ctx context.Context, logger logging.Logger) (gpuMonitor, error) {
	if sbcidentify.IsBoardType(boardtype.NVIDIA) {
		return newJetsonGpuMonitor(ctx, logger)
	}
	return nil, ErrUnsupportedBoard
}
