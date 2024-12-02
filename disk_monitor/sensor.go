package disk_monitor

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/shirou/gopsutil/disk"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "disk_monitor")
	API         = sensor.API
	PrettyName  = "Disk Monitor"
	Description = "A sensor that reports disk information including usage and IO counters"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu                sync.RWMutex
	logger            logging.Logger
	cancelCtx         context.Context
	cancelFunc        func()
	disks             []*localDisk
	includeIOCounters bool
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
	disks, err := getDisks(ctx, newConf.Disks)
	if err != nil {
		return err
	}
	c.disks = disks
	c.includeIOCounters = newConf.IncludeIOCounters

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ret := make(map[string]interface{})
	if c.includeIOCounters {
		devices := make([]string, 0)
		for _, d := range c.disks {
			devices = append(devices, d.Device)
		}
		iocounters, err := disk.IOCountersWithContext(ctx, devices...)
		if err != nil {
			return nil, err
		}

		for name, ioc := range iocounters {
			ret[name+"_read_count"] = ioc.ReadCount
			ret[name+"_read_count_merged"] = ioc.MergedReadCount
			ret[name+"_write_count"] = ioc.WriteCount
			ret[name+"_write_count_merged"] = ioc.MergedWriteCount
			ret[name+"_read_bytes"] = ioc.ReadBytes
			ret[name+"_write_bytes"] = ioc.WriteBytes
			ret[name+"_read_time"] = ioc.ReadTime
			ret[name+"_write_time"] = ioc.WriteTime
			ret[name+"_iops_in_progress"] = ioc.IopsInProgress
			ret[name+"_io_time"] = ioc.IoTime
			ret[name+"_weighted_io_time"] = ioc.WeightedIO
			ret[name+"_serial_number"] = ioc.SerialNumber
			ret[name+"_label"] = ioc.Label
		}
	}
	for _, d := range c.disks {
		name := d.Device

		usage, err := disk.UsageWithContext(ctx, d.Mountpoint)
		if err != nil {
			return nil, err
		}

		ret[name+"_total"] = usage.Total
		ret[name+"_used"] = usage.Used
		ret[name+"_free"] = usage.Free
		ret[name+"_used_percent"] = usage.UsedPercent
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

func getDisks(ctx context.Context, confDisks []string) ([]*localDisk, error) {
	disks := make([]*localDisk, 0)
	if len(confDisks) == 0 {
		realDisks, err := getRealDisks(ctx)
		if err != nil {
			return nil, err
		}
		disks = append(disks, realDisks...)
	} else {
		for _, d := range confDisks {
			realDisks, err := getRealDisks(ctx)
			if err != nil {
				return nil, err
			}
			for _, disk := range realDisks {
				if disk.Device == d || filepath.Base(disk.Device) == d || disk.Device == filepath.Base(d) {
					disks = append(disks, disk)
				}
				if disk.Mountpoint == d {
					disks = append(disks, disk)
				}
			}
		}
	}
	return disks, nil
}

func getRealDisks(ctx context.Context) ([]*localDisk, error) {
	realDisks := make([]*localDisk, 0)
	parts, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	for _, part := range parts {
		switch part.Fstype {
		case "ext4", "ext3", "ext2", "xfs", "btrfs", "reiserfs", "jfs", "ntfs", "fat32", "fat16", "fat12", "hfs", "hfsplus", "apfs", "ufs", "zfs":
			realDisks = append(realDisks, &localDisk{Device: filepath.Base(part.Device), Mountpoint: part.Mountpoint})
		default:
			continue
		}
	}
	return realDisks, nil
}

type localDisk struct {
	Device     string
	Mountpoint string
}
