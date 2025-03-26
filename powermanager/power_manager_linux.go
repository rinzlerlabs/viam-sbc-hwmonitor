package powermanager

import (
	"errors"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-raspi-sensors/powermanager/jetson"
	"github.com/rinzlerlabs/viam-raspi-sensors/powermanager/raspberrypi"
	"go.viam.com/rdk/logging"
)

var (
	ErrBoardMismatch    = errors.New("board does not match configuration")
	ErrNoConfigForBoard = errors.New("no configuration for board")
)

func newPowerManager(config *ComponentConfig, logger logging.Logger) (powerManager PowerManager, err error) {
	if sbcidentify.IsJetson() {
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
