//go:build L1

package orchestrator

import (
	"bytes"
	"log"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/stretchr/testify/assert"
)

func TestDisplayManager_FormatLockInfo_Empty(t *testing.T) {
	registry := locktracker.NewRegistry()
	result := formatLockInfo(registry)
	assert.Equal(t, "", result)
}

func TestDisplayManager_FormatLockInfo_WithSemaphores(t *testing.T) {
	registry := locktracker.NewRegistry()

	// Register a semaphore with usage
	registry.Register(&locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "component-scheduler",
		Capacity: 8,
		Used:     6,
		Waiting:  2,
	})

	result := formatLockInfo(registry)
	assert.Contains(t, result, "slots:6/8")
	assert.Contains(t, result, "wait:2")
}

func TestDisplayManager_FormatLockInfo_NoWaiting(t *testing.T) {
	registry := locktracker.NewRegistry()

	// Register a semaphore without waiting
	registry.Register(&locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "component-scheduler",
		Capacity: 8,
		Used:     3,
		Waiting:  0,
	})

	result := formatLockInfo(registry)
	assert.Contains(t, result, "slots:3/8")
	assert.NotContains(t, result, "wait:")
}

func TestDisplayManager_FormatLockInfo_WithFileLocks(t *testing.T) {
	registry := locktracker.NewRegistry()

	// Register file locks
	registry.Register(&locktracker.LockInfo{
		ID:   "file-1",
		Type: locktracker.LockTypeFileLock,
		Name: "module:eac-core",
		Used: 1,
	})
	registry.Register(&locktracker.LockInfo{
		ID:   "file-2",
		Type: locktracker.LockTypeFileLock,
		Name: "module:docs",
		Used: 1,
	})

	result := formatLockInfo(registry)
	assert.Contains(t, result, "locks:2")
}

func TestDisplayManager_StatusWithLockInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	// Create registry with test data
	registry := locktracker.NewRegistry()
	registry.Register(&locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "scheduler",
		Capacity: 8,
		Used:     4,
		Waiting:  1,
	})

	dm := newDisplayManager(logger, "building", 5, 1000, false, registry)
	dm.running["test-module"] = true
	dm.completed = 2
	dm.total = 5

	dm.displayStatus()

	output := buf.String()
	assert.Contains(t, output, "2/5 completed")
	assert.Contains(t, output, "1 running")
	assert.Contains(t, output, "slots:4/8")
	assert.Contains(t, output, "wait:1")
}
