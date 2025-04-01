package sensors

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

type SystemTemperatures struct {
	CPU   *float64
	GPU   *float64
	Extra map[string]float64
}

type TemperatureReader interface {
	Name() string
	Read(context.Context) (float64, error)
}

func NewFileTemperatureSensor(name, path string) TemperatureReader {
	return &FileTemperatureSensor{name: name, path: path}
}

type FileTemperatureSensor struct {
	name string
	path string
}

func (t *FileTemperatureSensor) Read(ctx context.Context) (float64, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	data, err := utils.ReadFileWithContext(timeout, t.path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
}

func (t *FileTemperatureSensor) Name() string {
	return t.name
}
