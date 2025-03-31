package voltages

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"go.viam.com/rdk/logging"
)

type raspberryPiPowerSensor struct {
	logger logging.Logger
	mu     sync.RWMutex
	name   string
}

func (s *raspberryPiPowerSensor) Close() error {
	return nil
}

func (s *raspberryPiPowerSensor) GetReading() (voltage, current, power float64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	voltage, err = getRaspberryPiComponentVoltage(s.name)
	return
}

func (s *raspberryPiPowerSensor) GetReadingMap() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	voltage, err := getRaspberryPiComponentVoltage(s.name)
	return map[string]interface{}{
		"voltage": voltage,
	}, err
}

func (s *raspberryPiPowerSensor) GetName() string {
	return s.name
}

func newRaspberryPiPowerSensor(ctx context.Context, logger logging.Logger, name string) (*raspberryPiPowerSensor, error) {
	logger.Infof("Creating Raspberry Pi power sensor for %s", name)
	s := &raspberryPiPowerSensor{
		logger: logger,
		name:   name,
	}
	return s, nil
}

func getRaspberryPiPowerSensors(ctx context.Context, logger logging.Logger) ([]powerSensor, error) {
	components := []string{"core", "sdram_c", "sdram_i", "sdram_p"}
	sensors := make([]powerSensor, 0)
	for _, component := range components {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			s, err := newRaspberryPiPowerSensor(ctx, logger, component)
			if err != nil {
				return nil, err
			}
			sensors = append(sensors, s)
		}
	}
	return sensors, nil
}

func getRaspberryPiComponentVoltage(component string) (Voltage float64, Err error) {
	proc := exec.Command("vcgencmd", "measure_volts", component)
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	output := string(outputBytes)
	return parseVcgencmdVoltage(output)
}

func parseVcgencmdVoltage(output string) (Voltage float64, Err error) {
	parts := strings.Split(output, "=")
	if len(parts) != 2 {
		return 0, errors.New("unexpected output from vcgencmd")
	}
	return strconv.ParseFloat(strings.TrimSpace(strings.Replace(parts[1], "V", "", 1)), 64)
}
