package voltages

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.viam.com/rdk/logging"
)

type raspberryPiPowerSensor struct {
	logger     logging.Logger
	mu         sync.RWMutex
	wg         sync.WaitGroup
	name       string
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	updateTask func()
	voltage    float64
	current    float64
	power      float64
}

func (s *raspberryPiPowerSensor) StartUpdating() error {
	updateInterval := 1 * time.Second
	s.updateTask = func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for {
			select {
			case <-s.cancelCtx.Done():
				return
			case <-time.After(updateInterval):
				voltage, err := getRaspberryPiComponentVoltage(s.name)
				if err != nil {
					s.logger.Errorf("failed to get voltage: %v", err)
					continue
				}
				s.mu.Lock()
				s.voltage = voltage
				s.mu.Unlock()
			}
		}
	}
	go s.updateTask()
	return nil
}

func (s *raspberryPiPowerSensor) Close() {
}

func (s *raspberryPiPowerSensor) GetReading() (voltage, current, power float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.voltage, s.current, s.power
}

func (s *raspberryPiPowerSensor) GetReadingMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"voltage": s.voltage,
	}
}

func (s *raspberryPiPowerSensor) GetName() string {
	return s.name
}

func newRaspberryPiPowerSensor(ctx context.Context, logger logging.Logger, name string) (*raspberryPiPowerSensor, error) {
	logger.Infof("Creating Raspberry Pi power sensor for %s", name)
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	s := &raspberryPiPowerSensor{
		logger:     logger,
		name:       name,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
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
