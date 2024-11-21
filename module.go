package main

import (
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/rinzlerlabs/viam-raspi-sensors/clocks"
	"github.com/rinzlerlabs/viam-raspi-sensors/cpu_manager"
	"github.com/rinzlerlabs/viam-raspi-sensors/pwm_fan"
	"github.com/rinzlerlabs/viam-raspi-sensors/temperatures"
	"github.com/rinzlerlabs/viam-raspi-sensors/throttling"
	"github.com/rinzlerlabs/viam-raspi-sensors/voltages"

	raspiutils "github.com/rinzlerlabs/viam-raspi-sensors/utils"
	moduleutils "github.com/thegreatco/viamutils/module"
)

func main() {
	logger := module.NewLoggerFromArgs(raspiutils.LoggerName)
	logger.Infof("Starting RinzlerLabs RaspiSensors Module %v", raspiutils.Version)
	moduleutils.AddModularResource(clocks.API, clocks.Model)
	moduleutils.AddModularResource(cpu_manager.API, cpu_manager.Model)
	moduleutils.AddModularResource(temperatures.API, temperatures.Model)
	moduleutils.AddModularResource(throttling.API, throttling.Model)
	moduleutils.AddModularResource(voltages.API, voltages.Model)
	moduleutils.AddModularResource(pwm_fan.API, pwm_fan.Model)
	utils.ContextualMain(moduleutils.RunModule, logger)
}
