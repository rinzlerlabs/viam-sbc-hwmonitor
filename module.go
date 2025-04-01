//go:build linux
// +build linux

package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/module"
	viamutils "go.viam.com/utils"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/clocks"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/cpumanager"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/cpumonitor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/diskmonitor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/gpumonitor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/memorymonitor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/powermanager"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/processmonitor"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/pwmfan"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/temperatures"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/throttling"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/voltages"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/wifimonitor"
)

func main() {
	logger := module.NewLoggerFromArgs(utils.LoggerName)
	logger.Infof("Starting RinzlerLabs SBC Sensors Module %v", utils.Version)
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
	viamutils.ContextualMain(moduleutils.RunModule, logger)
}
