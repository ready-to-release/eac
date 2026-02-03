package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/adapters/tui"
	"github.com/ready-to-release/eac/go/clibase/ansi"
	"github.com/ready-to-release/eac/go/clibase/capacity"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/shirou/gopsutil/v3/mem"
)

// UnitExtras holds additional data passed from workers to unit results.
// Used for test-specific fields that need to flow from the test runner to the summary.
type UnitExtras struct {
	TestsTotal   int
	TestsPassed  int
	TestsFailed  int
	TestsSkipped int
}

// UnitScheduler manages parallel execution of work units
// with weighted resource control and dependency ordering.
type UnitScheduler struct {
	config    Config
	semaphore *capacity.GlobalSemaphore // Global cross-process semaphore
	registry  *locktracker.Registry     // Lock tracking registry for TUI visualization

	// Dynamic capacity management
	capacityTicker *time.Ticker  // Recalculates capacity every 2 seconds
	capacityStop   chan struct{} // Signal to stop the capacity ticker
	configMax      int           // Maximum capacity from config (ceiling)
	turbo          float64       // Turbo multiplier (1.0x, 1.25x, 2.0x, etc.)

	// Module completion tracking for inter-module dependencies
	moduleCompleteMu sync.RWMutex
	moduleComplete   map[string]bool          // module -> all components done
	moduleCompleteCh map[string]chan struct{} // broadcast when module completes
	moduleCompCount  map[string]int           // module -> number of components
	moduleCompDone   map[string]int           // module -> components completed

	// Unit completion tracking for intra-module dependencies (build_after)
	unitCompleteMu   sync.RWMutex
	unitComplete     map[string]map[string]bool          // module -> component -> done
	unitCompleteCh   map[string]map[string]chan struct{} // module -> component -> broadcast
	unitHandlerCount map[string]map[string]int           // module -> component -> total handlers
	unitHandlerDone  map[string]map[string]int           // module -> component -> handlers completed

	// Unit failure tracking
	unitFailedMu sync.RWMutex
	unitFailed   map[string]map[string]bool // module -> component -> failed

	// Unit extras tracking (test counts, etc.)
	unitExtrasMu sync.RWMutex
	unitExtras   map[string]map[string]UnitExtras // module -> component -> extras

	// TUI console for real-time output display (legacy - kept for lifecycle)
	tuiConsole tui.Console
	tuiCtx     interface{} // context.Context but we avoid import cycle

	// Observer pattern - emit function and writer factory
	emitFunc      func(interfaces.ExecutionEvent)
	writerFactory interfaces.WriterFactory

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int

	// Capacity tracking for three-value model
	roof int // Hard ceiling - pool size set at scheduler start

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

	// Early cache detection results for worker short-circuit
	// Background thread populates this; workers check before executing
	earlyCached sync.Map // key: moniker (string), value: EarlyCacheInfo

	// Cache detection configuration (optional, set via SetCacheDetection)
	cacheDetectionMu       sync.Mutex
	cacheDetectionVerifier CacheVerifier
	cacheDetectionModules  map[string]bool
}

// EarlyCacheInfo holds cache verification result from background detection.
// Stored in earlyCached map for worker short-circuit.
type EarlyCacheInfo struct {
	Module    string
	Component string
	Handler   string
	CacheTime time.Time
}

// NewUnitScheduler creates a new scheduler with the given configuration.
// If registry is non-nil, the semaphore will be tracked for lock visualization.
// Starts a dynamic capacity ticker that adjusts capacity based on available system resources.
// Uses a GLOBAL semaphore shared across all processes (build, test, lint, scan).
//
// emitFunc is called to emit events to observers (can be nil for legacy mode).
// writerFactory creates writers for unit output (can be nil, falls back to log files).
func NewUnitScheduler(config *Config, tuiConsole tui.Console, registry *locktracker.Registry, emitFunc func(interfaces.ExecutionEvent), writerFactory interfaces.WriterFactory) *UnitScheduler {
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

	us := &UnitScheduler{
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
		unitComplete:     make(map[string]map[string]bool),
		unitCompleteCh:   make(map[string]map[string]chan struct{}),
		unitHandlerCount: make(map[string]map[string]int),
		unitHandlerDone:  make(map[string]map[string]int),
		unitFailed:       make(map[string]map[string]bool),
		unitExtras:       make(map[string]map[string]UnitExtras),
		activeContainers: make(map[string]int),
		usedContainers:   make(map[string]bool),
		activeSystem:     make(map[string]int),
		usedSystem:       make(map[string]bool),
		tuiConsole:       tuiConsole,
		emitFunc:         emitFunc,
		writerFactory:    writerFactory,
	}

	// Configure output writer
	if config.TUI {
		us.orchestratorOut = io.Discard
	} else {
		us.orchestratorOut = os.Stdout
	}

	// Start dynamic capacity ticker
	us.startCapacityTicker()

	return us
}

// startCapacityTicker starts a goroutine that recalculates capacity every 2 seconds
// based on available system resources. The config max is used as a ceiling.
func (us *UnitScheduler) startCapacityTicker() {
	us.capacityTicker = time.NewTicker(config.CapacityRecalcInterval())

	go func() {
		for {
			select {
			case <-us.capacityTicker.C:
				newCap := detectAvailableCapacity(us.configMax, us.turbo)
				us.semaphore.SetCapacity(newCap)
			case <-us.capacityStop:
				us.capacityTicker.Stop()
				return
			}
		}
	}()
}

// StopCapacityTicker stops the dynamic capacity ticker.
// Should be called when the scheduler is no longer needed.
func (us *UnitScheduler) StopCapacityTicker() {
	if us.capacityStop != nil {
		close(us.capacityStop)
		us.capacityStop = nil
	}
}

// Close releases resources held by the scheduler.
// Should be called when the scheduler is no longer needed.
func (us *UnitScheduler) Close() {
	us.StopCapacityTicker()
	if us.semaphore != nil {
		us.semaphore.Close()
	}
}

// emit sends an event to observers via the emit callback.
func (us *UnitScheduler) emit(event interfaces.ExecutionEvent) {
	if us.emitFunc != nil {
		us.emitFunc(event)
	}
}

// emitResourceStatus sends the current lock/resource status to observers.
// This should be called after semaphore acquire/release to update the TUI's Resources pane.
func (us *UnitScheduler) emitResourceStatus() {
	if us.emitFunc == nil || us.registry == nil {
		return
	}

	snapshot := us.registry.Snapshot()
	resources := make([]interfaces.ResourceInfo, 0, len(snapshot))
	for _, lock := range snapshot {
		resources = append(resources, interfaces.ResourceInfo{
			Name:     lock.Name,
			Type:     string(lock.Type),
			Capacity: int(lock.Capacity),
			Used:     int(lock.Used),
			Waiting:  int(lock.Waiting),
		})
	}

	us.emit(interfaces.ResourceStatusEvent{
		Time:      time.Now(),
		Resources: resources,
	})
}

// SetCacheTimes sets the cache times for modules that are up-to-date.
// These times are displayed in the TUI for cached modules to show when
// the cached artifacts were originally built.
func (us *UnitScheduler) SetCacheTimes(times map[string]time.Time) {
	us.cacheTimesMu.Lock()
	defer us.cacheTimesMu.Unlock()
	us.cacheTimes = times
}

// getCacheTime returns the cache time for a module, or zero time if not cached.
func (us *UnitScheduler) getCacheTime(module string) time.Time {
	us.cacheTimesMu.RLock()
	defer us.cacheTimesMu.RUnlock()
	if us.cacheTimes == nil {
		return time.Time{}
	}
	return us.cacheTimes[module]
}

// SetSummaryBuilder sets the summary builder for incremental summary computation.
// The builder receives component results as they complete.
func (us *UnitScheduler) SetSummaryBuilder(builder SummaryBuilder) {
	us.summaryBuilder = builder
}

// CacheVerifier is an alias for execution.CacheVerifier.
// Kept for backward compatibility with existing code.
// New code should use execution.CacheVerifier directly.
type CacheVerifier = execution.CacheVerifier

// SetCacheDetection configures background cache detection for RunUnits.
// When set, RunUnits will start background detection after creating tabs,
// causing cached tabs to progressively "light up" blue.
//
// The verifier function checks cache status for a component.
// cachedModules is the pre-computed set of modules known to be cached.
// cacheTimes comes from SetCacheTimes() - already set on the scheduler.
//
// Call this before RunUnits to enable early cache detection.
func (us *UnitScheduler) SetCacheDetection(verifier CacheVerifier, cachedModules map[string]bool) {
	us.cacheDetectionMu.Lock()
	defer us.cacheDetectionMu.Unlock()
	us.cacheDetectionVerifier = verifier
	us.cacheDetectionModules = cachedModules
}

// StartBackgroundCacheDetection spawns goroutines to verify cache status in parallel.
// Uses I/O-appropriate parallelism (NOT the weighted semaphore - this is lightweight).
// Updates TUI and marks items as early-cached for worker short-circuit.
//
// The verifier interface checks cache status for a component and returns
// whether it's cached and when. This allows different commands (build, test, etc.)
// to provide their own cache verification logic via the execution.CacheVerifier interface.
//
// Background detection does NOT report to summaryBuilder - workers are the sole
// source of truth for summarization. This prevents duplicate results.
func (us *UnitScheduler) StartBackgroundCacheDetection(
	work []workunit.UnitSpec,
	cachedModules map[string]bool,
	cacheTimes map[string]time.Time,
	verifier CacheVerifier,
) {
	if verifier == nil || len(work) == 0 {
		return
	}

	go func() {
		// Panic recovery - don't crash if detection fails
		// Workers will still check cache normally
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "[cache-detect] panic recovered: %v\n", r)
			}
		}()

		// I/O-appropriate parallelism (NOT weighted semaphore)
		// Cache checks are file reads + hash computation, not heavy builds
		maxParallel := 32
		if len(work) < maxParallel {
			maxParallel = len(work)
		}
		sem := make(chan struct{}, maxParallel)

		var wg sync.WaitGroup

		// Create a context for cache verification
		// Background context is appropriate since detection runs to completion
		ctx := context.Background()

		for _, spec := range work {
			wg.Add(1)
			go func(s workunit.UnitSpec) {
				defer wg.Done()

				// Acquire I/O slot
				sem <- struct{}{}
				defer func() { <-sem }()

				// Use provided verification logic via interface
				result, err := verifier.Verify(ctx, s)
				if err != nil {
					// Log error but continue - workers will still check cache normally
					logging.C().Debugf("[cache-detect] verification error for %s: %v", s.ID.Longname(), err)
					return
				}

				if result.Cached {
					moniker := s.ID.Longname()

					// 1. Emit event to observers - tab lights up blue
					us.emit(interfaces.UnitCompletedEvent{
						Time:     time.Now(),
						ID:       moniker,
						ExitCode: -1, // Exit code -1 = cached
					})

					// 2. Mark as early-cached so worker can short-circuit
					us.earlyCached.Store(moniker, EarlyCacheInfo{
						Module:    s.ID.Module,
						Component: s.ID.Component,
						Handler:   s.ID.Tool,
						CacheTime: result.CacheTime,
					})

					// NOTE: Do NOT increment tuiCompleted here!
					// Workers are the sole source of truth for completion counting.
					// This prevents a race where background marks item cached after
					// worker started, causing worker to skip its counter increment.
				}

				// NOTE: Do NOT report to summaryBuilder here!
				// Workers are the sole source of truth for summarization.
				// This prevents duplicate results.
			}(spec)
		}

		wg.Wait()
	}()
}

// detectAvailableCapacity calculates the pressure roof for parallel builds.
// If --roof is set, uses that value directly. Otherwise auto-detects from system resources.
func detectAvailableCapacity(configMax int, turbo float64) int {
	// Use Docker/WSL-aware CPU detection
	cpuCount := GetEffectiveCPUs()
	if cpuCount < 1 {
		cpuCount = 4 // Fallback
	}

	// Prioritize Docker → WSL → Host memory detection
	var ramGB int
	effectiveMem := GetEffectiveMemoryBytes()
	if effectiveMem > 0 {
		ramGB = int(effectiveMem / (1024 * 1024 * 1024))
	} else {
		// Fallback: use host available RAM (not total) for safety
		memInfo, err := mem.VirtualMemory()
		if err == nil {
			// Use AVAILABLE, not TOTAL (critical on Windows/WSL)
			ramGB = int(memInfo.Available / (1024 * 1024 * 1024))
		} else {
			ramGB = 8 // Hard fallback
		}
	}

	capacity := calculateCapacity(cpuCount, ramGB, configMax, turbo)

	logging.C().Debugf("[scheduler] Capacity calculation: cpuCount=%d ramGB=%d configMax=%d turbo=%.2f → capacity=%d",
		cpuCount, ramGB, configMax, turbo, capacity)

	return capacity
}

// calculateCapacity computes the effective parallelism capacity.
//
// If roof > 0: use roof directly (--roof flag overrides everything)
// If roof == 0: auto-detect from system resources
func calculateCapacity(cpuCount, ramGB, roof int, turbo float64) int {
	// --roof overrides all
	if roof > 0 {
		logging.C().Debugf("[scheduler] Using explicit roof: %d", roof)
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
		logging.C().Debugf("[scheduler] RAM-limited: ramCap=%d < cpuCount=%d, using ramCap", ramCap, cpuCount)
		base = ramCap
	} else {
		logging.C().Debugf("[scheduler] CPU-limited: ramCap=%d >= cpuCount=%d, using cpuCount", ramCap, cpuCount)
	}

	capacity := int(float64(base) * turbo)
	logging.C().Debugf("[scheduler] Base capacity=%d x turbo=%.2f = %d", base, turbo, capacity)

	// Cap at RAM limit FIRST (safety), then CPU limit
	// This prevents turbo from overcommitting memory
	ramMax := ramGB / 3
	if capacity > ramMax && ramMax > 0 {
		logging.C().Debugf("[scheduler] Turbo exceeded RAM limit: %d -> %d (ramGB=%d)", capacity, ramMax, ramGB)
		capacity = ramMax
	}

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
// Must be called before RunUnits.
func (us *UnitScheduler) InitializeWork(work []workunit.UnitSpec) {
	// Reset state
	us.moduleComplete = make(map[string]bool)
	us.moduleCompleteCh = make(map[string]chan struct{})
	us.moduleCompCount = make(map[string]int)
	us.moduleCompDone = make(map[string]int)
	us.unitComplete = make(map[string]map[string]bool)
	us.unitCompleteCh = make(map[string]map[string]chan struct{})
	us.unitHandlerCount = make(map[string]map[string]int)
	us.unitHandlerDone = make(map[string]map[string]int)
	us.unitFailed = make(map[string]map[string]bool)
	us.unitExtras = make(map[string]map[string]UnitExtras)
	us.activeContainers = make(map[string]int)
	us.usedContainers = make(map[string]bool)
	us.activeSystem = make(map[string]int)
	us.usedSystem = make(map[string]bool)

	// Count components per module and initialize tracking maps
	for _, w := range work {
		module := w.ID.Module
		component := w.ID.Component

		us.moduleCompCount[module]++

		// Initialize module completion channel if needed
		if _, ok := us.moduleCompleteCh[module]; !ok {
			us.moduleCompleteCh[module] = make(chan struct{})
		}

		// Initialize component tracking for this module if needed
		if us.unitComplete[module] == nil {
			us.unitComplete[module] = make(map[string]bool)
			us.unitCompleteCh[module] = make(map[string]chan struct{})
			us.unitHandlerCount[module] = make(map[string]int)
			us.unitHandlerDone[module] = make(map[string]int)
			us.unitFailed[module] = make(map[string]bool)
		}

		// Count handlers per component and create channel only once per component
		us.unitHandlerCount[module][component]++
		if _, exists := us.unitCompleteCh[module][component]; !exists {
			us.unitCompleteCh[module][component] = make(chan struct{})
		}
	}

	// Initialize TUI status
	us.tuiTotal = len(work)
	us.tuiCompleted = 0
	us.tuiRunning = nil
}

// RunUnits executes work units with worker pool scheduling.
// Uses LPT (Longest Processing Time First) - heaviest jobs scheduled first.
// Spawns a pool of worker goroutines that pull from the queue concurrently.
// Returns results in the same order as the input work items.
func (us *UnitScheduler) RunUnits(work []workunit.UnitSpec, worker UnitWorkerFunc) []UnitResult {
	results := make([]UnitResult, len(work))
	var resultsMu sync.Mutex
	var wg sync.WaitGroup

	// Determine layer mode from config:
	// - Layered=true → LayerModeStrict (module + component + DependsOn enforced)
	// - Layered=false → LayerModeNone (component + DependsOn only)
	mode := execution.LayerModeNone
	if us.config.Layered {
		mode = execution.LayerModeStrict
	}

	// Create work queue with LayerPolicy for scheduling
	queue, err := NewWorkQueueWithPolicy(work, mode)
	if err != nil {
		// Circular dependencies are a configuration error.
		// Return early with error results for all work items.
		logging.C().Errorf("[scheduler] Failed to create work queue: %v", err)
		for i, w := range work {
			results[i] = UnitResult{
				Module:    w.ID.Module,
				Component: w.ID.Component,
				Handler:   w.ID.Tool,
				ExitCode:  1,
				Errors:    []string{fmt.Sprintf("circular dependency: %v", err)},
			}
		}
		return results
	}
	us.workQueue = queue // Store for stats access

	// Create all tabs as QUEUED upfront (positions are immutable)
	for _, w := range work {
		moniker := w.ID.Longname()
		displayName := w.Shortname()
		us.tuiMarkQueued(moniker, displayName, w.Weight)
	}

	// Emit initial resource status to show the Resources pane in TUI
	us.emitResourceStatus()

	// Start background cache detection if configured
	// Timing is critical: tabs must exist first so they can "light up"
	us.cacheDetectionMu.Lock()
	verifier := us.cacheDetectionVerifier
	cachedModules := us.cacheDetectionModules
	us.cacheDetectionMu.Unlock()

	if verifier != nil && cachedModules != nil {
		us.cacheTimesMu.RLock()
		cacheTimes := us.cacheTimes
		us.cacheTimesMu.RUnlock()
		us.StartBackgroundCacheDetection(work, cachedModules, cacheTimes, verifier)
	}

	// Determine worker pool size: min(work items, capacity)
	// This ensures we can saturate capacity immediately
	poolSize := len(work)
	capacity := us.semaphore.Capacity()
	if capacity > 0 && poolSize > capacity {
		poolSize = capacity
	}

	// Store roof (hard ceiling) for three-value capacity model
	// Roof is the actual peak allocation - workers spawned at start
	us.roof = poolSize

	// Store total capacity for over-capacity detection
	totalCapacity := capacity

	// Spawn worker pool - all workers start immediately and compete for queue items
	// Workers use bin-packing: check available capacity, pop item that fits
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				// Check available capacity for bin-packing
				// This is approximate (race-safe) - actual acquisition may differ
				available := us.semaphore.Available()
				if available < 1 {
					available = 1 // At minimum, look for weight-1 items
				}

				// Pop item that fits available capacity (bin-packing)
				// Prioritizes heavier items that fit, enabling parallel small jobs
				spec := queue.PopReadyWithBudget(available, totalCapacity)
				if spec == nil {
					return // queue exhausted
				}

				weight := spec.Weight
				if weight <= 0 {
					weight = 1
				}

				// Check if this is an over-capacity item (weight > total capacity)
				// Over-capacity items skip semaphore and execute serially via queue lock
				isOverCapacity := totalCapacity > 0 && weight > totalCapacity

				// Acquire capacity (blocks until slot available)
				// CRITICAL: Skip semaphore for over-capacity items to prevent forever-blocking
				if !isOverCapacity {
					us.semaphore.Acquire(weight)
				}
				us.emitResourceStatus() // Update TUI Resources pane

				// Update TUI: queued -> running (uses moniker for matching)
				us.tuiMarkRunning(spec.ID.Longname())

				var result UnitResult

				// Check if any dependency failed - if so, skip execution and fail immediately
				if queue.HasFailedDependency(spec.ID) {
					result = UnitResult{
						Module:    spec.ID.Module,
						Component: spec.ID.Component,
						Handler:   spec.ID.Tool,
						ExitCode:  1,
						Errors:    []string{"Skipped: dependency failed"},
					}
					// Release capacity immediately since we're not executing
					if !isOverCapacity {
						us.semaphore.Release(weight)
					}
					us.emitResourceStatus()
				} else {
					// Execute work
					result = us.executeWorker(*spec, worker)

					// Release capacity
					if !isOverCapacity {
						us.semaphore.Release(weight)
					}
					us.emitResourceStatus() // Update TUI Resources pane
				}

				// Store result by original index
				resultsMu.Lock()
				results[spec.Index] = result
				resultsMu.Unlock()

				// Notify queue: mark as completed or failed based on result
				if result.ExitCode > 0 {
					queue.MarkFailed(spec.ID)
				} else {
					queue.MarkComplete(spec.ID)
				}

				// Release over-capacity lock if this was an over-capacity item
				if isOverCapacity {
					queue.ReleaseOverCapacity(spec.ID)
				}

				// Mark component complete (broadcasts to legacy channels, updates TUI)
				us.markUnitComplete(*spec, &result)
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()

	// Defensive check: verify queue is fully drained
	// If items remain, workers exited prematurely (indicates a bug)
	if remaining := queue.Len(); remaining > 0 {
		logging.C().Warnf("[scheduler] BUG: queue has %d items remaining after all workers exited", remaining)
	}

	// Close queue
	queue.Close()

	return results
}

// markUnitComplete marks a component as done and broadcasts to waiters.
func (us *UnitScheduler) markUnitComplete(spec workunit.UnitSpec, result *UnitResult) {
	module := spec.ID.Module
	component := spec.ID.Component

	// Track handler completion and only close channel when all handlers are done
	us.unitCompleteMu.Lock()
	if result.ExitCode != 0 {
		us.unitFailed[module][component] = true
	}
	us.unitHandlerDone[module][component]++
	handlersDone := us.unitHandlerDone[module][component]
	handlersTotal := us.unitHandlerCount[module][component]
	allHandlersDone := handlersDone >= handlersTotal
	var ch chan struct{}
	if allHandlersDone {
		us.unitComplete[module][component] = true
		ch = us.unitCompleteCh[module][component]
	}
	us.unitCompleteMu.Unlock()

	// Broadcast completion only when all handlers for this component are done
	if allHandlersDone {
		close(ch)
	}

	// Update module completion tracking
	us.moduleCompleteMu.Lock()
	us.moduleCompDone[module]++
	if us.moduleCompDone[module] >= us.moduleCompCount[module] {
		us.moduleComplete[module] = true
		close(us.moduleCompleteCh[module])
	}
	us.moduleCompleteMu.Unlock()

	// Send result to summary builder for incremental summary computation
	if us.summaryBuilder != nil {
		us.summaryBuilder.AddResult(*result)
	}

	// Update TUI with exit code
	// Use Longname() to match the ID format used in SetInitSummary
	moniker := spec.ID.Longname()
	us.tuiMarkCompleted(moniker, result.ExitCode)
}

// WaitForModule blocks until all components of a module are complete.
// Used for inter-module dependencies.
func (us *UnitScheduler) WaitForModule(module string) {
	us.moduleCompleteMu.RLock()
	ch, exists := us.moduleCompleteCh[module]
	us.moduleCompleteMu.RUnlock()

	if exists {
		<-ch
	}
}

// IsModuleFailed returns true if any component in the module failed.
func (us *UnitScheduler) IsModuleFailed(module string) bool {
	us.unitFailedMu.RLock()
	defer us.unitFailedMu.RUnlock()

	if failures, ok := us.unitFailed[module]; ok {
		for _, failed := range failures {
			if failed {
				return true
			}
		}
	}
	return false
}

// SetUnitExtras stores additional data for a component result.
// This is called by workers to pass test counts or other data that will be
// merged into the UnitResult when processing completes.
func (us *UnitScheduler) SetUnitExtras(module, component string, extras UnitExtras) {
	us.unitExtrasMu.Lock()
	defer us.unitExtrasMu.Unlock()

	if us.unitExtras[module] == nil {
		us.unitExtras[module] = make(map[string]UnitExtras)
	}
	us.unitExtras[module][component] = extras
}

// getUnitExtras retrieves the extras for a component, if any.
func (us *UnitScheduler) getUnitExtras(module, component string) (UnitExtras, bool) {
	us.unitExtrasMu.RLock()
	defer us.unitExtrasMu.RUnlock()

	if moduleExtras, ok := us.unitExtras[module]; ok {
		if extras, ok := moduleExtras[component]; ok {
			return extras, true
		}
	}
	return UnitExtras{}, false
}

// addActiveTool increments the usage count for a tool/handler.
// Also tracks docker as active when container tools are used.
func (us *UnitScheduler) addActiveTool(handler string, isContainer bool) {
	if handler == "" {
		return
	}

	us.toolsMu.Lock()
	if isContainer {
		us.activeContainers[handler]++
		us.usedContainers[handler] = true
		// Container tools require docker - track it as active system tool
		us.activeSystem["docker"]++
		us.usedSystem["docker"] = true
	} else {
		us.activeSystem[handler]++
		us.usedSystem[handler] = true
	}
	us.toolsMu.Unlock()
}

// removeActiveTool decrements the usage count for a tool/handler.
// Also decrements docker count when container tools finish.
func (us *UnitScheduler) removeActiveTool(handler string, isContainer bool) {
	if handler == "" {
		return
	}

	us.toolsMu.Lock()
	if isContainer {
		if us.activeContainers[handler] > 0 {
			us.activeContainers[handler]--
			if us.activeContainers[handler] == 0 {
				delete(us.activeContainers, handler)
			}
		}
		// Decrement docker usage count
		if us.activeSystem["docker"] > 0 {
			us.activeSystem["docker"]--
			if us.activeSystem["docker"] == 0 {
				delete(us.activeSystem, "docker")
			}
		}
	} else {
		if us.activeSystem[handler] > 0 {
			us.activeSystem[handler]--
			if us.activeSystem[handler] == 0 {
				delete(us.activeSystem, handler)
			}
		}
	}
	us.toolsMu.Unlock()
}

// getActiveContainersList returns a sorted list of currently active container tools.
func (us *UnitScheduler) getActiveContainersList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.activeContainers) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.activeContainers))
	for tool := range us.activeContainers {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedContainersList returns a sorted list of all container tools that have been used.
func (us *UnitScheduler) getUsedContainersList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.usedContainers) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.usedContainers))
	for tool := range us.usedContainers {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getActiveSystemToolsList returns a sorted list of currently active system tools.
func (us *UnitScheduler) getActiveSystemToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.activeSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.activeSystem))
	for tool := range us.activeSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// getUsedSystemToolsList returns a sorted list of all system tools that have been used.
func (us *UnitScheduler) getUsedSystemToolsList() []string {
	us.toolsMu.Lock()
	defer us.toolsMu.Unlock()

	if len(us.usedSystem) == 0 {
		return nil
	}

	tools := make([]string, 0, len(us.usedSystem))
	for tool := range us.usedSystem {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// tuiMarkPending creates a component tab in pending state (scheduled, waiting for slot).
// moniker is the globally unique ID (Longname) for matching; displayName is for tab labels.
// Emits UnitQueuedEvent to observers.
func (us *UnitScheduler) tuiMarkPending(moniker, displayName string, weight int) {
	// Emit event to observers
	us.emit(interfaces.UnitQueuedEvent{
		Time:        time.Now(),
		ID:          moniker,
		DisplayName: displayName,
		Weight:      weight,
	})
}

// tuiMarkQueued creates a component tab in queued state (waiting in work queue).
// This is the same visual as pending - tabs are created upfront.
func (us *UnitScheduler) tuiMarkQueued(moniker, displayName string, weight int) {
	us.tuiMarkPending(moniker, displayName, weight)
}

// executeWorker runs the actual work for a component.
// Called by dispatcher after capacity is acquired and dependencies satisfied.
// This is the core execution extracted from processComponent without dep-waiting or semaphore handling.
func (us *UnitScheduler) executeWorker(spec workunit.UnitSpec, worker UnitWorkerFunc) UnitResult {
	module := spec.ID.Module
	component := spec.ID.Component
	tool := spec.ID.Tool

	result := UnitResult{
		Module:    module,
		Component: component,
		Handler:   tool,
	}

	moniker := spec.ID.Longname()
	displayName := spec.Shortname()

	// FAST PATH: Check if background already verified cached
	// This enables fast termination - workers don't re-do cache checks
	// TUI was already updated by background detection - just return result for summary
	if cacheInfo, ok := us.earlyCached.Load(moniker); ok {
		info := cacheInfo.(EarlyCacheInfo)
		return UnitResult{
			Module:    info.Module,
			Component: info.Component,
			Handler:   info.Handler,
			ExitCode:  -1, // Cached
			Duration:  0,  // Instant - no actual work done
		}
	}

	// Start timing - duration measures actual execution time
	startTime := time.Now()

	// Track active tool usage
	us.addActiveTool(tool, spec.Container)

	// Create output directory for this component
	// Structure: out/build/<module>/<component> (e.g., out/build/books/howto)
	sanitizedModule := sanitizePathForFS(output.PackageDisplayName(module))
	sanitizedComponent := sanitizePathForFS(component)
	componentOutputDir := filepath.Join(us.config.WorkspaceRoot, us.config.OutputBaseDir, sanitizedModule, sanitizedComponent)

	// Relative log path for result reporting: out/build/<module>/<component>/build.log
	relLogPath := filepath.Join(us.config.OutputBaseDir, sanitizedModule, sanitizedComponent, us.config.LogFileName)

	if err := os.MkdirAll(componentOutputDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		us.removeActiveTool(tool, spec.Container)
		return result
	}

	// Create log file
	logPath := filepath.Join(componentOutputDir, us.config.LogFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create log file: %v", err)}
		result.LogPath = relLogPath
		result.Duration = time.Since(startTime)
		us.removeActiveTool(tool, spec.Container)
		return result
	}

	// Wrap log file with bad ANSI filter - preserves colors, strips control sequences
	filteredLogFile := ansi.NewBadOnlyFilter(logFile, displayName)

	// Create writer for worker - use writerFactory if available (e.g., TUIObserver)
	// Use moniker (longname) to match TUI tab identification
	var workerWriter io.Writer
	if us.writerFactory != nil {
		workerWriter = us.writerFactory.NewWriter(moniker, filteredLogFile)
	} else {
		workerWriter = filteredLogFile
	}

	// Execute work with memory instrumentation
	memBefore := GetMemoryStats()
	fmt.Fprintf(logFile, "[memory] before: used=%s avail=%s total=%s (%.1f%%)\n",
		FormatBytes(memBefore.UsedBytes), FormatBytes(memBefore.AvailableBytes),
		FormatBytes(memBefore.TotalBytes), memBefore.UsedPercent)

	// Get worker timeout from config (default 5m if not set)
	timeout := config.WorkerTimeout()
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Run worker with timeout enforcement
	var exitCode int
	resultCh := make(chan int, 1)
	go func() {
		// Combine component and tool for worker (e.g., "go:golangci-lint")
		// When tool is empty (tests), workerComponent equals component
		workerComponent := component
		if tool != "" {
			workerComponent = component + ":" + tool
		}
		resultCh <- worker(module, workerComponent, workerWriter)
	}()

	select {
	case exitCode = <-resultCh:
		// Worker completed normally
	case <-time.After(timeout):
		// Worker timed out - kill it
		exitCode = workunit.ExitCodeTimeout
		fmt.Fprintf(os.Stderr, "TIMEOUT: %s killed after %v (limit: %v)\n", displayName, time.Since(startTime), timeout)
		fmt.Fprintf(logFile, "\n[TIMEOUT] Worker killed after %v (limit: %v)\n", time.Since(startTime), timeout)
	}

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
	if extras, ok := us.getUnitExtras(module, component); ok {
		result.TestsTotal = extras.TestsTotal
		result.TestsPassed = extras.TestsPassed
		result.TestsFailed = extras.TestsFailed
		result.TestsSkipped = extras.TestsSkipped
	}

	// Remove tool from active list
	us.removeActiveTool(tool, spec.Container)

	return result
}

// tuiMarkRunning marks a component as actively running (slot acquired).
// If the item was already marked complete by background cache detection, this is a no-op.
// moniker is the globally unique ID (Longname) for matching.
// Emits UnitStartedEvent and ProgressUpdateEvent to observers.
func (us *UnitScheduler) tuiMarkRunning(moniker string) {
	// Skip if already marked complete by background cache detection
	// This prevents blue tabs from flashing orange
	if _, wasEarlyCached := us.earlyCached.Load(moniker); wasEarlyCached {
		return
	}

	us.tuiMu.Lock()
	us.tuiRunning = append(us.tuiRunning, moniker)
	running := make([]string, len(us.tuiRunning))
	copy(running, us.tuiRunning)
	completed := us.tuiCompleted
	total := us.tuiTotal
	us.tuiMu.Unlock()

	// Emit events to observers
	us.emit(interfaces.UnitStartedEvent{
		Time: time.Now(),
		ID:   moniker,
	})
	pressureTarget := 0
	if us.semaphore != nil {
		pressureTarget = us.semaphore.Capacity()
	}
	us.emit(interfaces.ProgressUpdateEvent{
		Time:           time.Now(),
		Running:        running,
		Completed:      completed,
		Total:          total,
		Roof:           us.roof,
		PressureTarget: pressureTarget,
	})
}

// tuiMarkCompleted removes a component from running, increments completed, and reports exit code.
// Counter updates happen regardless of observer presence - workers are the sole source of truth.
// moniker is the globally unique ID (Longname) for matching.
// Emits UnitCompletedEvent and ProgressUpdateEvent to observers.
func (us *UnitScheduler) tuiMarkCompleted(moniker string, exitCode int) {
	// Update counters under lock (always)
	// Workers are the sole source of truth for completion counting.
	// Background cache detection marks items visually but does NOT count.
	us.tuiMu.Lock()
	// Remove from running list
	for i, m := range us.tuiRunning {
		if m == moniker {
			us.tuiRunning = append(us.tuiRunning[:i], us.tuiRunning[i+1:]...)
			break
		}
	}
	// Always increment counter - background detection no longer counts
	us.tuiCompleted++
	running := make([]string, len(us.tuiRunning))
	copy(running, us.tuiRunning)
	completed := us.tuiCompleted
	total := us.tuiTotal
	us.tuiMu.Unlock()

	// Emit events to observers
	us.emit(interfaces.UnitCompletedEvent{
		Time:     time.Now(),
		ID:       moniker,
		ExitCode: exitCode,
	})
	pressureTarget := 0
	if us.semaphore != nil {
		pressureTarget = us.semaphore.Capacity()
	}
	us.emit(interfaces.ProgressUpdateEvent{
		Time:           time.Now(),
		Running:        running,
		Completed:      completed,
		Total:          total,
		Roof:           us.roof,
		PressureTarget: pressureTarget,
	})
}

// AggregateToWorkResults converts component results to module-level WorkResults.
// This maintains compatibility with existing code that expects WorkResult.
func AggregateToWorkResults(compResults []UnitResult, work []workunit.UnitSpec) []WorkResult {
	// Group results by module
	moduleResults := make(map[string][]UnitResult)
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

// AggregateToModuleResultSets groups component results by module.
// Results are sorted alphabetically by module name, and components within each module
// are sorted alphabetically by component name.
// Status and Duration are computed for each module.
func AggregateToModuleResultSets(results []UnitResult) []ModuleResultSet {
	if len(results) == 0 {
		return []ModuleResultSet{}
	}

	// Group results by module
	moduleResults := make(map[string][]UnitResult)
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
	resultSets := make([]ModuleResultSet, 0, len(modules))
	for _, module := range modules {
		components := moduleResults[module]

		// Sort units alphabetically by component name
		sortedUnits := make([]UnitResult, len(components))
		copy(sortedUnits, components)
		sort.Slice(sortedUnits, func(i, j int) bool {
			return sortedUnits[i].Component < sortedUnits[j].Component
		})

		// Compute max duration
		var maxDuration time.Duration
		for _, c := range sortedUnits {
			if c.Duration > maxDuration {
				maxDuration = c.Duration
			}
		}

		// Create result set and derive status
		rs := ModuleResultSet{
			Module:   module,
			Units:    sortedUnits,
			Duration: maxDuration,
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
