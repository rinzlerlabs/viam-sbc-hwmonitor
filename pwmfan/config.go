package pwmfan

import (
	"errors"

	sbc "github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
)

type CloudConfig struct {
	FanPin           string             `json:"fan_pin"`
	TemperatureTable map[string]float64 `json:"temperature_table"`
	BoardName        string             `json:"board_name"`
	UseInternalFan   bool               `json:"use_internal_fan"`
}

func (conf *CloudConfig) Validate(path string) ([]string, []string, error) {
	if conf.UseInternalFan {
		if !sbc.IsBoardType(boardtype.RaspberryPi5) {
			return nil, nil, errors.New("use_internal_fan is only supported on Raspberry Pi 5")
		}
	} else {
		if conf.FanPin == "" {
			return nil, nil, errors.New("fan_pin is required")
		}

		if conf.BoardName == "" {
			return nil, nil, errors.New("board_name is required")
		}
	}

	if conf.TemperatureTable == nil {
		return nil, nil, errors.New("temperature_table is required")
	}

	if len(conf.TemperatureTable) == 0 {
		return nil, nil, errors.New("temperature_table must have at least one entry")
	}

	// We need to make sure fan speed is between 0 and 100
	// We don't need to check temperature because very low temperatures are possible on robots exposed to the elements
	for _, speed := range conf.TemperatureTable {
		if speed < 0 || speed > 100 {
			return nil, nil, errors.New("temperature_table must have speeds between 0 and 100")
		}
	}

	// Declare the board as a required dependency so it is constructed first and
	// available in deps. If it does not exist, the framework reports a clear
	// missing-dependency error instead of failing later when building the fan.
	var deps []string
	if !conf.UseInternalFan {
		deps = append(deps, conf.BoardName)
	}
	return deps, nil, nil
}
