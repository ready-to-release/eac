package orchestrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/locktracker"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
)

// ComponentScheduler manages parallel execution of component work items
// with weighted resource control and dependency ordering.
type ComponentScheduler struct {
	config    Config
	semaphore *WeightedSemaphore

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
func NewComponentScheduler(config Config, tuiConsole *tui.Console, registry *locktracker.Registry) *ComponentScheduler {
	var sem *WeightedSemaphore
	if registry != nil {
		sem = NewWeightedSemaphoreWithRegistry("component-scheduler", config.MaxConcurrency, registry)
	} else {
		sem = NewWeightedSemaphore(config.MaxConcurrency)
	}

	cs := &ComponentScheduler{
		config:           config,
		semaphore:        sem,
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

	return cs
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
// Returns results in the same order as the input work items.
func (cs *ComponentScheduler) RunComponents(work []ComponentWork, worker ComponentWorkerFunc) []ComponentResult {
	results := make([]ComponentResult, len(work))
	var wg sync.WaitGroup

	// Launch all components - they will wait for their dependencies
	for _, w := range work {
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

	// 2. Acquire weighted slot
	weight := work.Weight
	if weight <= 0 {
		weight = 1
	}
	cs.semaphore.Acquire(weight)
	defer cs.semaphore.Release(weight)

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
		cs.markComponentComplete(work, &result)
		return result
	}

	// Mark as running
	cs.tuiMarkRunning(work.Module, work.Component)

	// Create writer for worker
	var workerWriter io.Writer
	if cs.tuiConsole != nil {
		displayName := fmt.Sprintf("%s:%s", work.Module, work.Component)
		workerWriter = cs.tuiConsole.NewWriter(displayName, logFile)
	} else {
		workerWriter = logFile
	}

	// 4. Execute build
	exitCode := worker(work.Module, work.Component, workerWriter)
	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = relLogPath
	result.Duration = time.Since(startTime)

	// Mark as completed
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

	// Update TUI
	cs.tuiMarkCompleted(work.Module, work.Component)
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

// tuiMarkRunning adds a component to the running list.
func (cs *ComponentScheduler) tuiMarkRunning(module, component string) {
	if cs.tuiConsole == nil {
		return
	}

	displayName := fmt.Sprintf("%s:%s", module, component)

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
	})
}

// tuiMarkCompleted removes a component from running and increments completed.
func (cs *ComponentScheduler) tuiMarkCompleted(module, component string) {
	if cs.tuiConsole == nil {
		return
	}

	displayName := fmt.Sprintf("%s:%s", module, component)

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
	})
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
