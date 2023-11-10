package pwm_fan

import "errors"

type CloudConfig struct {
	FanPin           string             `json:"fan_pin"`
	TemperatureTable map[string]float64 `json:"temperature_table"`
	BoardName        string             `json:"board_name"`
}

func (conf *CloudConfig) Validate(path string) ([]string, error) {
	if conf.FanPin == "" {
		return nil, errors.New("fan_pin is required")
	}

	if conf.TemperatureTable == nil {
		return nil, errors.New("temperature_table is required")
	}

	if len(conf.TemperatureTable) == 0 {
		return nil, errors.New("temperature_table must have at least one entry")
	}

	if conf.BoardName == "" {
		return nil, errors.New("board_name is required")
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
