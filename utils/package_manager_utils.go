package utils

import (
	"errors"
	"os/exec"
	"strings"
)

var ErrNoPackageManagerFound = errors.New("no package manager found")

func InstallPackage(packageName string) error {
	if isAptInstalled() {
		proc := exec.Command("apt", "install", packageName, "-y")
		_, err := proc.Output()
		if err != nil {
			return err
		}
		return nil
	}

	if isYumInstalled() {
		proc := exec.Command("yum", "install", packageName, "-y")
		_, err := proc.Output()
		if err != nil {
			return err
		}

		return nil
	}
	return ErrNoPackageManagerFound
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
