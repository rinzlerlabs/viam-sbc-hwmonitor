package wifi_monitor

import (
	"errors"
	"runtime"
)

type ComponentConfig struct {
	Adapter string `json:"adapter"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	if conf.Adapter == "" {
		return nil, errors.New("adapter is required")
	}
	if runtime.GOOS != "linux" {
		return nil, errors.New("only linux is supported")
	}
	return nil, nil
}
