package resource

// PoolCapacity represents the capacity of a resource pool.
type PoolCapacity struct {
	// Pool identifies which pool this capacity is for.
	Pool PoolType

	// Total is the maximum capacity (weight units).
	// Calculated from available memory / weight-to-memory ratio.
	Total int

	// Used is the current allocated capacity.
	Used int

	// Waiting is the number of work units waiting for capacity.
	Waiting int
}

// Available returns the remaining capacity.
func (c PoolCapacity) Available() int {
	avail := c.Total - c.Used
	if avail < 0 {
		return 0
	}
	return avail
}

// SystemResources holds detected system resource information.
type SystemResources struct {
	// HostMemoryBytes is total host RAM.
	HostMemoryBytes uint64

	// DockerMemoryBytes is Docker daemon memory limit.
	// Zero if Docker is unavailable.
	DockerMemoryBytes uint64

	// CPUCount is the number of available CPU cores.
	CPUCount int
}

// CapacityConfig holds configuration for capacity calculation.
type CapacityConfig struct {
	// HostRoof is an explicit host capacity override (--roof flag).
	// Zero means auto-detect.
	HostRoof int

	// DockerRoof is an explicit docker capacity override.
	// Zero means auto-detect from Docker memory.
	DockerRoof int

	// Turbo is the parallelism multiplier (1.0, 1.25, 2.0, etc.).
	Turbo float64

	// BytesPerWeight is how many bytes each weight unit represents.
	// Default: 256MB for host pool, 1GB for docker pool.
	HostBytesPerWeight   uint64
	DockerBytesPerWeight uint64
}

// DefaultCapacityConfig returns the default capacity configuration.
func DefaultCapacityConfig() CapacityConfig {
	return CapacityConfig{
		Turbo:                1.0,
		HostBytesPerWeight:   256 * 1024 * 1024,  // 256MB per weight unit
		DockerBytesPerWeight: 1024 * 1024 * 1024, // 1GB per weight unit (heavier builds)
	}
}
