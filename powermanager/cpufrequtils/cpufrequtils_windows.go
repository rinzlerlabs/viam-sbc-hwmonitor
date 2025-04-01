package cpufrequtils

import "github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"

func SetGovernor(governor string) error {
	return utils.ErrPlatformNotSupported
}
func SetFrequency(frequency int) error {
	return utils.ErrPlatformNotSupported
}

func SetFrequencyLimits(minimum int, maximum int) error {
	return utils.ErrPlatformNotSupported
}

func GetAvailableGovernors() ([]string, error) {
	return nil, utils.ErrPlatformNotSupported
}

func GetFrequencyLimits() (MinimumFrequency int, MaximumFrequency int, Err error) {
	return -1, -1, utils.ErrPlatformNotSupported
}

func GetCurrentPolicy() (CurrentFrequency int, MaximumFrequency int, Governor string, Err error) {
	return -1, -1, "", utils.ErrPlatformNotSupported
}

func GetCurrentFrequency() (Frequency int, Err error) {
	return -1, utils.ErrPlatformNotSupported
}
