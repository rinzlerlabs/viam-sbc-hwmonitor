package wifi_monitor

import (
	"context"
	"errors"
	"runtime"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "wifi_monitor")
	API         = sensor.API
	PrettyName  = "WiFi Monitor Sensor"
	Description = "A sensor that reports the status of the WiFi connection"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu          sync.Mutex
	logger      logging.Logger
	cancelCtx   context.Context
	cancelFunc  func()
	wifiMonitor wifiMonitor
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
		mu:         sync.Mutex{},
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

	// There is no conf for this sensor
	newConf, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	if runtime.GOOS == "linux" {
		mon := newLinuxWifiMonitor(newConf.Adapter)
		if mon == nil {
			return errors.New("no suitable wifi monitor found")
		}
		c.wifiMonitor = mon
	} else if runtime.GOOS == "darwin" {
		c.wifiMonitor = &macOsWifiMonitor{}
	} else {
		c.wifiMonitor = nil
	}

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ret := make(map[string]interface{})
	if c.wifiMonitor != nil {
		status, err := c.wifiMonitor.GetNetworkStatus()
		if err != nil {
			return nil, err
		}
		ret["network"] = status.NetworkName
		ret["signal_strength"] = status.SignalStrength
		ret["link_speed_mbps"] = status.LinkSpeedMbps
	} else {
		ret["network"] = "unknown"
	}

	return ret, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	c.cancelFunc()
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
