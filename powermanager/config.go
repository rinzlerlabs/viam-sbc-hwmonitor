package powermanager

import (
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/raspberrypi"
)

type ComponentConfig struct {
	Jetson *jetson.PowerManagerConfig      `json:"jetson"`
	Raspi  *raspberrypi.PowerManagerConfig `json:"raspi"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	return nil, nil
}
