package powermanager

import (
	"errors"

	"go.viam.com/rdk/logging"
)

func newPowerManager(_ *ComponentConfig, logger logging.Logger) (powerManager PowerManager, err error) {
	logger.Errorf("Power manager not implemented on windows")
	return nil, errors.New("not implemented on windows")
}
