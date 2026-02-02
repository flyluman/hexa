package tuning

import (
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	maxMemPercentage = 0.90
	minMemBytes      = 64 * 1024 * 1024
)

type Logger interface {
	Info(msg string, keysAndValues ...any)
}

func Tune(logger Logger) {
	tuneCPU(logger)
	tuneMemory(logger)
}

func tuneCPU(logger Logger) {
	quota := getCgroupCPUQuota()

	var procs int
	if quota > 0 {
		procs = int(math.Floor(quota))
		procs = clamp(procs, 1, runtime.NumCPU())
		logger.Info("tuning: tuned cpu from cgroup quota", "gomaxprocs", procs, "quota", quota)
	} else {
		procs = runtime.NumCPU()
		logger.Info("tuning: cgroup cpu quota not found, using default", "gomaxprocs", procs)
	}

	runtime.GOMAXPROCS(procs)
}

func tuneMemory(logger Logger) {
	limit := getCgroupMemoryLimit()
	var memLimit int64

	if limit > 0 {
		memLimit = int64(float64(limit) * maxMemPercentage)
		logger.Info("tuning: tuned memory from cgroup limit", "gomemlimit_bytes", memLimit, "cgroup_limit_bytes", limit)
	} else {
		hostMem := getHostMemory()
		if hostMem <= 0 {
			logger.Info("tuning: unable to determine host memory")
			return
		}
		memLimit = max(int64(float64(hostMem)*maxMemPercentage), minMemBytes)
		logger.Info("tuning: cgroup memory limit not found, using default",
			"gomemlimit_bytes", memLimit,
			"host_memory_bytes", hostMem,
			"percentage", int(maxMemPercentage*100),
		)
	}

	debug.SetMemoryLimit(memLimit)
}

func getCgroupCPUQuota() float64 {
	// CGroup v2
	if data, err := os.ReadFile("/sys/fs/cgroup/cpu.max"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 2 && parts[0] != "max" {
			if q, err := strconv.ParseFloat(parts[0], 64); err == nil {
				if p, err := strconv.ParseFloat(parts[1], 64); err == nil && p > 0 {
					return q / p
				}
			}
		}
	}

	// CGroup v1
	quotaData, errQ := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	periodData, errP := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if errQ == nil && errP == nil {
		if q, err := strconv.ParseFloat(strings.TrimSpace(string(quotaData)), 64); err == nil {
			if p, err := strconv.ParseFloat(strings.TrimSpace(string(periodData)), 64); err == nil && p > 0 {
				return q / p
			}
		}
	}

	return -1
}

func getCgroupMemoryLimit() int64 {
	// CGroup v2
	if data, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		str := strings.TrimSpace(string(data))
		if str != "max" {
			if v, err := strconv.ParseInt(str, 10, 64); err == nil && v > 0 {
				return v
			}
		}
	}

	// CGroup v1
	if data, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		if v, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil && v > 0 && v < math.MaxInt64/2 {
			return v
		}
	}

	return -1
}

func getHostMemory() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return -1
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if kb, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					return kb * 1024
				}
			}
		}
	}

	return -1
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
