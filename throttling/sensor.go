package throttling

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var Model = resource.NewModel("viam-soleng", "raspi", "throttling")
var PrettyName = "Raspberry Pi Throttling Sensor"
var Description = "A sensor that reports the throttling state of the Raspberry Pi."
var Version = "v0.0.1"

type Config struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *ComponentConfig]{Constructor: NewSensor})
}

func NewSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting %s %s", PrettyName, Version)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	b := Config{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		mu:         sync.RWMutex{},
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Debugf("Reconfiguring %s", PrettyName)

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	undervolt, armFrequencyCapped, currentlyThrottled, softTempLimitActive, underVoltOccurred, armFrequencyCapOccurred, throttlingOccurred, softTempLimitOccurred, err := getThrottlingStates()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"undervolt":                 undervolt,
		"arm_frequency_capped":      armFrequencyCapped,
		"currently_throttled":       currentlyThrottled,
		"soft_temp_limit_active":    softTempLimitActive,
		"under_volt_occurred":       underVoltOccurred,
		"arm_frequency_cap_occured": armFrequencyCapOccurred,
		"throttling_occurred":       throttlingOccurred,
		"soft_temp_limit_occurred":  softTempLimitOccurred,
	}, nil
}

func (c *Config) Close(ctx context.Context) error {
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
