package workunit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// LockInfo Type Tests
// =============================================================================

func TestLockInfo_FieldsExist(t *testing.T) {
	now := time.Now()
	info := LockInfo{
		PID:       12345,
		StartedAt: now,
		UnitID:    "test:core:go:gotest:unit",
	}

	assert.Equal(t, 12345, info.PID)
	assert.Equal(t, now, info.StartedAt)
	assert.Equal(t, "test:core:go:gotest:unit", info.UnitID)
}

func TestLockInfo_ZeroValue(t *testing.T) {
	var info LockInfo

	assert.Equal(t, 0, info.PID)
	assert.True(t, info.StartedAt.IsZero())
	assert.Empty(t, info.UnitID)
}

// =============================================================================
// LockInfo JSON Marshaling Tests
// =============================================================================

func TestLockInfo_MarshalJSON(t *testing.T) {
	info := LockInfo{
		PID:       12345,
		StartedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UnitID:    "build:core:go:go",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"pid"`)
	assert.Contains(t, jsonStr, `"started_at"`)
	assert.Contains(t, jsonStr, `"unit_id"`)
	assert.Contains(t, jsonStr, "12345")
	assert.Contains(t, jsonStr, "build:core:go:go")
}

func TestLockInfo_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"pid": 67890,
		"started_at": "2024-01-15T10:30:00Z",
		"unit_id": "test:core:go:gotest:unit"
	}`

	var info LockInfo
	err := json.Unmarshal([]byte(jsonData), &info)
	require.NoError(t, err)

	assert.Equal(t, 67890, info.PID)
	assert.Equal(t, "test:core:go:gotest:unit", info.UnitID)
	assert.Equal(t, 2024, info.StartedAt.Year())
}

func TestLockInfo_MarshalUnmarshalRoundTrip(t *testing.T) {
	original := LockInfo{
		PID:       99999,
		StartedAt: time.Date(2024, 6, 20, 14, 45, 30, 0, time.UTC),
		UnitID:    "lint:eac:go:golangci-lint",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored LockInfo
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.PID, restored.PID)
	assert.Equal(t, original.UnitID, restored.UnitID)
	assert.True(t, original.StartedAt.Equal(restored.StartedAt))
}

// =============================================================================
// UnitID.Lock() Tests
// =============================================================================

func TestUnitID_Lock_CreatesLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Create the output directory structure expected by LockFile()
	// Lock() should create directories if needed
	lockPath := filepath.Join(tmpDir, unitID.LockFile())
	lockDir := filepath.Dir(lockPath)
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err, "setup: failed to create lock directory")

	// Use a workspace-rooted version or set working dir
	// For this test, we'll work relative to tmpDir
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Acquire lock
	err = Lock(unitID)
	require.NoError(t, err, "Lock() should succeed")

	// Verify lock file exists
	_, err = os.Stat(unitID.LockFile())
	assert.NoError(t, err, "Lock file should exist after Lock()")

	// Cleanup: release the lock
	_ = Unlock(unitID)
}

func TestUnitID_Lock_FailsIfAlreadyLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool:  "gotest",
		Extra: map[string]string{"testset": "unit"},
	}

	// Setup directory and change to tmpDir
	lockDir := filepath.Dir(filepath.Join(tmpDir, unitID.LockFile()))
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err)

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// First lock should succeed
	err = Lock(unitID)
	require.NoError(t, err, "First Lock() should succeed")

	// Second lock should fail
	err = Lock(unitID)
	assert.Error(t, err, "Second Lock() should fail when already locked")

	// Cleanup
	_ = Unlock(unitID)
}

func TestUnitID_Lock_ContainsValidLockInfo(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionLint,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "golangci-lint",
	}

	// Setup
	lockDir := filepath.Dir(filepath.Join(tmpDir, unitID.LockFile()))
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err)

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	beforeLock := time.Now()

	err = Lock(unitID)
	require.NoError(t, err)

	afterLock := time.Now()

	// Read lock file and verify contents
	data, err := os.ReadFile(unitID.LockFile())
	require.NoError(t, err, "Should be able to read lock file")

	var info LockInfo
	err = json.Unmarshal(data, &info)
	require.NoError(t, err, "Lock file should contain valid JSON")

	// Verify PID is the current process
	assert.Equal(t, os.Getpid(), info.PID, "Lock should contain current process PID")

	// Verify StartedAt is reasonable
	assert.False(t, info.StartedAt.IsZero(), "StartedAt should be set")
	assert.True(t, info.StartedAt.After(beforeLock.Add(-time.Second)), "StartedAt should be after lock acquisition started")
	assert.True(t, info.StartedAt.Before(afterLock.Add(time.Second)), "StartedAt should be before lock acquisition ended")

	// Verify UnitID is set
	assert.Equal(t, unitID.Longname(), info.UnitID, "Lock should contain the unit ID")

	// Cleanup
	_ = Unlock(unitID)
}

// =============================================================================
// UnitID.Unlock() Tests
// =============================================================================

func TestUnitID_Unlock_RemovesLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionScan,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "trivy-vuln",
	}

	// Setup
	lockDir := filepath.Dir(filepath.Join(tmpDir, unitID.LockFile()))
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err)

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Acquire and then release lock
	err = Lock(unitID)
	require.NoError(t, err)

	err = Unlock(unitID)
	require.NoError(t, err, "Unlock() should succeed")

	// Verify lock file is removed
	_, err = os.Stat(unitID.LockFile())
	assert.True(t, os.IsNotExist(err), "Lock file should not exist after Unlock()")
}

func TestUnitID_Unlock_IsIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Setup - create directory but NO lock file
	lockDir := filepath.Dir(filepath.Join(tmpDir, unitID.LockFile()))
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err)

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Unlock without ever locking should succeed (idempotent)
	err = Unlock(unitID)
	assert.NoError(t, err, "Unlock() should succeed even when no lock exists (idempotent)")

	// Multiple unlocks should all succeed
	err = Unlock(unitID)
	assert.NoError(t, err, "Second Unlock() should also succeed (idempotent)")
}

func TestUnitID_Unlock_SucceedsAfterMultipleLockUnlockCycles(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        "cycle-test",
		ComponentType: "go", ComponentName: "go",
		Tool:  "gotest",
		Extra: map[string]string{"testset": "unit"},
	}

	// Setup
	lockDir := filepath.Dir(filepath.Join(tmpDir, unitID.LockFile()))
	err := os.MkdirAll(lockDir, 0755)
	require.NoError(t, err)

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Multiple lock/unlock cycles should work
	for i := 0; i < 3; i++ {
		err = Lock(unitID)
		require.NoError(t, err, "Lock() should succeed in cycle %d", i)

		err = Unlock(unitID)
		require.NoError(t, err, "Unlock() should succeed in cycle %d", i)
	}
}

// =============================================================================
// Multiple UnitIDs Lock Concurrently Tests
// =============================================================================

func TestUnitID_MultipleUnitsCanBeLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unit1 := UnitID{
		Action:        core.ActionBuild,
		Module:        "module-a",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	unit2 := UnitID{
		Action:        core.ActionBuild,
		Module:        "module-b",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	unit3 := UnitID{
		Action:        core.ActionTest,
		Module:        "module-a",
		ComponentType: "go", ComponentName: "go",
		Tool:  "gotest",
		Extra: map[string]string{"testset": "unit"},
	}

	// Setup directories for all units
	for _, u := range []UnitID{unit1, unit2, unit3} {
		lockDir := filepath.Dir(filepath.Join(tmpDir, u.LockFile()))
		err := os.MkdirAll(lockDir, 0755)
		require.NoError(t, err)
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// All three units should be lockable simultaneously
	err = Lock(unit1)
	require.NoError(t, err, "unit1.Lock() should succeed")

	err = Lock(unit2)
	require.NoError(t, err, "unit2.Lock() should succeed while unit1 is locked")

	err = Lock(unit3)
	require.NoError(t, err, "unit3.Lock() should succeed while unit1 and unit2 are locked")

	// Verify all lock files exist
	_, err = os.Stat(unit1.LockFile())
	assert.NoError(t, err, "unit1 lock file should exist")

	_, err = os.Stat(unit2.LockFile())
	assert.NoError(t, err, "unit2 lock file should exist")

	_, err = os.Stat(unit3.LockFile())
	assert.NoError(t, err, "unit3 lock file should exist")

	// Cleanup
	_ = Unlock(unit1)
	_ = Unlock(unit2)
	_ = Unlock(unit3)
}

func TestUnitID_SameModuleDifferentContextsCanBeLocked(t *testing.T) {
	tmpDir := t.TempDir()

	buildUnit := UnitID{
		Action:        core.ActionBuild,
		Module:        "shared-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	testUnit := UnitID{
		Action:        core.ActionTest,
		Module:        "shared-module",
		ComponentType: "go", ComponentName: "go",
		Tool:  "gotest",
		Extra: map[string]string{"testset": "unit"},
	}

	lintUnit := UnitID{
		Action:        core.ActionLint,
		Module:        "shared-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "golangci-lint",
	}

	// Setup directories
	for _, u := range []UnitID{buildUnit, testUnit, lintUnit} {
		lockDir := filepath.Dir(filepath.Join(tmpDir, u.LockFile()))
		err := os.MkdirAll(lockDir, 0755)
		require.NoError(t, err)
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Different contexts for same module should have independent locks
	err = Lock(buildUnit)
	require.NoError(t, err)

	err = Lock(testUnit)
	require.NoError(t, err, "Test unit should be lockable while build is locked")

	err = Lock(lintUnit)
	require.NoError(t, err, "Lint unit should be lockable while build and test are locked")

	// Cleanup
	_ = Unlock(buildUnit)
	_ = Unlock(testUnit)
	_ = Unlock(lintUnit)
}

// =============================================================================
// Lock Directory Creation Tests
// =============================================================================

func TestUnitID_Lock_CreatesDirectoriesIfNeeded(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        "deep-nested",
		ComponentType: "gherkin", ComponentName: "gherkin",
		Tool:  "godog",
		Extra: map[string]string{"testset": "acceptance"},
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Do NOT create directories beforehand - Lock() should create them
	lockDir := filepath.Dir(unitID.LockFile())
	_, err = os.Stat(lockDir)
	assert.True(t, os.IsNotExist(err), "Directory should not exist before Lock()")

	// Lock should create directory and lock file
	err = Lock(unitID)
	require.NoError(t, err, "Lock() should create directories and succeed")

	// Verify lock file exists
	_, err = os.Stat(unitID.LockFile())
	assert.NoError(t, err, "Lock file should exist after Lock()")

	// Cleanup
	_ = Unlock(unitID)
}

// =============================================================================
// ErrLocked Error Type Tests
// =============================================================================

func TestErrLocked_ErrorWithHeldBy(t *testing.T) {
	heldBy := &LockInfo{
		PID:       12345,
		StartedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UnitID:    "test:core:go:gotest",
	}

	err := &ErrLocked{
		UnitID:   "test:core:go:gotest",
		HeldBy:   heldBy,
		LockPath: "out/test/core/go/gotest/lock.json",
	}

	errMsg := err.Error()
	assert.Contains(t, errMsg, "test:core:go:gotest")
	assert.Contains(t, errMsg, "12345")
	assert.Contains(t, errMsg, "locked")
}

func TestErrLocked_ErrorWithoutHeldBy(t *testing.T) {
	err := &ErrLocked{
		UnitID:   "test:core:go:gotest",
		HeldBy:   nil,
		LockPath: "out/test/core/go/gotest/lock.json",
	}

	errMsg := err.Error()
	assert.Contains(t, errMsg, "test:core:go:gotest")
	assert.Contains(t, errMsg, "locked")
}

func TestUnitID_Lock_ReturnsErrLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "gotest",
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// First lock should succeed
	err = Lock(unitID)
	require.NoError(t, err)
	defer Unlock(unitID)

	// Second lock should return ErrLocked
	err = Lock(unitID)
	require.Error(t, err)

	var lockErr *ErrLocked
	assert.ErrorAs(t, err, &lockErr)
	assert.Equal(t, unitID.Longname(), lockErr.UnitID)
	assert.NotNil(t, lockErr.HeldBy)
	assert.Equal(t, os.Getpid(), lockErr.HeldBy.PID)
}

// =============================================================================
// LockWithRoot/UnlockWithRoot Tests
// =============================================================================

func TestUnitID_LockWithRoot_CreatesLockFileInRoot(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Acquire lock with explicit root
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)

	// Verify lock file exists under the root
	lockPath := filepath.Join(tmpDir, unitID.LockFile())
	_, err = os.Stat(lockPath)
	assert.NoError(t, err, "Lock file should exist under root")

	// Cleanup
	_ = UnlockWithRoot(unitID, tmpDir)
}

func TestUnitID_UnlockWithRoot_RemovesLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Acquire and release lock with root
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)

	err = UnlockWithRoot(unitID, tmpDir)
	require.NoError(t, err)

	// Verify lock file is removed
	lockPath := filepath.Join(tmpDir, unitID.LockFile())
	_, err = os.Stat(lockPath)
	assert.True(t, os.IsNotExist(err), "Lock file should not exist after unlock")
}

func TestUnitID_LockWithRoot_FailsIfAlreadyLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionTest,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "gotest",
	}

	// First lock
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)
	defer UnlockWithRoot(unitID, tmpDir)

	// Second lock should fail with ErrLocked
	err = LockWithRoot(unitID, tmpDir)
	require.Error(t, err)

	var lockErr *ErrLocked
	assert.ErrorAs(t, err, &lockErr)
}

// =============================================================================
// IsLocked/IsLockedWithRoot Tests
// =============================================================================

func TestUnitID_IsLocked_ReturnsTrueWhenLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Not locked initially
	assert.False(t, IsLocked(unitID))

	// Lock
	err = Lock(unitID)
	require.NoError(t, err)

	// Should be locked now
	assert.True(t, IsLocked(unitID))

	// Unlock
	_ = Unlock(unitID)

	// Should be unlocked again
	assert.False(t, IsLocked(unitID))
}

func TestUnitID_IsLockedWithRoot(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Not locked initially
	assert.False(t, IsLockedWithRoot(unitID, tmpDir))

	// Lock with root
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)

	// Should be locked now
	assert.True(t, IsLockedWithRoot(unitID, tmpDir))

	// Different root should not be locked
	otherDir := t.TempDir()
	assert.False(t, IsLockedWithRoot(unitID, otherDir))

	// Cleanup
	_ = UnlockWithRoot(unitID, tmpDir)
}

// =============================================================================
// ReadLockInfo/ReadLockInfoWithRoot Tests
// =============================================================================

func TestUnitID_ReadLockInfo_ReturnsNilWhenNotLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	info := ReadLockInfo(unitID)
	assert.Nil(t, info)
}

func TestUnitID_ReadLockInfo_ReturnsInfoWhenLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Lock
	beforeLock := time.Now()
	err = Lock(unitID)
	require.NoError(t, err)
	defer Unlock(unitID)

	info := ReadLockInfo(unitID)
	require.NotNil(t, info)
	assert.Equal(t, os.Getpid(), info.PID)
	assert.Equal(t, unitID.Longname(), info.UnitID)
	assert.True(t, info.StartedAt.After(beforeLock.Add(-time.Second)))
}

func TestUnitID_ReadLockInfoWithRoot(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Lock with root
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)
	defer UnlockWithRoot(unitID, tmpDir)

	info := ReadLockInfoWithRoot(unitID, tmpDir)
	require.NotNil(t, info)
	assert.Equal(t, os.Getpid(), info.PID)
	assert.Equal(t, unitID.Longname(), info.UnitID)

	// Different root should return nil
	otherDir := t.TempDir()
	otherInfo := ReadLockInfoWithRoot(unitID, otherDir)
	assert.Nil(t, otherInfo)
}

// =============================================================================
// LockWaitConfig Tests
// =============================================================================

func TestLockWaitConfig_DefaultValues(t *testing.T) {
	cfg := DefaultLockWaitConfig()
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.Equal(t, 200*time.Millisecond, cfg.PollInterval)
}

// =============================================================================
// LockWithWait Tests
// =============================================================================

func TestUnitID_LockWithWait_ImmediateSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := t.Context()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	cfg := LockWaitConfig{
		Timeout:      1 * time.Second,
		PollInterval: 50 * time.Millisecond,
	}

	// Lock should succeed immediately when not already locked
	start := time.Now()
	err := LockWithWait(unitID, ctx, tmpDir, cfg)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, elapsed, 100*time.Millisecond, "Should succeed immediately")

	// Cleanup
	_ = UnlockWithRoot(unitID, tmpDir)
}

func TestUnitID_LockWithWait_WaitsAndAcquires(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := t.Context()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// First, acquire the lock
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)

	cfg := LockWaitConfig{
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	}

	// Release the lock after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = UnlockWithRoot(unitID, tmpDir)
	}()

	// LockWithWait should wait and then succeed
	start := time.Now()
	err = LockWithWait(unitID, ctx, tmpDir, cfg)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Greater(t, elapsed, 150*time.Millisecond, "Should have waited")
	assert.Less(t, elapsed, 1*time.Second, "Should not timeout")

	// Cleanup
	_ = UnlockWithRoot(unitID, tmpDir)
}

func TestUnitID_LockWithWait_TimeoutReturnsErrLocked(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := t.Context()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// First, acquire the lock and don't release it
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)
	defer UnlockWithRoot(unitID, tmpDir)

	cfg := LockWaitConfig{
		Timeout:      200 * time.Millisecond,
		PollInterval: 50 * time.Millisecond,
	}

	// LockWithWait should timeout
	start := time.Now()
	err = LockWithWait(unitID, ctx, tmpDir, cfg)
	elapsed := time.Since(start)

	require.Error(t, err)
	var lockErr *ErrLocked
	assert.ErrorAs(t, err, &lockErr)
	assert.Greater(t, elapsed, 150*time.Millisecond, "Should have waited for timeout")
}

func TestUnitID_LockWithWait_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	ctx, cancel := t.Context(), func() {}
	ctx, cancel = contextWithCancel(ctx)

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// First, acquire the lock and don't release it
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)
	defer UnlockWithRoot(unitID, tmpDir)

	cfg := LockWaitConfig{
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// LockWithWait should be cancelled
	start := time.Now()
	err = LockWithWait(unitID, ctx, tmpDir, cfg)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.Less(t, elapsed, 500*time.Millisecond, "Should cancel quickly")
}

// contextWithCancel wraps context.WithCancel for test compatibility
func contextWithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

func TestUnitID_LockWithWait_UsesDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := t.Context()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Pass zero config - should use defaults
	cfg := LockWaitConfig{}

	// Should succeed immediately with defaults
	err := LockWithWait(unitID, ctx, tmpDir, cfg)
	require.NoError(t, err)

	// Cleanup
	_ = UnlockWithRoot(unitID, tmpDir)
}

// =============================================================================
// TryBreakStaleLock Tests
// =============================================================================

func TestUnitID_TryBreakStaleLock_ReturnsFalseWhenNotLocked(t *testing.T) {
	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// No lock exists
	broken := TryBreakStaleLock(unitID, tmpDir)
	assert.False(t, broken, "Should return false when no lock exists")
}

func TestUnitID_TryBreakStaleLock_ReturnsFalseForLiveProcess(t *testing.T) {
	// Skip on Windows: process.Signal(nil) isn't supported on Windows,
	// so TryBreakStaleLock cannot reliably detect live processes.
	// The function is documented as "best-effort" with platform limitations.
	if os.PathSeparator == '\\' {
		t.Skip("Skipping on Windows: Signal-based process detection not supported")
	}

	tmpDir := t.TempDir()

	unitID := UnitID{
		Action:        core.ActionBuild,
		Module:        "test-module",
		ComponentType: "go", ComponentName: "go",
		Tool: "go",
	}

	// Create a lock held by current process (which is alive)
	err := LockWithRoot(unitID, tmpDir)
	require.NoError(t, err)
	defer UnlockWithRoot(unitID, tmpDir)

	// Should not break lock for live process
	broken := TryBreakStaleLock(unitID, tmpDir)
	assert.False(t, broken, "Should not break lock held by live process")

	// Lock should still exist
	assert.True(t, IsLockedWithRoot(unitID, tmpDir))
}
