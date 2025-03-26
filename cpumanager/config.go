//go:build linux
// +build linux

package cpumanager

import (
	"errors"
	"slices"
	"strconv"
)

type ComponentConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	if conf.Governor != "" {
		availableGovernors, err := getAvailableGovernors()
		if err != nil {
			return nil, err
		}
		if !slices.Contains(availableGovernors, conf.Governor) {
			return nil, errors.New("unknown governor")
		}
	}

	if conf.Frequency != 0 {
		min, max, err := getFrequencyLimits()
		if err != nil {
			return nil, err
		}
		if conf.Frequency > max || conf.Frequency < min {
			return nil, errors.New("frequency out of range. valid range: " + strconv.Itoa(min) + " - " + strconv.Itoa(max))
		}
	}
	return nil, nil
}
