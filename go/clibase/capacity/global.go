// Package capacity provides a global weighted semaphore shared across all processes.
// Uses file locking to coordinate capacity allocation between concurrent build/test/lint/scan commands.
package capacity

import (
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	stateDir      = "out"
	stateFile     = ".global-capacity.json"
	stateLockFile = ".global-capacity.lock"
)

// State is the persisted global capacity state.
type State struct {
	Capacity    int                    `json:"capacity"`
	Allocations map[string]*Allocation `json:"allocations"`
}

// Allocation represents capacity held by a process.
type Allocation struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	Weight    int       `json:"weight"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// GlobalSemaphore is a cross-process weighted semaphore using file-based coordination.
type GlobalSemaphore struct {
	workspaceRoot string
	capacity      int
	registry      *locktracker.Registry
	lockID        string

	// Custom file names (defaults to stateFile/stateLockFile if empty)
	stateFileName string
	lockFileName  string

	mu          sync.Mutex
	allocations map[string]int // local: allocID -> weight
	closed      bool

	// For dynamic capacity updates
	capacityMu sync.RWMutex

	// Signal handling for graceful shutdown
	sigChan  chan os.Signal
	sigClose chan struct{}
}

// NewGlobalSemaphore creates a global semaphore coordinated via filesystem.
// Automatically registers a signal handler to clean up allocations on SIGINT/SIGTERM.
func NewGlobalSemaphore(workspaceRoot string, capacity int, registry *locktracker.Registry) *GlobalSemaphore {
	gs := &GlobalSemaphore{
		workspaceRoot: workspaceRoot,
		capacity:      capacity,
		registry:      registry,
		lockID:        uuid.New().String(),
		allocations:   make(map[string]int),
		sigChan:       make(chan os.Signal, 1),
		sigClose:      make(chan struct{}),
	}

	// Ensure state directory exists
	os.MkdirAll(filepath.Join(workspaceRoot, stateDir), 0o755)

	// Register with lock tracker for TUI display
	if registry != nil {
		registry.Register(&locktracker.LockInfo{
			ID:       gs.lockID,
			Type:     locktracker.LockTypeWeighted,
			Name:     "component-scheduler",
			Capacity: int64(capacity),
			Used:     0,
		})
	}

	// Clean stale allocations on startup (including dead PIDs)
	gs.cleanStale()

	// Register signal handler for graceful shutdown
	signal.Notify(gs.sigChan, os.Interrupt, syscall.SIGTERM)
	go gs.signalHandler()

	return gs
}

// signalHandler listens for signals and cleans up allocations.
func (gs *GlobalSemaphore) signalHandler() {
	// Capture the channel reference under the mutex
	gs.mu.Lock()
	sigClose := gs.sigClose
	gs.mu.Unlock()

	if sigClose == nil {
		return
	}

	select {
	case <-gs.sigChan:
		// On signal, release all allocations then stop listening
		gs.Close()
		signal.Stop(gs.sigChan)
	case <-sigClose:
		// Normal close, stop listening
		signal.Stop(gs.sigChan)
	}
}

// Acquire blocks until the requested weight is available.
func (gs *GlobalSemaphore) Acquire(weight int) {
	gs.capacityMu.RLock()
	currentCap := gs.capacity
	gs.capacityMu.RUnlock()

	// Update waiting count
	gs.updateRegistry(0, 1)

	for {
		allocID, acquired, used := gs.tryAcquireInternal(weight, currentCap)
		if acquired {
			gs.mu.Lock()
			gs.allocations[allocID] = weight
			gs.mu.Unlock()
			gs.updateRegistry(int64(used), -1) // -1 to decrement waiting
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Release returns capacity to the global pool.
func (gs *GlobalSemaphore) Release(weight int) {
	gs.mu.Lock()
	// Find and remove one allocation with this weight
	var allocID string
	for id, w := range gs.allocations {
		if w == weight {
			allocID = id
			delete(gs.allocations, id)
			break
		}
	}
	gs.mu.Unlock()

	if allocID == "" {
		return
	}

	used := gs.releaseInternal(allocID)
	gs.updateRegistry(int64(used), 0)
}

// SetCapacity updates the capacity (for dynamic adjustment).
func (gs *GlobalSemaphore) SetCapacity(newCap int) {
	gs.capacityMu.Lock()
	gs.capacity = newCap
	gs.capacityMu.Unlock()

	if gs.registry != nil {
		gs.registry.UpdateSemaphoreCapacity(gs.lockID, int64(newCap))
	}
}

// Capacity returns current capacity.
func (gs *GlobalSemaphore) Capacity() int {
	gs.capacityMu.RLock()
	defer gs.capacityMu.RUnlock()
	return gs.capacity
}

// Used returns current global usage.
func (gs *GlobalSemaphore) Used() int {
	state := gs.readState()
	var used int
	for _, a := range state.Allocations {
		used += a.Weight
	}
	return used
}

// Available returns current available capacity (capacity - used).
func (gs *GlobalSemaphore) Available() int {
	avail := gs.Capacity() - gs.Used()
	if avail < 0 {
		return 0
	}
	return avail
}

// Close releases all local allocations and unregisters from tracking.
// Safe to call multiple times.
func (gs *GlobalSemaphore) Close() {
	gs.mu.Lock()
	if gs.closed {
		gs.mu.Unlock()
		return
	}
	gs.closed = true
	local := make(map[string]int)
	for k, v := range gs.allocations {
		local[k] = v
	}
	gs.allocations = make(map[string]int)

	// Stop signal handler
	if gs.sigClose != nil {
		close(gs.sigClose)
		gs.sigClose = nil
	}
	gs.mu.Unlock()

	// Release all local allocations
	for id := range local {
		gs.releaseInternal(id)
	}

	if gs.registry != nil {
		gs.registry.Unregister(gs.lockID)
	}
}

// getStateFileName returns the state file name to use.
func (gs *GlobalSemaphore) getStateFileName() string {
	if gs.stateFileName != "" {
		return gs.stateFileName
	}
	return stateFile
}

// getLockFileName returns the lock file name to use.
func (gs *GlobalSemaphore) getLockFileName() string {
	if gs.lockFileName != "" {
		return gs.lockFileName
	}
	return stateLockFile
}

// tryAcquireInternal attempts to acquire weight atomically.
func (gs *GlobalSemaphore) tryAcquireInternal(weight, capacity int) (string, bool, int) {
	lock := flock.New(filepath.Join(gs.workspaceRoot, stateDir, gs.getLockFileName()))
	if err := lock.Lock(); err != nil {
		return "", false, 0
	}
	defer lock.Unlock()

	state := gs.readState()
	gs.cleanStaleFromState(state)

	// Calculate current usage
	var used int
	for _, a := range state.Allocations {
		used += a.Weight
	}

	// Check capacity
	if used+weight > capacity {
		return "", false, used
	}

	// Allocate
	allocID := uuid.New().String()
	state.Allocations[allocID] = &Allocation{
		ID:        allocID,
		PID:       os.Getpid(),
		Weight:    weight,
		Timestamp: time.Now(),
	}

	gs.writeState(state)
	return allocID, true, used + weight
}

// releaseInternal removes an allocation from global state.
func (gs *GlobalSemaphore) releaseInternal(allocID string) int {
	lock := flock.New(filepath.Join(gs.workspaceRoot, stateDir, gs.getLockFileName()))
	if err := lock.Lock(); err != nil {
		return 0
	}
	defer lock.Unlock()

	state := gs.readState()
	delete(state.Allocations, allocID)

	var used int
	for _, a := range state.Allocations {
		used += a.Weight
	}

	gs.writeState(state)
	return used
}

// cleanStale removes stale allocations from dead processes.
func (gs *GlobalSemaphore) cleanStale() {
	lock := flock.New(filepath.Join(gs.workspaceRoot, stateDir, gs.getLockFileName()))
	if err := lock.Lock(); err != nil {
		return
	}
	defer lock.Unlock()

	state := gs.readState()
	gs.cleanStaleFromState(state)
	gs.writeState(state)
}

func (gs *GlobalSemaphore) cleanStaleFromState(state *State) {
	for id, a := range state.Allocations {
		// Remove if the owning process is no longer running
		if !isProcessAlive(a.PID) {
			delete(state.Allocations, id)
		}
	}
}

// isProcessAlive checks if a process with the given PID is still running.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Use gopsutil for cross-platform process checking
	exists, err := process.PidExists(int32(pid))
	if err != nil {
		// On error, assume process is alive to avoid premature cleanup
		return true
	}
	return exists
}

func (gs *GlobalSemaphore) readState() *State {
	path := filepath.Join(gs.workspaceRoot, stateDir, gs.getStateFileName())
	data, err := os.ReadFile(path)
	if err != nil {
		return &State{Allocations: make(map[string]*Allocation)}
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return &State{Allocations: make(map[string]*Allocation)}
	}
	if state.Allocations == nil {
		state.Allocations = make(map[string]*Allocation)
	}
	return &state
}

func (gs *GlobalSemaphore) writeState(state *State) {
	path := filepath.Join(gs.workspaceRoot, stateDir, gs.getStateFileName())
	data, _ := json.Marshal(state)
	os.WriteFile(path, data, 0o644)
}

func (gs *GlobalSemaphore) updateRegistry(used, waitingDelta int64) {
	if gs.registry == nil {
		return
	}

	// If used is 0, read current state
	if used == 0 {
		state := gs.readState()
		for _, a := range state.Allocations {
			used += int64(a.Weight)
		}
	}

	// Get current waiting and apply delta
	snapshot := gs.registry.Snapshot()
	var waiting int64
	if info, ok := snapshot[gs.lockID]; ok {
		waiting = info.Waiting + waitingDelta
	}
	if waiting < 0 {
		waiting = 0
	}

	gs.registry.UpdateSemaphore(gs.lockID, used, waiting)
}
