# Work Queue Dispatcher Implementation Plan

## Summary

**Goal**: Replace spawn-all-goroutines model with dispatcher that only spawns workers when capacity available.

**Approach**: Clean cut - no fallback, no legacy code. Full replacement.

**Key Points**:
- LPT scheduling (heaviest first) via max-heap
- Single dispatcher goroutine pulls from queue
- Workers spawned only when semaphore slot acquired
- No blocked goroutines = no misleading "waiting" counts
- Tabs created upfront, immutable positions
- Red lamps kept for diagnostics (dep-blocked items)

## Overview

Replace the current "spawn-all-goroutines" model with a dispatcher-based work queue that only spawns worker goroutines when capacity is available. Uses LPT (Longest Processing Time First) scheduling - heaviest jobs first.

## Current Problem

```go
// Current: spawns ALL goroutines immediately
for _, w := range sortedWork {
    wg.Add(1)
    go func(spec workunit.UnitSpec) {  // 170 goroutines spawned
        defer wg.Done()
        // ... waits on semaphore (154 blocked, showing red lamps)
    }(w)
}
```

## Target Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      WorkQueue                               │
│  ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐         │
│  │ w=8 │ w=6 │ w=4 │ w=4 │ w=2 │ w=2 │ w=1 │ ... │ (heap)  │
│  └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘         │
│         ▲                                                    │
│         │ sorted by weight DESC (LPT)                       │
└─────────┼───────────────────────────────────────────────────┘
          │
          │ Pop() when:
          │   1. capacity available
          │   2. dependencies satisfied
          │
┌─────────┴───────────────────────────────────────────────────┐
│                     Dispatcher                               │
│                  (single goroutine)                          │
│                                                              │
│  for !queue.Empty() {                                        │
│      spec := queue.PopReady()  // blocks until ready         │
│      semaphore.Acquire(spec.Weight)                          │
│      go worker(spec)           // spawn only when scheduled  │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
          │
          │ spawns workers on-demand
          ▼
    ┌──────────┐  ┌──────────┐  ┌──────────┐
    │ worker 1 │  │ worker 2 │  │ worker 3 │  ... (max N = capacity)
    └──────────┘  └──────────┘  └──────────┘
```

## New Types

### work_queue.go

```go
package orchestrator

import (
    "container/heap"
    "sync"

    "github.com/ready-to-release/eac/go/core/workunit"
)

// WorkQueue is a thread-safe priority queue for work units.
// Items are ordered by weight descending (LPT scheduling).
type WorkQueue struct {
    mu       sync.Mutex
    items    workHeap
    deps     *DepsTracker
    notEmpty *sync.Cond  // signaled when items added or deps satisfied
    closed   bool
}

// NewWorkQueue creates a new work queue with dependency tracking.
func NewWorkQueue(work []workunit.UnitSpec) *WorkQueue {
    q := &WorkQueue{
        items: make(workHeap, 0, len(work)),
        deps:  NewDepsTracker(work),
    }
    q.notEmpty = sync.NewCond(&q.mu)

    // Add all items to heap (already tracks deps internally)
    for _, w := range work {
        heap.Push(&q.items, w)
    }

    return q
}

// PopReady blocks until a ready item is available, then returns it.
// An item is ready when all its DependsOn components have completed.
// Returns nil when queue is closed and empty.
func (q *WorkQueue) PopReady() *workunit.UnitSpec {
    q.mu.Lock()
    defer q.mu.Unlock()

    for {
        // Find highest-weight ready item
        if spec := q.findReady(); spec != nil {
            return spec
        }

        // No ready items - wait for signal
        if q.closed && len(q.items) == 0 {
            return nil
        }
        q.notEmpty.Wait()
    }
}

// findReady finds and removes the highest-weight item with satisfied deps.
// Must be called with mu held.
func (q *WorkQueue) findReady() *workunit.UnitSpec {
    // Scan heap for first ready item (heap is sorted by weight)
    for i, item := range q.items {
        if q.deps.IsReady(item.ID) {
            // Remove from heap
            heap.Remove(&q.items, i)
            return &item
        }
    }
    return nil
}

// MarkComplete marks a component as done, potentially unblocking dependents.
func (q *WorkQueue) MarkComplete(id workunit.UnitID) {
    q.mu.Lock()
    q.deps.MarkComplete(id)
    q.notEmpty.Broadcast()  // wake dispatcher to check for newly ready items
    q.mu.Unlock()
}

// Close signals that no more items will be added.
func (q *WorkQueue) Close() {
    q.mu.Lock()
    q.closed = true
    q.notEmpty.Broadcast()
    q.mu.Unlock()
}

// Len returns current queue length.
func (q *WorkQueue) Len() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    return len(q.items)
}

// Stats returns queue statistics for TUI display.
func (q *WorkQueue) Stats() QueueStats {
    q.mu.Lock()
    defer q.mu.Unlock()

    ready := 0
    blocked := 0
    for _, item := range q.items {
        if q.deps.IsReady(item.ID) {
            ready++
        } else {
            blocked++
        }
    }
    return QueueStats{
        Total:   len(q.items),
        Ready:   ready,
        Blocked: blocked,
    }
}

// QueueStats holds queue statistics for display.
type QueueStats struct {
    Total   int  // total items in queue
    Ready   int  // items ready to run (deps satisfied)
    Blocked int  // items waiting on dependencies
}

// --- Heap implementation ---

type workHeap []workunit.UnitSpec

func (h workHeap) Len() int           { return len(h) }
func (h workHeap) Less(i, j int) bool { return h[i].Weight > h[j].Weight }  // Max heap (LPT)
func (h workHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *workHeap) Push(x any) {
    *h = append(*h, x.(workunit.UnitSpec))
}

func (h *workHeap) Pop() any {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}
```

### deps_tracker.go

```go
package orchestrator

import (
    "github.com/ready-to-release/eac/go/core/workunit"
)

// DepsTracker tracks component completion for dependency resolution.
type DepsTracker struct {
    // completed tracks which components have finished
    // key: "module:component"
    completed map[string]bool

    // depsOf maps each unit to its dependencies
    // key: unit longname, value: list of dep component names
    depsOf map[string][]string

    // moduleOf maps each unit to its module (for resolving dep component names)
    moduleOf map[string]string
}

// NewDepsTracker creates a tracker for the given work units.
func NewDepsTracker(work []workunit.UnitSpec) *DepsTracker {
    dt := &DepsTracker{
        completed: make(map[string]bool),
        depsOf:    make(map[string][]string),
        moduleOf:  make(map[string]string),
    }

    for _, w := range work {
        key := w.ID.Longname()
        dt.depsOf[key] = w.DependsOnComponents()
        dt.moduleOf[key] = w.ID.Module
    }

    return dt
}

// IsReady returns true if all dependencies for the unit are completed.
func (dt *DepsTracker) IsReady(id workunit.UnitID) bool {
    key := id.Longname()
    deps := dt.depsOf[key]
    module := dt.moduleOf[key]

    for _, depComp := range deps {
        depKey := module + ":" + depComp
        if !dt.completed[depKey] {
            return false
        }
    }
    return true
}

// MarkComplete marks a component as done.
func (dt *DepsTracker) MarkComplete(id workunit.UnitID) {
    // Mark by module:component (not full longname with tool)
    key := id.Module + ":" + id.Component
    dt.completed[key] = true
}
```

## Modified RunComponents

### component_scheduler.go changes

```go
// RunComponents executes work units with dispatcher-based scheduling.
// Uses LPT (Longest Processing Time First) - heaviest jobs scheduled first.
// Only spawns worker goroutines when capacity is available.
func (cs *ComponentScheduler) RunComponents(work []workunit.UnitSpec, worker ComponentWorkerFunc) []ComponentResult {
    results := make([]ComponentResult, len(work))
    var resultsMu sync.Mutex
    var wg sync.WaitGroup

    // Create work queue (items already sorted by weight via heap)
    queue := NewWorkQueue(work)
    defer queue.Close()

    // Create all tabs as QUEUED upfront
    for _, w := range work {
        displayName := cs.formatDisplayName(w)
        cs.tuiMarkQueued(displayName, w.Weight)
    }

    // Dispatcher goroutine - pulls from queue, spawns workers
    dispatcherDone := make(chan struct{})
    go func() {
        defer close(dispatcherDone)

        for {
            // Get next ready item (blocks until available)
            spec := queue.PopReady()
            if spec == nil {
                return  // queue exhausted
            }

            displayName := cs.formatDisplayName(*spec)

            // Acquire capacity (this is the only place we wait for semaphore)
            cs.semaphore.Acquire(spec.Weight)

            // Update TUI: queued -> running
            cs.tuiMarkRunning(displayName)

            // Spawn worker
            wg.Add(1)
            go func(s workunit.UnitSpec) {
                defer wg.Done()
                defer cs.semaphore.Release(s.Weight)

                result := cs.executeWorker(s, worker)

                // Store result
                resultsMu.Lock()
                results[s.Index] = result
                resultsMu.Unlock()

                // Notify queue that this component is done (unblocks dependents)
                queue.MarkComplete(s.ID)

                // Update TUI
                cs.tuiMarkCompleted(cs.formatDisplayName(s), result.ExitCode)
            }(s)
        }
    }()

    // Wait for dispatcher to finish (means queue is empty)
    <-dispatcherDone

    // Wait for all workers to complete
    wg.Wait()

    return results
}

// executeWorker runs the actual work (extracted from processComponent).
// Called only after semaphore is acquired.
func (cs *ComponentScheduler) executeWorker(spec workunit.UnitSpec, worker ComponentWorkerFunc) ComponentResult {
    module := spec.ID.Module
    component := spec.ID.Component
    tool := spec.ID.Tool

    result := ComponentResult{
        Module:    module,
        Component: component,
        Handler:   tool,
    }

    startTime := time.Now()

    // Track active tool usage
    cs.addActiveTool(tool, spec.IsContainer)
    defer cs.removeActiveTool(tool, spec.IsContainer)

    // Create output directory
    sanitizedModule := sanitizePathForFS(output.PackageDisplayName(module))
    sanitizedComponent := sanitizePathForFS(component)
    componentOutputDir := filepath.Join(cs.config.WorkspaceRoot, cs.config.OutputBaseDir,
        sanitizedModule, sanitizedComponent)
    relLogPath := filepath.Join(cs.config.OutputBaseDir, sanitizedModule, sanitizedComponent,
        cs.config.LogFileName)

    if err := os.MkdirAll(componentOutputDir, 0o755); err != nil {
        result.ExitCode = 1
        result.Errors = []string{fmt.Sprintf("Failed to create directory: %v", err)}
        result.LogPath = relLogPath
        result.Duration = time.Since(startTime)
        return result
    }

    // Create log file
    logPath := filepath.Join(componentOutputDir, cs.config.LogFileName)
    logFile, err := os.Create(logPath)
    if err != nil {
        result.ExitCode = 1
        result.Errors = []string{fmt.Sprintf("Failed to create log file: %v", err)}
        result.LogPath = relLogPath
        result.Duration = time.Since(startTime)
        return result
    }

    // Create writer
    var workerWriter io.Writer
    displayName := cs.formatDisplayName(spec)
    if cs.tuiConsole != nil {
        workerWriter = cs.tuiConsole.NewWriter(displayName, logFile)
    } else {
        workerWriter = logFile
    }

    // Execute work
    exitCode := worker(module, component, workerWriter)

    // Close writers
    if closer, ok := workerWriter.(io.Closer); ok {
        closer.Close()
    }
    logFile.Close()

    // Parse log
    warnings, errors := parseLogForIssues(logPath)

    result.ExitCode = exitCode
    result.Warnings = warnings
    result.Errors = errors
    result.LogPath = relLogPath
    result.Duration = time.Since(startTime)

    return result
}

// formatDisplayName returns the display name for a work unit.
func (cs *ComponentScheduler) formatDisplayName(spec workunit.UnitSpec) string {
    if spec.ID.Tool != "" {
        return fmt.Sprintf("%s:%s:%s", spec.ID.Module, spec.ID.Component, spec.ID.Tool)
    }
    return fmt.Sprintf("%s:%s", spec.ID.Module, spec.ID.Component)
}

// tuiMarkQueued creates a tab in queued state (grey, waiting in queue).
func (cs *ComponentScheduler) tuiMarkQueued(displayName string, weight int) {
    if cs.tuiConsole == nil {
        return
    }
    cs.tuiConsole.StartModule(displayName, weight)  // Same as pending for now
}
```

## TUI Changes

### Red Lamps: Only for dependency-blocked items

With dispatcher model, there's no "waiting for capacity" state to show.
Red lamps now only appear when items are blocked on DependsOn dependencies.

```go
// view.go - lamps only for dependency-blocked items
if blocked := m.queueStats.Blocked; blocked > 0 {
    redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
    lamp := "●"
    displayCount := min(blocked, 16)
    lamps := strings.Repeat(lamp, displayCount)
    var waitStr string
    if blocked > displayCount {
        waitStr = fmt.Sprintf("%s+%d", lamps, blocked-displayCount)
    } else {
        waitStr = lamps
    }
    content += sep + white.Render("Locks: ") + redStyle.Render(waitStr)
}
```

### Tabs: Created upfront, positions immutable

All tabs created at start of RunComponents. Tab positions never change.
Only status/color changes as work progresses:
- Grey (queued) → Orange (running) → Green/Red (complete)

## Design Decisions (Resolved)

| Question | Decision |
|----------|----------|
| Show queue depth vs lamps? | **Keep lamps** - only for dep-blocked items (queue depth = future) |
| Tabs for queued items? | **Yes** - create all upfront, positions immutable |
| Handle turbo mid-run? | **No** - turbo sets ceiling at start only |

## Messages Update

### messages.go

```go
type Status struct {
    // ... existing fields ...

    // Queue statistics (for dependency blocking display)
    QueueStats QueueStats
}
```

## Files to Modify

| File | Action |
|------|--------|
| `orchestrator/work_queue.go` | **NEW** - WorkQueue with LPT heap |
| `orchestrator/deps_tracker.go` | **NEW** - Dependency tracking |
| `orchestrator/component_scheduler.go` | Refactor RunComponents to use dispatcher |
| `orchestrator/types.go` | Add QueueStats |
| `tui/messages.go` | Add QueueStats to Status |
| `tui/console/view.go` | Show queue depth instead of lock lamps |
| `tui/console/model.go` | Add queueStats field |
| `tui/console/update.go` | Handle QueueStats in status updates |

## Benefits

1. **No blocked goroutines** - only active workers exist as goroutines
2. **No misleading "locks" display** - queue depth is clearer metric
3. **LPT preserved** - heap ensures heaviest jobs pop first
4. **Dependencies handled cleanly** - tracker notifies queue when deps satisfied
5. **Simpler resource model** - goroutine count = actual parallelism

## Implementation Order

1. **deps_tracker.go** - New file, dependency tracking
2. **work_queue.go** - New file, LPT heap + queue logic
3. **component_scheduler.go** - Replace `RunComponents` with dispatcher model
4. **types.go** - Add `QueueStats`
5. **tui/messages.go** - Add `QueueStats` to `Status`
6. **tui/console/model.go** - Add `queueStats` field
7. **tui/console/update.go** - Handle `QueueStats`
8. **tui/console/view.go** - Update lamps to show dep-blocked only

## Testing

1. Verify LPT order: heaviest items scheduled first
2. Verify dependencies: blocked items wait for deps
3. Verify capacity: only N workers active at once
4. Verify TUI: lamps only for dep-blocked items
5. Verify completion: all results collected correctly
6. Verify tabs: all created upfront, positions stable
