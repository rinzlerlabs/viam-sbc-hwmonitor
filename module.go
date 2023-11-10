package main

import (
	"context"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"

	"github.com/viam-soleng/viam-raspi-sensors/clocks"
	"github.com/viam-soleng/viam-raspi-sensors/cpu_manager"
	"github.com/viam-soleng/viam-raspi-sensors/temperatures"
	"github.com/viam-soleng/viam-raspi-sensors/throttling"
	"github.com/viam-soleng/viam-raspi-sensors/voltages"
)

func main() {
	utils.ContextualMain(mainWithArgs, module.NewLoggerFromArgs("viam-raspi-utils"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) (err error) {
	custom_module, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, clocks.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, cpu_manager.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, temperatures.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, throttling.Model)
	if err != nil {
		return err
	}

	err = custom_module.AddModelFromRegistry(ctx, sensor.API, voltages.Model)
	if err != nil {
		return err
	}

	err = custom_module.Start(ctx)
	defer custom_module.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
