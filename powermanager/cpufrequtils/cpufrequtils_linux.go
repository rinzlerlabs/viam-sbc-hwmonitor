package cpufrequtils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

// cpufreqBasePath is the sysfs cpufreq interface for the first CPU. Reading
// these kernel-provided files avoids parsing the human-formatted output of
// cpufreq-info/cpupower, which differs between tool versions.
const cpufreqBasePath = "/sys/devices/system/cpu/cpu0/cpufreq"

// backend names the CLI tool used to apply a governor/frequency policy, and
// the package(s) that provide it.
type backend struct {
	binary   string
	packages []string
}

// selectBackend decides which cpufreq tool to install and invoke for the
// host distro. Debian/Raspberry Pi OS Trixie and newer dropped the legacy
// cpufrequtils package in favor of linux-cpupower (the maintained cpupower
// frontend). cpufrequtils remains the right choice everywhere else,
// including Ubuntu-based Jetson/L4T images: Ubuntu's linux-tools-<version>
// packages only cover the generic Ubuntu kernel build, not the vendor's
// custom Tegra kernel, so cpupower isn't installable there — confirmed on a
// Jetson AGX Orin (Ubuntu 24.04 "noble"), where linux-cpupower has no
// installation candidate at all and linux-tools-generic only ships tools for
// the generic 6.8.0 kernel, not the running 6.8.12-tegra build.
func selectBackend(isDebianTrixieOrNewer bool) backend {
	if isDebianTrixieOrNewer {
		return backend{binary: "cpupower", packages: []string{"linux-cpupower", "libcpupower1"}}
	}
	return backend{binary: "cpufreq-set", packages: []string{"cpufrequtils"}}
}

// Install ensures the CPU frequency tooling ApplyPolicy needs is present for
// the host distro.
func Install() error {
	b := selectBackend(utils.IsDebianTrixieOrNewer())
	if err := utils.InstallPackage(b.packages...); err != nil {
		return fmt.Errorf("error installing %s: %w", strings.Join(b.packages, ", "), err)
	}
	return nil
}

func readCPUFreqString(basePath, name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(basePath, name))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func readCPUFreqInt(basePath, name string) (int, error) {
	value, err := readCPUFreqString(basePath, name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// buildCpupowerArgs and buildCpufreqSetArgs translate a governor/frequency
// policy into the CLI arguments for `cpupower frequency-set` and
// `cpufreq-set` respectively. Split out from ApplyPolicy so the argument
// mapping can be tested without executing either tool.
func buildCpupowerArgs(governor string, frequency, minimum, maximum int) []string {
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
	return args
}

func buildCpufreqSetArgs(governor string, frequency, minimum, maximum int) []string {
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
	return args
}

// ApplyPolicy sets the CPU frequency policy in a single call, using whichever
// tool Install would have installed for this distro (see selectBackend). The
// combined command output is returned for logging.
func ApplyPolicy(governor string, frequency, minimum, maximum int) (string, error) {
	b := selectBackend(utils.IsDebianTrixieOrNewer())
	var cmd *exec.Cmd
	if b.binary == "cpupower" {
		cmd = exec.Command("cpupower", buildCpupowerArgs(governor, frequency, minimum, maximum)...)
	} else {
		cmd = exec.Command("cpufreq-set", buildCpufreqSetArgs(governor, frequency, minimum, maximum)...)
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
	governors, err := readCPUFreqString(cpufreqBasePath, "scaling_available_governors")
	if err != nil {
		return nil, err
	}
	return strings.Fields(governors), nil
}

func GetGovernor() (string, error) {
	return readCPUFreqString(cpufreqBasePath, "scaling_governor")
}

// GetHardwareMinimumFrequency and GetHardwareMaximumFrequency report the
// frequency range the CPU is capable of, regardless of the current policy.
func GetHardwareMinimumFrequency() (int, error) {
	return readCPUFreqInt(cpufreqBasePath, "cpuinfo_min_freq")
}

func GetHardwareMaximumFrequency() (int, error) {
	return readCPUFreqInt(cpufreqBasePath, "cpuinfo_max_freq")
}

// GetPolicyMinimumFrequency and GetPolicyMaximumFrequency report the bounds
// of the currently applied scaling policy, which may be narrower than the
// hardware range.
func GetPolicyMinimumFrequency() (int, error) {
	return readCPUFreqInt(cpufreqBasePath, "scaling_min_freq")
}

func GetPolicyMaximumFrequency() (int, error) {
	return readCPUFreqInt(cpufreqBasePath, "scaling_max_freq")
}

func GetCurrentFrequency() (Frequency int, Err error) {
	return readCPUFreqInt(cpufreqBasePath, "scaling_cur_freq")
}
