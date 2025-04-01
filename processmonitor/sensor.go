package processmonitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	viamutils "go.viam.com/utils"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/internal/sensors"
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	Model       = resource.NewModel(utils.Namespace, "hwmonitor", "process_monitor")
	API         = sensor.API
	PrettyName  = "SBC Process Monitor"
	Description = "A sensor that reports process information"
	Version     = utils.Version
)

type Config struct {
	resource.Named
	configLock      sync.Mutex
	readingsLock    sync.RWMutex // to protect the readings map
	logger          logging.Logger
	info            *procInfo
	currentReadings map[string]interface{}
	workers         *viamutils.StoppableWorkers
	sleepTime       time.Duration
}

type procInfo struct {
	Name                 string
	ExecutablePath       string
	IncludeEnv           bool
	IncludeCmdline       bool
	IncludeCwd           bool
	IncludeOpenFiles     bool
	IncludeUlimits       bool
	IncludeNetStats      bool
	IncludeMemInfo       bool
	IncludeOpenFileCount bool
}

func init() {
	resource.RegisterComponent(
		API,
		Model,
		resource.Registration[sensor.Sensor, *ComponentConfig]{Constructor: NewSensor})
}

func NewSensor(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting %s %s", PrettyName, Version)

	b := Config{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, rawConf resource.Config) error {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.logger.Debugf("Reconfiguring %s", PrettyName)
	if c.workers != nil {
		c.logger.Debug("Stopping background worker")
		c.workers.Stop()
		c.logger.Debugf("Background worker stopped")
	}

	conf, err := resource.NativeConfig[*ComponentConfig](rawConf)
	if err != nil {
		return err
	}

	if conf.Name == "" && conf.ExecutablePath == "" {
		return errors.New("either name or executable path must be set")
	}

	c.info = &procInfo{
		Name:                 conf.Name,
		ExecutablePath:       conf.ExecutablePath,
		IncludeEnv:           conf.IncludeEnv,
		IncludeCmdline:       conf.IncludeCmdline,
		IncludeCwd:           conf.IncludeCwd,
		IncludeOpenFileCount: conf.IncludeOpenFileCount,
		IncludeMemInfo:       conf.IncludeMemInfo,
		IncludeUlimits:       conf.IncludeUlimits,
		IncludeOpenFiles:     conf.IncludeOpenFiles,
		IncludeNetStats:      conf.IncludeNetStats,
	}

	// In case the module has changed name
	c.Named = rawConf.ResourceName().AsNamed()
	if conf.SleepTimeMs <= 0 {
		// Default to 1000ms if no sleep time is provided
		c.logger.Warnf("Invalid sleep time %d, defaulting to 1000ms", conf.SleepTimeMs)
		conf.SleepTimeMs = 1000 // Default to 1 second
	}
	c.sleepTime = time.Duration(conf.SleepTimeMs * int(time.Millisecond))
	c.workers = viamutils.NewBackgroundStoppableWorkers(c.startUpdating)

	if c.currentReadings == nil {
		// Initialize the current readings map if it is nil, this shouldn't happen but just in case
		c.currentReadings = make(map[string]interface{})
	}

	return nil
}

// Get the process to monitor, if it hasn't already been found, or is no longer running, try to find it (again)
func getMatchingProcesses(ctx context.Context, exePath, name string) (utils.OrderedMap[int, *sensors.Process], error) {
	searchTerm := ""
	if exePath != "" {
		searchTerm = exePath
	} else if name != "" {
		searchTerm = name
	} else {
		return nil, errors.New("no process specified")
	}

	return sensors.GetProcessesWithContext(ctx, searchTerm)
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.readingsLock.RLock()
	defer c.readingsLock.RUnlock()
	return c.currentReadings, nil
}

func (c *Config) startUpdating(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// exit the loop if the context is done
			c.logger.Infof("Stopping %s update loop: %v", PrettyName, ctx.Err())
			return
		case <-time.After(c.sleepTime):
			readings, err := c.getCPUStats(ctx)
			if err != nil {
				// log the error but continue the loop
				c.logger.Warnf("Failed to get readings: %v", err)
				c.updateCurrentReadings(make(map[string]interface{}))
				continue
			}
			// Update the readings in the sensor
			c.updateCurrentReadings(readings)
			// log the successful update
			c.logger.Debugf("Successfully updated readings for %s: %v", PrettyName, readings)
		}
	}
}

func (c *Config) updateCurrentReadings(newReadings map[string]interface{}) {
	c.readingsLock.Lock()
	defer c.readingsLock.Unlock()
	c.currentReadings = newReadings
}

func (c *Config) getCPUStats(ctx context.Context) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	procs, err := getMatchingProcesses(ctx, c.info.ExecutablePath, c.info.Name)
	if err != nil {
		c.logger.Warnf("Error getting process: %v", err)
		return nil, err
	}

	for _, proc := range procs.AllFromFront() {
		ret := make(map[string]interface{})
		if c.info.Name != "" {
			ret["name"] = proc.Name
		}
		if c.info.ExecutablePath != "" {
			exe, err := proc.Exe()
			if err != nil {
				c.logger.Warnf("Error getting executable path for process %d: %v", proc.PID, err)
			} else {
				ret["exe"] = exe
			}
		}

		ret["pid"] = proc.PID
		if cpu, err := proc.CPUPercentWithContext(ctx); err == nil {
			ret["cpu"] = cpu
		} else {
			c.logger.Debugf("Failed to get CPU percent for process %d: %v", proc.PID, err)
		}

		// ret["cpu_since_boot"] = proc.CPUPercentSinceBoot()
		if numThreads, err := proc.NumThreadsWithContext(ctx); err == nil {
			ret["threads"] = numThreads
		} else {
			c.logger.Debugf("Failed to get number of threads for process %d: %v", proc.PID, err)
		}

		if c.info.IncludeCwd {
			if cwd, err := proc.CwdWithContext(ctx); err == nil {
				ret["cwd"] = cwd
			} else {
				c.logger.Debugf("Failed to get current working directory for process %d: %v", proc.PID, err)
			}
		}
		if c.info.IncludeCmdline {
			if cmdline, err := proc.CmdlineWithContext(ctx); err == nil {
				ret["cmdline"] = cmdline
			} else {
				c.logger.Debugf("Failed to get command line for process %d: %v", proc.PID, err)
			}
		}
		if c.info.IncludeOpenFileCount {
			if openFiles, err := proc.OpenFilesWithContext(ctx); err == nil {
				ret["open_files"] = len(openFiles)
			} else {
				c.logger.Debugf("Failed to get open files for process %d: %v", proc.PID, err)
			}
		}
		if c.info.IncludeEnv {
			if env, err := proc.EnvironWithContext(ctx); err == nil {
				ret["env"] = env
			} else {
				c.logger.Debugf("Failed to get environment variables for process %d: %v", proc.PID, err)
			}
		}
		if c.info.IncludeMemInfo {
			if mem, err := proc.MemoryInfoWithContext(ctx); err == nil {
				ret["mem_rss"] = mem.RSS
				ret["mem_hwm"] = mem.HWM
				ret["mem_data"] = mem.Data
				ret["mem_stack"] = mem.Stack
				ret["mem_swap"] = mem.Swap
				ret["mem_size"] = mem.VMS
			} else {
				c.logger.Debugf("Failed to get memory info for process %d: %v", proc.PID, err)
			}
		}
		resp[fmt.Sprintf("%d", proc.Pid)] = ret
	}
	return resp, nil
}

// func resourceToString(r int32) string {
// 	switch r {
// 	case process.RLIMIT_AS:
// 		return "as"
// 	case process.RLIMIT_CORE:
// 		return "core"
// 	case process.RLIMIT_CPU:
// 		return "cpu"
// 	case process.RLIMIT_DATA:
// 		return "data"
// 	case process.RLIMIT_FSIZE:
// 		return "fsize"
// 	case process.RLIMIT_LOCKS:
// 		return "locks"
// 	case process.RLIMIT_MEMLOCK:
// 		return "memlock"
// 	case process.RLIMIT_MSGQUEUE:
// 		return "msgqueue"
// 	case process.RLIMIT_NICE:
// 		return "nice"
// 	case process.RLIMIT_NOFILE:
// 		return "nofile"
// 	case process.RLIMIT_NPROC:
// 		return "nproc"
// 	case process.RLIMIT_RSS:
// 		return "rss"
// 	case process.RLIMIT_RTPRIO:
// 		return "rtprio"
// 	case process.RLIMIT_RTTIME:
// 		return "rttime"
// 	case process.RLIMIT_SIGPENDING:
// 		return "sigpending"
// 	case process.RLIMIT_STACK:
// 		return "stack"
// 	default:
// 		return fmt.Sprintf("unknown_%d", r)
// 	}
// }

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	c.workers.Stop()
	c.logger.Infof("%s Shutdown complete", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
