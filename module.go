//go:build linux
// +build linux

package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/rinzlerlabs/viam-raspi-sensors/clocks"
	"github.com/rinzlerlabs/viam-raspi-sensors/cpumanager"
	"github.com/rinzlerlabs/viam-raspi-sensors/cpumonitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/diskmonitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/gpumonitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/memorymonitor"
	powermanager "github.com/rinzlerlabs/viam-raspi-sensors/powermanager"
	"github.com/rinzlerlabs/viam-raspi-sensors/processmonitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/pwmfan"
	"github.com/rinzlerlabs/viam-raspi-sensors/temperatures"
	"github.com/rinzlerlabs/viam-raspi-sensors/throttling"
	raspiutils "github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"github.com/rinzlerlabs/viam-raspi-sensors/voltages"
	"github.com/rinzlerlabs/viam-raspi-sensors/wifimonitor"
)

func main() {
	logger := module.NewLoggerFromArgs(raspiutils.LoggerName)
	logger.Infof("Starting RinzlerLabs SBC Sensors Module %v", raspiutils.Version)
	moduleutils.AddModularResource(clocks.API, clocks.Model)
	moduleutils.AddModularResource(cpumanager.API, cpumanager.Model)
	moduleutils.AddModularResource(temperatures.API, temperatures.Model)
	moduleutils.AddModularResource(throttling.API, throttling.Model)
	moduleutils.AddModularResource(voltages.API, voltages.Model)
	moduleutils.AddModularResource(pwmfan.API, pwmfan.Model)
	moduleutils.AddModularResource(cpumonitor.API, cpumonitor.Model)
	moduleutils.AddModularResource(gpumonitor.API, gpumonitor.Model)
	moduleutils.AddModularResource(memorymonitor.API, memorymonitor.Model)
	moduleutils.AddModularResource(processmonitor.API, processmonitor.Model)
	moduleutils.AddModularResource(diskmonitor.API, diskmonitor.Model)
	moduleutils.AddModularResource(wifimonitor.API, wifimonitor.Model)
	moduleutils.AddModularResource(powermanager.API, powermanager.Model)
	utils.ContextualMain(moduleutils.RunModule, logger)
}
