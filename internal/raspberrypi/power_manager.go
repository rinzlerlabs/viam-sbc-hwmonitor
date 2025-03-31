package raspberrypi

import "go.viam.com/rdk/logging"

type PowerManagerConfig struct {
	Governor  string `json:"governor"`
	Frequency int    `json:"frequency"`
	Minimum   int    `json:"minimum"`
	Maximum   int    `json:"maximum"`
}

type raspiPowerManager struct {
	config *PowerManagerConfig
	logger logging.Logger
}
