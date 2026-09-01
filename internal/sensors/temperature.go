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

// thermalClassPath is the sysfs directory containing kernel thermal zones.
// These are standard across Linux and present on most SBCs (Jetson, x86,
// etc.).
const thermalClassPath = "/sys/class/thermal"

// ThermalZoneReader polls kernel thermal zones repeatedly without re-globbing
// sysfs or re-reading each zone's static "type" file on every call — that
// discovery happens once, in NewThermalZoneReader. Use this instead of
// ReadSysfsThermalZones for anything that polls on a fixed interval (e.g.
// pwm_fan's control loop); ReadSysfsThermalZones redoes the discovery work on
// every single call, which is wasted I/O for data that never changes at
// runtime.
type ThermalZoneReader struct {
	zones []thermalZoneSource
}

type thermalZoneSource struct {
	name     string // cleaned zone type, ex: "CPU0"
	tempPath string
}

// NewThermalZoneReader discovers all kernel thermal zones once, mapping each
// to its cleaned type name (ex: "cpu-thermal" -> "CPU"). Zones whose type
// can't be read are skipped.
func NewThermalZoneReader(ctx context.Context) (*ThermalZoneReader, error) {
	return newThermalZoneReader(ctx, thermalClassPath)
}

// newThermalZoneReader does the work for NewThermalZoneReader against an
// injectable sysfs class directory, so tests can point it at a fake directory
// tree instead of the real sysfs.
func newThermalZoneReader(ctx context.Context, classPath string) (*ThermalZoneReader, error) {
	paths, err := filepath.Glob(filepath.Join(classPath, "thermal_zone*"))
	if err != nil {
		return nil, err
	}

	zones := make([]thermalZoneSource, 0, len(paths))
	for _, zone := range paths {
		name, err := readThermalZoneType(ctx, filepath.Join(zone, "type"))
		if err != nil {
			continue
		}
		zones = append(zones, thermalZoneSource{name: name, tempPath: filepath.Join(zone, "temp")})
	}
	return &ThermalZoneReader{zones: zones}, nil
}

// Read polls the current temperature of every zone discovered at
// construction time, mapping each to CPU/GPU/Extra based on its type (ex:
// "CPU" -> CPU). Zones that fail to read are skipped. The returned map may be
// empty if no zones are readable.
func (r *ThermalZoneReader) Read(ctx context.Context) (*SystemTemperatures, error) {
	systemTemps := &SystemTemperatures{Extra: make(map[string]float64)}
	for _, zone := range r.zones {
		temp, err := NewFileTemperatureSensor(zone.name, zone.tempPath).Read(ctx)
		if err != nil {
			continue
		}
		temp = float64(int((temp/1000)*100)) / 100

		lowerName := strings.ToLower(zone.name)
		switch {
		case strings.Contains(lowerName, "cpu"):
			// A board can expose more than one CPU-type zone (e.g. per-cluster
			// zones on big.LITTLE SoCs: cpu0-thermal, cpu1-thermal). Report the
			// hottest as the single CPU reading, since that's what matters for
			// both general reporting and driving fan control, and keep every
			// zone individually in Extra so none of them are lost.
			if systemTemps.CPU == nil || temp > *systemTemps.CPU {
				cpu := temp
				systemTemps.CPU = &cpu
			}
			systemTemps.Extra[zone.name] = temp
		case strings.Contains(lowerName, "gpu"):
			if systemTemps.GPU == nil || temp > *systemTemps.GPU {
				gpu := temp
				systemTemps.GPU = &gpu
			}
			systemTemps.Extra[zone.name] = temp
		default:
			systemTemps.Extra[zone.name] = temp
		}
	}
	return systemTemps, nil
}

// ReadSysfsThermalZones discovers and reads all kernel thermal zones in a
// single call. Prefer ThermalZoneReader for anything that polls repeatedly.
func ReadSysfsThermalZones(ctx context.Context) (*SystemTemperatures, error) {
	return readSysfsThermalZones(ctx, thermalClassPath)
}

// readSysfsThermalZones does the work for ReadSysfsThermalZones against an
// injectable sysfs class directory, so tests can point it at a fake directory
// tree instead of the real sysfs.
func readSysfsThermalZones(ctx context.Context, classPath string) (*SystemTemperatures, error) {
	r, err := newThermalZoneReader(ctx, classPath)
	if err != nil {
		return nil, err
	}
	return r.Read(ctx)
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
