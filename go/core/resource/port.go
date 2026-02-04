package resource

import "context"

// CapacityDetector detects system resource capacity.
// This is a PORT interface - implementations are adapters.
type CapacityDetector interface {
	// DetectResources returns current system resource information.
	DetectResources(ctx context.Context) (SystemResources, error)
}

// CapacityCalculator computes pool capacities from system resources.
// This is pure domain logic with no external dependencies.
type CapacityCalculator interface {
	// CalculateCapacity computes pool capacities from resources and config.
	CalculateCapacity(resources SystemResources, config CapacityConfig) (host, docker PoolCapacity)
}

// PoolAcquirer manages capacity allocation for resource pools.
// This is a PORT interface - implementations handle actual semaphores.
type PoolAcquirer interface {
	// Acquire blocks until the allocation can be satisfied from all required pools.
	// Returns true if acquired, false if context cancelled.
	Acquire(ctx context.Context, alloc PoolAllocation) bool

	// Release returns capacity to the pools.
	Release(alloc PoolAllocation)

	// HostCapacity returns current host pool status.
	HostCapacity() PoolCapacity

	// DockerCapacity returns current docker pool status.
	DockerCapacity() PoolCapacity

	// SetCapacity updates pool capacities (for dynamic adjustment).
	SetCapacity(host, docker int)

	// Close releases all resources.
	Close()
}
