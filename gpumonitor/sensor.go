package gpumonitor

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
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
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	gpuMonitor gpuMonitor
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

	c.cancelCtx, c.cancelFunc = context.WithCancel(context.Background())

	_, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()
	c.gpuMonitor, err = newGpuMonitor(c.logger)
	if err != nil {
		return err
	}
	c.logger.Debugf("reconfigure complete %s", PrettyName)
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]interface{})
	sample, err := c.gpuMonitor.GetGPUStats(ctx)
	if err != nil {
		return nil, err
	}
	for key, typedStats := range sample {
		stats := make(map[string]interface{}, len(typedStats))
		for _, stat := range typedStats {
			if stat.Type == "" {
				continue
			}
			stats[string(stat.Type)] = stat.Value
		}
		m[key] = stats
	}

	return m, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Info("shutting down")
	c.cancelFunc()
	c.logger.Info("shutdown complete")
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
