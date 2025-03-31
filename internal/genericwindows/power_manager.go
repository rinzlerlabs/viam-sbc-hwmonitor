package genericwindows

import "go.viam.com/rdk/logging"

type WindowsConfig struct {
}

type windowsPowerManager struct {
	config *WindowsConfig
	logger logging.Logger
}
