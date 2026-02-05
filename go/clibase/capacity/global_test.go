package capacity

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsProcessAlive(t *testing.T) {
	// Current process should be alive
	if !isProcessAlive(os.Getpid()) {
		t.Error("current process should be reported as alive")
	}

	// PID 0 should not be alive
	if isProcessAlive(0) {
		t.Error("PID 0 should not be reported as alive")
	}

	// Negative PID should not be alive
	if isProcessAlive(-1) {
		t.Error("negative PID should not be reported as alive")
	}

	// Very high PID is likely not running (but we can't guarantee this)
	// This is a weak test - the PID might exist on some systems
	// We test this by using a PID that's extremely unlikely to exist
	if isProcessAlive(999999999) {
		t.Log("PID 999999999 is unexpectedly alive (this may happen on some systems)")
	}
}

func TestCleanStaleFromState_DeadProcess(t *testing.T) {
	tmpDir := t.TempDir()

	gs := &GlobalSemaphore{
		workspaceRoot: tmpDir,
		allocations:   make(map[string]int),
	}

	// Ensure state directory exists
	os.MkdirAll(filepath.Join(tmpDir, stateDir), 0o755)

	// Create a state with a dead PID (using an unlikely PID)
	state := &State{
		Allocations: map[string]*Allocation{
			"dead-alloc": {
				ID:        "dead-alloc",
				PID:       999999999, // Very unlikely to exist
				Weight:    4,
				Timestamp: time.Now(), // Not stale by time
			},
			"alive-alloc": {
				ID:        "alive-alloc",
				PID:       os.Getpid(), // Current process
				Weight:    2,
				Timestamp: time.Now(),
			},
		},
	}

	// Clean stale allocations
	gs.cleanStaleFromState(state)

	// Dead process allocation should be removed
	if _, ok := state.Allocations["dead-alloc"]; ok {
		t.Error("allocation for dead process should be removed")
	}

	// Alive process allocation should remain
	if _, ok := state.Allocations["alive-alloc"]; !ok {
		t.Error("allocation for alive process should remain")
	}
}

func TestGlobalSemaphore_CloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	gs := NewGlobalSemaphore(tmpDir, 8, nil)

	// First close should work
	gs.Close()

	// Second close should not panic
	gs.Close()

	// Third close should still not panic
	gs.Close()
}

func TestGlobalSemaphore_AcquireRelease(t *testing.T) {
	tmpDir := t.TempDir()

	gs := NewGlobalSemaphore(tmpDir, 8, nil)
	defer gs.Close()

	// Acquire some capacity
	gs.Acquire(3)

	// Check usage
	used := gs.Used()
	if used != 3 {
		t.Errorf("expected used=3, got %d", used)
	}

	// Release
	gs.Release(3)

	// Check usage is back to 0
	used = gs.Used()
	if used != 0 {
		t.Errorf("expected used=0 after release, got %d", used)
	}
}
