package main

import (
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/viam-soleng/viam-raspi-sensors/clocks"
	"github.com/viam-soleng/viam-raspi-sensors/cpu_manager"
	"github.com/viam-soleng/viam-raspi-sensors/pwm_fan"
	"github.com/viam-soleng/viam-raspi-sensors/temperatures"
	"github.com/viam-soleng/viam-raspi-sensors/throttling"
	"github.com/viam-soleng/viam-raspi-sensors/voltages"

	moduleutils "github.com/thegreatco/viamutils/module"
	raspiutils "github.com/viam-soleng/viam-raspi-sensors/utils"
)

func main() {
	logger := module.NewLoggerFromArgs(raspiutils.LoggerName)
	logger.Infof("Starting RaspiSensors Module %v", raspiutils.Version)
	moduleutils.AddModularResource(clocks.API, clocks.Model)
	moduleutils.AddModularResource(cpu_manager.API, cpu_manager.Model)
	moduleutils.AddModularResource(temperatures.API, temperatures.Model)
	moduleutils.AddModularResource(throttling.API, throttling.Model)
	moduleutils.AddModularResource(voltages.API, voltages.Model)
	moduleutils.AddModularResource(pwm_fan.API, pwm_fan.Model)
	utils.ContextualMain(moduleutils.RunModule, logger)
}
