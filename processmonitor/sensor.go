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
	goutils "go.viam.com/utils"

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
	mu         sync.Mutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	info       *procInfo
	processes  utils.OrderedMap[int, *process]
	workers    *goutils.StoppableWorkers
	readings   map[string]interface{}
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
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	b := Config{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		mu:         sync.Mutex{},
	}

	if err := b.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Config) Reconfigure(ctx context.Context, _ resource.Dependencies, conf resource.Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger.Debugf("Reconfiguring %s", PrettyName)

	newConf, err := resource.NativeConfig[*ComponentConfig](conf)
	if err != nil {
		return err
	}

	if newConf.Name == "" && newConf.ExecutablePath == "" {
		return errors.New("either name or executable path must be set")
	}

	c.info = &procInfo{
		Name:                 newConf.Name,
		ExecutablePath:       newConf.ExecutablePath,
		IncludeEnv:           newConf.IncludeEnv,
		IncludeCmdline:       newConf.IncludeCmdline,
		IncludeCwd:           newConf.IncludeCwd,
		IncludeOpenFileCount: newConf.IncludeOpenFileCount,
		IncludeMemInfo:       newConf.IncludeMemInfo,
		// IncludeUlimits:       newConf.IncludeUlimits,
		// IncludeOpenFiles:     newConf.IncludeOpenFiles,
		// IncludeNetStats:      newConf.IncludeNetStats,
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	c.workers = goutils.NewBackgroundStoppableWorkers(c.Update)

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
func (c *Config) getProcesses() (utils.OrderedMap[int, *process], error) {
	procs := utils.NewOrderedMap[int, *process]()
	for pid, proc := range c.processes.AllFromFront() {
		if procExists(proc) {
			procs.Set(pid, proc)
		}
	}

	var newProcs utils.OrderedMap[int, *process]
	var err error
	if c.info.ExecutablePath != "" {
		newProcs, err = getProcessesByExe(c.info.ExecutablePath)
	} else if c.info.Name != "" {
		newProcs, err = getProcessesByName(c.info.Name)
	} else {
		return nil, errors.New("no process specified")
	}
	if err != nil {
		return nil, err
	}
	for newProcPid, newProc := range newProcs.AllFromFront() {
		if procs.Has(newProcPid) {
			continue
		}
		procs.Set(newProcPid, newProc)
	}
	c.processes = procs
	return c.processes, nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readings, nil
}

func (c *Config) Update(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			resp := make(map[string]interface{})

			procs, err := c.getProcesses()
			if err != nil {
				c.logger.Warnf("Error getting process: %v", err)
				continue
			}

			for _, proc := range procs.AllFromFront() {
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

			c.readings = resp
		}
	}
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
	c.cancelFunc()
	c.workers.Stop()
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
