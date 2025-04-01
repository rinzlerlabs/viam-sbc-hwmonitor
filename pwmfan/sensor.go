package pwmfan

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	viam_utils "go.viam.com/utils"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "pwm_fan")
	API         = sensor.API
	PrettyName  = "SBC PWM Fan Speed Controller"
	Description = "A module to control the speed of a PWM fan connected to the SBC based on a temperature table"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	mu               sync.RWMutex
	logger           logging.Logger
	cancelCtx        context.Context
	cancelFunc       func()
	fan              *fan
	temperatureTable map[float64]float64
	temps            []float64
	temperatureFunc  func(ctx context.Context) (*sensors.SystemTemperatures, error)
	worker           *viam_utils.StoppableWorkers
}

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[sensor.Sensor, *CloudConfig]{Constructor: NewSensor})
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

func (c *Config) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Infof("Reconfiguring %s", PrettyName)
	c.logger.Debug("Stopping background worker")
	c.worker.Stop()
	c.logger.Debugf("Background worker stopped")

	newConf, err := resource.NativeConfig[*CloudConfig](conf)
	if err != nil {
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()
	fan, err := newFan(deps, newConf.BoardName, newConf.FanPin, newConf.UseInternalFan)
	if err != nil {
		return err
	}
	c.fan = fan

	tempTable := make(map[float64]float64)
	temps := make([]float64, 0, len(newConf.TemperatureTable))
	for ts, speed := range newConf.TemperatureTable {
		temp, err := strconv.ParseFloat(ts, 64)
		if err != nil {
			c.logger.Errorf("Error parsing temperature: %s", err)
			return err
		}
		if speed > 1 {
			speed = speed / float64(100)
		}
		tempTable[temp] = speed
		temps = append(temps, temp)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(temps)))

	c.temps = temps
	c.temperatureTable = tempTable
	tempFunc, err := GetTemperatureFunc()
	if err != nil {
		return err
	}
	c.temperatureFunc = tempFunc
	c.worker = viam_utils.NewBackgroundStoppableWorkers(c.startUpdating)

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	temperatures, err := c.temperatureFunc(ctx)
	if err != nil {
		c.logger.Errorf("Error getting board temperatures: %s", err)
		return nil, err
	}

	if temperatures.CPU == nil {
		c.logger.Errorf("Error getting CPU temperature")
		return nil, err
	}
	currentTemp := *temperatures.CPU

	fan_speed, err := c.fan.GetSpeed(ctx)
	if err != nil {
		c.logger.Errorf("Error getting fan speed: %s", err)
		return nil, err
	}

	return map[string]interface{}{
		"temperature":   currentTemp,
		"fan_speed_pct": fan_speed * 100,
	}, nil
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	c.logger.Debugf("Notifying monitor to shut down")
	c.worker.Stop()
	c.logger.Info("Monitor shut down")
	c.fan.Close()
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}

func (c *Config) startUpdating(ctx context.Context) {
	seepTime := 100 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(seepTime):
			temperatures, err := c.temperatureFunc(ctx)
			if err != nil {
				c.logger.Errorf("Error getting SoC temperature: %s", err)
				break
			}
			if temperatures.CPU == nil {
				c.logger.Errorf("Error getting CPU temperature")
				break
			}
			currentTemp := *temperatures.CPU
			var desiredSpeed float64
			for _, targetTemp := range c.temps {
				if currentTemp >= targetTemp {
					desiredSpeed = c.temperatureTable[targetTemp]
					break
				}
			}

			c.logger.Debugf("Current temperature: %f, desired speed: %f", currentTemp, desiredSpeed)
			err = c.fan.SetSpeed(ctx, desiredSpeed)
			if err != nil {
				c.logger.Errorf("Error setting fan speed: %s", err)
			}
		}
	}
}
