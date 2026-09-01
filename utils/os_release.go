package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const osReleasePath = "/etc/os-release"

// parseOSRelease reads /etc/os-release into a key/value map. It returns an empty
// map when the file is missing or unreadable (e.g. on non-Linux hosts).
func parseOSRelease() map[string]string {
	return parseOSReleaseFrom(osReleasePath)
}

// parseOSReleaseFrom does the work of parseOSRelease against an injectable
// path, so tests can point it at a fake os-release file.
func parseOSReleaseFrom(path string) map[string]string {
	result := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		result[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return result
}

// IsDebianTrixieOrNewer reports whether the host is running Debian 13 ("trixie")
// or newer, including Raspberry Pi OS / Raspbian derivatives. The legacy
// cpufrequtils package was removed in Trixie, so callers should install
// linux-cpupower (the maintained replacement) instead. Non-Debian-family
// distros (including Ubuntu, e.g. Jetson/L4T) always report false here:
// cpufrequtils remains the right choice for them too, since their
// kernel-version-pinned cpupower packages generally aren't installable for a
// vendor kernel build (confirmed on a Jetson AGX Orin running Ubuntu 24.04).
func IsDebianTrixieOrNewer() bool {
	return isDebianTrixieOrNewer(parseOSRelease())
}

// isDebianTrixieOrNewer does the work of IsDebianTrixieOrNewer against an
// already-parsed os-release map, so tests can exercise it directly.
func isDebianTrixieOrNewer(osRelease map[string]string) bool {
	id := strings.ToLower(osRelease["ID"])
	idLike := strings.ToLower(osRelease["ID_LIKE"])
	if id != "debian" && id != "raspbian" && !strings.Contains(idLike, "debian") {
		return false
	}

	// Stable releases carry a numeric VERSION_ID; Trixie is 13.
	if v, err := strconv.Atoi(osRelease["VERSION_ID"]); err == nil {
		return v >= 13
	}

	// testing/unstable have no VERSION_ID, so fall back to the codename.
	switch strings.ToLower(osRelease["VERSION_CODENAME"]) {
	case "trixie", "forky", "duke", "sid":
		return true
	}
	return false
}
