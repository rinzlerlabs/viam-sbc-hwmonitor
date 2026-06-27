//go:build linux
// +build linux

package cpumanager

import (
	"context"
	"sync"

	"github.com/rinzlerlabs/sbcidentify"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
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
	logger.Errorf("This component is deprecated and support will be removed in a subsequent release. Please migrate to the new %v component.", powermanager.Model)

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

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	if !sbcidentify.IsRaspberryPi() {
		c.logger.Errorf("This sensor is only supported on Raspberry Pi")
		return utils.ErrBoardNotSupported
	}

	// cpufrequtils was removed in Debian Trixie; install its maintained
	// replacement (linux-cpupower, plus its libcpupower1 runtime library) there
	// and keep cpufrequtils on older systems.
	cpuFreqPackages := []string{"cpufrequtils"}
	if utils.IsDebianTrixieOrNewer() {
		cpuFreqPackages = []string{"linux-cpupower", "libcpupower1"}
	}
	if err = utils.InstallPackage(cpuFreqPackages...); err != nil {
		c.logger.Errorf("Error installing %v: %s", cpuFreqPackages, err)
		return err
	}

	c.Governor = newConf.Governor
	c.Frequency = newConf.Frequency
	c.Minimum = newConf.Minimum
	c.Maximum = newConf.Maximum

	if c.Governor == "" && c.Frequency == 0 && c.Minimum == 0 && c.Maximum == 0 {
		c.logger.Info("No configuration changes made")
		return nil
	}

	output, err := applyCPUPolicy(c.Governor, c.Frequency, c.Minimum, c.Maximum)
	if err != nil {
		c.logger.Errorf("Error configuring CPU: %s: %s", err, output)
		return err
	}
	c.logger.Infof("CPU configured: %s", output)

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
