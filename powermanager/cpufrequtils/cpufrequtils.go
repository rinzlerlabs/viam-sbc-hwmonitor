package cpufrequtils

import (
	"os/exec"
	"strconv"
	"strings"
)

func SetGovernor(governor string) error {
	proc := exec.Command("cpufreq-set", "-g", governor)
	outputBytes, err := proc.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(outputBytes)) != "" {
		return nil
	}
	return nil
}
func SetFrequency(frequency int) error {
	proc := exec.Command("cpufreq-set", "-f", strconv.Itoa(frequency))
	outputBytes, err := proc.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(outputBytes)) != "" {
		return nil
	}
	return nil
}

func SetFrequencyLimits(minimum int, maximum int) error {
	proc := exec.Command("cpufreq-set", "-l", strconv.Itoa(minimum), strconv.Itoa(maximum))
	outputBytes, err := proc.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(outputBytes)) != "" {
		return nil
	}
	return nil
}

func GetAvailableGovernors() ([]string, error) {
	proc := exec.Command("cpufreq-info", "--governors")
	outputBytes, err := proc.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(outputBytes), " "), nil
}

func GetFrequencyLimits() (MinimumFrequency int, MaximumFrequency int, Err error) {
	proc := exec.Command("cpufreq-info", "-l")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, 0, err
	}
	frequencies := strings.Split(string(outputBytes), " ")
	min, err := strconv.Atoi(frequencies[0])
	if err != nil {
		return 0, 0, err
	}
	max, err := strconv.Atoi(frequencies[1])
	if err != nil {
		return 0, 0, err
	}
	return min, max, nil
}

func GetCurrentPolicy() (CurrentFrequency int, MaximumFrequency int, Governor string, Err error) {
	proc := exec.Command("cpufreq-info", "-p")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, 0, "", err
	}
	policy := strings.Split(string(outputBytes), " ")
	min, err := strconv.Atoi(policy[0])
	if err != nil {
		return 0, 0, "", err
	}
	max, err := strconv.Atoi(policy[1])
	if err != nil {
		return 0, 0, "", err
	}
	return min, max, strings.TrimSpace(policy[2]), nil
}

func GetCurrentFrequency() (Frequency int, Err error) {
	proc := exec.Command("cpufreq-info", "-f")
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	frequency, err := strconv.Atoi(strings.TrimSpace(string(outputBytes)))
	if err != nil {
		return 0, err
	}
	return frequency, nil
}
