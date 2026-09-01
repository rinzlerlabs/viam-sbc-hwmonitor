package sensors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeThermalZone creates a fake sysfs thermal zone directory under root,
// ex: root/thermal_zone0/{type,temp}.
func writeThermalZone(t *testing.T, root, zoneDir, zoneType, tempMilliC string) {
	t.Helper()
	dir := filepath.Join(root, zoneDir)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "type"), []byte(zoneType+"\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "temp"), []byte(tempMilliC+"\n"), 0644))
}

func TestReadSysfsThermalZones_MultipleCPUZonesReportMaxAndKeepEachInExtra(t *testing.T) {
	root := t.TempDir()
	writeThermalZone(t, root, "thermal_zone0", "cpu0-thermal", "45000")
	writeThermalZone(t, root, "thermal_zone1", "cpu1-thermal", "55000")

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.NotNil(t, temps.CPU)
	require.Equal(t, 55.0, *temps.CPU, "CPU should report the hottest of the two zones")
	require.Equal(t, 45.0, temps.Extra["CPU0"], "cooler zone should still be preserved individually")
	require.Equal(t, 55.0, temps.Extra["CPU1"], "hotter zone should also be preserved individually")
}

func TestReadSysfsThermalZones_MultipleGPUZonesReportMax(t *testing.T) {
	root := t.TempDir()
	writeThermalZone(t, root, "thermal_zone0", "gpu0-thermal", "60000")
	writeThermalZone(t, root, "thermal_zone1", "gpu1-thermal", "40000")

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.NotNil(t, temps.GPU)
	require.Equal(t, 60.0, *temps.GPU)
	require.Equal(t, 60.0, temps.Extra["GPU0"])
	require.Equal(t, 40.0, temps.Extra["GPU1"])
}

func TestReadSysfsThermalZones_SingleCPUZoneAlsoLandsInExtra(t *testing.T) {
	root := t.TempDir()
	writeThermalZone(t, root, "thermal_zone0", "cpu-thermal", "50000")

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.NotNil(t, temps.CPU)
	require.Equal(t, 50.0, *temps.CPU)
	require.Equal(t, 50.0, temps.Extra["CPU"])
}

func TestReadSysfsThermalZones_UnrecognizedZoneGoesToExtraOnly(t *testing.T) {
	root := t.TempDir()
	writeThermalZone(t, root, "thermal_zone0", "pmic-thermal", "35000")

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.Nil(t, temps.CPU)
	require.Nil(t, temps.GPU)
	require.Equal(t, 35.0, temps.Extra["PMIC"])
}

func TestReadSysfsThermalZones_SkipsUnreadableZoneButKeepsOthers(t *testing.T) {
	root := t.TempDir()
	// thermal_zone0 has a type but no temp file, so it should be skipped.
	require.NoError(t, os.MkdirAll(filepath.Join(root, "thermal_zone0"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "thermal_zone0", "type"), []byte("cpu-thermal\n"), 0644))
	writeThermalZone(t, root, "thermal_zone1", "gpu-thermal", "42000")

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.Nil(t, temps.CPU)
	require.NotNil(t, temps.GPU)
	require.Equal(t, 42.0, *temps.GPU)
}

func TestReadSysfsThermalZones_NoZonesReturnsEmptyNonNilExtra(t *testing.T) {
	root := t.TempDir()

	temps, err := readSysfsThermalZones(context.Background(), root)
	require.NoError(t, err)

	require.Nil(t, temps.CPU)
	require.Nil(t, temps.GPU)
	require.NotNil(t, temps.Extra)
	require.Empty(t, temps.Extra)
}
