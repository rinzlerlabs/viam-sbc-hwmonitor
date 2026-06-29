package powermanager

import (
	"errors"
	"strings"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/linux/raspberrypi"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/logging"
)

var (
	ErrBoardMismatch    = errors.New("board does not match configuration")
	ErrNoConfigForBoard = errors.New("no configuration for board")
)

func newPowerManager(config *ComponentConfig, logger logging.Logger) (PowerManager, error) {
	// cpufrequtils was removed in Debian Trixie; install its maintained
	// replacement (linux-cpupower, plus its libcpupower1 runtime library) there
	// and keep cpufrequtils on older systems.
	cpuFreqPackages := []string{"cpufrequtils"}
	if utils.IsDebianTrixieOrNewer() {
		cpuFreqPackages = []string{"linux-cpupower", "libcpupower1"}
	}
	if err := utils.InstallPackage(cpuFreqPackages...); err != nil {
		return nil, errors.Join(err, errors.New("error installing "+strings.Join(cpuFreqPackages, ", ")))
	}

	// Detect a Tegra GPU directly in addition to board identification, since
	// sbcidentify can fail to identify some Jetsons (e.g. Orin).
	if sbcidentify.IsJetson() || jetson.HasJetsonGpu() {
		if config.Jetson == nil {
			return nil, ErrNoConfigForBoard
		}
		return jetson.NewPowerManager(config.Jetson, logger)
	} else if sbcidentify.IsRaspberryPi() {
		if config.Raspi == nil {
			return nil, ErrNoConfigForBoard
		}
		return raspberrypi.NewPowerManager(config.Raspi, logger)
	}

	return nil, errors.New("unknown power mode")
}
