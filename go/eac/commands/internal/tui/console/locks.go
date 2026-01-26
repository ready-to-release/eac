package console

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
)

// LockDisplay renders lock status for TUI visualization.
type LockDisplay struct {
	registry *locktracker.Registry
}

// NewLockDisplay creates a new lock display with the given registry.
func NewLockDisplay(registry *locktracker.Registry) *LockDisplay {
	return &LockDisplay{
		registry: registry,
	}
}

// RenderCompact returns a single-line lock summary for header/footer display.
// Format: "slots:6/8 wait:2 locks:3"
// Returns empty string if no locks are registered.
func (ld *LockDisplay) RenderCompact() string {
	summary := ld.registry.Summary()

	if summary.Total == 0 {
		return ""
	}

	var parts []string

	// Show semaphore/weighted usage (combined capacity)
	if summary.TotalCapacity > 0 {
		parts = append(parts, fmt.Sprintf("slots:%d/%d",
			summary.TotalUsed, summary.TotalCapacity))
	}

	// Show waiting if any
	if summary.TotalWaiting > 0 {
		parts = append(parts, fmt.Sprintf("wait:%d", summary.TotalWaiting))
	}

	// Show file lock count
	fileLocks := summary.ByType[locktracker.LockTypeFileLock]
	if fileLocks > 0 {
		parts = append(parts, fmt.Sprintf("locks:%d", fileLocks))
	}

	return strings.Join(parts, " ")
}

// RenderDetailed returns multi-line lock details for expanded view.
// Groups locks by type and shows progress bars for semaphores.
func (ld *LockDisplay) RenderDetailed() []string {
	locks := ld.registry.Snapshot()

	if len(locks) == 0 {
		return []string{"No active locks"}
	}

	lines := []string{
		fmt.Sprintf("Active Locks (%d):", len(locks)),
		"",
	}

	// Group by type
	byType := make(map[locktracker.LockType][]locktracker.LockInfo)
	for _, lock := range locks {
		byType[lock.Type] = append(byType[lock.Type], lock)
	}

	// Render semaphores first (most interesting for concurrency visualization)
	if sems, ok := byType[locktracker.LockTypeSemaphore]; ok && len(sems) > 0 {
		lines = append(lines, "Semaphores:")
		// Sort by name for consistent output
		sort.Slice(sems, func(i, j int) bool {
			return sems[i].Name < sems[j].Name
		})
		for _, sem := range sems {
			bar := renderProgressBar(sem.Used, sem.Capacity, 10)
			lines = append(lines, fmt.Sprintf("  %s: %s %d/%d",
				sem.Name, bar, sem.Used, sem.Capacity))
			if sem.Waiting > 0 {
				lines = append(lines, fmt.Sprintf("    waiting: %d", sem.Waiting))
			}
		}
	}

	// Render weighted semaphores
	if weighted, ok := byType[locktracker.LockTypeWeighted]; ok && len(weighted) > 0 {
		lines = append(lines, "Weighted:")
		// Sort by name for consistent output
		sort.Slice(weighted, func(i, j int) bool {
			return weighted[i].Name < weighted[j].Name
		})
		for _, w := range weighted {
			bar := renderProgressBar(w.Used, w.Capacity, 10)
			lines = append(lines, fmt.Sprintf("  %s: %s %d/%d",
				w.Name, bar, w.Used, w.Capacity))
			if w.Waiting > 0 {
				lines = append(lines, fmt.Sprintf("    waiting: %d", w.Waiting))
			}
		}
	}

	// Render file locks
	if files, ok := byType[locktracker.LockTypeFileLock]; ok && len(files) > 0 {
		lines = append(lines, "File Locks:")
		// Sort by name for consistent output
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
		for _, f := range files {
			if !f.AcquiredAt.IsZero() {
				age := time.Since(f.AcquiredAt).Round(time.Second)
				lines = append(lines, fmt.Sprintf("  %s (%s)", f.Name, age))
			} else {
				lines = append(lines, fmt.Sprintf("  %s", f.Name))
			}
		}
	}

	// Render mutexes if any (less common to track)
	if mutexes, ok := byType[locktracker.LockTypeMutex]; ok && len(mutexes) > 0 {
		lines = append(lines, "Mutexes:")
		for _, m := range mutexes {
			lines = append(lines, fmt.Sprintf("  %s", m.Name))
		}
	}

	// Render RW mutexes if any
	if rwmutexes, ok := byType[locktracker.LockTypeRWMutex]; ok && len(rwmutexes) > 0 {
		lines = append(lines, "RW Mutexes:")
		for _, m := range rwmutexes {
			lines = append(lines, fmt.Sprintf("  %s", m.Name))
		}
	}

	return lines
}

// renderProgressBar creates a visual progress bar.
// Example: [=====     ] for 50% full
func renderProgressBar(used, capacity int64, width int) string {
	if capacity == 0 {
		return "[" + strings.Repeat("-", width) + "]"
	}

	filled := int((used * int64(width)) / capacity)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	return "[" + strings.Repeat("=", filled) + strings.Repeat(" ", width-filled) + "]"
}
