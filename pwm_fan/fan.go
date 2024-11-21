package pwm_fan

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/resource"
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
