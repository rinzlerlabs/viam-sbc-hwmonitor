package gpu_monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "gpu_monitor")
	API         = sensor.API
	PrettyName  = "SBC GPU Monitor Sensor"
	Description = "A sensor that reports the GPU usage of an SBC"
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
	stats      *utils.CappedCollection[sample]
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
	collectionSize := int(1 / c.sleepTime.Seconds())
	if collectionSize < 1 {
		collectionSize = 1
	}
	c.logger.Infof("Sleep time: %v, Collection size: %v", c.sleepTime, collectionSize)
	c.stats = utils.NewCappedCollection[sample](collectionSize)
	c.task = c.captureGPUStats
	go c.task()
	c.logger.Debugf("reconfigure complete %s", PrettyName)
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string][]gpuDeviceStats)
	for _, sample := range c.stats.Items() {
		for _, stats := range sample.DeviceStats {
			m[stats.Name] = append(m[stats.Name], stats)
		}
	}

	tmp := make(map[string]float64)
	for name, stats := range m {
		for _, stat := range stats {
			tmp[fmt.Sprintf("%s-current_frequency", name)] += float64(stat.CurrentFrequency)
			tmp[fmt.Sprintf("%s-max_frequency", name)] += float64(stat.MaxFrequency)
			tmp[fmt.Sprintf("%s-min_frequency", name)] += float64(stat.MinFrequency)
			tmp[fmt.Sprintf("%s-load", name)] += stat.Load
		}
		tmp[fmt.Sprintf("%s-current_frequency", name)] /= float64(len(stats))
		tmp[fmt.Sprintf("%s-max_frequency", name)] /= float64(len(stats))
		tmp[fmt.Sprintf("%s-min_frequency", name)] /= float64(len(stats))
		tmp[fmt.Sprintf("%s-load", name)] /= float64(len(stats))
	}

	ret := make(map[string]interface{})
	for k, v := range tmp {
		ret[k] = v
	}

	return ret, nil
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

func (c *Config) captureGPUStats() {
	c.wg.Add(1)
	defer c.wg.Done()
	gpuMonitor, err := newGpuMonitor(c.cancelCtx, c.logger)
	if err != nil {
		c.logger.Warnf("Failed to initialize GPU monitor: %v", err)
		return
	}
	defer gpuMonitor.Close()
	c.logger.Debug("starting GPU stats main loop")
	for {
		select {
		case <-c.cancelCtx.Done():
			return
		case <-time.After(c.sleepTime):
			currStats, err := gpuMonitor.GetGPUStats(c.cancelCtx)
			if err != nil {
				c.logger.Warnf("Failed to read GPU stats, skipping iteration: %v", err)
				continue
			}
			sample := sample{DeviceStats: currStats}
			c.stats.Push(sample)
		}
	}
}

type sample struct {
	DeviceStats []gpuDeviceStats
}
