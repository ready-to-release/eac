package orchestrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/shirou/gopsutil/v3/mem"
)

// ComponentScheduler manages parallel execution of component work items
// with weighted resource control and dependency ordering.
type ComponentScheduler struct {
	config    Config
	semaphore *WeightedSemaphore
	registry  *locktracker.Registry // Lock tracking registry for TUI visualization

	// Dynamic capacity management
	capacityTicker *time.Ticker    // Recalculates capacity every 2 seconds
	capacityStop   chan struct{}   // Signal to stop the capacity ticker
	configMax      int             // Maximum capacity from config (ceiling)
	turbo          int             // Turbo multiplier (1x, 2x, 4x, etc.)

	// Module completion tracking for inter-module dependencies
	moduleCompleteMu sync.RWMutex
	moduleComplete   map[string]bool           // module -> all components done
	moduleCompleteCh map[string]chan struct{}  // broadcast when module completes
	moduleCompCount  map[string]int            // module -> number of components
	moduleCompDone   map[string]int            // module -> components completed

	// Component completion tracking for intra-module dependencies (build_after)
	compCompleteMu sync.RWMutex
	compComplete   map[string]map[string]bool          // module -> component -> done
	compCompleteCh map[string]map[string]chan struct{} // module -> component -> broadcast

	// Component failure tracking
	compFailedMu sync.RWMutex
	compFailed   map[string]map[string]bool // module -> component -> failed

	// TUI console for real-time output display
	tuiConsole *tui.Console
	tuiCtx     interface{} // context.Context but we avoid import cycle

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int

	// Output infrastructure
	orchestratorOut io.Writer
}

// NewComponentScheduler creates a new scheduler with the given configuration.
// If registry is non-nil, the semaphore will be tracked for lock visualization.
// Starts a dynamic capacity ticker that adjusts capacity based on available system resources.
func NewComponentScheduler(config *Config, tuiConsole *tui.Console, registry *locktracker.Registry) *ComponentScheduler {
	// Calculate initial capacity based on available resources
	// Turbo multiplies the pressure roof: 1=normal, 2=2x, 4=4x
	// If turbo flag is set without a value, default to 4x
	turbo := config.Turbo
	if turbo < 1 {
		turbo = 1
	}
	initialCap := detectAvailableCapacity(config.MaxConcurrency, turbo)

	var sem *WeightedSemaphore
	if registry != nil {
		sem = NewWeightedSemaphoreWithRegistry("component-scheduler", initialCap, registry)
	} else {
		sem = NewWeightedSemaphore(initialCap)
	}

	cs := &ComponentScheduler{
		config:           *config,
		semaphore:        sem,
		registry:         registry,
		configMax:        config.MaxConcurrency,
		turbo:            turbo,
		capacityStop:     make(chan struct{}),
		moduleComplete:   make(map[string]bool),
		moduleCompleteCh: make(map[string]chan struct{}),
		moduleCompCount:  make(map[string]int),
		moduleCompDone:   make(map[string]int),
		compComplete:     make(map[string]map[string]bool),
		compCompleteCh:   make(map[string]map[string]chan struct{}),
		compFailed:       make(map[string]map[string]bool),
		tuiConsole:       tuiConsole,
	}

	// Configure output writer
	if config.TUI {
		cs.orchestratorOut = io.Discard
	} else {
		cs.orchestratorOut = os.Stdout
	}

	// Start dynamic capacity ticker
	cs.startCapacityTicker()

	return cs
}

// startCapacityTicker starts a goroutine that recalculates capacity every 2 seconds
// based on available system resources. The config max is used as a ceiling.
func (cs *ComponentScheduler) startCapacityTicker() {
	cs.capacityTicker = time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-cs.capacityTicker.C:
				newCap := detectAvailableCapacity(cs.configMax, cs.turbo)
				cs.semaphore.SetCapacity(newCap)
			case <-cs.capacityStop:
				cs.capacityTicker.Stop()
				return
			}
		}
	}()
}

// StopCapacityTicker stops the dynamic capacity ticker.
// Should be called when the scheduler is no longer needed.
func (cs *ComponentScheduler) StopCapacityTicker() {
	if cs.capacityStop != nil {
		close(cs.capacityStop)
		cs.capacityStop = nil
	}
}

// Close releases resources held by the scheduler.
// Should be called when the scheduler is no longer needed.
func (cs *ComponentScheduler) Close() {
	cs.StopCapacityTicker()
	if cs.semaphore != nil {
		cs.semaphore.Close()
	}
}

// detectAvailableCapacity calculates capacity as percentage of RAM.
// Normal: 100% of (RAM/1GB) slots
// detectAvailableCapacity calculates the pressure roof for parallel builds.
//
// Formula: min(CPU count, RAM_GB / 2) × turbo
//
// This gives predictable results:
//   - 16 CPU, 32 GB RAM → min(16, 16) = 16 base slots
//   - 8 CPU, 16 GB RAM → min(8, 8) = 8 base slots
//   - 4 CPU, 8 GB RAM → min(4, 4) = 4 base slots
//
// With turbo=2: double the slots (for I/O bound builds)
// With turbo=4: quadruple the slots
//
// configMax is only used as ceiling if > 0 (user explicitly set it).
// Returns at least 1.
func detectAvailableCapacity(configMax int, turbo int) int {
	// Use HOST resources (not Docker limits) since orchestrator runs on host
	cpuCount := runtime.NumCPU()
	if cpuCount < 1 {
		cpuCount = 4 // Fallback
	}

	// Get host total RAM in GB
	var ramGB int
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		ramGB = int(memInfo.Total / (1024 * 1024 * 1024))
	} else {
		ramGB = 8 // Fallback: assume 8GB
	}

	// Base capacity: min(CPU count, RAM_GB / 2)
	// This ensures we don't over-subscribe either resource
	ramBasedCap := ramGB / 2
	if ramBasedCap < 1 {
		ramBasedCap = 1
	}

	baseCapacity := cpuCount
	if ramBasedCap < baseCapacity {
		baseCapacity = ramBasedCap
	}

	// Apply turbo multiplier
	capacity := baseCapacity * turbo

	// Cap at reasonable maximum:
	// - Without turbo (turbo=1): cap at CPU count
	// - With turbo: cap at 2x CPU count (max 64)
	var maxCapacity int
	if turbo <= 1 {
		maxCapacity = cpuCount
	} else {
		maxCapacity = cpuCount * 2
		if maxCapacity > 64 {
			maxCapacity = 64
		}
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}

	// Ensure at least 1
	if capacity < 1 {
		capacity = 1
	}

	// Only apply configMax ceiling if user explicitly set it (> 0)
	if configMax > 0 && capacity > configMax {
		return configMax
	}

	return capacity
}

// InitializeWork prepares the scheduler for a batch of component work items.
// Must be called before RunComponents.
func (cs *ComponentScheduler) InitializeWork(work []ComponentWork) {
	// Reset state
	cs.moduleComplete = make(map[string]bool)
	cs.moduleCompleteCh = make(map[string]chan struct{})
	cs.moduleCompCount = make(map[string]int)
	cs.moduleCompDone = make(map[string]int)
	cs.compComplete = make(map[string]map[string]bool)
	cs.compCompleteCh = make(map[string]map[string]chan struct{})
	cs.compFailed = make(map[string]map[string]bool)

	// Count components per module and initialize tracking maps
	for _, w := range work {
		cs.moduleCompCount[w.Module]++

		// Initialize module completion channel if needed
		if _, ok := cs.moduleCompleteCh[w.Module]; !ok {
			cs.moduleCompleteCh[w.Module] = make(chan struct{})
		}

		// Initialize component tracking for this module if needed
		if cs.compComplete[w.Module] == nil {
			cs.compComplete[w.Module] = make(map[string]bool)
			cs.compCompleteCh[w.Module] = make(map[string]chan struct{})
			cs.compFailed[w.Module] = make(map[string]bool)
		}

		// Initialize component completion channel
		cs.compCompleteCh[w.Module][w.Component] = make(chan struct{})
	}

	// Initialize TUI status
	cs.tuiTotal = len(work)
	cs.tuiCompleted = 0
	cs.tuiRunning = nil
}

// RunComponents executes component work items in parallel with weighted scheduling.
// Components respect intra-module dependencies (BuildAfter) and weighted resource limits.
// Uses LPT (Longest Processing Time First) scheduling: heaviest jobs start first
// to ensure heavy jobs are distributed throughout execution.
// Returns results in the same order as the input work items.
func (cs *ComponentScheduler) RunComponents(work []ComponentWork, worker ComponentWorkerFunc) []ComponentResult {
	results := make([]ComponentResult, len(work))
	var wg sync.WaitGroup

	// Sort by weight descending (LPT scheduling) - heaviest jobs first
	// This ensures heavy jobs start early and are distributed across execution time
	sortedWork := make([]ComponentWork, len(work))
	copy(sortedWork, work)
	sort.Slice(sortedWork, func(i, j int) bool {
		return sortedWork[i].Weight > sortedWork[j].Weight
	})

	// Launch components in LPT order - heaviest first
	// Results are placed by original Index, so output order is preserved
	for _, w := range sortedWork {
		wg.Add(1)
		go func(cw ComponentWork) {
			defer wg.Done()
			result := cs.processComponent(cw, worker)
			results[cw.Index] = result
		}(w)
	}

	wg.Wait()
	return results
}

// processComponent handles a single component build with dependency waiting and weighted scheduling.
func (cs *ComponentScheduler) processComponent(work ComponentWork, worker ComponentWorkerFunc) ComponentResult {
	startTime := time.Now()
	result := ComponentResult{
		Module:    work.Module,
		Component: work.Component,
	}

	// 1. Wait for intra-module component dependencies (build_after)
	for _, depComp := range work.BuildAfter {
		// Check if dependency component exists in this module
		cs.compCompleteMu.RLock()
		ch, exists := cs.compCompleteCh[work.Module][depComp]
		cs.compCompleteMu.RUnlock()

		if exists {
			// Wait for dependency to complete
			<-ch

			// Check if dependency failed - skip if so
			cs.compFailedMu.RLock()
			depFailed := cs.compFailed[work.Module][depComp]
			cs.compFailedMu.RUnlock()

			if depFailed {
				result.ExitCode = 1
				result.Errors = []string{fmt.Sprintf("Skipped: dependency %s failed", depComp)}
				result.Duration = time.Since(startTime)
				cs.markComponentComplete(work, &result)
				return result
			}
		}
	}

	// Display name for TUI (module:component)
	displayName := fmt.Sprintf("%s:%s", work.Module, work.Component)

	// 2. Acquire weighted slot
	weight := work.Weight
	if weight <= 0 {
		weight = 1
	}

	// Create tab as PENDING before acquiring slot (shows scheduled but waiting)
	cs.tuiMarkPending(displayName, weight)

	cs.semaphore.Acquire(weight)
	// Note: We explicitly release the semaphore BEFORE marking complete in TUI
	// to ensure the pressure gauge and running tabs stay in sync.
	// Do not use defer here - it would release after TUI update.

	// Mark as RUNNING after acquiring slot
	cs.tuiMarkRunning(displayName)

	// 3. Create output directory for this component
	// Structure: out/build/<module>/<component> (e.g., out/build/books/howto)
	sanitizedModule := sanitizePathForFS(output.PackageDisplayName(work.Module))
	sanitizedComponent := sanitizePathForFS(work.Component)
	componentOutputDir := filepath.Join(cs.config.WorkspaceRoot, cs.config.OutputBaseDir, sanitizedModule, sanitizedComponent)

	// Relative log path for result reporting: out/build/<module>/<component>/build.log
	relLogPath := filepath.Join(cs.config.OutputBaseDir, sanitizedModule, sanitizedComponent, cs.config.LogFileName)

	if err := os.MkdirAll(componentOutputDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		cs.semaphore.Release(weight) // Release before TUI update
		cs.markComponentComplete(work, &result)
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
		cs.semaphore.Release(weight) // Release before TUI update
		cs.markComponentComplete(work, &result)
		return result
	}

	// Create writer for worker
	var workerWriter io.Writer
	if cs.tuiConsole != nil {
		workerWriter = cs.tuiConsole.NewWriter(displayName, logFile)
	} else {
		workerWriter = logFile
	}

	// 4. Execute build with memory instrumentation
	memBefore := GetMemoryStats()
	fmt.Fprintf(logFile, "[memory] before: used=%s avail=%s total=%s (%.1f%%)\n",
		FormatBytes(memBefore.UsedBytes), FormatBytes(memBefore.AvailableBytes),
		FormatBytes(memBefore.TotalBytes), memBefore.UsedPercent)

	exitCode := worker(work.Module, work.Component, workerWriter)

	memAfter := GetMemoryStats()
	memDelta := int64(memAfter.UsedBytes) - int64(memBefore.UsedBytes)
	deltaSign := "+"
	if memDelta < 0 {
		deltaSign = ""
	}
	fmt.Fprintf(logFile, "[memory] after: used=%s avail=%s total=%s (%.1f%%) delta=%s%s\n",
		FormatBytes(memAfter.UsedBytes), FormatBytes(memAfter.AvailableBytes),
		FormatBytes(memAfter.TotalBytes), memAfter.UsedPercent,
		deltaSign, FormatBytes(uint64(abs64(memDelta))))

	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = relLogPath
	result.Duration = time.Since(startTime)

	// Release semaphore BEFORE marking complete in TUI
	// This ensures pressure gauge reflects the release immediately
	cs.semaphore.Release(weight)

	// Mark as completed (TUI will now show correct pressure)
	cs.markComponentComplete(work, &result)

	return result
}

// markComponentComplete marks a component as done and broadcasts to waiters.
func (cs *ComponentScheduler) markComponentComplete(work ComponentWork, result *ComponentResult) {
	// Mark component complete
	cs.compCompleteMu.Lock()
	cs.compComplete[work.Module][work.Component] = true
	if result.ExitCode != 0 {
		cs.compFailed[work.Module][work.Component] = true
	}
	ch := cs.compCompleteCh[work.Module][work.Component]
	cs.compCompleteMu.Unlock()

	// Broadcast completion
	close(ch)

	// Update module completion tracking
	cs.moduleCompleteMu.Lock()
	cs.moduleCompDone[work.Module]++
	if cs.moduleCompDone[work.Module] >= cs.moduleCompCount[work.Module] {
		cs.moduleComplete[work.Module] = true
		close(cs.moduleCompleteCh[work.Module])
	}
	cs.moduleCompleteMu.Unlock()

	// Update TUI with exit code
	displayName := fmt.Sprintf("%s:%s", work.Module, work.Component)
	cs.tuiMarkCompleted(displayName, result.ExitCode)
}

// WaitForModule blocks until all components of a module are complete.
// Used for inter-module dependencies.
func (cs *ComponentScheduler) WaitForModule(module string) {
	cs.moduleCompleteMu.RLock()
	ch, exists := cs.moduleCompleteCh[module]
	cs.moduleCompleteMu.RUnlock()

	if exists {
		<-ch
	}
}

// IsModuleFailed returns true if any component in the module failed.
func (cs *ComponentScheduler) IsModuleFailed(module string) bool {
	cs.compFailedMu.RLock()
	defer cs.compFailedMu.RUnlock()

	if failures, ok := cs.compFailed[module]; ok {
		for _, failed := range failures {
			if failed {
				return true
			}
		}
	}
	return false
}

// tuiMarkPending creates a component tab in pending state (scheduled, waiting for slot).
func (cs *ComponentScheduler) tuiMarkPending(displayName string, weight int) {
	if cs.tuiConsole == nil {
		return
	}
	// Create the tab with pending status
	cs.tuiConsole.StartModule(displayName, weight)
}

// tuiMarkRunning marks a component as actively running (slot acquired).
func (cs *ComponentScheduler) tuiMarkRunning(displayName string) {
	if cs.tuiConsole == nil {
		return
	}

	// Update status to running
	cs.tuiConsole.MarkModuleRunning(displayName)

	cs.tuiMu.Lock()
	cs.tuiRunning = append(cs.tuiRunning, displayName)
	running := make([]string, len(cs.tuiRunning))
	copy(running, cs.tuiRunning)
	completed := cs.tuiCompleted
	total := cs.tuiTotal
	cs.tuiMu.Unlock()

	cs.tuiConsole.UpdateStatus(tui.Status{
		Phase:     capitalize(cs.config.ActionVerb),
		Running:   running,
		Completed: completed,
		Total:     total,
		Locks:     cs.getLockStatuses(),
	})
}

// tuiMarkCompleted removes a component from running, increments completed, and reports exit code.
func (cs *ComponentScheduler) tuiMarkCompleted(displayName string, exitCode int) {
	if cs.tuiConsole == nil {
		return
	}

	// First, send the completion message with exit code so TUI knows success/failure
	cs.tuiConsole.MarkModuleComplete(displayName, exitCode)

	cs.tuiMu.Lock()
	// Remove from running list
	for i, m := range cs.tuiRunning {
		if m == displayName {
			cs.tuiRunning = append(cs.tuiRunning[:i], cs.tuiRunning[i+1:]...)
			break
		}
	}
	cs.tuiCompleted++
	running := make([]string, len(cs.tuiRunning))
	copy(running, cs.tuiRunning)
	completed := cs.tuiCompleted
	total := cs.tuiTotal
	cs.tuiMu.Unlock()

	cs.tuiConsole.UpdateStatus(tui.Status{
		Phase:     capitalize(cs.config.ActionVerb),
		Running:   running,
		Completed: completed,
		Total:     total,
		Locks:     cs.getLockStatuses(),
	})
}

// getLockStatuses returns current lock states from the registry.
func (cs *ComponentScheduler) getLockStatuses() []tui.LockStatus {
	if cs.registry == nil {
		return nil
	}

	snapshot := cs.registry.Snapshot()
	if len(snapshot) == 0 {
		return nil
	}

	locks := make([]tui.LockStatus, 0, len(snapshot))
	for _, info := range snapshot {
		locks = append(locks, tui.LockStatus{
			Name:     info.Name,
			Type:     string(info.Type),
			Capacity: int(info.Capacity),
			Used:     int(info.Used),
			Waiting:  int(info.Waiting),
		})
	}
	return locks
}

// AggregateToWorkResults converts component results to module-level WorkResults.
// This maintains compatibility with existing code that expects WorkResult.
func AggregateToWorkResults(compResults []ComponentResult, work []ComponentWork) []WorkResult {
	// Group results by module
	moduleResults := make(map[string][]ComponentResult)
	for _, r := range compResults {
		moduleResults[r.Module] = append(moduleResults[r.Module], r)
	}

	// Get unique modules in order
	seenModules := make(map[string]bool)
	var moduleOrder []string
	for _, w := range work {
		if !seenModules[w.Module] {
			seenModules[w.Module] = true
			moduleOrder = append(moduleOrder, w.Module)
		}
	}

	// Create WorkResult for each module
	results := make([]WorkResult, 0, len(moduleOrder))
	for i, module := range moduleOrder {
		compResults := moduleResults[module]
		if len(compResults) == 0 {
			continue
		}

		// Aggregate component results
		wr := WorkResult{
			Moniker: module,
			Index:   i,
		}

		var maxDuration time.Duration
		var firstFailedLogPath string
		for _, cr := range compResults {
			// Any component failure = module failure
			if cr.ExitCode != 0 {
				wr.ExitCode = 1
				// Track first failed component's log path
				if firstFailedLogPath == "" {
					firstFailedLogPath = cr.LogPath
				}
			}
			// Collect all warnings and errors
			wr.Warnings = append(wr.Warnings, cr.Warnings...)
			wr.Errors = append(wr.Errors, cr.Errors...)
			// Track max duration (since components run in parallel)
			if cr.Duration > maxDuration {
				maxDuration = cr.Duration
			}
		}
		wr.Duration = maxDuration

		// Use first failed component's log path, or first component's log if all succeeded
		if firstFailedLogPath != "" {
			wr.LogPath = firstFailedLogPath
		} else if len(compResults) > 0 {
			wr.LogPath = compResults[0].LogPath
		}

		results = append(results, wr)
	}

	return results
}

// AggregateToComponentResultSets groups component results by module.
// Results are sorted alphabetically by module name, and components within each module
// are sorted alphabetically by component name.
// Status and Duration are computed for each module.
func AggregateToComponentResultSets(results []ComponentResult) []ComponentResultSet {
	if len(results) == 0 {
		return []ComponentResultSet{}
	}

	// Group results by module
	moduleResults := make(map[string][]ComponentResult)
	for _, r := range results {
		moduleResults[r.Module] = append(moduleResults[r.Module], r)
	}

	// Get sorted module names
	modules := make([]string, 0, len(moduleResults))
	for m := range moduleResults {
		modules = append(modules, m)
	}
	// Sort modules alphabetically
	sort.Strings(modules)

	// Build result sets
	resultSets := make([]ComponentResultSet, 0, len(modules))
	for _, module := range modules {
		components := moduleResults[module]

		// Sort components alphabetically by name
		sortedComponents := make([]ComponentResult, len(components))
		copy(sortedComponents, components)
		sort.Slice(sortedComponents, func(i, j int) bool {
			return sortedComponents[i].Component < sortedComponents[j].Component
		})

		// Compute max duration
		var maxDuration time.Duration
		for _, c := range sortedComponents {
			if c.Duration > maxDuration {
				maxDuration = c.Duration
			}
		}

		// Create result set and derive status
		rs := ComponentResultSet{
			Module:     module,
			Components: sortedComponents,
			Duration:   maxDuration,
		}
		rs.Status = rs.DeriveStatus()

		resultSets = append(resultSets, rs)
	}

	return resultSets
}

// Note: sanitizePathForFS is defined in orchestrator.go

// abs64 returns the absolute value of an int64.
func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
