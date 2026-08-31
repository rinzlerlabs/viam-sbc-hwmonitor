package cpufrequtils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// cpufreqBasePath is the sysfs cpufreq interface for the first CPU. Reading
// these kernel-provided files avoids parsing the human-formatted output of
// cpufreq-info/cpupower, which differs between tool versions.
const cpufreqBasePath = "/sys/devices/system/cpu/cpu0/cpufreq"

func readCPUFreqString(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(cpufreqBasePath, name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func readCPUFreqInt(name string) (int, error) {
	value, err := readCPUFreqString(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// ApplyPolicy sets the CPU frequency policy in a single call. It prefers
// cpupower (from linux-cpupower), the maintained replacement for the obsolete
// cpufrequtils that was removed in Debian Trixie, and falls back to cpufreq-set
// on older systems where only cpufrequtils is available. The combined command
// output is returned for logging.
func ApplyPolicy(governor string, frequency, minimum, maximum int) (string, error) {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("cpupower"); err == nil {
		args := []string{"frequency-set"}
		if governor != "" {
			args = append(args, "-g", governor)
		}
		if frequency != 0 {
			args = append(args, "-f", strconv.Itoa(frequency))
		}
		if minimum != 0 {
			args = append(args, "-d", strconv.Itoa(minimum))
		}
		if maximum != 0 {
			args = append(args, "-u", strconv.Itoa(maximum))
		}
		cmd = exec.Command("cpupower", args...)
	} else {
		args := make([]string, 0)
		if governor != "" {
			args = append(args, "--governor", governor)
		}
		if frequency != 0 {
			args = append(args, "--freq", strconv.Itoa(frequency))
		}
		if minimum != 0 {
			args = append(args, "--min", strconv.Itoa(minimum))
		}
		if maximum != 0 {
			args = append(args, "--max", strconv.Itoa(maximum))
		}
		cmd = exec.Command("cpufreq-set", args...)
	}

	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func SetGovernor(governor string) error {
	_, err := ApplyPolicy(governor, 0, 0, 0)
	return err
}

func SetFrequency(frequency int) error {
	_, err := ApplyPolicy("", frequency, 0, 0)
	return err
}

func SetFrequencyLimits(minimum int, maximum int) error {
	_, err := ApplyPolicy("", 0, minimum, maximum)
	return err
}

func GetAvailableGovernors() ([]string, error) {
	governors, err := readCPUFreqString("scaling_available_governors")
	if err != nil {
		return nil, err
	}
	return strings.Fields(governors), nil
}

func GetFrequencyLimits() (MinimumFrequency int, MaximumFrequency int, Err error) {
	min, err := readCPUFreqInt("cpuinfo_min_freq")
	if err != nil {
		return 0, 0, err
	}
	max, err := readCPUFreqInt("cpuinfo_max_freq")
	if err != nil {
		return 0, 0, err
	}
	return min, max, nil
}

func GetCurrentPolicy() (CurrentFrequency int, MaximumFrequency int, Governor string, Err error) {
	min, err := readCPUFreqInt("scaling_min_freq")
	if err != nil {
		return 0, 0, "", err
	}
	max, err := readCPUFreqInt("scaling_max_freq")
	if err != nil {
		return 0, 0, "", err
	}
	governor, err := readCPUFreqString("scaling_governor")
	if err != nil {
		return 0, 0, "", err
	}
	return min, max, governor, nil
}

func GetCurrentFrequency() (Frequency int, Err error) {
	return readCPUFreqInt("scaling_cur_freq")
}
