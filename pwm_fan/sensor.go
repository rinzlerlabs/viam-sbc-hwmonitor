package pwm_fan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	viam_utils "go.viam.com/utils"

	"github.com/viam-soleng/viam-raspi-sensors/utils"
)

var (
	Model       = resource.NewModel("viam-soleng", "raspi", "pwm_fan")
	API         = sensor.API
	PrettyName  = "Raspberry Pi PWM Fan Speed Controller"
	Description = "A module to control the speed of a PWM fan connected to the Raspberry Pi based on a temperature table"
	Version     = utils.Version
)

type fan struct {
	pin         board.GPIOPin
	internalFan *os.File
}

func newFan(deps resource.Dependencies, boardName string, pin string, useInternalFan bool) (*fan, error) {
	if useInternalFan {
		matches, err := filepath.Glob("/sys/class/hwmon/hwmon*/pwm1")
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no pwm1 file found in /sys/class/hwmon/hwmon*/")
		}
		internalFan, err := os.OpenFile(matches[0], os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return &fan{
			internalFan: internalFan,
			pin:         nil,
		}, nil
	}

	b, err := board.FromDependencies(deps, boardName)
	if err != nil {
		return nil, err
	}

	fanPin, err := b.GPIOPinByName(pin)
	if err != nil {
		return nil, err
	}

	return &fan{
		internalFan: nil,
		pin:         fanPin,
	}, nil
}

func (f *fan) SetSpeed(ctx context.Context, speed float64) error {
	if f.internalFan != nil {
		actualSpeed := int(speed * 255)
		if actualSpeed > 255 {
			actualSpeed = 255
		}
		fmt.Printf("Setting internal fan speed to %v\n", actualSpeed)
		f.internalFan.Seek(0, 0)
		_, err := f.internalFan.Write([]byte(strconv.Itoa(actualSpeed)))
		if err != nil {
			return err
		}
		return nil
	}
	return f.pin.SetPWM(ctx, speed, nil)
}

func (f *fan) GetSpeed(ctx context.Context) (float64, error) {
	if f.internalFan != nil {
		f.internalFan.Seek(0, 0)
		b := make([]byte, 10)
		count, err := f.internalFan.Read(b)
		if err != nil {
			return 0, err
		}
		speed, err := strconv.ParseFloat(strings.TrimSpace(string(b[:count])), 64)
		if err != nil {
			return 0, err
		}
		return speed / 255, nil
	}
	return f.pin.PWM(ctx, nil)
}

func (f *fan) Close() {
	if f.internalFan != nil {
		f.internalFan.Close()
	}
}

type Config struct {
	resource.Named
	mu               sync.RWMutex
	logger           logging.Logger
	cancelCtx        context.Context
	cancelFunc       func()
	fan              *fan
	temperatureTable map[float64]float64
	temps            []float64
	monitor          func()
	done             chan bool
	wg               sync.WaitGroup
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

				select {
				case <-time.After(100 * time.Millisecond):
					continue
				case <-c.done:
					return
				}
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
	c.done <- true
	c.logger.Infof("Notifying monitor to shut down")
	c.wg.Wait()
	c.logger.Info("Monitor shut down")
	c.fan.Close()
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
