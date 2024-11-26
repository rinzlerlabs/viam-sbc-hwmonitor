package pwm_fan

import (
	"errors"

	sbc "github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/raspberrypi"
)

type CloudConfig struct {
	FanPin           string             `json:"fan_pin"`
	TemperatureTable map[string]float64 `json:"temperature_table"`
	BoardName        string             `json:"board_name"`
	UseInternalFan   bool               `json:"use_internal_fan"`
}

func (conf *CloudConfig) Validate(path string) ([]string, error) {
	if conf.UseInternalFan {
		if !sbc.IsBoardType(raspberrypi.RaspberryPi5) {
			return nil, errors.New("use_internal_fan is only supported on Raspberry Pi 5")
		}
	} else {
		if conf.FanPin == "" {
			return nil, errors.New("fan_pin is required")
		}

		if conf.BoardName == "" {
			return nil, errors.New("board_name is required")
		}
	}

	if conf.TemperatureTable == nil {
		return nil, errors.New("temperature_table is required")
	}

	if len(conf.TemperatureTable) == 0 {
		return nil, errors.New("temperature_table must have at least one entry")
	}

	// We need to make sure fan speed is between 0 and 100
	// We don't need to check temperature because very low temperatures are possible on robots exposed to the elements
	for _, speed := range conf.TemperatureTable {
		if speed < 0 || speed > 100 {
			return nil, errors.New("temperature_table must have speeds between 0 and 100")
		}
	}

	return nil, nil
}
