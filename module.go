package main

import (
	moduleutils "github.com/thegreatco/viamutils/module"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/rinzlerlabs/viam-raspi-sensors/clocks"
	"github.com/rinzlerlabs/viam-raspi-sensors/cpu_manager"
	"github.com/rinzlerlabs/viam-raspi-sensors/cpu_monitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/gpu_monitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/memory_monitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/process_monitor"
	"github.com/rinzlerlabs/viam-raspi-sensors/pwm_fan"
	"github.com/rinzlerlabs/viam-raspi-sensors/temperatures"
	"github.com/rinzlerlabs/viam-raspi-sensors/throttling"
	raspiutils "github.com/rinzlerlabs/viam-raspi-sensors/utils"
	"github.com/rinzlerlabs/viam-raspi-sensors/voltages"
)

func main() {
	logger := module.NewLoggerFromArgs(raspiutils.LoggerName)
	logger.Infof("Starting RinzlerLabs SBC Sensors Module %v", raspiutils.Version)
	moduleutils.AddModularResource(clocks.API, clocks.Model)
	moduleutils.AddModularResource(cpu_manager.API, cpu_manager.Model)
	moduleutils.AddModularResource(temperatures.API, temperatures.Model)
	moduleutils.AddModularResource(throttling.API, throttling.Model)
	moduleutils.AddModularResource(voltages.API, voltages.Model)
	moduleutils.AddModularResource(pwm_fan.API, pwm_fan.Model)
	moduleutils.AddModularResource(cpu_monitor.API, cpu_monitor.Model)
	moduleutils.AddModularResource(gpu_monitor.API, gpu_monitor.Model)
	moduleutils.AddModularResource(memory_monitor.API, memory_monitor.Model)
	moduleutils.AddModularResource(process_monitor.API, process_monitor.Model)
	utils.ContextualMain(moduleutils.RunModule, logger)
}
