package resource

import (
	"testing"
)

func TestCalculateCapacity_HostOnly(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 32 * 1024 * 1024 * 1024, // 32GB
		CPUCount:        8,
		DockerMemoryBytes: 0, // No Docker
	}
	cfg := DefaultCapacityConfig()

	host, docker := calc.CalculateCapacity(res, cfg)

	// With 32GB RAM and 8 CPUs:
	// RAM capacity = 32GB / 256MB = 128 slots
	// CPU capacity = 8
	// Base = min(128, 8) = 8
	if host.Total != 8 {
		t.Errorf("host.Total = %d, want 8 (min of CPU and RAM capacity)", host.Total)
	}
	if docker.Total != 0 {
		t.Errorf("docker.Total = %d, want 0 (no docker)", docker.Total)
	}
	if host.Pool != PoolHost {
		t.Errorf("host.Pool = %v, want %v", host.Pool, PoolHost)
	}
	if docker.Pool != PoolDocker {
		t.Errorf("docker.Pool = %v, want %v", docker.Pool, PoolDocker)
	}
}

func TestCalculateCapacity_WithDocker(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes:   32 * 1024 * 1024 * 1024, // 32GB
		DockerMemoryBytes: 6 * 1024 * 1024 * 1024,  // 6GB
		CPUCount:          8,
	}
	cfg := DefaultCapacityConfig()

	host, docker := calc.CalculateCapacity(res, cfg)

	if host.Total != 8 {
		t.Errorf("host.Total = %d, want 8", host.Total)
	}
	// Docker: 6GB / 1GB = 6 slots
	if docker.Total != 6 {
		t.Errorf("docker.Total = %d, want 6 (6GB / 1GB)", docker.Total)
	}
}

func TestCalculateCapacity_WithTurbo(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes:   32 * 1024 * 1024 * 1024,
		DockerMemoryBytes: 6 * 1024 * 1024 * 1024,
		CPUCount:          8,
	}
	cfg := DefaultCapacityConfig()
	cfg.Turbo = 2.0

	host, docker := calc.CalculateCapacity(res, cfg)

	// Host: 8 * 2.0 = 16 (capped at 2x CPU = 16)
	if host.Total != 16 {
		t.Errorf("host.Total = %d, want 16 (8 * 2.0, capped at 2x CPU)", host.Total)
	}
	// Docker: 6 * 2.0 = 12
	if docker.Total != 12 {
		t.Errorf("docker.Total = %d, want 12 (6 * 2.0)", docker.Total)
	}
}

func TestCalculateCapacity_WithHostRoof(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 32 * 1024 * 1024 * 1024,
		CPUCount:        8,
	}
	cfg := DefaultCapacityConfig()
	cfg.HostRoof = 4 // Explicit override

	host, _ := calc.CalculateCapacity(res, cfg)

	if host.Total != 4 {
		t.Errorf("host.Total = %d, want 4 (roof override)", host.Total)
	}
}

func TestCalculateCapacity_WithDockerRoof(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes:   32 * 1024 * 1024 * 1024,
		DockerMemoryBytes: 6 * 1024 * 1024 * 1024,
		CPUCount:          8,
	}
	cfg := DefaultCapacityConfig()
	cfg.DockerRoof = 2 // Explicit override

	_, docker := calc.CalculateCapacity(res, cfg)

	if docker.Total != 2 {
		t.Errorf("docker.Total = %d, want 2 (roof override)", docker.Total)
	}
}

func TestCalculateCapacity_LowMemoryHost(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 1 * 1024 * 1024 * 1024, // Only 1GB
		CPUCount:        8,
	}
	cfg := DefaultCapacityConfig()

	host, _ := calc.CalculateCapacity(res, cfg)

	// RAM capacity = 1GB / 256MB = 4 slots
	// CPU capacity = 8
	// Base = min(4, 8) = 4
	if host.Total != 4 {
		t.Errorf("host.Total = %d, want 4 (limited by RAM)", host.Total)
	}
}

func TestCalculateCapacity_SmallDockerMemory(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes:   32 * 1024 * 1024 * 1024,
		DockerMemoryBytes: 512 * 1024 * 1024, // Only 512MB
		CPUCount:          8,
	}
	cfg := DefaultCapacityConfig()

	_, docker := calc.CalculateCapacity(res, cfg)

	// 512MB / 1GB = 0.5, rounds down but minimum is 1
	if docker.Total != 1 {
		t.Errorf("docker.Total = %d, want 1 (minimum)", docker.Total)
	}
}

func TestCalculateCapacity_MinimumValues(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes:   100 * 1024 * 1024, // 100MB (very low)
		DockerMemoryBytes: 100 * 1024 * 1024, // 100MB
		CPUCount:          1,
	}
	cfg := DefaultCapacityConfig()

	host, docker := calc.CalculateCapacity(res, cfg)

	// Should enforce minimum of 1
	if host.Total < 1 {
		t.Errorf("host.Total = %d, want at least 1", host.Total)
	}
	if docker.Total < 1 {
		t.Errorf("docker.Total = %d, want at least 1 (when docker available)", docker.Total)
	}
}

func TestCalculateCapacity_TurboLessThanOne(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 32 * 1024 * 1024 * 1024,
		CPUCount:        8,
	}
	cfg := DefaultCapacityConfig()
	cfg.Turbo = 0.5 // Should be treated as 1.0

	host, _ := calc.CalculateCapacity(res, cfg)

	if host.Total != 8 {
		t.Errorf("host.Total = %d, want 8 (turbo < 1.0 should default to 1.0)", host.Total)
	}
}

func TestCalculateCapacity_HighTurboCapped(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 64 * 1024 * 1024 * 1024, // 64GB
		CPUCount:        4,
	}
	cfg := DefaultCapacityConfig()
	cfg.Turbo = 10.0 // Very high turbo

	host, _ := calc.CalculateCapacity(res, cfg)

	// With 4 CPUs, cap is 2x = 8, or max 64
	// 4 * 10.0 = 40, capped at 8
	if host.Total != 8 {
		t.Errorf("host.Total = %d, want 8 (capped at 2x CPU)", host.Total)
	}
}

func TestCalculateCapacity_HighCPUCountCapped(t *testing.T) {
	calc := NewCalculator()
	res := SystemResources{
		HostMemoryBytes: 256 * 1024 * 1024 * 1024, // 256GB
		CPUCount:        128,                       // 128 cores (high-end server)
	}
	cfg := DefaultCapacityConfig()
	cfg.Turbo = 2.0

	host, _ := calc.CalculateCapacity(res, cfg)

	// RAM capacity = 256GB / 256MB = 1024 slots
	// CPU capacity = 128
	// Base = min(1024, 128) = 128
	// With turbo: 128 * 2.0 = 256, capped at min(2*128=256, 64) = 64
	if host.Total != 64 {
		t.Errorf("host.Total = %d, want 64 (max cap)", host.Total)
	}
}

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()
	if calc == nil {
		t.Error("NewCalculator() returned nil")
	}
}
