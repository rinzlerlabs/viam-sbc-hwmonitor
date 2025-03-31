package genericlinux

import "go.viam.com/rdk/logging"

type LinuxConfig struct {
}

type linuxPowerManager struct {
	config *LinuxConfig
	logger logging.Logger
}
