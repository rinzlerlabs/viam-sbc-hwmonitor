package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrNoPackageManagerFound = errors.New("no package manager found")

func InstallPackage(packageNames ...string) error {
	if len(packageNames) == 0 {
		return nil
	}

	args := append([]string{"install"}, packageNames...)
	args = append(args, "-y")

	switch {
	case isAptInstalled():
		return runInstall("apt", args...)
	case isYumInstalled():
		return runInstall("yum", args...)
	default:
		return ErrNoPackageManagerFound
	}
}

// runInstall executes a package-manager command, capturing stderr so a failure
// surfaces the underlying message (e.g. apt's "Permission denied" or "Unable to
// locate package") instead of a bare "exit status 100".
func runInstall(name string, args ...string) error {
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func isAptInstalled() bool {
	proc := exec.Command("apt", "-v")
	outputBytes, err := proc.Output()
	if err != nil {
		return false
	}

	if strings.Contains(string(outputBytes), "apt") {
		return true
	}
	return false
}

func isYumInstalled() bool {
	proc := exec.Command("yum", "-v")
	outputBytes, err := proc.Output()
	if err != nil {
		return false
	}

	if strings.Contains(string(outputBytes), "yum") {
		return true
	}
	return false
}
