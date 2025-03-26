package diskmonitor

type ComponentConfig struct {
	Disks             []string `json:"disks"`
	IncludeIOCounters bool     `json:"include_io_counters"`
}

func (conf *ComponentConfig) Validate(path string) ([]string, error) {
	return nil, nil
}
