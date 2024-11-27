package cpu_monitor

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/rinzlerlabs/viam-raspi-sensors/utils"
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

func readCPUStats() (map[string]CPUCoreStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	data, err := utils.ReadFileWithContext(ctx, "/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	stats := make(map[string]CPUCoreStats)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 8 || !strings.HasPrefix(fields[0], "cpu") {
			continue
		}

		// Parse CPU stats
		core := fields[0]
		user, _ := strconv.ParseUint(fields[1], 10, 64)
		nice, _ := strconv.ParseUint(fields[2], 10, 64)
		system, _ := strconv.ParseUint(fields[3], 10, 64)
		idle, _ := strconv.ParseUint(fields[4], 10, 64)
		iowait, _ := strconv.ParseUint(fields[5], 10, 64)
		irq, _ := strconv.ParseUint(fields[6], 10, 64)
		softirq, _ := strconv.ParseUint(fields[7], 10, 64)
		steal := uint64(0)
		if len(fields) > 8 {
			steal, _ = strconv.ParseUint(fields[8], 10, 64)
		}

		stats[core] = CPUCoreStats{
			User:    user,
			Nice:    nice,
			System:  system,
			Idle:    idle,
			IOWait:  iowait,
			IRQ:     irq,
			SoftIRQ: softirq,
			Steal:   steal,
		}
	}

	return stats, nil
}

// calculateUsage calculates CPU usage percentages
func calculateUsage(prev, curr CPUCoreStats) float64 {
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
