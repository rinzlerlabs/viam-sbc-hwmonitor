package jetson

import "go.viam.com/rdk/logging"

type PowerManagerConfig struct {
	PowerMode int    `json:"power_mode"`
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

type jetsonPowerManager struct {
	config *PowerManagerConfig
	logger logging.Logger
}
