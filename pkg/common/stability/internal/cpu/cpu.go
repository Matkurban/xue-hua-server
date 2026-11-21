package cpu

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	// Sampling interval
	interval time.Duration = time.Millisecond * 500
)

var (
	stats CPU
	usage uint64
)

// CPU defines CPU statistics interface
type CPU interface {
	// Get CPU usage rate, returned value is 1000 times the actual usage rate (e.g., 50.5% returns 505)
	Usage() (u uint64, e error)
	// Get CPU information
	Info() Info
}

// Info CPU information structure
type Info struct {
	Frequency uint64  // CPU frequency (Hz)
	Quota     float64 // CPU quota (if limited)
	Cores     int     // Logical core count
}

// Stat CPU usage statistics
type Stat struct {
	Usage uint64 // CPU usage rate multiplied by 1000
}

type noopCPU struct {
	cores int
}

func (n *noopCPU) Usage() (uint64, error) { return 0, nil }

func (n *noopCPU) Info() Info {
	return Info{
		Cores: n.cores,
		Quota: float64(n.cores),
	}
}

// Initialize CPU monitoring
func init() {
	var err error

	// Choose appropriate CPU stats implementation based on platform
	stats, err = newCPUStats()
	if err != nil {
		// Gracefully degrade (e.g. macOS with CGO_ENABLED=0 or restricted sandboxes)
		// instead of panicking during package initialization.
		stats = &noopCPU{cores: runtime.NumCPU()}
		return
	}

	// Start background monitoring goroutine
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			u, err := stats.Usage()
			if err == nil && u != 0 {
				atomic.StoreUint64(&usage, u)
			}
		}
	}()
}

// newCPUStats creates appropriate CPU monitoring implementation
func newCPUStats() (CPU, error) {
	// Try cgroup implementation first on Linux
	if runtime.GOOS == "linux" {
		cgroupCPU, err := NewCgroupCPU()
		if err == nil {
			return cgroupCPU, nil
		}
		// Log the error but continue to fallback
		fmt.Printf("Cgroup CPU monitoring not available: %v, falling back to psutil\n", err)
	}

	// Fallback to cross-platform psutil implementation
	return NewPsutilCPU(interval)
}

// ReadStat reads current CPU usage
func ReadStat(stat *Stat) {
	stat.Usage = atomic.LoadUint64(&usage)
}

// GetInfo gets CPU information
func GetInfo() Info {
	return stats.Info()
}
