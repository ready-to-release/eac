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
var HostOnlyAllocation = core.HostOnlyAllocation

// ContainerAllocation creates an allocation for container work.
// Container work requires capacity from both pools.
var ContainerAllocation = core.ContainerAllocation
