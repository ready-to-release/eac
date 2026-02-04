package resource

import (
	"testing"
)

func TestPoolType_Constants(t *testing.T) {
	// Verify pool type constants are properly defined
	if PoolHost != "host" {
		t.Errorf("PoolHost = %q, want %q", PoolHost, "host")
	}
	if PoolDocker != "docker" {
		t.Errorf("PoolDocker = %q, want %q", PoolDocker, "docker")
	}
}

func TestPoolAllocation_IsContainer(t *testing.T) {
	tests := []struct {
		name   string
		alloc  PoolAllocation
		want   bool
	}{
		{
			name:   "host only allocation",
			alloc:  PoolAllocation{HostWeight: 2, DockerWeight: 0},
			want:   false,
		},
		{
			name:   "container allocation",
			alloc:  PoolAllocation{HostWeight: 2, DockerWeight: 2},
			want:   true,
		},
		{
			name:   "container with different weights",
			alloc:  PoolAllocation{HostWeight: 4, DockerWeight: 2},
			want:   true,
		},
		{
			name:   "zero weights",
			alloc:  PoolAllocation{HostWeight: 0, DockerWeight: 0},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.IsContainer(); got != tt.want {
				t.Errorf("IsContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoolAllocation_TotalWeight(t *testing.T) {
	tests := []struct {
		name   string
		alloc  PoolAllocation
		want   int
	}{
		{
			name:   "host only",
			alloc:  PoolAllocation{HostWeight: 4, DockerWeight: 0},
			want:   4,
		},
		{
			name:   "docker higher",
			alloc:  PoolAllocation{HostWeight: 2, DockerWeight: 4},
			want:   4,
		},
		{
			name:   "host higher",
			alloc:  PoolAllocation{HostWeight: 6, DockerWeight: 2},
			want:   6,
		},
		{
			name:   "equal weights",
			alloc:  PoolAllocation{HostWeight: 3, DockerWeight: 3},
			want:   3,
		},
		{
			name:   "zero weights",
			alloc:  PoolAllocation{HostWeight: 0, DockerWeight: 0},
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alloc.TotalWeight(); got != tt.want {
				t.Errorf("TotalWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPoolAllocation_HostOnlyAllocation(t *testing.T) {
	alloc := HostOnlyAllocation(3)

	if alloc.HostWeight != 3 {
		t.Errorf("HostWeight = %d, want 3", alloc.HostWeight)
	}
	if alloc.DockerWeight != 0 {
		t.Errorf("DockerWeight = %d, want 0", alloc.DockerWeight)
	}
	if alloc.IsContainer() {
		t.Error("HostOnlyAllocation should not be a container")
	}
}

func TestPoolAllocation_ContainerAllocation(t *testing.T) {
	alloc := ContainerAllocation(2, 4)

	if alloc.HostWeight != 2 {
		t.Errorf("HostWeight = %d, want 2", alloc.HostWeight)
	}
	if alloc.DockerWeight != 4 {
		t.Errorf("DockerWeight = %d, want 4", alloc.DockerWeight)
	}
	if !alloc.IsContainer() {
		t.Error("ContainerAllocation should be a container")
	}
}
