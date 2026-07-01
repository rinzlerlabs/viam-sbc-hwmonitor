package wifimonitor

import (
	"errors"
	"runtime"
)

type ComponentConfig struct {
	Adapter string `json:"adapter"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, []string, error) {
	if conf.Adapter == "" {
		return nil, nil, errors.New("adapter is required")
	}
	if runtime.GOOS != "linux" {
		return nil, nil, errors.New("only linux is supported")
	}
	return nil, nil, nil
}
