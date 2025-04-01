package sensors

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
)

var (
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
	PID     int32
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

type ProcessMonitor struct {
	Processes         utils.OrderedMap[int32, *Process] // List of processes to monitor
	lastSync          time.Time
	name              string
	disablePidCachine bool
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

func NewProcessMonitor(name string, disablePidCaching bool) *ProcessMonitor {
	// Initialize a new process monitor
	pm := &ProcessMonitor{
		Processes:         utils.NewOrderedMap[int32, *Process](),
		name:              name,
		disablePidCachine: disablePidCaching,
	}
	return pm
}

func (p *ProcessMonitor) GetProcessesWithContext(ctx context.Context) (utils.OrderedMap[int32, *Process], error) {
	if !p.disablePidCachine && p.lastSync.Add(10*time.Second).After(time.Now()) || p.Processes.Len() > 0 {
		return p.Processes, nil // Return cached processes if within the sync interval
	}

	for _, pid := range slices.Collect(p.Processes.Keys()) {
		if ret, err := process.PidExistsWithContext(ctx, int32(pid)); err != nil || !ret {
			// Remove the process from the map if it no longer exists
			p.Processes.Delete(pid) // Keep the process in the list if it still exists
		}
	}

	process.EnableBootTimeCache(true)
	ret := utils.NewOrderedMap[int32, *Process]()
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, errors.Join(errors.New("failed to get processes"), err)
	}
	for _, proc := range procs {
		procName, err := getProcName(proc) // Get the process name
		if err != nil {
			continue
		}
		if procName == p.name {
			ret.Set(proc.Pid, &Process{Process: proc, PID: proc.Pid, Name: procName}) // Store the process in the ordered map
			continue
		}
	}
	p.Processes = ret
	p.lastSync = time.Now() // Update the last sync time
	return ret, nil
}

func getProcName(proc *process.Process) (string, error) {
	pid := proc.Pid
	statPath := filepath.Join("/proc", strconv.Itoa(int(pid)), "comm")
	contents, err := os.ReadFile(statPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(contents), "\n"), nil
}
