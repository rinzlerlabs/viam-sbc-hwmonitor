package pwm_fan

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	viam_utils "go.viam.com/utils"

	"github.com/viam-soleng/viam-raspi-sensors/utils"
)

var Model = resource.NewModel("viam-soleng", "raspi", "pwm_fan")
var PrettyName = "Raspberry Pi PWM Fan Speed Controller"
var Description = "A module to control the speed of a PWM fan connected to the Raspberry Pi based on a temperature table"
var Version = utils.Version

type Config struct {
	resource.Named
	mu               sync.RWMutex
	logger           logging.Logger
	cancelCtx        context.Context
	cancelFunc       func()
	FanPin           board.GPIOPin
	Board            *board.Board
	TemperatureTable map[float64]float64
	Temps            []float64
	monitor          func()
	done             chan bool
	wg               sync.WaitGroup
}

func init() {
	resource.RegisterComponent(
		sensor.API,
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
		done:       make(chan bool, 1),
		wg:         sync.WaitGroup{},
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Debugf("Reconfiguring %s", PrettyName)

	newConf, err := resource.NativeConfig[*CloudConfig](conf)
	if err != nil {
		return err
	}

	untypedBoard, err := deps.Lookup(resource.NewName(board.API, newConf.BoardName))
	if err != nil {
		c.logger.Errorf("Error looking up board: %s", err)
		return err
	}

	board := untypedBoard.(board.Board)
	fanPin, err := board.GPIOPinByName(newConf.FanPin)
	if err != nil {
		c.logger.Errorf("Error looking up fan pin: %s", err)
		return err
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()
	c.FanPin = fanPin
	c.Board = &board

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

	c.Temps = temps
	c.TemperatureTable = tempTable
	if fanPin.SetPWMFreq(ctx, 1000, nil) != nil {
		c.logger.Errorf("Error setting PWM frequency: %s", err)
		return err
	}

	if c.monitor == nil {
		c.monitor = func() {
			ctx := context.Background()
			c.wg.Add(1)
			defer c.wg.Done()
			for {
				select {
				case <-c.done:
					return
				default:
					currentTemp, err := utils.GetSoCTemperature()
					if err != nil {
						c.logger.Errorf("Error getting SoC temperature: %s", err)
						break
					}
					var desiredSpeed float64
					for _, targetTemp := range c.Temps {
						if currentTemp >= targetTemp {
							desiredSpeed = c.TemperatureTable[targetTemp]
							break
						}
					}

					c.logger.Debugf("Current temperature: %f, desired speed: %f", currentTemp, desiredSpeed)
					err = c.FanPin.SetPWM(ctx, desiredSpeed, nil)
					if err != nil {
						c.logger.Errorf("Error setting fan speed: %s", err)
					}
				}

				time.Sleep(100 * time.Millisecond)
			}
		}

		viam_utils.PanicCapturingGo(c.monitor)
	}

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	currentTemp, err := utils.GetSoCTemperature()
	if err != nil {
		c.logger.Errorf("Error getting SoC temperature: %s", err)
		return nil, err
	}

	fan_speed, err := c.FanPin.PWM(ctx, nil)
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
	c.done <- true
	c.logger.Infof("Notifying monitor to shut down")
	c.wg.Wait()
	c.logger.Info("Monitor shut down")
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
