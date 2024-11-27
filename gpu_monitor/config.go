package gpu_monitor

type ComponentConfig struct {
	SleepTimeMs int `json:"sleep_time_ms"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	return nil, nil
}
