//go:build linux
// +build linux

package cpumanager

import (
	"context"
	"errors"
	"slices"
	"strconv"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/cpufrequtils"
)

type ComponentConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, []string, error) {
	// This component only supports the Raspberry Pi: Reconfigure (sensor.go)
	// marks itself unsupported and no-ops on any other board. Skip cpufreq
	// validation here too, so Validate doesn't hard-fail on boards with no
	// cpufreq driver, or silently pass on boards that happen to expose
	// generic cpufreq sysfs (e.g. Jetson) only to have the config ignored at
	// runtime anyway.
	if !sbcidentify.IsRaspberryPi() {
		return nil, nil, nil
	}

	// Validate's signature (fixed by the resource.Validator interface) carries
	// no context, so there's no caller-provided deadline/cancellation to
	// propagate here — background is the honest choice, not a shortcut.
	ctx := context.Background()

	if conf.Governor != "" {
		availableGovernors, err := cpufrequtils.GetAvailableGovernors(ctx)
		if err != nil {
			return nil, nil, err
		}
		if !slices.Contains(availableGovernors, conf.Governor) {
			return nil, nil, errors.New("unknown governor")
		}
	}

	if conf.Frequency != 0 {
		min, err := cpufrequtils.GetHardwareMinimumFrequency(ctx)
		if err != nil {
			return nil, nil, err
		}
		max, err := cpufrequtils.GetHardwareMaximumFrequency(ctx)
		if err != nil {
			return nil, nil, err
		}
		if conf.Frequency > max || conf.Frequency < min {
			return nil, nil, errors.New("frequency out of range. valid range: " + strconv.Itoa(min) + " - " + strconv.Itoa(max))
		}
	}
	return nil, nil, nil
}
