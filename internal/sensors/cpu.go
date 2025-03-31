package sensors

import (
	"github.com/rinzlerlabs/viam-sbc-hwmonitor/utils"
	"github.com/shirou/gopsutil/v4/cpu"
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
