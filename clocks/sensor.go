//go:build linux
// +build linux

package clocks

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "clocks")
	API         = sensor.API
	PrettyName  = "SBC Clock Sensor"
	Description = "A sensor that reports the clock frequencies of an SBC"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	sensors    []clockSensor
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
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	if c.sensors != nil {
		for _, s := range c.sensors {
			s.Close()
		}
	}

	c.cancelCtx, c.cancelFunc = context.WithCancel(context.Background())
	sensors, err := getClockSensors(c.cancelCtx, c.logger)
	if err != nil {
		return err
	}
	c.sensors = sensors

	for _, sensor := range c.sensors {
		if err := sensor.StartUpdating(); err != nil {
			return err
		}
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	readings := make(map[string]interface{})
	for _, s := range c.sensors {
		for k, v := range s.GetReadingMap() {
			readings[k] = v
		}
	}
	return readings, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	c.cancelFunc()
	for _, s := range c.sensors {
		s.Close()
	}
	c.logger.Infof("Shutdown complete")
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
