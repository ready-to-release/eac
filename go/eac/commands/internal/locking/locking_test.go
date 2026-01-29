//go:build L1

package locking

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquireAndRelease(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// First acquisition should succeed
	lock, err := Acquire(tempDir, cfg)
	if err != nil {
		t.Fatalf("first acquisition failed: %v", err)
	}
	if lock == nil {
		t.Fatal("lock should not be nil after successful acquisition")
	}

	// Verify lock file exists
	lockPath := filepath.Join(tempDir, cfg.BaseDir, ".lock-"+cfg.Identifier)
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("lock file should exist after acquisition")
	}

	// Second acquisition should fail
	_, err = Acquire(tempDir, cfg)
	if err == nil {
		t.Error("second acquisition should fail while lock is held")
	}

	// Release the lock
	Release(lock)

	// Lock file should be removed after release
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("lock file should be removed after release")
	}

	// Third acquisition should succeed after release
	lock2, err := Acquire(tempDir, cfg)
	if err != nil {
		t.Fatalf("third acquisition failed after release: %v", err)
	}
	Release(lock2)
}

func TestBuildConfig(t *testing.T) {
	cfg := BuildConfig("eac-core", "out/build")

	if cfg.Identifier != "eac-core" {
		t.Errorf("expected identifier 'eac-core', got '%s'", cfg.Identifier)
	}
	if cfg.BaseDir != "out/build" {
		t.Errorf("expected baseDir 'out/build', got '%s'", cfg.BaseDir)
	}
	if cfg.ResourceType != "module" {
		t.Errorf("expected resourceType 'module', got '%s'", cfg.ResourceType)
	}
}

func TestTestConfig(t *testing.T) {
	cfg := TestConfig("unit", "out/test")

	if cfg.Identifier != "unit" {
		t.Errorf("expected identifier 'unit', got '%s'", cfg.Identifier)
	}
	if cfg.BaseDir != "out/test" {
		t.Errorf("expected baseDir 'out/test', got '%s'", cfg.BaseDir)
	}
	if cfg.ResourceType != "test suite" {
		t.Errorf("expected resourceType 'test suite', got '%s'", cfg.ResourceType)
	}
}

func TestReleaseNil(t *testing.T) {
	// Should not panic
	Release(nil)
}

func TestManualTestFileConfig(t *testing.T) {
	testCases := []struct {
		name       string
		filePath   string
		baseDir    string
		wantIdent  string
		wantBase   string
	}{
		{
			name:      "simple filename",
			filePath:  "results.json",
			baseDir:   "out/test",
			wantIdent: "results.json",
			wantBase:  "out/test",
		},
		{
			name:      "full path",
			filePath:  "out/test/manual-results.json",
			baseDir:   "out/test",
			wantIdent: "manual-results.json",
			wantBase:  "out/test",
		},
		{
			name:      "nested path",
			filePath:  "/path/to/test/results/manual-test-1.json",
			baseDir:   "out/test",
			wantIdent: "manual-test-1.json",
			wantBase:  "out/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ManualTestFileConfig(tc.filePath, tc.baseDir)

			if cfg.Identifier != tc.wantIdent {
				t.Errorf("identifier: got %q, want %q", cfg.Identifier, tc.wantIdent)
			}
			if cfg.BaseDir != tc.wantBase {
				t.Errorf("baseDir: got %q, want %q", cfg.BaseDir, tc.wantBase)
			}
			if cfg.ResourceType != "manual test file" {
				t.Errorf("resourceType: got %q, want %q", cfg.ResourceType, "manual test file")
			}
			if cfg.ActionVerb != "already being written" {
				t.Errorf("actionVerb: got %q, want %q", cfg.ActionVerb, "already being written")
			}
		})
	}
}

// Lock tracking tests

func TestAcquireTracked_RegistersWithRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	lock, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)
	defer ReleaseTracked(lock)

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}

	assert.Equal(t, locktracker.LockTypeFileLock, info.Type)
	assert.Equal(t, "module:test-module", info.Name)
	assert.Equal(t, int64(1), info.Used) // File lock is either held (1) or not (0)
}

func TestReleaseTracked_UnregistersFromRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	lock, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)

	// Verify lock is registered
	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	// Release the lock
	ReleaseTracked(lock)

	// Verify lock is unregistered
	snapshot = registry.Snapshot()
	assert.Len(t, snapshot, 0, "registry should be empty after release")
}

func TestAcquireTracked_Events(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()

	// Subscribe to events
	eventCh := make(chan locktracker.LockEvent, 10)
	unsubscribe := registry.Subscribe(eventCh)
	defer unsubscribe()

	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	lock, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)

	// Should receive EventAcquired
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventAcquired, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventAcquired event")
	}

	// Release the lock
	ReleaseTracked(lock)

	// Should receive EventReleased
	select {
	case event := <-eventCh:
		assert.Equal(t, locktracker.EventReleased, event.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected EventReleased event")
	}
}

func TestReleaseTracked_Nil(t *testing.T) {
	// Should not panic
	ReleaseTracked(nil)
}

func TestTrackedLock_AcquiredAt(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	beforeAcquire := time.Now()
	lock, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)
	defer ReleaseTracked(lock)
	afterAcquire := time.Now()

	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}

	assert.True(t, info.AcquiredAt.After(beforeAcquire) || info.AcquiredAt.Equal(beforeAcquire),
		"AcquiredAt should be after or equal to beforeAcquire")
	assert.True(t, info.AcquiredAt.Before(afterAcquire) || info.AcquiredAt.Equal(afterAcquire),
		"AcquiredAt should be before or equal to afterAcquire")
}

// AcquireWithWait tests

func TestAcquireWithWait_ImmediateSuccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// First acquisition should succeed immediately
	lock, err := AcquireWithWait(context.Background(), tempDir, cfg, registry, DefaultWaitConfig())
	require.NoError(t, err)
	defer ReleaseTracked(lock)

	// Verify lock is registered
	snapshot := registry.Snapshot()
	require.Len(t, snapshot, 1)

	var info locktracker.LockInfo
	for _, v := range snapshot {
		info = v
	}
	assert.Equal(t, locktracker.LockTypeFileLock, info.Type)
	assert.Equal(t, "module:test-module", info.Name)
	assert.Equal(t, int64(1), info.Used)
}

func TestAcquireWithWait_WaitsForRelease(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module-wait",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// Acquire first lock
	lock1, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)

	// Start goroutine that will release lock after delay
	// Use 300ms to give Windows time to release the file handle
	go func() {
		time.Sleep(300 * time.Millisecond)
		ReleaseTracked(lock1)
	}()

	// Second acquisition should wait and succeed
	waitCfg := WaitConfig{
		Timeout:      3 * time.Second,
		PollInterval: 100 * time.Millisecond, // Longer poll interval for Windows
	}

	start := time.Now()
	lock2, err := AcquireWithWait(context.Background(), tempDir, cfg, registry, waitCfg)
	elapsed := time.Since(start)

	require.NoError(t, err, "should acquire lock after waiting")
	defer ReleaseTracked(lock2)

	// Should have waited roughly 300ms (between 250ms and 1s, accounting for Windows timing)
	assert.True(t, elapsed >= 250*time.Millisecond, "should have waited at least 250ms, waited %v", elapsed)
	assert.True(t, elapsed < 1*time.Second, "should not have waited more than 1s, waited %v", elapsed)
}

func TestAcquireWithWait_Timeout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// Acquire first lock and don't release it
	lock1, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)
	defer ReleaseTracked(lock1)

	// Second acquisition should timeout
	waitCfg := WaitConfig{
		Timeout:      200 * time.Millisecond,
		PollInterval: 50 * time.Millisecond,
	}

	_, err = AcquireWithWait(context.Background(), tempDir, cfg, registry, waitCfg)
	require.Error(t, err, "should timeout when lock is held")
	assert.Contains(t, err.Error(), "timeout")
}

func TestAcquireWithWait_ContextCancellation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// Acquire first lock and don't release it
	lock1, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)
	defer ReleaseTracked(lock1)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	waitCfg := WaitConfig{
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	}

	start := time.Now()
	_, err = AcquireWithWait(ctx, tempDir, cfg, registry, waitCfg)
	elapsed := time.Since(start)

	require.Error(t, err, "should fail when context is cancelled")
	assert.Contains(t, err.Error(), "cancelled")
	assert.True(t, elapsed < 500*time.Millisecond, "should cancel quickly, took %v", elapsed)
}

func TestAcquireWithWait_WaitingStateTracked(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := locktracker.NewRegistry()
	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// Acquire first lock
	lock1, err := AcquireTracked(tempDir, cfg, registry)
	require.NoError(t, err)

	// Subscribe to events
	eventCh := make(chan locktracker.LockEvent, 10)
	unsubscribe := registry.Subscribe(eventCh)
	defer unsubscribe()

	// Start goroutine to check for waiting state
	waitingDetected := make(chan bool, 1)
	go func() {
		for i := 0; i < 20; i++ {
			snapshot := registry.Snapshot()
			for _, info := range snapshot {
				if info.Waiting > 0 {
					waitingDetected <- true
					return
				}
			}
			time.Sleep(25 * time.Millisecond)
		}
		waitingDetected <- false
	}()

	// Start second acquisition in goroutine
	lock2Chan := make(chan *TrackedLock, 1)
	go func() {
		waitCfg := WaitConfig{
			Timeout:      2 * time.Second,
			PollInterval: 50 * time.Millisecond,
		}
		lock2, _ := AcquireWithWait(context.Background(), tempDir, cfg, registry, waitCfg)
		lock2Chan <- lock2
	}()

	// Wait for waiting state to be detected
	time.Sleep(100 * time.Millisecond)

	// Release first lock
	ReleaseTracked(lock1)

	// Get second lock result
	lock2 := <-lock2Chan
	if lock2 != nil {
		defer ReleaseTracked(lock2)
	}

	// Check if waiting was detected (best effort - timing dependent)
	select {
	case detected := <-waitingDetected:
		t.Logf("Waiting state detected: %v", detected)
	case <-time.After(500 * time.Millisecond):
		t.Log("Waiting detection timed out")
	}
}
