// Package resource provides domain types for resource pool management.
// This is a pure domain package with no external dependencies.
package resource

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// PoolType identifies which resource pool a work unit uses.
// Aliased from contracts/core/0.1.0 — the canonical definition.
type PoolType = core.PoolType

const (
	// PoolHost represents the host system resource pool.
	PoolHost = core.PoolHost

	// PoolDocker represents the Docker/container resource pool.
	PoolDocker = core.PoolDocker
)

// PoolAllocation describes the resource pools a work unit needs.
// Aliased from contracts/core/0.1.0 — the canonical definition.
type PoolAllocation = core.PoolAllocation

// HostOnlyAllocation creates an allocation for host-only work.
func HostOnlyAllocation(weight int) PoolAllocation {
	return core.HostOnlyAllocation(weight)
}

// ContainerAllocation creates an allocation for container work.
// Container work requires capacity from both pools.
func ContainerAllocation(hostWeight, dockerWeight int) PoolAllocation {
	return core.ContainerAllocation(hostWeight, dockerWeight)
}

// AllocationForWeight creates a PoolAllocation from a weight and container flag.
func AllocationForWeight(weight int, isContainer bool) PoolAllocation {
	return core.AllocationForWeight(weight, isContainer)
}
