// Package resource provides domain types for resource pool management.
// This is a pure domain package with no external dependencies.
package resource

// PoolType identifies which resource pool a work unit uses.
type PoolType string

const (
	// PoolHost represents the host system resource pool.
	// Used by native tools like Go compiler, npm, linters.
	PoolHost PoolType = "host"

	// PoolDocker represents the Docker/container resource pool.
	// A subset of host memory. Container tools use BOTH pools.
	PoolDocker PoolType = "docker"
)

// PoolAllocation describes the resource pools a work unit needs.
type PoolAllocation struct {
	// HostWeight is the weight to acquire from the host pool.
	// All work units consume host resources.
	HostWeight int

	// DockerWeight is the weight to acquire from the docker pool.
	// Only container tools consume docker resources.
	// Zero means host-only execution.
	DockerWeight int
}

// GetHostWeight returns the weight to acquire from the host pool.
func (a PoolAllocation) GetHostWeight() int {
	return a.HostWeight
}

// GetDockerWeight returns the weight to acquire from the docker pool.
func (a PoolAllocation) GetDockerWeight() int {
	return a.DockerWeight
}

// IsContainer returns true if this allocation requires docker resources.
func (a PoolAllocation) IsContainer() bool {
	return a.DockerWeight > 0
}

// TotalWeight returns the effective scheduling weight.
// For bin-packing, we use the maximum of the two weights
// since both must be satisfied simultaneously.
func (a PoolAllocation) TotalWeight() int {
	if a.DockerWeight > a.HostWeight {
		return a.DockerWeight
	}
	return a.HostWeight
}

// HostOnlyAllocation creates an allocation for host-only work.
func HostOnlyAllocation(weight int) PoolAllocation {
	return PoolAllocation{
		HostWeight:   weight,
		DockerWeight: 0,
	}
}

// ContainerAllocation creates an allocation for container work.
// Container work requires capacity from both pools.
func ContainerAllocation(hostWeight, dockerWeight int) PoolAllocation {
	return PoolAllocation{
		HostWeight:   hostWeight,
		DockerWeight: dockerWeight,
	}
}
