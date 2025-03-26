package powermanager

import (
	"context"
	"sync"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager/cpufrequtils"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "power_manager")
	API         = sensor.API
	PrettyName  = "Power Manager"
	Description = "A sensor that reports and manages the CPU configuration of an SBC"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	pm         PowerManager
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

	newConfig, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	pm, err := newPowerManager(newConfig, c.logger)
	if err != nil {
		return err
	}
	requiresReboot, err := pm.ApplyPowerMode()
	if err != nil {
		c.logger.Errorf("Failed to apply power mode: %v", err)
		return err
	}
	c.logger.Infof("Successfully applied power mode: %v", pm)
	if requiresReboot {
		c.logger.Info("Reboot required, rebooting soon")
	}
	c.pm = pm
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	minFreq, maxFreq, err := cpufrequtils.GetFrequencyLimits()
	if err != nil {
		return nil, err
	}

	currentFreq, _, governor, err := cpufrequtils.GetCurrentPolicy()
	if err != nil {
		return nil, err
	}
	ret := map[string]interface{}{"MinimumFrequency": minFreq, "MaximumFrequency": maxFreq, "CurrentFrequency": currentFreq, "Governor": governor}
	powerMode, err := c.pm.GetCurrentPowerMode()
	if err != nil {
		return nil, err
	}
	if powerMode != nil {
		ret["PowerMode"] = powerMode
	}
	return ret, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}
