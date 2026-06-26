package cpumonitor

type ComponentConfig struct {
	SleepTimeMs int `json:"sleep_time_ms"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, []string, error) {
	return nil, nil, nil
}
