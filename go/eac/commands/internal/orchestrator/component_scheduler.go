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

	"github.com/ready-to-release/eac/go/eac/commands/internal/capacity"
	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/workunit"
	"github.com/shirou/gopsutil/v3/mem"
)

// ComponentExtras holds additional data passed from workers to component results.
// Used for test-specific fields that need to flow from the test runner to the summary.
type ComponentExtras struct {
	TestsTotal   int
	TestsPassed  int
	TestsFailed  int
	TestsSkipped int
}

// ComponentScheduler manages parallel execution of component work items
// with weighted resource control and dependency ordering.
type ComponentScheduler struct {
	config    Config
	semaphore *capacity.GlobalSemaphore // Global cross-process semaphore
	registry  *locktracker.Registry     // Lock tracking registry for TUI visualization

	// Dynamic capacity management
	capacityTicker *time.Ticker    // Recalculates capacity every 2 seconds
	capacityStop   chan struct{}   // Signal to stop the capacity ticker
	configMax      int             // Maximum capacity from config (ceiling)
	turbo          float64         // Turbo multiplier (1.0x, 1.25x, 2.0x, etc.)

	// Module completion tracking for inter-module dependencies
	moduleCompleteMu sync.RWMutex
	moduleComplete   map[string]bool           // module -> all components done
	moduleCompleteCh map[string]chan struct{}  // broadcast when module completes
	moduleCompCount  map[string]int            // module -> number of components
	moduleCompDone   map[string]int            // module -> components completed

	// Component completion tracking for intra-module dependencies (build_after)
	compCompleteMu   sync.RWMutex
	compComplete     map[string]map[string]bool          // module -> component -> done
	compCompleteCh   map[string]map[string]chan struct{} // module -> component -> broadcast
	compHandlerCount map[string]map[string]int           // module -> component -> total handlers
	compHandlerDone  map[string]map[string]int           // module -> component -> handlers completed

	// Component failure tracking
	compFailedMu sync.RWMutex
	compFailed   map[string]map[string]bool // module -> component -> failed

	// Component extras tracking (test counts, etc.)
	compExtrasMu sync.RWMutex
	compExtras   map[string]map[string]ComponentExtras // module -> component -> extras

	// TUI console for real-time output display
	tuiConsole *tui.Console
	tuiCtx     interface{} // context.Context but we avoid import cycle

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int

	// Tools tracking (protected by toolsMu) - separated by type
	toolsMu          sync.Mutex
	activeContainers map[string]int  // container tool -> count of running components
	usedContainers   map[string]bool // all container tools ever used
	activeSystem     map[string]int  // system tool -> count of running components
	usedSystem       map[string]bool // all system tools ever used

	// Output infrastructure
	orchestratorOut io.Writer

	// Cache times for displaying when cached artifacts were built
	cacheTimesMu sync.RWMutex
	cacheTimes   map[string]time.Time // module -> time when artifact was last built

	// Summary builder for incremental summary computation
	summaryBuilder SummaryBuilder

	// Work queue for dispatcher-based scheduling
	workQueue *WorkQueue
}

// NewComponentScheduler creates a new scheduler with the given configuration.
// If registry is non-nil, the semaphore will be tracked for lock visualization.
// Starts a dynamic capacity ticker that adjusts capacity based on available system resources.
// Uses a GLOBAL semaphore shared across all processes (build, test, lint, scan).
func NewComponentScheduler(config *Config, tuiConsole *tui.Console, registry *locktracker.Registry) *ComponentScheduler {
	// Calculate initial capacity based on available resources
	// Turbo multiplies the pressure roof: 1.0=normal, 1.25=+25%, 2.0=2x
	// If turbo flag is set without a value, default to 1.25x
	turbo := config.Turbo
	if turbo < 1.0 {
		turbo = 1.0
	}
	initialCap := detectAvailableCapacity(config.MaxConcurrency, turbo)

	// Create GLOBAL semaphore - shared across all processes via filesystem
	sem := capacity.NewGlobalSemaphore(config.WorkspaceRoot, initialCap, registry)

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
		compHandlerCount: make(map[string]map[string]int),
		compHandlerDone:  make(map[string]map[string]int),
		compFailed:       make(map[string]map[string]bool),
		compExtras:       make(map[string]map[string]ComponentExtras),
		activeContainers: make(map[string]int),
		usedContainers:   make(map[string]bool),
		activeSystem:     make(map[string]int),
		usedSystem:       make(map[string]bool),
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
	cs.capacityTicker = time.NewTicker(config.CapacityRecalcInterval())

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

// SetCacheTimes sets the cache times for modules that are up-to-date.
// These times are displayed in the TUI for cached modules to show when
// the cached artifacts were originally built.
func (cs *ComponentScheduler) SetCacheTimes(times map[string]time.Time) {
	cs.cacheTimesMu.Lock()
	defer cs.cacheTimesMu.Unlock()
	cs.cacheTimes = times
}

// getCacheTime returns the cache time for a module, or zero time if not cached.
func (cs *ComponentScheduler) getCacheTime(module string) time.Time {
	cs.cacheTimesMu.RLock()
	defer cs.cacheTimesMu.RUnlock()
	if cs.cacheTimes == nil {
		return time.Time{}
	}
	return cs.cacheTimes[module]
}

// SetSummaryBuilder sets the summary builder for incremental summary computation.
// The builder receives component results as they complete.
func (cs *ComponentScheduler) SetSummaryBuilder(builder SummaryBuilder) {
	cs.summaryBuilder = builder
}

// detectAvailableCapacity calculates the pressure roof for parallel builds.
// If --roof is set, uses that value directly. Otherwise auto-detects from system resources.
func detectAvailableCapacity(configMax int, turbo float64) int {
	// Get actual system resources
	cpuCount := runtime.NumCPU()
	if cpuCount < 1 {
		cpuCount = 4 // Fallback
	}

	var ramGB int
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		ramGB = int(memInfo.Total / (1024 * 1024 * 1024))
	} else {
		ramGB = 8 // Fallback: assume 8GB
	}

	return calculateCapacity(cpuCount, ramGB, configMax, turbo)
}

// calculateCapacity computes the effective parallelism capacity.
//
// If roof > 0: use roof directly (--roof flag overrides everything)
// If roof == 0: auto-detect from system resources
func calculateCapacity(cpuCount, ramGB, roof int, turbo float64) int {
	// --roof overrides all
	if roof > 0 {
		return roof
	}

	// Auto-detect: min(CPU, RAM/3) × turbo
	// RAM/3 because each weight unit uses ~2.5GB + overhead
	// This ensures we don't overcommit memory on lower-spec machines
	ramCap := ramGB / 3
	if ramCap < 1 {
		ramCap = 1
	}

	base := cpuCount
	if ramCap < base {
		base = ramCap
	}

	capacity := int(float64(base) * turbo)

	// Cap at CPU count (or 2x with turbo, max 64)
	maxCap := cpuCount
	if turbo > 1.0 {
		maxCap = cpuCount * 2
		if maxCap > 64 {
			maxCap = 64
		}
	}
	if capacity > maxCap {
		capacity = maxCap
	}

	if capacity < 1 {
		capacity = 1
	}

	return capacity
}

// InitializeWork prepares the scheduler for a batch of work units.
// Must be called before RunComponents.
func (cs *ComponentScheduler) InitializeWork(work []workunit.UnitSpec) {
	// Reset state
	cs.moduleComplete = make(map[string]bool)
	cs.moduleCompleteCh = make(map[string]chan struct{})
	cs.moduleCompCount = make(map[string]int)
	cs.moduleCompDone = make(map[string]int)
	cs.compComplete = make(map[string]map[string]bool)
	cs.compCompleteCh = make(map[string]map[string]chan struct{})
	cs.compHandlerCount = make(map[string]map[string]int)
	cs.compHandlerDone = make(map[string]map[string]int)
	cs.compFailed = make(map[string]map[string]bool)
	cs.compExtras = make(map[string]map[string]ComponentExtras)
	cs.activeContainers = make(map[string]int)
	cs.usedContainers = make(map[string]bool)
	cs.activeSystem = make(map[string]int)
	cs.usedSystem = make(map[string]bool)

	// Count components per module and initialize tracking maps
	for _, w := range work {
		module := w.ID.Module
		component := w.ID.Component

		cs.moduleCompCount[module]++

		// Initialize module completion channel if needed
		if _, ok := cs.moduleCompleteCh[module]; !ok {
			cs.moduleCompleteCh[module] = make(chan struct{})
		}

		// Initialize component tracking for this module if needed
		if cs.compComplete[module] == nil {
			cs.compComplete[module] = make(map[string]bool)
			cs.compCompleteCh[module] = make(map[string]chan struct{})
			cs.compHandlerCount[module] = make(map[string]int)
			cs.compHandlerDone[module] = make(map[string]int)
			cs.compFailed[module] = make(map[string]bool)
		}

		// Count handlers per component and create channel only once per component
		cs.compHandlerCount[module][component]++
		if _, exists := cs.compCompleteCh[module][component]; !exists {
			cs.compCompleteCh[module][component] = make(chan struct{})
		}
	}

	// Initialize TUI status
	cs.tuiTotal = len(work)
	cs.tuiCompleted = 0
	cs.tuiRunning = nil
}

// RunComponents executes work units with worker pool scheduling.
// Uses LPT (Longest Processing Time First) - heaviest jobs scheduled first.
// Spawns a pool of worker goroutines that pull from the queue concurrently.
// Returns results in the same order as the input work items.
func (cs *ComponentScheduler) RunComponents(work []workunit.UnitSpec, worker ComponentWorkerFunc) []ComponentResult {
	results := make([]ComponentResult, len(work))
	var resultsMu sync.Mutex
	var wg sync.WaitGroup

	// Create work queue (items ordered by weight via LPT heap)
	queue := NewWorkQueue(work)
	cs.workQueue = queue // Store for stats access

	// Create all tabs as QUEUED upfront (positions are immutable)
	for _, w := range work {
		displayName := cs.formatDisplayName(w)
		cs.tuiMarkQueued(displayName, w.Weight)
	}

	// Determine worker pool size: min(work items, capacity)
	// This ensures we can saturate capacity immediately
	poolSize := len(work)
	capacity := cs.semaphore.Capacity()
	if capacity > 0 && poolSize > capacity {
		poolSize = capacity
	}

	// Spawn worker pool - all workers start immediately and compete for queue items
	// Workers use bin-packing: check available capacity, pop item that fits
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				// Check available capacity for bin-packing
				// This is approximate (race-safe) - actual acquisition may differ
				available := cs.semaphore.Available()
				if available < 1 {
					available = 1 // At minimum, look for weight-1 items
				}

				// Pop item that fits available capacity (bin-packing)
				// Prioritizes heavier items that fit, enabling parallel small jobs
				spec := queue.PopReadyWithBudget(available)
				if spec == nil {
					return // queue exhausted
				}

				weight := spec.Weight
				if weight <= 0 {
					weight = 1
				}

				// Acquire capacity (blocks until slot available)
				cs.semaphore.Acquire(weight)

				// Update TUI: queued -> running
				cs.tuiMarkRunning(cs.formatDisplayName(*spec))

				// Execute work
				result := cs.executeWorker(*spec, worker)

				// Release capacity
				cs.semaphore.Release(weight)

				// Store result by original index
				resultsMu.Lock()
				results[spec.Index] = result
				resultsMu.Unlock()

				// Notify queue that this component is done (unblocks dependents)
				queue.MarkComplete(spec.ID)

				// Mark component complete (broadcasts to legacy channels, updates TUI)
				cs.markComponentComplete(*spec, &result)
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()

	// Close queue
	queue.Close()

	return results
}

// markComponentComplete marks a component as done and broadcasts to waiters.
func (cs *ComponentScheduler) markComponentComplete(spec workunit.UnitSpec, result *ComponentResult) {
	module := spec.ID.Module
	component := spec.ID.Component
	tool := spec.ID.Tool

	// Track handler completion and only close channel when all handlers are done
	cs.compCompleteMu.Lock()
	if result.ExitCode != 0 {
		cs.compFailed[module][component] = true
	}
	cs.compHandlerDone[module][component]++
	handlersDone := cs.compHandlerDone[module][component]
	handlersTotal := cs.compHandlerCount[module][component]
	allHandlersDone := handlersDone >= handlersTotal
	var ch chan struct{}
	if allHandlersDone {
		cs.compComplete[module][component] = true
		ch = cs.compCompleteCh[module][component]
	}
	cs.compCompleteMu.Unlock()

	// Broadcast completion only when all handlers for this component are done
	if allHandlersDone {
		close(ch)
	}

	// Update module completion tracking
	cs.moduleCompleteMu.Lock()
	cs.moduleCompDone[module]++
	if cs.moduleCompDone[module] >= cs.moduleCompCount[module] {
		cs.moduleComplete[module] = true
		close(cs.moduleCompleteCh[module])
	}
	cs.moduleCompleteMu.Unlock()

	// Send result to summary builder for incremental summary computation
	if cs.summaryBuilder != nil {
		cs.summaryBuilder.AddResult(*result)
	}

	// Update TUI with exit code
	// Use same format as tuiMarkPending/tuiMarkRunning: module:component:tool
	var displayName string
	if tool != "" {
		displayName = fmt.Sprintf("%s:%s:%s", module, component, tool)
	} else {
		displayName = fmt.Sprintf("%s:%s", module, component)
	}
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

// SetComponentExtras stores additional data for a component result.
// This is called by workers to pass test counts or other data that will be
// merged into the ComponentResult when processing completes.
func (cs *ComponentScheduler) SetComponentExtras(module, component string, extras ComponentExtras) {
	cs.compExtrasMu.Lock()
	defer cs.compExtrasMu.Unlock()

	if cs.compExtras[module] == nil {
		cs.compExtras[module] = make(map[string]ComponentExtras)
	}
	cs.compExtras[module][component] = extras
}

// getComponentExtras retrieves the extras for a component, if any.
func (cs *ComponentScheduler) getComponentExtras(module, component string) (ComponentExtras, bool) {
	cs.compExtrasMu.RLock()
	defer cs.compExtrasMu.RUnlock()

	if moduleExtras, ok := cs.compExtras[module]; ok {
		if extras, ok := moduleExtras[component]; ok {
			return extras, true
		}
	}
	return ComponentExtras{}, false
}

// addActiveTool increments the usage count for a tool/handler.
// Also tracks docker as active when container tools are used.
func (cs *ComponentScheduler) addActiveTool(handler string, isContainer bool) {
	if handler == "" {
		return
	}

	cs.toolsMu.Lock()
	if isContainer {
		cs.activeContainers[handler]++
		cs.usedContainers[handler] = true
		// Container tools require docker - track it as active system tool
		cs.activeSystem["docker"]++
		cs.usedSystem["docker"] = true
	} else {
		cs.activeSystem[handler]++
		cs.usedSystem[handler] = true
	}
	cs.toolsMu.Unlock()
}

// removeActiveTool decrements the usage count for a tool/handler.
// Also decrements docker count when container tools finish.
func (cs *ComponentScheduler) removeActiveTool(handler string, isContainer bool) {
	if handler == "" {
		return
	}

	cs.toolsMu.Lock()
	if isContainer {
		if cs.activeContainers[handler] > 0 {
			cs.activeContainers[handler]--
			if cs.activeContainers[handler] == 0 {
				delete(cs.activeContainers, handler)
			}
		}
		// Decrement docker usage count
		if cs.activeSystem["docker"] > 0 {
			cs.activeSystem["docker"]--
			if cs.activeSystem["docker"] == 0 {
				delete(cs.activeSystem, "docker")
			}
		}
	} else {
		if cs.activeSystem[handler] > 0 {
			cs.activeSystem[handler]--
			if cs.activeSystem[handler] == 0 {
				delete(cs.activeSystem, handler)
			}
		}
	}
	cs.toolsMu.Unlock()
}

// getActiveContainersList returns a sorted list of currently active container tools.
func (cs *ComponentScheduler) getActiveContainersList() []string {
	cs.toolsMu.Lock()
	defer cs.toolsMu.Unlock()

	if len(cs.activeContainers) == 0 {
		return nil
	}

	tools := make([]string, 0, len(cs.activeContainers))
	for tool := range cs.activeContainers {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedContainersList returns a sorted list of all container tools that have been used.
func (cs *ComponentScheduler) getUsedContainersList() []string {
	cs.toolsMu.Lock()
	defer cs.toolsMu.Unlock()

	if len(cs.usedContainers) == 0 {
		return nil
	}

	tools := make([]string, 0, len(cs.usedContainers))
	for tool := range cs.usedContainers {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getActiveSystemToolsList returns a sorted list of currently active system tools.
func (cs *ComponentScheduler) getActiveSystemToolsList() []string {
	cs.toolsMu.Lock()
	defer cs.toolsMu.Unlock()

	if len(cs.activeSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(cs.activeSystem))
	for tool := range cs.activeSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedSystemToolsList returns a sorted list of all system tools that have been used.
func (cs *ComponentScheduler) getUsedSystemToolsList() []string {
	cs.toolsMu.Lock()
	defer cs.toolsMu.Unlock()

	if len(cs.usedSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(cs.usedSystem))
	for tool := range cs.usedSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// tuiMarkPending creates a component tab in pending state (scheduled, waiting for slot).
func (cs *ComponentScheduler) tuiMarkPending(displayName string, weight int) {
	if cs.tuiConsole == nil {
		return
	}
	// Create the tab with pending status
	cs.tuiConsole.StartModule(displayName, weight)
}

// tuiMarkQueued creates a component tab in queued state (waiting in work queue).
// This is the same visual as pending - tabs are created upfront.
func (cs *ComponentScheduler) tuiMarkQueued(displayName string, weight int) {
	cs.tuiMarkPending(displayName, weight)
}

// formatDisplayName returns the display name for a work unit.
func (cs *ComponentScheduler) formatDisplayName(spec workunit.UnitSpec) string {
	if spec.ID.Tool != "" {
		return fmt.Sprintf("%s:%s:%s", spec.ID.Module, spec.ID.Component, spec.ID.Tool)
	}
	return fmt.Sprintf("%s:%s", spec.ID.Module, spec.ID.Component)
}

// executeWorker runs the actual work for a component.
// Called by dispatcher after capacity is acquired and dependencies satisfied.
// This is the core execution extracted from processComponent without dep-waiting or semaphore handling.
func (cs *ComponentScheduler) executeWorker(spec workunit.UnitSpec, worker ComponentWorkerFunc) ComponentResult {
	module := spec.ID.Module
	component := spec.ID.Component
	tool := spec.ID.Tool

	result := ComponentResult{
		Module:    module,
		Component: component,
		Handler:   tool,
	}

	displayName := cs.formatDisplayName(spec)

	// Start timing - duration measures actual execution time
	startTime := time.Now()

	// Track active tool usage
	cs.addActiveTool(tool, spec.IsContainer)

	// Create output directory for this component
	// Structure: out/build/<module>/<component> (e.g., out/build/books/howto)
	sanitizedModule := sanitizePathForFS(output.PackageDisplayName(module))
	sanitizedComponent := sanitizePathForFS(component)
	componentOutputDir := filepath.Join(cs.config.WorkspaceRoot, cs.config.OutputBaseDir, sanitizedModule, sanitizedComponent)

	// Relative log path for result reporting: out/build/<module>/<component>/build.log
	relLogPath := filepath.Join(cs.config.OutputBaseDir, sanitizedModule, sanitizedComponent, cs.config.LogFileName)

	if err := os.MkdirAll(componentOutputDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		cs.removeActiveTool(tool, spec.IsContainer)
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
		cs.removeActiveTool(tool, spec.IsContainer)
		return result
	}

	// Create writer for worker
	var workerWriter io.Writer
	if cs.tuiConsole != nil {
		workerWriter = cs.tuiConsole.NewWriter(displayName, logFile)
	} else {
		workerWriter = logFile
	}

	// Execute work with memory instrumentation
	memBefore := GetMemoryStats()
	fmt.Fprintf(logFile, "[memory] before: used=%s avail=%s total=%s (%.1f%%)\n",
		FormatBytes(memBefore.UsedBytes), FormatBytes(memBefore.AvailableBytes),
		FormatBytes(memBefore.TotalBytes), memBefore.UsedPercent)

	exitCode := worker(module, component, workerWriter)

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

	// Close TUI writer first (flushes pipe), then log file
	if closer, ok := workerWriter.(io.Closer); ok {
		closer.Close()
	}
	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = relLogPath
	result.Duration = time.Since(startTime)

	// Merge any extras set by the worker (e.g., test counts)
	if extras, ok := cs.getComponentExtras(module, component); ok {
		result.TestsTotal = extras.TestsTotal
		result.TestsPassed = extras.TestsPassed
		result.TestsFailed = extras.TestsFailed
		result.TestsSkipped = extras.TestsSkipped
	}

	// Remove tool from active list
	cs.removeActiveTool(tool, spec.IsContainer)

	return result
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
		Phase:             capitalize(cs.config.ActionVerb),
		Running:           running,
		Completed:         completed,
		Total:             total,
		Locks:             cs.getLockStatuses(),
		ActiveContainers:  cs.getActiveContainersList(),
		UsedContainers:    cs.getUsedContainersList(),
		ActiveSystemTools: cs.getActiveSystemToolsList(),
		UsedSystemTools:   cs.getUsedSystemToolsList(),
	})
}

// tuiMarkCompleted removes a component from running, increments completed, and reports exit code.
func (cs *ComponentScheduler) tuiMarkCompleted(displayName string, exitCode int) {
	if cs.tuiConsole == nil {
		return
	}

	// DEBUG: Log exit code for TUI caching visibility investigation
	// Exit codes: 0=success(green), <0=skipped(blue), >0=failed(red)
	// Note: This will only fire if something actually returns -1, which currently only mkdocs does

	// For cached modules (exitCode < 0), include when the artifact was last built
	if exitCode < 0 {
		// Extract module name from displayName (format: "module:component")
		moduleName := displayName
		if idx := len(displayName) - 1; idx > 0 {
			// displayName is typically "module:component", we need just the module
			for i := 0; i < len(displayName); i++ {
				if displayName[i] == ':' {
					moduleName = displayName[:i]
					break
				}
			}
		}
		cacheTime := cs.getCacheTime(moduleName)
		cs.tuiConsole.MarkModuleCompleteWithCacheInfo(displayName, exitCode, cacheTime, "")
	} else {
		cs.tuiConsole.MarkModuleComplete(displayName, exitCode)
	}

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
		Phase:             capitalize(cs.config.ActionVerb),
		Running:           running,
		Completed:         completed,
		Total:             total,
		Locks:             cs.getLockStatuses(),
		ActiveContainers:  cs.getActiveContainersList(),
		UsedContainers:    cs.getUsedContainersList(),
		ActiveSystemTools: cs.getActiveSystemToolsList(),
		UsedSystemTools:   cs.getUsedSystemToolsList(),
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
func AggregateToWorkResults(compResults []ComponentResult, work []workunit.UnitSpec) []WorkResult {
	// Group results by module
	moduleResults := make(map[string][]ComponentResult)
	for _, r := range compResults {
		moduleResults[r.Module] = append(moduleResults[r.Module], r)
	}

	// Get unique modules in order
	seenModules := make(map[string]bool)
	var moduleOrder []string
	for _, w := range work {
		module := w.ID.Module
		if !seenModules[module] {
			seenModules[module] = true
			moduleOrder = append(moduleOrder, module)
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
		allSkipped := true
		anyFailed := false
		for _, cr := range compResults {
			// Component failure (exit code > 0) = module failure
			// Exit code -1 means cached/skipped, not failure
			if cr.ExitCode > 0 {
				anyFailed = true
				wr.ExitCode = 1
				// Track first failed component's log path
				if firstFailedLogPath == "" {
					firstFailedLogPath = cr.LogPath
				}
			}
			// Track if all components were skipped (cached)
			if cr.ExitCode != -1 {
				allSkipped = false
			}
			// Collect all warnings and errors
			wr.Warnings = append(wr.Warnings, cr.Warnings...)
			wr.Errors = append(wr.Errors, cr.Errors...)
			// Track max duration (since components run in parallel)
			if cr.Duration > maxDuration {
				maxDuration = cr.Duration
			}
		}
		// If all components returned -1 (cached), module result is -1
		if !anyFailed && allSkipped && len(compResults) > 0 {
			wr.ExitCode = -1
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
