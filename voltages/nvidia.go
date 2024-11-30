package voltages

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"go.viam.com/rdk/logging"
)

var (
	ErrIgnoredSensor = errors.New("ignored sensor")
)

type powerSensor interface {
	StartUpdating() error
	Close()
	GetReading() (voltage, current, power float64)
	GetReadingMap() map[string]interface{}
	GetName() string
}

type jetsonPowerSensor struct {
	logger                       logging.Logger
	mu                           sync.RWMutex
	wg                           sync.WaitGroup
	index                        int
	name                         string
	cancelCtx                    context.Context
	cancelFunc                   context.CancelFunc
	updateTask                   func()
	voltageFile                  string
	currentFile                  string
	overCurrentAlarmFile         string
	criticalOverCurrentAlarmFile string
	voltage                      float64
	current                      float64
	power                        float64
	overCurrentAlarm             bool
	criticalOverCurrentAlarm     bool
}

func (s *jetsonPowerSensor) StartUpdating() error {
	updateInterval := 1 * time.Second
	ui, err := utils.ReadInt64FromFileWithContext(s.cancelCtx, "/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/update_interval")
	if err == nil {
		updateInterval = time.Duration(ui) * time.Second
	}
	s.updateTask = func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for {
			select {
			case <-s.cancelCtx.Done():
				s.logger.Infof("Stopping background update for %s", s.name)
				return
			case <-time.After(updateInterval):
				rawVoltage, vErr := utils.ReadInt64FromFileWithContext(s.cancelCtx, s.voltageFile)
				rawCurrent, cErr := utils.ReadInt64FromFileWithContext(s.cancelCtx, s.currentFile)
				overCurrentAlarm, ocaErr := utils.ReadBoolFromFileWithContext(s.cancelCtx, s.overCurrentAlarmFile)
				criticalOverCurrentAlarm, cocaErr := utils.ReadBoolFromFileWithContext(s.cancelCtx, s.criticalOverCurrentAlarmFile)
				current := float64(rawCurrent) / 1000
				voltage := float64(rawVoltage) / 1000
				s.logger.Infof("Voltage: %v, Current: %v, OverCurrentAlarm: %v, CriticalOverCurrentAlarm: %v", voltage, current, overCurrentAlarm, criticalOverCurrentAlarm)
				s.mu.Lock()
				if vErr == nil {
					s.voltage = voltage
				}
				if cErr == nil {
					s.current = current
				}
				if vErr == nil && cErr == nil {
					s.power = math.Round(voltage*current*100) / 100
				}
				if ocaErr == nil {
					s.overCurrentAlarm = overCurrentAlarm
				}
				if cocaErr == nil {
					s.criticalOverCurrentAlarm = criticalOverCurrentAlarm
				}
				s.mu.Unlock()
			}
		}

	}
	go s.updateTask()
	return nil
}

func (s *jetsonPowerSensor) Close() {
	s.logger.Infof("Shutting down %s", s.name)
	s.cancelFunc()
	s.wg.Wait()
	s.logger.Infof("Shutdown complete")
}

func (s *jetsonPowerSensor) GetName() string {
	return s.name
}

func (s *jetsonPowerSensor) GetReading() (voltage, current, power float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.voltage, s.current, s.power
}

func (s *jetsonPowerSensor) GetReadingMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"voltage":               s.voltage,
		"current":               s.current,
		"power":                 s.power,
		"over_current_alarm":    s.overCurrentAlarm,
		"critical_over_current": s.criticalOverCurrentAlarm,
	}
}

func newJetsonPowerSensor(ctx context.Context, logger logging.Logger, index int) (*jetsonPowerSensor, error) {
	name, err := utils.ReadFileWithContext(ctx, fmt.Sprintf("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in%v_label", index))
	if err != nil {
		return nil, err
	}
	if strings.Contains(name, "sum") {
		return nil, ErrIgnoredSensor
	}
	logger.Infof("Creating Jetson Power Sensor: %s", name)
	ctx, cancel := context.WithCancel(ctx)
	return &jetsonPowerSensor{
		logger:                       logger.Sublogger(name),
		index:                        index,
		name:                         name,
		cancelCtx:                    ctx,
		cancelFunc:                   cancel,
		overCurrentAlarmFile:         fmt.Sprintf("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/curr%v_alarm", index),
		criticalOverCurrentAlarmFile: fmt.Sprintf("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/curr%v_crit_alarm", index),
		voltageFile:                  fmt.Sprintf("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/in%v_input", index),
		currentFile:                  fmt.Sprintf("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon1/curr%v_input", index),
	}, nil
}

func getJetsonPowerSensors(ctx context.Context, logger logging.Logger) ([]powerSensor, error) {
	sensors := make([]powerSensor, 0)
	matches, err := filepath.Glob("/sys/bus/i2c/drivers/ina3221/1-0040/hwmon/hwmon*/in*_label")
	if err != nil {
		return nil, err
	}
	for _, match := range matches {
		base := filepath.Base(match)
		var index int
		_, err := fmt.Sscanf(base, "in%d_label", &index)
		if err != nil {
			return nil, err
		}
		sensor, err := newJetsonPowerSensor(ctx, logger, index)
		if err == ErrIgnoredSensor {
			logger.Debugf("Ignoring sensor %s", index)
			continue
		}
		if err != nil {
			return nil, err
		}
		sensor.StartUpdating()
		sensors = append(sensors, sensor)
	}
	return sensors, nil
}
