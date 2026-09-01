package sensors

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

// thermalZoneGlob matches the kernel thermal zones exposed by sysfs. These are
// standard across Linux and present on most SBCs (Jetson, x86, etc.).
const thermalZoneGlob = "/sys/class/thermal/thermal_zone*"

// ReadSysfsThermalZones discovers and reads all kernel thermal zones, mapping
// each to CPU/GPU/Extra based on its reported type (ex: "cpu-thermal" -> CPU).
// Zones that cannot be read are skipped. The returned map may be empty if no
// zones are readable.
func ReadSysfsThermalZones(ctx context.Context) (*SystemTemperatures, error) {
	zones, err := filepath.Glob(thermalZoneGlob)
	if err != nil {
		return nil, err
	}

	systemTemps := &SystemTemperatures{Extra: make(map[string]float64)}
	for _, zone := range zones {
		name, err := readThermalZoneType(ctx, filepath.Join(zone, "type"))
		if err != nil {
			continue
		}

		temp, err := NewFileTemperatureSensor(name, filepath.Join(zone, "temp")).Read(ctx)
		if err != nil {
			continue
		}
		temp = float64(int((temp/1000)*100)) / 100

		lowerName := strings.ToLower(name)
		switch {
		case strings.Contains(lowerName, "cpu"):
			cpu := temp
			systemTemps.CPU = &cpu
		case strings.Contains(lowerName, "gpu"):
			gpu := temp
			systemTemps.GPU = &gpu
		default:
			systemTemps.Extra[name] = temp
		}
	}

	return systemTemps, nil
}

// readThermalZoneType returns the cleaned type name of a thermal zone, ex: a
// zone reporting "cpu-thermal" becomes "CPU".
func readThermalZoneType(ctx context.Context, path string) (string, error) {
	zoneType, err := utils.ReadFileWithContext(ctx, path)
	if err != nil {
		return "", err
	}
	zoneType = strings.TrimSpace(zoneType)
	zoneType = strings.TrimSuffix(zoneType, "-thermal")
	zoneType = strings.TrimSuffix(zoneType, "-therm")
	if zoneType == "" {
		return "", errors.New("empty thermal zone type")
	}
	return strings.ToUpper(zoneType), nil
}

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
	// Thermal sysfs reads can be slow on some boards (e.g. Jetson reads go
	// through the BPMP), so allow a generous timeout rather than dropping them.
	timeout, cancel := context.WithTimeout(ctx, time.Second)
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
