package cpufrequtils

import (
	"context"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

func Install() error {
	return utils.ErrPlatformNotSupported
}

func ApplyPolicy(governor string, frequency, minimum, maximum int) (string, error) {
	return "", utils.ErrPlatformNotSupported
}

func SetGovernor(governor string) error {
	return utils.ErrPlatformNotSupported
}
func SetFrequency(frequency int) error {
	return utils.ErrPlatformNotSupported
}

func SetFrequencyLimits(minimum int, maximum int) error {
	return utils.ErrPlatformNotSupported
}

func GetAvailableGovernors(ctx context.Context) ([]string, error) {
	return nil, utils.ErrPlatformNotSupported
}

func GetGovernor(ctx context.Context) (string, error) {
	return "", utils.ErrPlatformNotSupported
}

func GetHardwareMinimumFrequency(ctx context.Context) (int, error) {
	return -1, utils.ErrPlatformNotSupported
}

func GetHardwareMaximumFrequency(ctx context.Context) (int, error) {
	return -1, utils.ErrPlatformNotSupported
}

func GetPolicyMinimumFrequency(ctx context.Context) (int, error) {
	return -1, utils.ErrPlatformNotSupported
}

func GetPolicyMaximumFrequency(ctx context.Context) (int, error) {
	return -1, utils.ErrPlatformNotSupported
}

func GetCurrentFrequency(ctx context.Context) (int, error) {
	return -1, utils.ErrPlatformNotSupported
}
