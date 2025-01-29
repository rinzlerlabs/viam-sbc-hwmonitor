package utils

import (
	"context"
	"errors"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/sbcidentify"
	"github.com/rinzlerlabs/sbcidentify/boardtype"
)

var (
	ErrBoardNotSupported = errors.New("board not supported")
)

func ReadFileWithContext(ctx context.Context, path string) (string, error) {
	fileChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			errChan <- err
			return
		}
		fileChan <- data
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case data := <-fileChan:
		return strings.TrimSpace(string(data)), nil
	case err := <-errChan:
		return "", err
	}
}

func ReadInt64FromFileWithContext(ctx context.Context, path string) (int64, error) {
	data, err := ReadFileWithContext(ctx, path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(data), 10, 64)
}

func ReadBoolFromFileWithContext(ctx context.Context, path string) (bool, error) {
	data, err := ReadFileWithContext(ctx, path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(data) == "1", nil
}

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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	data, err := ReadFileWithContext(timeout, t.path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
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
	if sbcidentify.IsBoardType(boardtype.RaspberryPi) {
		return rpi_getTemperatures(ctx)
	} else if sbcidentify.IsBoardType(boardtype.Jetson) {
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
		temp = float64(int((temp/1000)*100)) / 100
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

func RoundValue(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(f*shift) / shift
}
