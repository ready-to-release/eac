package orchestrator

import (
	"time"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/shirou/gopsutil/v3/mem"
)

// startCapacityTicker starts a goroutine that recalculates capacity every 2 seconds
// based on available system resources. The config max is used as a ceiling.
func (us *UnitScheduler) startCapacityTicker() {
	us.capacityTicker = time.NewTicker(config.CapacityRecalcInterval())

	go func() {
		for {
			select {
			case <-us.capacityTicker.C:
				hostCap := detectAvailableCapacityWith(us.detector, us.configMax, us.turbo)
				dockerCap := detectDockerCapacity(us.turbo)
				us.semaphore.SetCapacity(hostCap, dockerCap)
			case <-us.capacityStop:
				us.capacityTicker.Stop()
				return
			}
		}
	}()
}

// StopCapacityTicker stops the dynamic capacity ticker.
// Safe to call multiple times; the channel is closed exactly once.
func (us *UnitScheduler) StopCapacityTicker() {
	us.capacityOnce.Do(func() {
		close(us.capacityStop)
	})
}

// CapacityDetector abstracts system resource detection for testability.
type CapacityDetector interface {
	EffectiveCPUs() int
	EffectiveMemoryGB() int
}

// systemCapacityDetector uses real system detection.
type systemCapacityDetector struct{}

func (d systemCapacityDetector) EffectiveCPUs() int {
	return GetEffectiveCPUs()
}

func (d systemCapacityDetector) EffectiveMemoryGB() int {
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		return int(memInfo.Total / (1024 * 1024 * 1024))
	}
	return 8 // Hard fallback
}

// newSystemCapacityDetector returns the default production capacity detector.
func newSystemCapacityDetector() CapacityDetector {
	return systemCapacityDetector{}
}

// detectAvailableCapacity calculates the pressure roof for parallel builds
// using the default system detector.
// If --roof is set, uses that value directly. Otherwise auto-detects from system resources.
func detectAvailableCapacity(configMax int, turbo float64) int {
	return detectAvailableCapacityWith(newSystemCapacityDetector(), configMax, turbo)
}

// detectAvailableCapacityWith is the testable implementation that accepts an injected detector.
func detectAvailableCapacityWith(detector CapacityDetector, configMax int, turbo float64) int {
	cpuCount := detector.EffectiveCPUs()
	if cpuCount < 1 {
		cpuCount = 4 // Fallback
	}

	ramGB := detector.EffectiveMemoryGB()
	if ramGB < 1 {
		ramGB = 8 // Hard fallback
	}

	capacity := calculateCapacity(cpuCount, ramGB, configMax, turbo)

	logging.C().Debugf("[scheduler] Capacity calculation: cpuCount=%d ramGB=%d configMax=%d turbo=%.2f → capacity=%d",
		cpuCount, ramGB, configMax, turbo, capacity)

	return capacity
}

// calculateCapacity computes the effective parallelism capacity.
//
// If roof > 0: use roof directly (--roof flag overrides everything)
// If roof == 0: auto-detect from system resources
func calculateCapacity(cpuCount, ramGB, roof int, turbo float64) int {
	// --roof overrides all
	if roof > 0 {
		logging.C().Debugf("[scheduler] Using explicit roof: %d", roof)
		return roof
	}

	// Auto-detect: min(CPU, RAM/2) × turbo
	// RAM/2 because each parallel unit uses ~1.5-2GB for typical Go builds
	// This balances parallelism with memory safety on constrained systems
	ramCap := ramGB / 2
	if ramCap < 1 {
		ramCap = 1
	}

	base := cpuCount
	if ramCap < base {
		logging.C().Debugf("[scheduler] RAM-limited: ramCap=%d < cpuCount=%d, using ramCap", ramCap, cpuCount)
		base = ramCap
	} else {
		logging.C().Debugf("[scheduler] CPU-limited: ramCap=%d >= cpuCount=%d, using cpuCount", ramCap, cpuCount)
	}

	capacity := int(float64(base) * turbo)
	logging.C().Debugf("[scheduler] Base capacity=%d x turbo=%.2f = %d", base, turbo, capacity)

	// Cap at RAM limit FIRST (safety), then CPU limit
	// This prevents turbo from overcommitting memory
	ramMax := ramGB / 2
	if capacity > ramMax && ramMax > 0 {
		logging.C().Debugf("[scheduler] Turbo exceeded RAM limit: %d -> %d (ramGB=%d)", capacity, ramMax, ramGB)
		capacity = ramMax
	}

	// Cap at CPU count (or 2x with turbo, max 64)
	maxCap := cpuCount
	if turbo > 1.0 {
		maxCap = cpuCount * 2
		if maxCap > 64 {
			maxCap = 64
		}
	}
	if capacity > maxCap {
		capacity = maxCap
	}

	if capacity < 1 {
		capacity = 1
	}

	return capacity
}

// detectDockerCapacity calculates the docker pool capacity.
// Uses Docker daemon's memory limit via `docker info`.
// Returns 0 if Docker is unavailable or has no memory limit.
func detectDockerCapacity(turbo float64) int {
	dockerMem := GetDockerMemoryBytes()
	if dockerMem == 0 {
		logging.C().Debugf("[scheduler] Docker unavailable or no memory limit, docker capacity = 0")
		return 0
	}

	// Docker pool uses 1GB per weight unit (heavier builds)
	// This is more conservative than host pool (256MB per unit)
	const dockerBytesPerWeight = 1024 * 1024 * 1024 // 1GB

	dockerCapGB := int(dockerMem / dockerBytesPerWeight)
	if dockerCapGB < 1 {
		dockerCapGB = 1
	}

	// Apply turbo multiplier but cap at reasonable level
	if turbo > 1.0 {
		dockerCapGB = int(float64(dockerCapGB) * turbo)
	}

	// Cap at 16 concurrent docker builds (practical limit)
	if dockerCapGB > 16 {
		dockerCapGB = 16
	}

	logging.C().Debugf("[scheduler] Docker capacity: dockerMem=%d bytes → capacity=%d",
		dockerMem, dockerCapGB)

	return dockerCapGB
}

