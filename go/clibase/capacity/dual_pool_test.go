package capacity

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/resource"
)

func TestDualPoolSemaphore_AcquireHostOnly(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)
	defer dps.Close()

	alloc := resource.PoolAllocation{HostWeight: 2, DockerWeight: 0}
	acquired := dps.Acquire(context.Background(), alloc)

	if !acquired {
		t.Error("Acquire should return true")
	}

	hostCap := dps.HostCapacity()
	if hostCap.Used != 2 {
		t.Errorf("host.Used = %d, want 2", hostCap.Used)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Used != 0 {
		t.Errorf("docker.Used = %d, want 0", dockerCap.Used)
	}
}

func TestDualPoolSemaphore_AcquireContainer(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)
	defer dps.Close()

	alloc := resource.PoolAllocation{HostWeight: 2, DockerWeight: 2}
	acquired := dps.Acquire(context.Background(), alloc)

	if !acquired {
		t.Error("Acquire should return true")
	}

	hostCap := dps.HostCapacity()
	if hostCap.Used != 2 {
		t.Errorf("host.Used = %d, want 2", hostCap.Used)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Used != 2 {
		t.Errorf("docker.Used = %d, want 2", dockerCap.Used)
	}
}

func TestDualPoolSemaphore_Release(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)
	defer dps.Close()

	alloc := resource.PoolAllocation{HostWeight: 3, DockerWeight: 2}
	dps.Acquire(context.Background(), alloc)
	dps.Release(alloc)

	hostCap := dps.HostCapacity()
	if hostCap.Used != 0 {
		t.Errorf("host.Used = %d, want 0 after release", hostCap.Used)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Used != 0 {
		t.Errorf("docker.Used = %d, want 0 after release", dockerCap.Used)
	}
}

func TestDualPoolSemaphore_DockerPoolDisabled(t *testing.T) {
	dir := t.TempDir()

	// Docker pool disabled (capacity=0)
	dps := NewDualPoolSemaphore(dir, 10, 0, nil)
	defer dps.Close()

	// Host-only allocation should work
	alloc := resource.PoolAllocation{HostWeight: 2, DockerWeight: 0}
	acquired := dps.Acquire(context.Background(), alloc)

	if !acquired {
		t.Error("Acquire should return true")
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Total != 0 {
		t.Errorf("docker.Total = %d, want 0 (disabled)", dockerCap.Total)
	}
}

func TestDualPoolSemaphore_Capacities(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 8, 4, nil)
	defer dps.Close()

	hostCap := dps.HostCapacity()
	if hostCap.Total != 8 {
		t.Errorf("host.Total = %d, want 8", hostCap.Total)
	}
	if hostCap.Pool != resource.PoolHost {
		t.Errorf("host.Pool = %v, want %v", hostCap.Pool, resource.PoolHost)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Total != 4 {
		t.Errorf("docker.Total = %d, want 4", dockerCap.Total)
	}
	if dockerCap.Pool != resource.PoolDocker {
		t.Errorf("docker.Pool = %v, want %v", dockerCap.Pool, resource.PoolDocker)
	}
}

func TestDualPoolSemaphore_SetCapacity(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 8, 4, nil)
	defer dps.Close()

	dps.SetCapacity(12, 6)

	hostCap := dps.HostCapacity()
	if hostCap.Total != 12 {
		t.Errorf("host.Total = %d, want 12", hostCap.Total)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Total != 6 {
		t.Errorf("docker.Total = %d, want 6", dockerCap.Total)
	}
}

func TestDualPoolSemaphore_MultipleAllocations(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 6, nil)
	defer dps.Close()

	// Host-only allocation
	alloc1 := resource.PoolAllocation{HostWeight: 2, DockerWeight: 0}
	dps.Acquire(context.Background(), alloc1)

	// Container allocation
	alloc2 := resource.PoolAllocation{HostWeight: 3, DockerWeight: 4}
	dps.Acquire(context.Background(), alloc2)

	// Another host-only
	alloc3 := resource.PoolAllocation{HostWeight: 1, DockerWeight: 0}
	dps.Acquire(context.Background(), alloc3)

	hostCap := dps.HostCapacity()
	if hostCap.Used != 6 { // 2 + 3 + 1
		t.Errorf("host.Used = %d, want 6", hostCap.Used)
	}

	dockerCap := dps.DockerCapacity()
	if dockerCap.Used != 4 { // only alloc2
		t.Errorf("docker.Used = %d, want 4", dockerCap.Used)
	}

	// Release container allocation
	dps.Release(alloc2)

	hostCap = dps.HostCapacity()
	if hostCap.Used != 3 { // 2 + 1
		t.Errorf("host.Used after release = %d, want 3", hostCap.Used)
	}

	dockerCap = dps.DockerCapacity()
	if dockerCap.Used != 0 {
		t.Errorf("docker.Used after release = %d, want 0", dockerCap.Used)
	}
}

func TestDualPoolSemaphore_SeparateStateFiles(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)
	defer dps.Close()

	// Make some allocations
	dps.Acquire(context.Background(), resource.PoolAllocation{HostWeight: 2, DockerWeight: 2})

	// Verify separate state files exist
	hostStateFile := filepath.Join(dir, stateDir, ".global-capacity.json")
	dockerStateFile := filepath.Join(dir, stateDir, ".global-docker-capacity.json")

	if _, err := os.Stat(hostStateFile); os.IsNotExist(err) {
		t.Error("host state file should exist")
	}

	if _, err := os.Stat(dockerStateFile); os.IsNotExist(err) {
		t.Error("docker state file should exist")
	}
}

func TestDualPoolSemaphore_CloseMultipleTimes(t *testing.T) {
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)

	// Close multiple times should not panic
	dps.Close()
	dps.Close()
	dps.Close()
}

func TestDualPoolSemaphore_ZeroWeightEnforcesMinimum(t *testing.T) {
	// Verify that zero-weight allocations still acquire minimum weight of 1
	// This prevents accidental bypass of the semaphore
	dir := t.TempDir()

	dps := NewDualPoolSemaphore(dir, 10, 4, nil)
	defer dps.Close()

	// Zero-weight allocation should still acquire weight 1
	alloc := resource.PoolAllocation{HostWeight: 0, DockerWeight: 0}
	acquired := dps.Acquire(context.Background(), alloc)

	if !acquired {
		t.Error("Acquire should return true")
	}

	hostCap := dps.HostCapacity()
	if hostCap.Used != 1 {
		t.Errorf("host.Used = %d, want 1 (minimum enforced)", hostCap.Used)
	}

	// Release should also use minimum weight 1
	dps.Release(alloc)

	hostCap = dps.HostCapacity()
	if hostCap.Used != 0 {
		t.Errorf("host.Used after release = %d, want 0", hostCap.Used)
	}
}
