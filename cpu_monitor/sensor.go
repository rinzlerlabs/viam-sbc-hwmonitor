package cpu_monitor

import (
	"context"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
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
	wg         sync.WaitGroup
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	task       func()
	stats      map[string]interface{}
	sleepTime  time.Duration
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
		stats:      make(map[string]interface{}),
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}

	logger.Infof("Started %s %s", PrettyName, Version)
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Infof("Reconfiguring %s", PrettyName)
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.logger.Infof("Waiting for background task to stop")
	c.wg.Wait()

	c.cancelCtx, c.cancelFunc = context.WithCancel(context.Background())

	newConf, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()
	c.sleepTime = 1 * time.Second
	if newConf.SleepTimeMs > 0 {
		c.sleepTime = time.Duration(newConf.SleepTimeMs) * time.Millisecond
	}
	c.task = c.captureCPUStats
	go c.task()
	c.logger.Debugf("reconfigure complete %s", PrettyName)
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("shutting down")
	c.cancelFunc()
	c.wg.Wait()
	c.logger.Info("shutdown complete")
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}

func (c *Config) captureCPUStats() {
	c.wg.Add(1)
	defer c.wg.Done()
	lastStats, err := readCPUStats()
	if err != nil {
		c.logger.Errorf("Failed to read CPU stats: %v", err)
		panic(err)
	}
	c.logger.Debug("starting CPU stats main loop")
	for {
		select {
		case <-c.cancelCtx.Done():
			return
		case <-time.After(c.sleepTime):
			currStats, err := readCPUStats()
			if err != nil {
				c.logger.Warnf("Failed to read CPU stats, skipping iteration: %v", err)
				continue
			}
			newStats := make(map[string]interface{})
			for core, prev := range lastStats {
				curr, ok := currStats[core]
				if !ok {
					c.logger.Warnf("Core %s not found in current stats", core)
					continue
				}
				usage := calculateUsage(prev, curr)
				newStats[core] = usage
			}
			c.stats = newStats
		}
	}
}
