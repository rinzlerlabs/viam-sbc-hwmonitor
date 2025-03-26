package memorymonitor

import (
	"context"
	"math"
	"sync"

	"github.com/shirou/gopsutil/v4/mem"
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

	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	swap_devices, err := mem.SwapDevicesWithContext(ctx)
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
	ret["writeback"] = v.WriteBack
	ret["dirty"] = v.Dirty
	ret["writebacktmp"] = v.WriteBackTmp
	ret["shared"] = v.Shared
	ret["slab"] = v.Slab
	ret["sreclaimable"] = v.Sreclaimable
	ret["sunreclaim"] = v.Sunreclaim
	ret["pagetables"] = v.PageTables
	ret["swapcached"] = v.SwapCached
	ret["commitlimit"] = v.CommitLimit
	ret["committed_as"] = v.CommittedAS
	ret["high_total"] = v.HighTotal
	ret["high_free"] = v.HighFree
	ret["low_total"] = v.LowTotal
	ret["low_free"] = v.LowFree
	ret["mapped"] = v.Mapped
	ret["vmalloc_total"] = v.VmallocTotal
	ret["vmalloc_used"] = v.VmallocUsed
	ret["vmalloc_chunk"] = v.VmallocChunk
	ret["hugepages_total"] = v.HugePagesTotal
	ret["hugepages_free"] = v.HugePagesFree
	ret["hugepages_rsvd"] = v.HugePagesRsvd
	ret["hugepages_surp"] = v.HugePagesSurp
	ret["hugepages_size"] = v.HugePageSize
	ret["anonhugepages"] = v.AnonHugePages

	ret["swap_total"] = swap.Total
	ret["swap_used"] = swap.Used
	ret["swap_free"] = swap.Free
	ret["swap_used_percent"] = math.Round(swap.UsedPercent*100) / 100
	ret["swap_s_in"] = swap.Sin
	ret["swap_s_out"] = swap.Sout
	ret["swap_page_in"] = swap.PgIn
	ret["swap_page_out"] = swap.PgOut
	ret["swap_page_fault"] = swap.PgFault
	ret["swap_page_maj_fault"] = swap.PgMajFault

	for _, device := range swap_devices {
		total_swap := device.UsedBytes + device.FreeBytes
		ret["swap_device_"+device.Name+"_used"] = device.UsedBytes
		ret["swap_device_"+device.Name+"_free"] = device.FreeBytes
		ret["swap_device_"+device.Name+"_total"] = total_swap
		ret["swap_device_"+device.Name+"_used_percent"] = math.Round((float64(device.UsedBytes)/float64(total_swap))*100) / 100
	}

	return ret, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
