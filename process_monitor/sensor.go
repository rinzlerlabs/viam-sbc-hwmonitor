package process_monitor

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/shirou/gopsutil/process"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
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
	mu         sync.RWMutex
	logger     logging.Logger
	cancelCtx  context.Context
	cancelFunc func()
	process    *processConfig
}

type processConfig struct {
	Name             string
	ExecutablePath   string
	IncludeEnv       bool
	IncludeCmdline   bool
	IncludeCwd       bool
	IncludeOpenFiles bool
	IncludeUlimits   bool
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
		mu:         sync.RWMutex{},
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

	c.process = &processConfig{
		Name:             newConf.Name,
		ExecutablePath:   newConf.ExecutablePath,
		IncludeEnv:       newConf.IncludeEnv,
		IncludeCmdline:   newConf.IncludeCmdline,
		IncludeCwd:       newConf.IncludeCwd,
		IncludeOpenFiles: newConf.IncludeOpenFiles,
		IncludeUlimits:   newConf.IncludeUlimits,
	}

	// In case the module has changed name
	c.Named = conf.ResourceName().AsNamed()

	return nil
}

func (c *Config) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ret := make(map[string]interface{})
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, proc := range procs {
		exe, err := proc.ExeWithContext(ctx)
		if os.IsPermission(err) {
			continue
		}
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			c.logger.Warnf("Error getting process exe, skipping: %v", err)
			continue
		}
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			c.logger.Warnf("Error getting process name, skipping: %v", err)
			continue
		}

		if c.process.Name != "" && c.process.Name != name {
			continue
		}

		if c.process.ExecutablePath != "" && c.process.ExecutablePath != exe {
			continue
		}

		cpu, err := proc.CPUPercentWithContext(ctx)
		if err != nil {
			c.logger.Warnf("Error getting process cpu: %v", err)
		} else {
			ret["cpu"] = math.Round(cpu*100) / 100
		}
		mem, err := proc.MemoryInfoWithContext(ctx)
		if err != nil {
			c.logger.Warnf("Error getting process memory: %v", err)
		} else {
			ret["rss"] = mem.RSS
			ret["vms"] = mem.VMS
			ret["swap"] = mem.Swap
			ret["locked"] = mem.Locked
		}
		numThreads, err := proc.NumThreadsWithContext(ctx)
		if err != nil {
			c.logger.Warnf("Error getting process threads: %v", err)
		} else {
			ret["threads"] = numThreads
		}

		numOpenFiles, err := proc.NumFDsWithContext(ctx)
		if err != nil {
			c.logger.Warnf("Error getting process open files: %v", err)
		} else {
			ret["open_files"] = numOpenFiles
		}

		if c.process.IncludeEnv {
			env, err := proc.EnvironWithContext(ctx)
			if err != nil {
				c.logger.Warnf("Error getting process env: %v", err)
			} else {
				for _, e := range env {
					parts := strings.Split(e, "=")
					if len(parts) == 2 {
						ret[parts[0]] = parts[1]
					}
				}
			}
		}
		if c.process.IncludeCmdline {
			cmdline, err := proc.CmdlineWithContext(ctx)
			if err != nil {
				c.logger.Warnf("Error getting process cmdline: %v", err)
			} else {
				ret["cmdline"] = cmdline
			}
		}
		if c.process.IncludeCwd {
			cwd, err := proc.CwdWithContext(ctx)
			if err != nil {
				c.logger.Warnf("Error getting process cwd: %v", err)
			} else {
				ret["cwd"] = cwd
			}
		}
		if c.process.IncludeOpenFiles {
			openFiles, err := proc.OpenFilesWithContext(ctx)
			if err != nil {
				c.logger.Warnf("Error getting process open files: %v", err)
			} else {
				for i, f := range openFiles {
					ret[fmt.Sprintf("open_file_%d", i)] = f.Path
				}
			}
		}

		if c.process.IncludeUlimits {
			limits, err := proc.RlimitWithContext(ctx)
			if err != nil {
				c.logger.Warnf("Error getting process rlimits: %v", err)
			} else {
				for _, v := range limits {
					ret[fmt.Sprintf("rlimit_%s_hard", resourceToString(v.Resource))] = v.Hard
					ret[fmt.Sprintf("rlimit_%s_soft", resourceToString(v.Resource))] = v.Soft
					ret[fmt.Sprintf("rlimit_%s_used", resourceToString(v.Resource))] = v.Used
				}
			}
		}
	}

	return ret, nil
}

func resourceToString(r int32) string {
	switch r {
	case process.RLIMIT_AS:
		return "as"
	case process.RLIMIT_CORE:
		return "core"
	case process.RLIMIT_CPU:
		return "cpu"
	case process.RLIMIT_DATA:
		return "data"
	case process.RLIMIT_FSIZE:
		return "fsize"
	case process.RLIMIT_LOCKS:
		return "locks"
	case process.RLIMIT_MEMLOCK:
		return "memlock"
	case process.RLIMIT_MSGQUEUE:
		return "msgqueue"
	case process.RLIMIT_NICE:
		return "nice"
	case process.RLIMIT_NOFILE:
		return "nofile"
	case process.RLIMIT_NPROC:
		return "nproc"
	case process.RLIMIT_RSS:
		return "rss"
	case process.RLIMIT_RTPRIO:
		return "rtprio"
	case process.RLIMIT_RTTIME:
		return "rttime"
	case process.RLIMIT_SIGPENDING:
		return "sigpending"
	case process.RLIMIT_STACK:
		return "stack"
	default:
		return fmt.Sprintf("unknown_%d", r)
	}
}

func (c *Config) Close(ctx context.Context) error {
	c.logger.Infof("Shutting down %s", PrettyName)
	return nil
}

func (c *Config) Ready(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, nil
}
