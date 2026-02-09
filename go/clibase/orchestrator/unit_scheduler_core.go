package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/capacity"
	"github.com/ready-to-release/eac/go/core/resource"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/scheduling"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// Ensure execution package is used (for CacheVerifier alias)
var _ execution.CacheVerifier

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
	semaphore *capacity.DualPoolSemaphore // Dual-pool cross-process semaphore (host + docker)
	registry  *locktracker.Registry     // Lock tracking registry for TUI visualization

	// Dynamic capacity management
	capacityTicker *time.Ticker  // Recalculates capacity every 2 seconds
	capacityStop   chan struct{} // Signal to stop the capacity ticker
	capacityOnce   sync.Once    // Ensures capacityStop is closed exactly once
	configMax      int           // Maximum capacity from config (ceiling)
	turbo          float64       // Turbo multiplier (1.0x, 1.25x, 2.0x, etc.)

	// Unit extras tracking (test counts, etc.)
	unitExtrasMu sync.RWMutex
	unitExtras   map[string]map[string]UnitExtras // module -> component -> extras

	// TUI console for real-time output display (legacy - kept for lifecycle)
	tuiConsole display.Console
	tuiCtx     interface{} // context.Context but we avoid import cycle

	// Observer pattern - emit function and writer factory
	emitFunc      func(core.ExecutionEvent)
	writerFactory core.WriterFactory

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int

	// Capacity tracking for three-value model
	roof int // Hard ceiling - pool size set at scheduler start

	// Tools tracking (protected by toolsMu) - separated by type
	toolsMu              sync.Mutex
	activeContainerTools map[string]int  // container tool -> count of running components
	usedContainerTools   map[string]bool // all container tools ever used
	activeSystem         map[string]int  // system tool -> count of running components
	usedSystem           map[string]bool // all system tools ever used

	// Container instance tracking (protected by toolsMu)
	// For "Containers" lamps: each container gets its own lamp position
	// When started: new lamp lights up at next position
	// When stopped: that lamp goes dim but stays visible
	containerLamps    []bool         // true = running (lit), false = completed (dim)
	containerLampMap  map[string]int // moniker -> lamp index (to turn off correct lamp)

	// System tool instance tracking (protected by toolsMu)
	// For "Native" lamps: each native invocation gets its own lamp position
	systemLamps []bool // true = running (lit), false = completed (dim)

	// Output infrastructure
	orchestratorOut io.Writer

	// Cache times for displaying when cached artifacts were built
	cacheTimesMu sync.RWMutex
	cacheTimes   map[string]time.Time // module -> time when artifact was last built

	// Summary builder for incremental summary computation
	summaryBuilder SummaryBuilder

	// initSummaryEmitted tracks whether InitSummary was already sent to observers.
	// When true, tuiMarkPending skips emitting UnitQueuedEvent since the TUI
	// already pre-registered all tabs from InitSummaryMsg.
	initSummaryEmitted bool

	// Scheduler for dependency-aware LPT scheduling
	scheduler scheduling.WorkScheduler

	// Over-capacity coordination: serialize items with weight > totalCapacity
	// These items can't acquire semaphore (would block forever), so they bypass it
	// but must execute one at a time to prevent resource exhaustion
	overCapacityMu   sync.Mutex
	overCapacityItem *workunit.UnitID

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
func NewUnitScheduler(config *Config, tuiConsole display.Console, registry *locktracker.Registry, emitFunc func(core.ExecutionEvent), writerFactory core.WriterFactory) *UnitScheduler {
	// Calculate initial capacity based on available resources
	// Turbo multiplies the pressure roof: 1.0=normal, 1.25=+25%, 2.0=2x
	// If turbo flag is set without a value, default to 1.25x
	turbo := config.Turbo
	if turbo < 1.0 {
		turbo = 1.0
	}
	initialCap := detectAvailableCapacity(config.MaxConcurrency, turbo)
	dockerCap := detectDockerCapacity(turbo)

	// Create DUAL-POOL semaphore - shared across all processes via filesystem
	// Host pool: for all work units (host-native and containerized)
	// Docker pool: additional constraint for containerized work
	sem := capacity.NewDualPoolSemaphore(config.WorkspaceRoot, initialCap, dockerCap, registry)

	us := &UnitScheduler{
		config:           *config,
		semaphore:        sem,
		registry:         registry,
		configMax:        config.MaxConcurrency,
		turbo:            turbo,
		capacityStop:     make(chan struct{}),
		unitExtras:       make(map[string]map[string]UnitExtras),
		activeContainerTools: make(map[string]int),
		usedContainerTools:   make(map[string]bool),
		activeSystem:         make(map[string]int),
		usedSystem:           make(map[string]bool),
		containerLamps:       make([]bool, 0),
		containerLampMap:     make(map[string]int),
		systemLamps:          make([]bool, 0),
		tuiConsole:           tuiConsole,
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

// Close releases resources held by the scheduler.
// Should be called when the scheduler is no longer needed.
func (us *UnitScheduler) Close() {
	us.StopCapacityTicker()
	if us.semaphore != nil {
		us.semaphore.Close()
	}
}

// emit sends an event to observers via the emit callback.
func (us *UnitScheduler) emit(event core.ExecutionEvent) {
	if us.emitFunc != nil {
		us.emitFunc(event)
	}
}

// SetSummaryBuilder sets the summary builder for incremental summary computation.
// The builder receives component results as they complete.
func (us *UnitScheduler) SetSummaryBuilder(builder SummaryBuilder) {
	us.summaryBuilder = builder
}

// SetInitSummaryEmitted marks that InitSummary was already sent to observers.
// When set, RunUnits skips emitting UnitQueuedEvent since the TUI already
// pre-registered all tabs from InitSummaryMsg.
func (us *UnitScheduler) SetInitSummaryEmitted() {
	us.initSummaryEmitted = true
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

// InitializeWork prepares the scheduler for a batch of work units.
// Must be called before RunUnits.
func (us *UnitScheduler) InitializeWork(work []workunit.UnitSpec) {
	// Reset tool tracking
	us.unitExtras = make(map[string]map[string]UnitExtras)
	us.activeContainerTools = make(map[string]int)
	us.usedContainerTools = make(map[string]bool)
	us.activeSystem = make(map[string]int)
	us.usedSystem = make(map[string]bool)
	us.containerLamps = make([]bool, 0)
	us.containerLampMap = make(map[string]int)
	us.systemLamps = make([]bool, 0)

	// Initialize TUI status
	us.tuiTotal = len(work)
	us.tuiCompleted = 0
	us.tuiRunning = nil
}

// RunUnits executes work units with worker pool scheduling.
// Uses LPT (Longest Processing Time First) - heaviest jobs scheduled first.
// Spawns a pool of worker goroutines that pull from the scheduler concurrently.
// Returns results in the same order as the input work items.
func (us *UnitScheduler) RunUnits(work []workunit.UnitSpec, worker UnitWorkerFunc) []UnitResult {
	results := make([]UnitResult, len(work))
	var resultsMu sync.Mutex
	var wg sync.WaitGroup

	// Create scheduler with dependency-based LPT scheduling.
	// Skip cycle validation since module-level topological sort already validated the graph.
	sched, err := scheduling.NewDependencyScheduler(work, scheduling.WithSkipValidation())
	if err != nil {
		// Circular dependencies are a configuration error.
		// Return early with error results for all work items.
		logging.C().Errorf("[scheduler] Failed to create scheduler: %v", err)
		for i, w := range work {
			results[i] = UnitResult{
				Module:    w.ID.Module,
				Component: w.ID.ComponentName,
				Handler:   w.ID.Tool,
				ExitCode:  1,
				Errors:    []string{fmt.Sprintf("circular dependency: %v", err)},
			}
		}
		return results
	}
	us.scheduler = sched // Store for stats access

	// Create all tabs as QUEUED upfront (positions are immutable)
	for _, w := range work {
		moniker := w.ID.Longname()
		displayName := w.DisplayName()
		us.tuiMarkPending(moniker, displayName, w.Weight, w.Tags)
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

	// Determine worker pool size: min(work items, host capacity)
	// This ensures we can saturate capacity immediately
	poolSize := len(work)
	hostCap := us.semaphore.HostCapacity()
	dockerCap := us.semaphore.DockerCapacity()
	if hostCap.Total > 0 && poolSize > hostCap.Total {
		poolSize = hostCap.Total
	}

	// Store roof (hard ceiling) for three-value capacity model
	// Roof is the actual peak allocation - workers spawned at start
	us.roof = poolSize

	// Store total capacities for over-capacity detection
	totalHostCapacity := hostCap.Total
	totalDockerCapacity := dockerCap.Total

	// Spawn worker pool - all workers start immediately and compete for scheduler items
	// Scheduler returns heaviest ready item (LPT); orchestrator handles capacity via semaphore
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				// Get next ready item from scheduler (LPT ordering, deps satisfied)
				// Blocks until an item is ready or scheduler is exhausted
				spec := sched.WaitForReady()
				if spec == nil {
					return // scheduler exhausted
				}

				// Get pool allocation for this unit
				alloc := spec.GetPoolAllocation().(resource.PoolAllocation)

				// Check if this is an over-capacity item (weight exceeds pool capacity)
				// Over-capacity items skip semaphore (would block forever) and execute serially
				isOverCapacity := (totalHostCapacity > 0 && alloc.HostWeight > totalHostCapacity) ||
					(totalDockerCapacity > 0 && alloc.DockerWeight > 0 && alloc.DockerWeight > totalDockerCapacity)

				// For over-capacity items: serialize via mutex (only one at a time)
				// For normal items: acquire weighted semaphore capacity
				if isOverCapacity {
					us.overCapacityMu.Lock()
					us.overCapacityItem = &spec.ID
				} else {
					us.semaphore.Acquire(context.Background(), alloc)
				}
				us.emitResourceStatus() // Update TUI Resources pane

				// Update TUI: queued -> running (uses moniker for matching)
				us.tuiMarkRunning(spec.ID.Longname())

				var result UnitResult

				// Check if any dependency failed - if so, skip execution and fail immediately
				if sched.HasFailedDependency(spec.ID) {
					failedMods := sched.FailedDependencyModules(spec.ID)
					errMsg := "Skipped: dependency failed"
					if len(failedMods) > 0 {
						errMsg = fmt.Sprintf("Skipped: dependency failed (%s)", strings.Join(failedMods, ", "))
					}
					result = UnitResult{
						Module:    spec.ID.Module,
						Component: spec.ID.ComponentName,
						Handler:   spec.ID.Tool,
						ExitCode:  1,
						Errors:    []string{errMsg},
					}
				} else {
					// Execute work
					result = us.executeWorker(*spec, worker)
				}

				// Release capacity/mutex
				if isOverCapacity {
					us.overCapacityItem = nil
					us.overCapacityMu.Unlock()
				} else {
					us.semaphore.Release(alloc)
				}
				us.emitResourceStatus() // Update TUI Resources pane

				// Store result by original index
				resultsMu.Lock()
				results[spec.Index] = result
				resultsMu.Unlock()

				// Notify scheduler: mark as completed or failed based on result
				if result.ExitCode > 0 {
					// Cascade failure to all transitive dependents still in queue.
					// Returns specs removed from the queue — they are NOT executed.
					cascaded := sched.MarkFailedCascade(spec.ID)

					// Process cascade-failed items: store results, update TUI, track completion
					for ci := range cascaded {
						cascadeResult := UnitResult{
							Module:    cascaded[ci].ID.Module,
							Component: cascaded[ci].ID.ComponentName,
							Handler:   cascaded[ci].ID.Tool,
							ExitCode:  1,
							Errors:    []string{fmt.Sprintf("Skipped: dependency failed (%s)", spec.ID.Module)},
						}
						resultsMu.Lock()
						results[cascaded[ci].Index] = cascadeResult
						resultsMu.Unlock()

						us.markUnitComplete(cascaded[ci], &cascadeResult)
					}
				} else {
					sched.MarkComplete(spec.ID)
				}

				// Mark component complete (broadcasts to legacy channels, updates TUI)
				us.markUnitComplete(*spec, &result)
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()

	// Defensive check: verify scheduler is fully drained
	// If items remain, workers exited prematurely (indicates a bug)
	if remaining := sched.Len(); remaining > 0 {
		logging.C().Warnf("[scheduler] BUG: scheduler has %d items remaining after all workers exited", remaining)
	}

	// Close scheduler
	sched.Close()

	return results
}

// markUnitComplete marks a component as done, updates summary and TUI.
func (us *UnitScheduler) markUnitComplete(spec workunit.UnitSpec, result *UnitResult) {
	// Send result to summary builder for incremental summary computation
	if us.summaryBuilder != nil {
		us.summaryBuilder.AddResult(*result)
	}

	// Update TUI with exit code
	// Use Longname() to match the ID format used in SetInitSummary
	moniker := spec.ID.Longname()
	us.tuiMarkCompleted(moniker, result.ExitCode)
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
