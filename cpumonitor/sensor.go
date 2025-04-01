package cpumonitor

import (
	"context"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	viamutils "go.viam.com/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "cpu_monitor")
	API         = sensor.API
	PrettyName  = "SBC CPU Monitor Sensor"
	Description = "A sensor that reports the CPU usage of an SBC"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	readingsLock sync.RWMutex
	configLock   sync.Mutex
	logger       logging.Logger
	sleepTime    time.Duration
	workers      *viamutils.StoppableWorkers
	reading      map[string]interface{}
}

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[sensor.Sensor, *ComponentConfig]{Constructor: NewSensor})
}

func NewSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting %s %s", PrettyName, Version)
	b := Config{
		Named:        conf.ResourceName().AsNamed(),
		logger:       logger,
		readingsLock: sync.RWMutex{},
		configLock:   sync.Mutex{},
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}

	logger.Infof("Started %s %s", PrettyName, Version)
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, rawConf resource.Config) error {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.logger.Infof("Reconfiguring %s", PrettyName)
	c.logger.Debug("Stopping background worker")
	c.workers.Stop()
	c.logger.Debugf("Background worker stopped")

	conf, err := resource.NativeConfig[*ComponentConfig](rawConf)
	if err != nil {
		return err
	}

	// In case the component has changed name
	c.Named = rawConf.ResourceName().AsNamed()
	if conf.SleepTimeMs <= 0 {
		// Default to 1000ms if no sleep time is provided
		c.logger.Warnf("Invalid sleep time %d, defaulting to 1000ms", conf.SleepTimeMs)
		conf.SleepTimeMs = 1000 // Default to 1 second
	}
	c.sleepTime = time.Duration(conf.SleepTimeMs * int(time.Millisecond))
	c.workers = viamutils.NewBackgroundStoppableWorkers(c.startUpdating)

	c.logger.Debugf("Reconfigure complete %s", PrettyName)
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.readingsLock.RLock()
	defer c.readingsLock.RUnlock()
	return c.reading, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.logger.Infof("Shutting down %v", PrettyName)
	c.workers.Stop()
	c.logger.Infof("%v Shutdown complete", PrettyName)
	return nil
}

// startUpdating is a goroutine that updates the CPU stats every sleepTime
// It ensures if there are multiple readers of this sensor, it doesn't cause short samples
func (c *Config) startUpdating(ctx context.Context) {
	var err error
	var lastStats map[string]sensors.CPUCoreStats
	for {
		if lastStats == nil {
			lastStats, err = sensors.ReadCPUStats()
			if err != nil {
				c.logger.Warnf("Failed to read CPU stats, skipping iteration: %v", err)
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.sleepTime):
			currStats, err := sensors.ReadCPUStats()
			if err != nil {
				c.logger.Warnf("Failed to read CPU stats, skipping iteration: %v", err)
				continue
			}
			ret := make(map[string]interface{})
			for core, prev := range lastStats {
				curr, ok := currStats[core]
				if !ok {
					c.logger.Warnf("Core %s not found in current stats", core)
					continue
				}
				usage := sensors.CalculateUsage(prev, curr)
				ret[core] = usage
			}
			lastStats = currStats
			c.readingsLock.Lock()
			c.reading = ret
			c.readingsLock.Unlock()
		}
	}
}
