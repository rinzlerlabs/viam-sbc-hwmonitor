package memory_monitor

import (
	"context"
	"math"
	"sync"

	"github.com/shirou/gopsutil/mem"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "memory_monitor")
	API         = sensor.API
	PrettyName  = "SBC Memory Monitor"
	Description = "A sensor that reports memory usage of the SBC"
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

	// newConf, err := resource.NativeConfig[*ComponentConfig](conf)
	// if err != nil {
	// 	return err
	// }

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ret := make(map[string]interface{})
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	ret["total_memory"] = v.Total
	ret["available_memory"] = v.Available
	ret["used_memory"] = v.Used
	ret["used_percent"] = math.Round(v.UsedPercent*100) / 100
	ret["free_memory"] = v.Free
	ret["active_memory"] = v.Active
	ret["inactive_memory"] = v.Inactive
	ret["wired_memory"] = v.Wired
	ret["buffers_memory"] = v.Buffers
	ret["cached_memory"] = v.Cached
	ret["swap_total"] = v.SwapTotal
	ret["swap_free"] = v.SwapFree
	ret["dirty"] = v.Dirty
	ret["writeback"] = v.Writeback

	return ret, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
