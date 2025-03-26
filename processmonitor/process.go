package processmonitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
)

var (
	ErrProcessNotFound = errors.New("process not found")
)

type processStat struct {
	PID            int
	Comm           string
	State          string
	PPID           int
	PGRP           int
	Session        int
	TtyNr          int
	Tpgid          int
	Flags          int
	Minflt         int
	Cminflt        int
	Majflt         int
	Cmajflt        int
	Utime          float64
	Stime          float64
	Cutime         int
	Cstime         int
	Priority       int
	Nice           int
	NumThreads     int
	Itrealvalue    int
	Starttime      float64
	Vsize          int
	Rss            int
	Rlim           int
	Startcode      int
	Endcode        int
	Startstack     int
	Kstkeip        int
	Kstkesp        int
	Signal         int
	Blocked        int
	Sigignore      int
	Sigcatch       int
	Wchan          int
	Nswap          int
	Cnswap         int
	ExitSignal     int
	Processor      int
	RtPriority     int
	Policy         int
	DelayacctBlkio int
	GuestTime      int
	CguestTime     int
}

type processMemoryInfo struct {
	VmSize  int
	VmRSS   int
	VmData  int
	VmStack int
	VmSwap  int
	VmPeak  int
	VmHWM   int
}

type process struct {
	Pid     int
	Name    string
	Exe     string
	CmdLine string
	Cwd     string

	environment                string
	statsFilePath              string
	currentStats               *processStat
	currentStatsTime           time.Time
	currentSystemUptimeSeconds float64
	lastStats                  *processStat
	lastStatsTime              time.Time
	jiffiesPerSecondF          float64
	jiffiesPerSecondI          int64
	mu                         sync.Mutex
}

func (p *process) UpdateStats(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.currentStats != nil {
		p.lastStats = p.currentStats
		p.lastStatsTime = p.currentStatsTime
	}

	stats, err := p.parseStatFile()
	if err != nil {
		return err
	}

	systemUptimeSeconds, err := p.getSystemUptimeSeconds()
	if err != nil {
		return err
	}

	p.currentStatsTime = time.Now()
	p.currentStats = stats
	p.currentSystemUptimeSeconds = systemUptimeSeconds
	return nil
}

func (p *process) CPUPercentSinceBoot() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we don't have current stats, we can't calculate CPU usage
	if p.currentStats == nil {
		return 0.0
	}

	// All times are in jiffies, until they're not
	procStartTime := p.currentStats.Starttime
	systemUptime := p.currentSystemUptimeSeconds * p.jiffiesPerSecondF
	timeSinceProcStart := systemUptime - procStartTime

	// This shouldn't be possible, but just in case
	if timeSinceProcStart == 0 {
		return 0.0
	}

	// Calculate total process time
	totalProcTime := p.currentStats.Utime + p.currentStats.Stime

	// Calculate CPU usage as a percentage
	cpuUsage := (totalProcTime / timeSinceProcStart) * 100

	// Truncate to two decimal places
	return math.Trunc((cpuUsage)*100) / 100
}

func (p *process) CPUPercent() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we don't have stats, we can't calculate CPU usage
	if p.currentStats == nil || p.lastStats == nil {
		return 0.0
	}

	deltaUtime := p.currentStats.Utime - p.lastStats.Utime
	deltaStime := p.currentStats.Stime - p.lastStats.Stime
	deltaTime := p.currentStatsTime.Sub(p.lastStatsTime).Seconds()

	// If deltaTime is 0, we can't calculate CPU usage
	if deltaTime == 0 {
		return 0.0
	}

	// Convert from jiffies to seconds to match deltaTime
	totalDeltaTime := (deltaUtime + deltaStime) / p.jiffiesPerSecondF

	// Convert to percentage
	cpuUsage := (totalDeltaTime / deltaTime) * 100

	// Truncate to two decimal places
	return math.Trunc(cpuUsage*100) / 100
}

func (p *process) NumThreads() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastStats.NumThreads
}

func (p *process) GetEnv() (string, error) {
	if p.environment == "" {
		env, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", p.Pid))
		if err != nil {
			return "", err
		}
		p.environment = strings.ReplaceAll(string(env), "\x00", "\n")
	}
	return p.environment, nil
}

func (p *process) GetOpenFileCount() (int, error) {
	dir := fmt.Sprintf("/proc/%d/fd", p.Pid)
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func (p *process) GetMemoryInfo() (*processMemoryInfo, error) {
	return GetMemoryInfo(p.Pid)
}

func getCwd(pid int) (string, error) {
	link, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return "", err
	}
	return link, nil
}

func getCmdLine(pid int) (string, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	return scanner.Text(), nil
}

func getJiffiesPerSecond() (float64, error) {
	output, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return 0, err
	}

	jiffies, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, err
	}

	return jiffies, nil
}

func newProcessWithExe(pid int, exe string) (*process, error) {
	cmdLine, err := getCmdLine(pid)
	if err != nil {
		return nil, err
	}
	cwd, err := getCwd(pid)
	if err != nil {
		return nil, err
	}
	jiffies, err := getJiffiesPerSecond()
	if err != nil {
		return nil, err
	}
	proc := &process{
		Pid:               pid,
		Exe:               exe,
		Name:              filepath.Base(exe),
		CmdLine:           cmdLine,
		Cwd:               cwd,
		statsFilePath:     fmt.Sprintf("/proc/%d/stat", pid),
		jiffiesPerSecondF: jiffies,
		jiffiesPerSecondI: int64(jiffies),
	}
	firstStats, err := proc.parseStatFile()
	if err != nil {
		return nil, err
	}
	proc.lastStats = firstStats
	return proc, nil
}

func newProcessWithName(pid int, name string) (*process, error) {
	target, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return nil, err
	}
	cmdLine, err := getCmdLine(pid)
	if err != nil {
		return nil, err
	}
	cwd, err := getCwd(pid)
	if err != nil {
		return nil, err
	}
	jiffies, err := getJiffiesPerSecond()
	if err != nil {
		return nil, err
	}
	proc := &process{
		Pid:               pid,
		Name:              name,
		Exe:               target,
		CmdLine:           cmdLine,
		Cwd:               cwd,
		statsFilePath:     fmt.Sprintf("/proc/%d/stat", pid),
		jiffiesPerSecondF: jiffies,
		jiffiesPerSecondI: int64(jiffies),
	}
	firstStats, err := proc.parseStatFile()
	if err != nil {
		return nil, err
	}
	proc.lastStats = firstStats
	return proc, nil
}

func getProcessesByExe(exe string) (utils.OrderedMap[int, *process], error) {
	matches, err := filepath.Glob("/proc/[0-9]*/exe")
	if err != nil {
		return nil, err
	}

	procs := utils.NewOrderedMap[int, *process]()
	for _, exePath := range matches {
		target, err := os.Readlink(exePath)
		if err != nil {
			continue
		}
		if target == exe {
			pidStr := filepath.Base(filepath.Dir(exePath))
			pid, err := strconv.ParseInt(pidStr, 10, 32)
			if err != nil {
				return nil, err
			}
			proc, err := newProcessWithExe(int(pid), exe)
			if err != nil {
				// TODO: Log the error...
				continue
			}
			procs.Set(int(pid), proc)
		}
	}

	if procs.Len() == 0 {
		return nil, ErrProcessNotFound
	}
	return procs, nil
}

func getProcessesByName(name string) (utils.OrderedMap[int, *process], error) {
	matches, err := filepath.Glob("/proc/[0-9]*/cmdline")
	if err != nil {
		return nil, err
	}

	procs := utils.NewOrderedMap[int, *process]()
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			return nil, err
		}
		cmdline := string(data)
		target := strings.Split(cmdline, "\x00")[0]
		binaryName := filepath.Base(target)
		if binaryName == name {
			pidStr := filepath.Base(filepath.Dir(match))
			pid, err := strconv.ParseInt(pidStr, 10, 64)
			if err != nil {
				return nil, err
			}
			proc, err := newProcessWithName(int(pid), name)
			if err != nil {
				// TODO: Log the error...
				continue
			}
			procs.Set(int(pid), proc)
		}
	}

	if procs.Len() == 0 {
		return nil, ErrProcessNotFound
	}
	return procs, nil
}

func (p *process) parseStatFile() (*processStat, error) {
	filePath := fmt.Sprintf("/proc/%d/stat", p.Pid)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 52 {
			return nil, fmt.Errorf("unexpected number of fields in /proc/%d/stat", p.Pid)
		}

		processStat := &processStat{}
		processStat.PID, _ = strconv.Atoi(fields[0])
		processStat.Comm = strings.Trim(fields[1], "()")
		processStat.State = fields[2]
		processStat.PPID, _ = strconv.Atoi(fields[3])
		processStat.PGRP, _ = strconv.Atoi(fields[4])
		processStat.Session, _ = strconv.Atoi(fields[5])
		processStat.TtyNr, _ = strconv.Atoi(fields[6])
		processStat.Tpgid, _ = strconv.Atoi(fields[7])
		processStat.Flags, _ = strconv.Atoi(fields[8])
		processStat.Minflt, _ = strconv.Atoi(fields[9])
		processStat.Cminflt, _ = strconv.Atoi(fields[10])
		processStat.Majflt, _ = strconv.Atoi(fields[11])
		processStat.Cmajflt, _ = strconv.Atoi(fields[12])
		processStat.Utime, _ = strconv.ParseFloat(fields[13], 64)
		processStat.Stime, _ = strconv.ParseFloat(fields[14], 64)
		processStat.Cutime, _ = strconv.Atoi(fields[15])
		processStat.Cstime, _ = strconv.Atoi(fields[16])
		processStat.Priority, _ = strconv.Atoi(fields[17])
		processStat.Nice, _ = strconv.Atoi(fields[18])
		processStat.NumThreads, _ = strconv.Atoi(fields[19])
		processStat.Itrealvalue, _ = strconv.Atoi(fields[20])
		processStat.Starttime, _ = strconv.ParseFloat(fields[21], 64)
		processStat.Vsize, _ = strconv.Atoi(fields[22])
		processStat.Rss, _ = strconv.Atoi(fields[23])
		processStat.Rlim, _ = strconv.Atoi(fields[24])
		processStat.Startcode, _ = strconv.Atoi(fields[25])
		processStat.Endcode, _ = strconv.Atoi(fields[26])
		processStat.Startstack, _ = strconv.Atoi(fields[27])
		processStat.Kstkeip, _ = strconv.Atoi(fields[28])
		processStat.Kstkesp, _ = strconv.Atoi(fields[29])
		processStat.Signal, _ = strconv.Atoi(fields[30])
		processStat.Blocked, _ = strconv.Atoi(fields[31])
		processStat.Sigignore, _ = strconv.Atoi(fields[32])
		processStat.Sigcatch, _ = strconv.Atoi(fields[33])
		processStat.Wchan, _ = strconv.Atoi(fields[34])
		processStat.Nswap, _ = strconv.Atoi(fields[35])
		processStat.Cnswap, _ = strconv.Atoi(fields[36])
		processStat.ExitSignal, _ = strconv.Atoi(fields[37])
		processStat.Processor, _ = strconv.Atoi(fields[38])
		processStat.RtPriority, _ = strconv.Atoi(fields[39])
		processStat.Policy, _ = strconv.Atoi(fields[40])
		processStat.DelayacctBlkio, _ = strconv.Atoi(fields[41])
		processStat.GuestTime, _ = strconv.Atoi(fields[42])
		processStat.CguestTime, _ = strconv.Atoi(fields[43])

		return processStat, nil
	}

	return nil, fmt.Errorf("failed to read stat file for pid %d", p.Pid)
}

func GetMemoryInfo(pid int) (*processMemoryInfo, error) {
	// Path to the /proc/<pid>/status file
	statusFile := fmt.Sprintf("/proc/%d/status", pid)

	// Open the status file
	file, err := os.Open(statusFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", statusFile, err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	memInfo := &processMemoryInfo{}

	for scanner.Scan() {
		line := scanner.Text()

		// Look for memory fields and parse them
		switch {
		case strings.HasPrefix(line, "VmSize:"):
			memInfo.VmSize = parseMemoryField(line)
		case strings.HasPrefix(line, "VmRSS:"):
			memInfo.VmRSS = parseMemoryField(line)
		case strings.HasPrefix(line, "VmData:"):
			memInfo.VmData = parseMemoryField(line)
		case strings.HasPrefix(line, "VmStack:"):
			memInfo.VmStack = parseMemoryField(line)
		case strings.HasPrefix(line, "VmSwap:"):
			memInfo.VmSwap = parseMemoryField(line)
		case strings.HasPrefix(line, "VmPeak:"):
			memInfo.VmPeak = parseMemoryField(line)
		case strings.HasPrefix(line, "VmHWM:"):
			memInfo.VmHWM = parseMemoryField(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %v", statusFile, err)
	}

	return memInfo, nil
}

// parseMemoryField parses memory information fields (e.g., "VmSize: 1234 kB")
func parseMemoryField(line string) int {
	fields := strings.Fields(line)
	if len(fields) > 1 {
		// Convert memory value to integer (the second field contains the value)
		value, err := strconv.Atoi(fields[1])
		if err == nil {
			return value
		}
	}
	return 0
}

func (p *process) getSystemUptimeSeconds() (float64, error) {
	// Read the /proc/uptime file to get system uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	// Parse the first field (system uptime)
	fields := strings.Fields(string(data))
	systemUptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	// Convert system uptime from seconds to clock ticks (assuming 100Hz jiffies)
	// 1 second = 100 jiffies (on most Linux systems)
	// The system may be configured with a different tick rate, but 100 is common
	return systemUptime, nil
}
