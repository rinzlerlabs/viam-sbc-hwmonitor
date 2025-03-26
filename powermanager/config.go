package powermanager

import (
	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/jetson"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/raspberrypi"
)

type ComponentConfig struct {
	Jetson *jetson.JetsonConfig     `json:"jetson"`
	Raspi  *raspberrypi.RaspiConfig `json:"raspi"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	sbcidentify.IsBoardType(boardtype.JetsonOrinNX8GB)
	return nil, nil
}
