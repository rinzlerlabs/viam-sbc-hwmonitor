//go:build linux
// +build linux

package cpumanager

import (
	"errors"
	"slices"
	"strconv"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/cpufrequtils"
)

type ComponentConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, []string, error) {
	if conf.Governor != "" {
		availableGovernors, err := cpufrequtils.GetAvailableGovernors()
		if err != nil {
			return nil, nil, err
		}
		if !slices.Contains(availableGovernors, conf.Governor) {
			return nil, nil, errors.New("unknown governor")
		}
	}

	if conf.Frequency != 0 {
		min, max, err := cpufrequtils.GetFrequencyLimits()
		if err != nil {
			return nil, nil, err
		}
		if conf.Frequency > max || conf.Frequency < min {
			return nil, nil, errors.New("frequency out of range. valid range: " + strconv.Itoa(min) + " - " + strconv.Itoa(max))
		}
	}
	return nil, nil, nil
}
