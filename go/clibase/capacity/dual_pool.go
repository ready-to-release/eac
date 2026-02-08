package capacity

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/core/resource"
)

const (
	dockerStateFile     = ".global-docker-capacity.json"
	dockerStateLockFile = ".global-docker-capacity.lock"
)

// DualPoolSemaphore implements resource.PoolAcquirer with two coordinated semaphores.
// Container work must acquire from BOTH pools; host work uses only the host pool.
type DualPoolSemaphore struct {
	workspaceRoot string
	registry      *locktracker.Registry

	hostSem   *GlobalSemaphore // Existing implementation for host pool
	dockerSem *GlobalSemaphore // Docker-specific semaphore (may be nil if disabled)

	hostLockID   string
	dockerLockID string

	mu     sync.Mutex
	closed bool

	// Waiting counters for TUI display
	hostWaiting   atomic.Int32
	dockerWaiting atomic.Int32

	// Signal handling (shared between both pools)
	sigChan  chan os.Signal
	sigClose chan struct{}
}

// NewDualPoolSemaphore creates a dual-pool semaphore.
// If dockerCapacity is 0, docker pool is disabled (all work uses host only).
func NewDualPoolSemaphore(
	workspaceRoot string,
	hostCapacity int,
	dockerCapacity int,
	registry *locktracker.Registry,
) *DualPoolSemaphore {
	dps := &DualPoolSemaphore{
		workspaceRoot: workspaceRoot,
		registry:      registry,
		sigChan:       make(chan os.Signal, 1),
		sigClose:      make(chan struct{}),
	}

	// Ensure state directory exists
	os.MkdirAll(filepath.Join(workspaceRoot, stateDir), 0o755)

	// Create host semaphore using existing implementation (without signal handling)
	// Name "component-scheduler" preserved for TUI backward compatibility
	dps.hostSem = newGlobalSemaphoreInternal(workspaceRoot, hostCapacity, registry, "component-scheduler", stateFile, stateLockFile)
	dps.hostLockID = dps.hostSem.lockID

	// Create docker semaphore if enabled
	if dockerCapacity > 0 {
		dps.dockerSem = newGlobalSemaphoreInternal(workspaceRoot, dockerCapacity, registry, "docker-scheduler", dockerStateFile, dockerStateLockFile)
		dps.dockerLockID = dps.dockerSem.lockID
	}

	// Register unified signal handler for graceful shutdown
	signal.Notify(dps.sigChan, os.Interrupt, syscall.SIGTERM)
	go dps.signalHandler()

	return dps
}

// newGlobalSemaphoreInternal creates a GlobalSemaphore with custom file names.
// Does NOT register signal handler (caller handles signals).
func newGlobalSemaphoreInternal(workspaceRoot string, capacity int, registry *locktracker.Registry, name, stateFileName, lockFileName string) *GlobalSemaphore {
	gs := &GlobalSemaphore{
		workspaceRoot: workspaceRoot,
		capacity:      capacity,
		registry:      registry,
		lockID:        uuid.New().String(),
		allocations:   make(map[string]int),
		sigChan:       nil, // No signal handling
		sigClose:      nil,
	}

	// Store custom file names (we'll use these via method calls)
	gs.stateFileName = stateFileName
	gs.lockFileName = lockFileName

	// Register with lock tracker for TUI display
	if registry != nil {
		registry.Register(&locktracker.LockInfo{
			ID:       gs.lockID,
			Type:     locktracker.LockTypeWeighted,
			Name:     name,
			Capacity: int64(capacity),
			Used:     0,
		})
	}

	// Clean stale allocations on startup
	gs.cleanStale()

	return gs
}

// signalHandler listens for signals and cleans up allocations.
func (dps *DualPoolSemaphore) signalHandler() {
	// Capture the channel reference under the mutex
	dps.mu.Lock()
	sigClose := dps.sigClose
	dps.mu.Unlock()

	if sigClose == nil {
		return
	}

	select {
	case <-dps.sigChan:
		// On signal, release all allocations then stop listening
		dps.Close()
		signal.Stop(dps.sigChan)
	case <-sigClose:
		// Normal close, stop listening
		signal.Stop(dps.sigChan)
	}
}

// Acquire blocks until allocation can be satisfied from all required pools.
func (dps *DualPoolSemaphore) Acquire(ctx context.Context, alloc resource.PoolAllocation) bool {
	// All work must acquire from host pool - enforce minimum weight of 1
	// This prevents accidental bypass of the semaphore when HostWeight is 0
	hostWeight := alloc.HostWeight
	if hostWeight < 1 {
		hostWeight = 1
	}
	dps.hostWaiting.Add(1)
	dps.hostSem.Acquire(hostWeight)
	dps.hostWaiting.Add(-1)

	// If container work, also acquire docker capacity
	if alloc.DockerWeight > 0 && dps.dockerSem != nil {
		dps.dockerWaiting.Add(1)
		dps.dockerSem.Acquire(alloc.DockerWeight)
		dps.dockerWaiting.Add(-1)
	}

	return true
}

// Release returns capacity to the pools.
func (dps *DualPoolSemaphore) Release(alloc resource.PoolAllocation) {
	// Release in reverse order (docker first, then host)
	if alloc.DockerWeight > 0 && dps.dockerSem != nil {
		dps.dockerSem.Release(alloc.DockerWeight)
	}

	// All work acquired from host pool with minimum weight 1
	hostWeight := alloc.HostWeight
	if hostWeight < 1 {
		hostWeight = 1
	}
	dps.hostSem.Release(hostWeight)
}

// HostCapacity returns current host pool status.
func (dps *DualPoolSemaphore) HostCapacity() resource.PoolCapacity {
	return resource.PoolCapacity{
		Pool:    resource.PoolHost,
		Total:   dps.hostSem.Capacity(),
		Used:    dps.hostSem.Used(),
		Waiting: int(dps.hostWaiting.Load()),
	}
}

// DockerCapacity returns current docker pool status.
func (dps *DualPoolSemaphore) DockerCapacity() resource.PoolCapacity {
	if dps.dockerSem == nil {
		return resource.PoolCapacity{Pool: resource.PoolDocker}
	}
	return resource.PoolCapacity{
		Pool:    resource.PoolDocker,
		Total:   dps.dockerSem.Capacity(),
		Used:    dps.dockerSem.Used(),
		Waiting: int(dps.dockerWaiting.Load()),
	}
}

// SetCapacity updates pool capacities.
func (dps *DualPoolSemaphore) SetCapacity(host, docker int) {
	dps.hostSem.SetCapacity(host)
	if dps.dockerSem != nil && docker > 0 {
		dps.dockerSem.SetCapacity(docker)
	}
}

// Close releases all resources.
func (dps *DualPoolSemaphore) Close() {
	dps.mu.Lock()
	if dps.closed {
		dps.mu.Unlock()
		return
	}
	dps.closed = true

	// Stop signal handler
	if dps.sigClose != nil {
		close(dps.sigClose)
		dps.sigClose = nil
	}
	dps.mu.Unlock()

	// Close both semaphores
	if dps.hostSem != nil {
		dps.hostSem.Close()
	}
	if dps.dockerSem != nil {
		dps.dockerSem.Close()
	}
}
