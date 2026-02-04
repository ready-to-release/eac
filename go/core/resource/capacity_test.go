package resource

import (
	"testing"
)

func TestPoolCapacity_Available(t *testing.T) {
	tests := []struct {
		name     string
		capacity PoolCapacity
		want     int
	}{
		{
			name:     "normal capacity",
			capacity: PoolCapacity{Pool: PoolHost, Total: 10, Used: 3},
			want:     7,
		},
		{
			name:     "full capacity",
			capacity: PoolCapacity{Pool: PoolHost, Total: 8, Used: 8},
			want:     0,
		},
		{
			name:     "over capacity returns zero",
			capacity: PoolCapacity{Pool: PoolDocker, Total: 4, Used: 6},
			want:     0,
		},
		{
			name:     "empty pool",
			capacity: PoolCapacity{Pool: PoolHost, Total: 10, Used: 0},
			want:     10,
		},
		{
			name:     "zero total",
			capacity: PoolCapacity{Pool: PoolDocker, Total: 0, Used: 0},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.capacity.Available(); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultCapacityConfig(t *testing.T) {
	cfg := DefaultCapacityConfig()

	if cfg.Turbo != 1.0 {
		t.Errorf("Turbo = %v, want 1.0", cfg.Turbo)
	}
	if cfg.HostBytesPerWeight != 256*1024*1024 {
		t.Errorf("HostBytesPerWeight = %v, want 256MB", cfg.HostBytesPerWeight)
	}
	if cfg.DockerBytesPerWeight != 1024*1024*1024 {
		t.Errorf("DockerBytesPerWeight = %v, want 1GB", cfg.DockerBytesPerWeight)
	}
	if cfg.HostRoof != 0 {
		t.Errorf("HostRoof = %v, want 0 (auto)", cfg.HostRoof)
	}
	if cfg.DockerRoof != 0 {
		t.Errorf("DockerRoof = %v, want 0 (auto)", cfg.DockerRoof)
	}
}

func TestSystemResources_Defaults(t *testing.T) {
	res := SystemResources{}

	if res.HostMemoryBytes != 0 {
		t.Errorf("HostMemoryBytes = %v, want 0", res.HostMemoryBytes)
	}
	if res.DockerMemoryBytes != 0 {
		t.Errorf("DockerMemoryBytes = %v, want 0", res.DockerMemoryBytes)
	}
	if res.CPUCount != 0 {
		t.Errorf("CPUCount = %v, want 0", res.CPUCount)
	}
}
