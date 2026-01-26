//go:build L1

package console

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/stretchr/testify/assert"
)

func TestNewLockDisplay(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	assert.NotNil(t, ld)
	assert.Equal(t, registry, ld.registry)
}

func TestLockDisplay_RenderCompact_Empty(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	result := ld.RenderCompact()
	assert.Equal(t, "", result, "empty registry should return empty string")
}

func TestLockDisplay_RenderCompact_WithSemaphores(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register a semaphore with some usage
	registry.Register(locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "component-scheduler",
		Capacity: 8,
		Used:     6,
		Waiting:  2,
	})

	result := ld.RenderCompact()
	assert.Contains(t, result, "slots:6/8", "should show slots usage")
	assert.Contains(t, result, "wait:2", "should show waiting count")
}

func TestLockDisplay_RenderCompact_WithFileLocks(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register file locks
	registry.Register(locktracker.LockInfo{
		ID:   "file-1",
		Type: locktracker.LockTypeFileLock,
		Name: "module:eac-core",
		Used: 1,
	})
	registry.Register(locktracker.LockInfo{
		ID:   "file-2",
		Type: locktracker.LockTypeFileLock,
		Name: "module:docs",
		Used: 1,
	})

	result := ld.RenderCompact()
	assert.Contains(t, result, "locks:2", "should show file lock count")
}

func TestLockDisplay_RenderCompact_NoWaiting(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register a semaphore without waiting
	registry.Register(locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "component-scheduler",
		Capacity: 8,
		Used:     3,
		Waiting:  0,
	})

	result := ld.RenderCompact()
	assert.Contains(t, result, "slots:3/8", "should show slots usage")
	assert.NotContains(t, result, "wait:", "should not show wait when 0")
}

func TestLockDisplay_RenderDetailed_Empty(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	lines := ld.RenderDetailed()
	assert.Len(t, lines, 1)
	assert.Equal(t, "No active locks", lines[0])
}

func TestLockDisplay_RenderDetailed_WithSemaphores(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register semaphores
	registry.Register(locktracker.LockInfo{
		ID:       "sem-1",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "scanner-trivy",
		Capacity: 3,
		Used:     2,
		Waiting:  0,
	})
	registry.Register(locktracker.LockInfo{
		ID:       "sem-2",
		Type:     locktracker.LockTypeSemaphore,
		Name:     "pdf-export",
		Capacity: 4,
		Used:     4,
		Waiting:  2,
	})

	lines := ld.RenderDetailed()

	// Should have header
	found := false
	for _, line := range lines {
		if line == "Semaphores:" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have Semaphores header")

	// Should show progress bars and names
	hasScanner := false
	hasPdf := false
	for _, line := range lines {
		if contains(line, "scanner-trivy") && contains(line, "2/3") {
			hasScanner = true
		}
		if contains(line, "pdf-export") && contains(line, "4/4") {
			hasPdf = true
		}
	}
	assert.True(t, hasScanner, "should show scanner-trivy with 2/3")
	assert.True(t, hasPdf, "should show pdf-export with 4/4")
}

func TestLockDisplay_RenderDetailed_WithWeighted(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register weighted semaphore
	registry.Register(locktracker.LockInfo{
		ID:       "weighted-1",
		Type:     locktracker.LockTypeWeighted,
		Name:     "component-scheduler",
		Capacity: 8,
		Used:     6,
		Waiting:  1,
	})

	lines := ld.RenderDetailed()

	// Should have Weighted header
	found := false
	for _, line := range lines {
		if line == "Weighted:" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have Weighted header")

	// Should show the weighted semaphore
	hasScheduler := false
	for _, line := range lines {
		if contains(line, "component-scheduler") && contains(line, "6/8") {
			hasScheduler = true
		}
	}
	assert.True(t, hasScheduler, "should show component-scheduler with 6/8")
}

func TestLockDisplay_RenderDetailed_WithFileLocks(t *testing.T) {
	registry := locktracker.NewRegistry()
	ld := NewLockDisplay(registry)

	// Register file locks with acquired time
	registry.Register(locktracker.LockInfo{
		ID:         "file-1",
		Type:       locktracker.LockTypeFileLock,
		Name:       "module:eac-core",
		AcquiredAt: time.Now().Add(-23 * time.Second),
		Used:       1,
	})

	lines := ld.RenderDetailed()

	// Should have File Locks header
	found := false
	for _, line := range lines {
		if line == "File Locks:" {
			found = true
			break
		}
	}
	assert.True(t, found, "should have File Locks header")

	// Should show the file lock with age
	hasModule := false
	for _, line := range lines {
		if contains(line, "module:eac-core") {
			hasModule = true
		}
	}
	assert.True(t, hasModule, "should show module:eac-core")
}

func TestRenderProgressBar(t *testing.T) {
	tests := []struct {
		name     string
		used     int64
		capacity int64
		width    int
		want     string
	}{
		{
			name:     "empty",
			used:     0,
			capacity: 10,
			width:    10,
			want:     "[          ]",
		},
		{
			name:     "half",
			used:     5,
			capacity: 10,
			width:    10,
			want:     "[=====     ]",
		},
		{
			name:     "full",
			used:     10,
			capacity: 10,
			width:    10,
			want:     "[==========]",
		},
		{
			name:     "over capacity",
			used:     15,
			capacity: 10,
			width:    10,
			want:     "[==========]",
		},
		{
			name:     "zero capacity",
			used:     0,
			capacity: 0,
			width:    10,
			want:     "[----------]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderProgressBar(tt.used, tt.capacity, tt.width)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
