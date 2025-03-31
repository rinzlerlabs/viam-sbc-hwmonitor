package temperatures

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "temperatures")
	API         = sensor.API
	PrettyName  = "SBC Temperature Sensor"
	Description = "A sensor that reports the temperatures of the SBC, if available"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu              sync.RWMutex
	logger          logging.Logger
	cancelCtx       context.Context
	cancelFunc      func()
	temperatureFunc func(ctx context.Context) (*sensors.SystemTemperatures, error)
}

func init() {
	resource.RegisterComponent(
		API,
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

	temperatureFunc, err := GetTemperatureFunc()
	if err != nil {
		return err
	}
	c.temperatureFunc = temperatureFunc
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	temperatures, err := c.temperatureFunc(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	if temperatures.CPU != nil {
		res["CPU"] = *temperatures.CPU
	}

	if temperatures.GPU != nil {
		res["GPU"] = *temperatures.GPU
	}

	for key, value := range temperatures.Extra {
		res[key] = value
	}

	return res, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
