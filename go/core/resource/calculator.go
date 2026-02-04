package resource

// DefaultCalculator implements CapacityCalculator with standard formulas.
type DefaultCalculator struct{}

// NewCalculator creates a new capacity calculator.
func NewCalculator() *DefaultCalculator {
	return &DefaultCalculator{}
}

// CalculateCapacity computes pool capacities.
//
// Host capacity formula:
//   - If roof > 0: use roof directly
//   - Else: min(CPU, RAM/BytesPerWeight) * turbo
//
// Docker capacity formula:
//   - If roof > 0: use roof directly
//   - If no Docker: return 0 (docker pool disabled)
//   - Else: DockerMem/BytesPerWeight * turbo
func (c *DefaultCalculator) CalculateCapacity(res SystemResources, cfg CapacityConfig) (host, docker PoolCapacity) {
	host = PoolCapacity{Pool: PoolHost}
	docker = PoolCapacity{Pool: PoolDocker}

	turbo := cfg.Turbo
	if turbo < 1.0 {
		turbo = 1.0
	}

	// Host pool calculation
	if cfg.HostRoof > 0 {
		host.Total = cfg.HostRoof
	} else {
		// RAM-based capacity
		ramCap := int(res.HostMemoryBytes / cfg.HostBytesPerWeight)
		if ramCap < 1 {
			ramCap = 1
		}

		// CPU-based capacity
		cpuCap := res.CPUCount

		// Use minimum of RAM and CPU, then apply turbo
		base := cpuCap
		if ramCap < base {
			base = ramCap
		}
		host.Total = int(float64(base) * turbo)

		// Cap at reasonable maximum (2x CPU with turbo, max 64)
		maxCap := cpuCap * 2
		if maxCap > 64 {
			maxCap = 64
		}
		if host.Total > maxCap {
			host.Total = maxCap
		}
	}

	if host.Total < 1 {
		host.Total = 1
	}

	// Docker pool calculation
	if res.DockerMemoryBytes == 0 {
		// No Docker available - docker pool disabled
		docker.Total = 0
		return host, docker
	}

	if cfg.DockerRoof > 0 {
		docker.Total = cfg.DockerRoof
	} else {
		docker.Total = int(float64(res.DockerMemoryBytes/cfg.DockerBytesPerWeight) * turbo)
	}

	if docker.Total < 1 && res.DockerMemoryBytes > 0 {
		docker.Total = 1
	}

	return host, docker
}
