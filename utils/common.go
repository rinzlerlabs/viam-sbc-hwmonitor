package utils

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/nvidia"
	"github.com/rinzlerlabs/sbcidentify/raspberrypi"
)

var (
	ErrBoardNotSupported = errors.New("board not supported")
)

type systemTemperatures struct {
	CPU   *float64
	GPU   *float64
	Extra map[string]float64
}

type temperatureReader interface {
	Name() string
	Read(context.Context) (float64, error)
}

type fileTemperatureSensor struct {
	name string
	path string
}

func (t *fileTemperatureSensor) Read(ctx context.Context) (float64, error) {
	tempChan := make(chan float64)
	errChan := make(chan error)
	go func() {
		data, err := os.ReadFile(t.path)
		if err != nil {
			errChan <- err
			return
		}
		temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			errChan <- err
			return
		}
		tempChan <- (temp / 1000)
	}()
	select {
	case temp := <-tempChan:
		return temp, nil
	case err := <-errChan:
		return 0, err
	case <-time.After(100 * time.Millisecond):
		return 0, errors.New("timeout")
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (t *fileTemperatureSensor) Name() string {
	return t.name
}

type vcgencmdSensor struct {
	name       string
	subcommand string
}

func (t *vcgencmdSensor) Read(ctx context.Context) (float64, error) {
	proc := exec.CommandContext(ctx, "vcgencmd", "measure_temp", t.subcommand)
	outputBytes, err := proc.Output()
	if err != nil {
		return 0, err
	}
	return parseTemperature(string(outputBytes))
}

func (t *vcgencmdSensor) Name() string {
	return t.name
}

var jetsonTemperatureSensors = []temperatureReader{
	&fileTemperatureSensor{name: "CPU", path: "/sys/devices/virtual/thermal/thermal_zone0/temp"},
	&fileTemperatureSensor{name: "GPU", path: "/sys/devices/virtual/thermal/thermal_zone1/temp"},
	&fileTemperatureSensor{name: "CV0", path: "/sys/devices/virtual/thermal/thermal_zone2/temp"},
	&fileTemperatureSensor{name: "CV1", path: "/sys/devices/virtual/thermal/thermal_zone3/temp"},
	&fileTemperatureSensor{name: "CV2", path: "/sys/devices/virtual/thermal/thermal_zone4/temp"},
	&fileTemperatureSensor{name: "SOC0", path: "/sys/devices/virtual/thermal/thermal_zone5/temp"},
	&fileTemperatureSensor{name: "SOC1", path: "/sys/devices/virtual/thermal/thermal_zone6/temp"},
	&fileTemperatureSensor{name: "SOC2", path: "/sys/devices/virtual/thermal/thermal_zone7/temp"},
	&fileTemperatureSensor{name: "TJ", path: "/sys/devices/virtual/thermal/thermal_zone8/temp"},
}

var raspberryPiTemperatureSensors = []temperatureReader{
	&vcgencmdSensor{name: "CPU", subcommand: ""},
	&vcgencmdSensor{name: "PMIC", subcommand: "pmic"},
}

func GetTemperatures(ctx context.Context) (*systemTemperatures, error) {
	if sbcidentify.IsBoardType(raspberrypi.RaspberryPi) {
		return rpi_getTemperatures(ctx)
	} else if sbcidentify.IsBoardType(nvidia.Jetson) {
		return jetson_getTemperatures(ctx)
	} else {
		return nil, ErrBoardNotSupported
	}
}

func rpi_getTemperatures(ctx context.Context) (*systemTemperatures, error) {
	systemTemps := &systemTemperatures{Extra: make(map[string]float64)}
	for _, sensor := range raspberryPiTemperatureSensors {
		temp, err := sensor.Read(ctx)
		if err != nil {
			continue
		}
		switch sensor.Name() {
		case "CPU":
			systemTemps.CPU = &temp
		case "GPU":
			systemTemps.GPU = &temp
		default:
			systemTemps.Extra[sensor.Name()] = temp
		}
	}

	return systemTemps, nil
}

func parseTemperature(output string) (Temperature float64, Err error) {
	t := strings.Split(output, "=")
	t1 := strings.TrimSuffix(t[1], "'C\n")
	return strconv.ParseFloat(strings.TrimSpace(t1), 64)
}

func jetson_getTemperatures(ctx context.Context) (*systemTemperatures, error) {
	systemTemps := &systemTemperatures{Extra: make(map[string]float64)}
	for _, sensor := range jetsonTemperatureSensors {
		temp, err := sensor.Read(ctx)
		if err != nil {
			continue
		}
		switch sensor.Name() {
		case "CPU":
			systemTemps.CPU = &temp
		case "GPU":
			systemTemps.GPU = &temp
		default:
			systemTemps.Extra[sensor.Name()] = temp
		}
	}

	return systemTemps, nil
}
