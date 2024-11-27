package cpu_manager

import (
	"context"
	"os/exec"
	"strconv"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "cpu_manager")
	API         = sensor.API
	PrettyName  = "SBC CPU Manager"
	Description = "A sensor that reports and manages the CPU configuration of an SBC"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	Governor   string
	Frequency  int
	Minimum    int
	Maximum    int
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

	newConf, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	err = utils.InstallPackage("cpufrequtils")
	if err != nil {
		c.logger.Errorf("Error installing cpufrequtils: %s", err)
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()
	c.Governor = newConf.Governor
	c.Frequency = newConf.Frequency
	c.Minimum = newConf.Minimum
	c.Maximum = newConf.Maximum

	args := make([]string, 0)
	if c.Governor != "" {
		args = append(args, "--governor", c.Governor)
	}
	if c.Frequency != 0 {
		args = append(args, "--freq", strconv.Itoa(c.Frequency))
	}
	if c.Minimum != 0 {
		args = append(args, "--min", strconv.Itoa(c.Minimum))
	}
	if c.Maximum != 0 {
		args = append(args, "--max", strconv.Itoa(c.Maximum))
	}

	if len(args) > 0 {
		proc := exec.Command("cpufreq-set", args...)

		outputBytes, err := proc.Output()
		if err != nil {
			c.logger.Errorf("Error configuring CPU: %s", err)
		}
		c.logger.Infof("CPU configured: %s", string(outputBytes))
	} else {
		c.logger.Info("No configuration changes made")
	}

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	min, max, governor, err := getCurrentPolicy()
	if err != nil {
		return nil, err

	}
	currentFrequency, err := getCurrentFrequency()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"current_frequency": currentFrequency,
		"minimum_frequency": min,
		"maximum_frequency": max,
		"governor":          governor,
	}, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
