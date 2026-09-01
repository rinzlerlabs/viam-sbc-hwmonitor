package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeOSRelease(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "os-release")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

// jetsonOrinOSRelease is the actual /etc/os-release captured from a Jetson
// AGX Orin Developer Kit running L4T's Ubuntu 24.04 ("noble") image.
const jetsonOrinOSRelease = `PRETTY_NAME="Ubuntu 24.04.4 LTS"
NAME="Ubuntu"
VERSION_ID="24.04"
VERSION="24.04.4 LTS (Noble Numbat)"
VERSION_CODENAME=noble
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=noble
LOGO=ubuntu-logo
`

func TestParseOSReleaseFromJetsonOrin(t *testing.T) {
	path := writeOSRelease(t, jetsonOrinOSRelease)
	osRelease := parseOSReleaseFrom(path)

	require.Equal(t, "ubuntu", osRelease["ID"])
	require.Equal(t, "debian", osRelease["ID_LIKE"])
	require.Equal(t, "24.04", osRelease["VERSION_ID"])
	require.Equal(t, "noble", osRelease["VERSION_CODENAME"])
}

// TestIsDebianTrixieOrNewer_JetsonOrin locks in that this AGX Orin's real
// os-release reports false: cpufrequtils remains the correct package there
// (see cpufrequtils.selectBackend), even though ID_LIKE contains "debian".
func TestIsDebianTrixieOrNewer_JetsonOrin(t *testing.T) {
	osRelease := parseOSReleaseFrom(writeOSRelease(t, jetsonOrinOSRelease))
	require.False(t, isDebianTrixieOrNewer(osRelease))
}

func TestIsDebianTrixieOrNewer_DebianTrixie(t *testing.T) {
	osRelease := parseOSReleaseFrom(writeOSRelease(t, `PRETTY_NAME="Debian GNU/Linux 13 (trixie)"
NAME="Debian GNU/Linux"
VERSION_ID="13"
VERSION="13 (trixie)"
VERSION_CODENAME=trixie
ID=debian
`))
	require.True(t, isDebianTrixieOrNewer(osRelease))
}

func TestIsDebianTrixieOrNewer_RaspberryPiOSBookworm(t *testing.T) {
	osRelease := parseOSReleaseFrom(writeOSRelease(t, `PRETTY_NAME="Raspbian GNU/Linux 12 (bookworm)"
NAME="Raspbian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=raspbian
ID_LIKE=debian
`))
	require.False(t, isDebianTrixieOrNewer(osRelease))
}

func TestIsDebianTrixieOrNewer_DebianSid(t *testing.T) {
	osRelease := parseOSReleaseFrom(writeOSRelease(t, `PRETTY_NAME="Debian GNU/Linux bookworm/sid"
NAME="Debian GNU/Linux"
ID=debian
VERSION_CODENAME=sid
`))
	require.True(t, isDebianTrixieOrNewer(osRelease))
}

func TestIsDebianTrixieOrNewer_MissingFile(t *testing.T) {
	osRelease := parseOSReleaseFrom(filepath.Join(t.TempDir(), "does-not-exist"))
	require.False(t, isDebianTrixieOrNewer(osRelease))
}
