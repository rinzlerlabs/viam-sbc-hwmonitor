package sensors

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

var (
	symlinkCache       = sync.Map{}
	ErrProcessNotFound = errors.New("process not found")
)

type CPUCoreStats struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
}

type Process struct {
	*process.Process
	PID     int
	Name    string // Name of the process (for convenience)
	exe     string // Path to the executable
	cmdline string // Command line of the process
}

func (p *Process) Exe() (string, error) {
	if p.exe != "" {
		return p.exe, nil
	}
	if p.Process == nil {
		return "", errors.New("process is nil")
	}
	exe, err := p.Process.Exe()
	if err != nil {
		return "", err
	}
	p.exe = exe // Cache the executable path
	return exe, nil
}

func (p *Process) Cmdline() (string, error) {
	if p.cmdline != "" {
		return p.cmdline, nil
	}
	if p.Process == nil {
		return "", errors.New("process is nil")
	}
	cmdline, err := p.Process.Cmdline()
	if err != nil {
		return "", err
	}
	p.cmdline = cmdline // Cache the command line
	return cmdline, nil
}

func ReadCPUStats() (map[string]CPUCoreStats, error) {
	rawStats, err := cpu.Times(true)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]CPUCoreStats)
	totalStats := CPUCoreStats{}
	for _, stat := range rawStats {
		// Add per-core stats
		stats[stat.CPU] = CPUCoreStats{
			User:    uint64(stat.User),
			Nice:    uint64(stat.Nice),
			System:  uint64(stat.System),
			Idle:    uint64(stat.Idle),
			IOWait:  uint64(stat.Iowait),
			IRQ:     uint64(stat.Irq),
			SoftIRQ: uint64(stat.Softirq),
			Steal:   uint64(stat.Steal),
		}

		// Add total stats
		totalStats.User += uint64(stat.User)
		totalStats.Nice += uint64(stat.Nice)
		totalStats.System += uint64(stat.System)
		totalStats.Idle += uint64(stat.Idle)
		totalStats.IOWait += uint64(stat.Iowait)
		totalStats.IRQ += uint64(stat.Irq)
		totalStats.SoftIRQ += uint64(stat.Softirq)
		totalStats.Steal += uint64(stat.Steal)
	}

	stats["cpu"] = totalStats

	return stats, nil
}

// CalculateUsage calculates CPU usage percentages
func CalculateUsage(prev, curr CPUCoreStats) float64 {
	prevIdle := prev.Idle + prev.IOWait
	currIdle := curr.Idle + curr.IOWait

	prevNonIdle := prev.User + prev.Nice + prev.System + prev.IRQ + prev.SoftIRQ + prev.Steal
	currNonIdle := curr.User + curr.Nice + curr.System + curr.IRQ + curr.SoftIRQ + curr.Steal

	prevTotal := prevIdle + prevNonIdle
	currTotal := currIdle + currNonIdle

	totalDelta := currTotal - prevTotal
	idleDelta := currIdle - prevIdle

	if totalDelta == 0 {
		return 0.0
	}

	return utils.RoundValue((float64(totalDelta-idleDelta)/float64(totalDelta))*100, 2)
}

func GetProcessesWithContext(ctx context.Context, name string) (utils.OrderedMap[int, *Process], error) {
	process.EnableBootTimeCache(true)
	resolvedName := name
	if !filepath.IsAbs(name) {
		fullPath, err := exec.LookPath(name)
		if err == nil {
			// Recursively resolve symlinks for the resolved name
			resolvedName, err = resolveSymlinkWithCache(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve symlink for %s: %w", name, err)
			}
		}
	} else {
		// Recursively resolve symlinks for the resolved name
		resolved, err := resolveSymlinkWithCache(resolvedName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlink for %s: %w", name, err)
		}
		resolvedName = resolved
	}

	ret := utils.NewOrderedMap[int, *Process]()
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, errors.Join(errors.New("failed to get processes"), err)
	}
	for _, proc := range procs {
		cmdLine, err := proc.CmdlineSliceWithContext(ctx)
		if err != nil {
			// If we can't get the command line, skip this process
			continue
		}
		if len(cmdLine) > 0 && filepath.Base(cmdLine[0]) == filepath.Base(name) {
			// If the command line matches the name, add it to the ordered map
			ret.Set(int(proc.Pid), &Process{Process: proc, PID: int(proc.Pid), Name: name})
			continue
		}

		exe, err := proc.Exe()
		if err != nil {
			// If we can't get the executable path, skip this process
			continue
		}
		if exe == resolvedName || filepath.Base(exe) == filepath.Base(resolvedName) {
			ret.Set(int(proc.Pid), &Process{Process: proc, PID: int(proc.Pid), exe: exe, Name: resolvedName}) // Store the process in the ordered map
			continue
		}
	}
	return ret, nil
}

func resolveSymlinkWithCache(path string) (string, error) {
	if cached, ok := symlinkCache.Load(path); ok {
		return cached.(string), nil
	}

	resolved, err := resolveSymlink(path)
	if err == nil {
		symlinkCache.Store(path, resolved)
	}
	return resolved, err
}

func resolveSymlink(path string) (string, error) {
	visited := make(map[string]bool) // To detect symlink loops
	for {
		if visited[path] {
			return "", fmt.Errorf("symlink loop detected for path: %s", path)
		}
		visited[path] = true

		resolved, err := os.Readlink(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("path does not exist: %s", path)
			}
			// If it's not a symlink, return the path as is
			if err.(*os.PathError).Err != syscall.EINVAL {
				return "", err
			}
			return path, nil
		}

		// If the resolved path is relative, resolve it relative to the current path
		if !filepath.IsAbs(resolved) {
			path = filepath.Join(filepath.Dir(path), resolved)
		} else {
			path = resolved
		}
	}
}
