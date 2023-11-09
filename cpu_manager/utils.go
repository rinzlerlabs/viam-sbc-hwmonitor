package cpu_manager

import (
	"os/exec"
	"strconv"
	"strings"
)

func getAvailableGovernors() ([]string, error) {
	proc := exec.Command("cpufreq-info", "--governors")
	outputBytes, err := proc.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(outputBytes), " "), nil
}

func getFrequencyLimits() (minimum int, maximum int, err error) {
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

func getCurrentPolicy() (MinimumFrequency int, MaximumFrequency int, Governor string, Err error) {
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
	return min, max, policy[2], nil
}

func getCurrentFrequency() (int, error) {
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
