package jetson

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.viam.com/rdk/logging"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	ErrIgnoredSensor = errors.New("ignored sensor")
)

type jetsonPowerSensor struct {
	logger                       logging.Logger
	mu                           sync.RWMutex
	index                        int
	name                         string
	cancelCtx                    context.Context
	cancelFunc                   context.CancelFunc
	voltageFile                  string
	currentFile                  string
	overCurrentAlarmFile         string
	criticalOverCurrentAlarmFile string
}

func (s *jetsonPowerSensor) Close() error {
	s.logger.Infof("Shutting down %s", s.name)
	s.cancelFunc()
	s.logger.Infof("Shutdown complete")
	return nil
}

func (s *jetsonPowerSensor) GetName() string {
	return s.name
}

func (s *jetsonPowerSensor) GetReading() (voltage, current, power float64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rawVoltage, err := utils.ReadInt64FromFileWithContext(s.cancelCtx, s.voltageFile)
	if err != nil {
		return 0, 0, 0, err
	}
	rawCurrent, err := utils.ReadInt64FromFileWithContext(s.cancelCtx, s.currentFile)
	if err != nil {
		return 0, 0, 0, err
	}

	current = float64(rawCurrent) / 1000
	voltage = float64(rawVoltage) / 1000
	return voltage, current, voltage * current, nil
}

func (s *jetsonPowerSensor) GetReadingMap() (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	s.mu.RLock()
	defer s.mu.RUnlock()
	voltage, current, power, err := s.GetReading()
	if err != nil {
		return nil, err
	}
	ret["voltage"] = voltage
	ret["current"] = current
	ret["power"] = power

	overCurrentAlarm, err := utils.ReadBoolFromFileWithContext(s.cancelCtx, s.overCurrentAlarmFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		ret["over_current_alarm"] = overCurrentAlarm // ensure we set this in the map if it was read successfully
	}
	criticalOverCurrentAlarm, err := utils.ReadBoolFromFileWithContext(s.cancelCtx, s.criticalOverCurrentAlarmFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		// ensure we set this in the map if it was read successfully
		ret["critical_over_current"] = criticalOverCurrentAlarm
	}
	return ret, nil
}

func newJetsonPowerSensor(ctx context.Context, logger logging.Logger, hwmonDir string, index int) (*jetsonPowerSensor, error) {
	name, err := utils.ReadFileWithContext(ctx, filepath.Join(hwmonDir, fmt.Sprintf("in%d_label", index)))
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
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
		overCurrentAlarmFile:         filepath.Join(hwmonDir, fmt.Sprintf("curr%d_alarm", index)),
		criticalOverCurrentAlarmFile: filepath.Join(hwmonDir, fmt.Sprintf("curr%d_crit_alarm", index)),
		voltageFile:                  filepath.Join(hwmonDir, fmt.Sprintf("in%d_input", index)),
		currentFile:                  filepath.Join(hwmonDir, fmt.Sprintf("curr%d_input", index)),
	}, nil
}

func GetPowerSensors(ctx context.Context, logger logging.Logger) ([]sensors.PowerSensor, error) {
	powerSensors := make([]sensors.PowerSensor, 0)
	// Discover INA3221 power monitors at any I2C address / hwmon index rather
	// than hardcoding 1-0040/hwmon1, which differs across Jetson boards.
	matches, err := filepath.Glob("/sys/bus/i2c/drivers/ina3221*/*/hwmon/hwmon*/in*_label")
	if err != nil {
		return nil, err
	}
	for _, match := range matches {
		hwmonDir := filepath.Dir(match)
		base := filepath.Base(match)
		var index int
		if _, err := fmt.Sscanf(base, "in%d_label", &index); err != nil {
			continue
		}
		sensor, err := newJetsonPowerSensor(ctx, logger, hwmonDir, index)
		if errors.Is(err, ErrIgnoredSensor) {
			logger.Debugf("Ignoring power sensor in%d_label", index)
			continue
		}
		if err != nil {
			logger.Warnf("failed to create power sensor from %s: %v", match, err)
			continue
		}
		powerSensors = append(powerSensors, sensor)
	}
	return powerSensors, nil
}
