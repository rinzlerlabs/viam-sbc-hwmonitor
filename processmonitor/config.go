package processmonitor

import (
	"errors"
	"fmt"
	"os"
)

type ComponentConfig struct {
	Name                 string `json:"name"`
	ExecutablePath       string `json:"executable_path"`
	IncludeEnv           bool   `json:"include_env"`
	IncludeCmdline       bool   `json:"include_cmdline"`
	IncludeCwd           bool   `json:"include_cwd"`
	IncludeOpenFileCount bool   `json:"include_open_file_count"`
	IncludeMemInfo       bool   `json:"include_mem_info"`
	IncludeOpenFiles     bool   `json:"include_open_files"`
	IncludeUlimits       bool   `json:"include_ulimits"`
	IncludeNetStats      bool   `json:"include_net_stats"`
	SleepTimeMs          int    `json:"sleep_time_ms"`       // Sleep time in milliseconds between process checks
	DisablePIDCaching    bool   `json:"disable_pid_caching"` // Enable caching of PID to avoid repeated lookups
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	if conf.ExecutablePath == "" && conf.Name == "" {
		return nil, errors.New("executable_path or name is required")
	}
	if conf.ExecutablePath != "" && conf.Name != "" {
		return nil, errors.New("only one of executable_path or name is allowed")
	}
	if conf.ExecutablePath != "" {
		if _, err := os.Stat(conf.ExecutablePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("executable_path does not exist: %s", conf.ExecutablePath)
		}
	}
	return nil, nil
}
