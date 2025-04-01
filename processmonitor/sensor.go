package processmonitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	viamutils "go.viam.com/utils"

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
	configLock   sync.Mutex
	readingsLock sync.RWMutex // to protect the readings map
	logger       logging.Logger
	info         *procInfo
	processes    utils.OrderedMap[int, *process]
	reading      map[string]interface{}
	workers      *viamutils.StoppableWorkers
	sleepTime    time.Duration
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
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		configLock: sync.Mutex{},
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

	return nil
}

// check if the PID exists in /proc, if it doesn't, the process is no longer running
func procExists(proc *process) bool {
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", proc.Pid)); os.IsNotExist(err) {
		return false
	}
	return true
}

// Get the process to monitor, if it hasn't already been found, or is no longer running, try to find it (again)
func (c *Config) updateProcessList() error {
	procs := utils.NewOrderedMap[int, *process]()
	if c.processes != nil {
		for pid, proc := range c.processes.AllFromFront() {
			if procExists(proc) {
				procs.Set(pid, proc)
			}
		}
	}

	var newProcs utils.OrderedMap[int, *process]
	var err error
	if c.info.ExecutablePath != "" {
		newProcs, err = getProcessesByExe(c.info.ExecutablePath)
	} else if c.info.Name != "" {
		newProcs, err = getProcessesByName(c.info.Name)
	} else {
		return errors.New("no process specified")
	}
	if err != nil {
		return err
	}
	for newProcPid, newProc := range newProcs.AllFromFront() {
		if procs.Has(newProcPid) {
			continue
		}
		procs.Set(newProcPid, newProc)
	}
	c.processes = procs
	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.readingsLock.RLock()
	defer c.readingsLock.RUnlock()
	return c.getReadings(ctx)
}

func (c *Config) startUpdating(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// exit the loop if the context is done
			c.logger.Infof("Stopping %s update loop: %v", PrettyName, ctx.Err())
			return
		case <-time.After(c.sleepTime):
			readings, err := c.getReadings(ctx)
			if err != nil {
				// log the error but continue the loop
				c.logger.Warnf("Failed to get readings for %s: %v", PrettyName, err)
				continue
			}
			// Update the readings in the sensor
			c.readingsLock.Lock()
			c.reading = readings
			c.readingsLock.Unlock()
			// log the successful update
			c.logger.Debugf("Successfully updated readings for %s: %v", PrettyName, readings)
		}
	}
}

func (c *Config) getReadings(ctx context.Context) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	err := c.updateProcessList()
	if err != nil {
		c.logger.Warnf("Error getting process: %v", err)
		return nil, err
	}

	for _, proc := range c.processes.AllFromFront() {
		if err := proc.UpdateStats(ctx); err != nil {
			c.logger.Warnf("Error updating process stats: %v", err)
		} else {
			ret := make(map[string]interface{})
			if c.info.Name != "" {
				ret["name"] = proc.Name
			}
			if c.info.ExecutablePath != "" {
				ret["exe"] = proc.Exe
			}

			ret["pid"] = proc.Pid
			ret["cpu"] = proc.CPUPercent()
			ret["cpu_since_boot"] = proc.CPUPercentSinceBoot()
			ret["threads"] = proc.NumThreads()

			if c.info.IncludeCwd {
				ret["cwd"] = proc.Cwd
			}
			if c.info.IncludeCmdline {
				ret["cmdline"] = proc.CmdLine
			}
			if c.info.IncludeOpenFileCount {
				fc, err := proc.GetOpenFileCount()
				if err != nil {
					c.logger.Warnf("Error getting process open file count: %v", err)
				} else {
					ret["open_files"] = fc
				}
			}
			if c.info.IncludeEnv {
				env, err := proc.GetEnv()
				if err != nil {
					c.logger.Warnf("Error getting process environment: %v", err)
				} else {
					ret["env"] = env
				}
			}
			if c.info.IncludeMemInfo {
				mem, err := proc.GetMemoryInfo()
				if err != nil {
					c.logger.Warnf("Error getting process memory: %v", err)
				} else {
					ret["mem_rss"] = mem.VmRSS
					ret["mem_hwm"] = mem.VmHWM
					ret["mem_data"] = mem.VmData
					ret["mem_stack"] = mem.VmStack
					ret["mem_swap"] = mem.VmSwap
					ret["mem_size"] = mem.VmSize
				}
			}
			resp[fmt.Sprintf("%d", proc.Pid)] = ret
		}
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
