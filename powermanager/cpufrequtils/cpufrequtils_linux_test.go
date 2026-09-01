package cpufrequtils

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeCPUFreqFile writes a single cpufreq sysfs file (ex: scaling_governor)
// under a fake cpuN/cpufreq directory.
func writeCPUFreqFile(t *testing.T, root, name, value string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(value+"\n"), 0644))
}

// fakeOrinCPUFreq builds a fake cpu0/cpufreq directory populated with values
// captured from a real Jetson AGX Orin Developer Kit (L4T, NV Power Mode
// MAXN): 12 cores in 3 policy groups (cpu0-3, cpu4-7, cpu8-11), governor
// schedutil, scaling range narrowed by nvpmodel to 729600-2201600 kHz within
// the hardware's 115200-2201600 kHz range.
func fakeOrinCPUFreq(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeCPUFreqFile(t, root, "scaling_available_governors", "ondemand userspace performance schedutil")
	writeCPUFreqFile(t, root, "scaling_governor", "schedutil")
	writeCPUFreqFile(t, root, "scaling_min_freq", "729600")
	writeCPUFreqFile(t, root, "scaling_max_freq", "2201600")
	writeCPUFreqFile(t, root, "scaling_cur_freq", "2201600")
	writeCPUFreqFile(t, root, "cpuinfo_min_freq", "115200")
	writeCPUFreqFile(t, root, "cpuinfo_max_freq", "2201600")
	return root
}

func TestReadCPUFreqAgainstRealOrinValues(t *testing.T) {
	ctx := context.Background()
	root := fakeOrinCPUFreq(t)

	governors, err := readCPUFreqString(ctx, root, "scaling_available_governors")
	require.NoError(t, err)
	require.Equal(t, "ondemand userspace performance schedutil", governors)

	governor, err := readCPUFreqString(ctx, root, "scaling_governor")
	require.NoError(t, err)
	require.Equal(t, "schedutil", governor)

	policyMin, err := readCPUFreqInt(ctx, root, "scaling_min_freq")
	require.NoError(t, err)
	require.Equal(t, 729600, policyMin)

	policyMax, err := readCPUFreqInt(ctx, root, "scaling_max_freq")
	require.NoError(t, err)
	require.Equal(t, 2201600, policyMax)

	current, err := readCPUFreqInt(ctx, root, "scaling_cur_freq")
	require.NoError(t, err)
	require.Equal(t, 2201600, current)

	hwMin, err := readCPUFreqInt(ctx, root, "cpuinfo_min_freq")
	require.NoError(t, err)
	require.Equal(t, 115200, hwMin)

	hwMax, err := readCPUFreqInt(ctx, root, "cpuinfo_max_freq")
	require.NoError(t, err)
	require.Equal(t, 2201600, hwMax)
}

func TestReadCPUFreqStringMissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := readCPUFreqString(context.Background(), root, "scaling_governor")
	require.Error(t, err)
}

// TestReadCPUFreqStringRespectsCancellation locks in that the shared
// context-aware utils.ReadFileWithContext helper is actually wired through:
// a canceled context should abort the read.
func TestReadCPUFreqStringRespectsCancellation(t *testing.T) {
	root := fakeOrinCPUFreq(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := readCPUFreqString(ctx, root, "scaling_governor")
	require.ErrorIs(t, err, context.Canceled)
}

// TestBuildCpupowerArgsForJetsonGovernorAndLimits locks in the CLI arguments
// generated for a governor+min/max policy (the shape used by the new Jetson
// ApplyPowerMode cpufreq call), using this AGX Orin's real hardware/scaling
// frequency values as the fixture.
func TestBuildCpupowerArgsForJetsonGovernorAndLimits(t *testing.T) {
	args := buildCpupowerArgs("performance", 0, 115200, 2201600)
	require.Equal(t, []string{"frequency-set", "-g", "performance", "-d", "115200", "-u", "2201600"}, args)
}

func TestBuildCpupowerArgsGovernorOnly(t *testing.T) {
	args := buildCpupowerArgs("schedutil", 0, 0, 0)
	require.Equal(t, []string{"frequency-set", "-g", "schedutil"}, args)
}

func TestBuildCpupowerArgsFrequencyOnly(t *testing.T) {
	args := buildCpupowerArgs("", 2201600, 0, 0)
	require.Equal(t, []string{"frequency-set", "-f", "2201600"}, args)
}

func TestBuildCpufreqSetArgsForJetsonGovernorAndLimits(t *testing.T) {
	args := buildCpufreqSetArgs("performance", 0, 115200, 2201600)
	require.Equal(t, []string{"--governor", "performance", "--min", "115200", "--max", "2201600"}, args)
}

func TestBuildCpufreqSetArgsAllUnset(t *testing.T) {
	args := buildCpufreqSetArgs("", 0, 0, 0)
	require.Empty(t, args)
}

// TestSelectBackendOnDebianTrixieOrNewer covers Debian/Raspberry Pi OS
// Trixie+, where cpufrequtils was dropped in favor of linux-cpupower.
func TestSelectBackendOnDebianTrixieOrNewer(t *testing.T) {
	b := selectBackend(true)
	require.Equal(t, "cpupower", b.binary)
	require.Equal(t, []string{"linux-cpupower", "libcpupower1"}, b.packages)
}

// TestSelectBackendOnJetsonOrin covers this AGX Orin's actual distro
// (Ubuntu 24.04 "noble"), which utils.IsDebianTrixieOrNewer reports false
// for. Verified live on the device: `apt-cache policy linux-cpupower` has no
// installation candidate at all, and `linux-tools-generic` only provides
// tools for Ubuntu's generic 6.8.0 kernel build, not the running
// 6.8.12-tegra kernel — cpufrequtils is the package that's actually
// installable and usable here (`apt-cache policy cpufrequtils` shows
// 008-2build2 available in noble/universe).
func TestSelectBackendOnJetsonOrin(t *testing.T) {
	b := selectBackend(false)
	require.Equal(t, "cpufreq-set", b.binary)
	require.Equal(t, []string{"cpufrequtils"}, b.packages)
}
