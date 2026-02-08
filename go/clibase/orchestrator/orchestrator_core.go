package orchestrator

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// Orchestrator manages parallel execution of work items.
type Orchestrator struct {
	config          Config
	worker          WorkerFunc
	display         *displayManager
	orchestratorOut io.Writer
	logger          *log.Logger // goroutine-safe logger

	// Lock tracking registry for visualization
	registry *locktracker.Registry

	// Current component scheduler (set during RunUnitsParallel)
	currentScheduler *UnitScheduler

	// TUI console for real-time output display
	tuiConsole display.Console
	tuiCtx     context.Context
	tuiCancel  context.CancelFunc

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int
	tuiStarted   bool // True once StartTUI has been called

	// Component results from last execution (for detailed reporting)
	lastUnitResults []UnitResult

	// Pending cache times to apply when scheduler is created
	pendingCacheTimes map[string]time.Time

	// Pending cache detection config to apply when scheduler is created
	pendingCacheVerifier      CacheVerifier
	pendingCachedModules      map[string]bool

	// Summary builder for incremental summary computation
	summaryBuilder SummaryBuilder

	// initSummaryEmitted tracks whether SetInitSummary was called.
	// Used to tell UnitScheduler to skip redundant UnitQueuedEvent emissions.
	initSummaryEmitted bool

	// Observer pattern fields
	observers     []core.ExecutionObserver
	observersMu   sync.RWMutex
	writerFactory core.WriterFactory
}

// New creates a new Orchestrator with the given configuration and worker function.
func New(config *Config, worker WorkerFunc) *Orchestrator {
	// MaxConcurrency=0 means dynamic (calculated from CPU×RAM).
	// Only use MaxConcurrency as ceiling if explicitly set by user.

	// Set default TUI height
	if config.TUIHeight <= 0 {
		config.TUIHeight = display.DefaultHeight
	}

	// Use global lock tracking registry (shared with component scheduler and other components)
	registry := locktracker.Get()

	o := &Orchestrator{
		config:   *config,
		worker:   worker,
		registry: registry,
	}

	// Note: TUI console is no longer created here.
	// Use AddObserver() to register a TUIObserver for TUI display.
	// The tuiConsole field is kept for backward compatibility with SetConsole().

	return o
}

// Init initializes the orchestrator's output infrastructure.
// This must be called before StartTUI or using phase methods.
// It's automatically called by Run/RunLayered if not called explicitly.
func (o *Orchestrator) Init() error {
	if o.orchestratorOut != nil {
		return nil // Already initialized
	}

	// Configure output writer based on TUI mode
	// When TUI is enabled, discard orchestrator-level output (TUI handles display)
	// When TUI is disabled, write to stdout
	var writer io.Writer
	if o.config.TUI {
		writer = io.Discard // TUI mode - TUI handles console display
	} else {
		writer = os.Stdout
	}

	o.logger = log.New(writer, "", 0)
	o.orchestratorOut = writer

	return nil
}

// RunLayered executes modules in dependency layers - layers run sequentially, modules within a layer run in parallel.
func (o *Orchestrator) RunLayered(layers [][]string) ([]WorkResult, error) {
	// Flatten layers to get total count
	var allMonikers []string
	for _, layer := range layers {
		allMonikers = append(allMonikers, layer...)
	}

	// Initialize if not already done
	if err := o.Init(); err != nil {
		return nil, err
	}

	// Initialize tracking state
	o.tuiMu.Lock()
	o.tuiTotal = len(allMonikers)
	o.tuiCompleted = 0
	o.tuiRunning = nil
	o.tuiMu.Unlock()

	// Start TUI console if enabled (lifecycle management)
	if o.tuiConsole != nil {
		o.tuiConsole.StartAsync(o.tuiCtx)
	}

	// Emit initial progress event to all observers
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(allMonikers),
	})

	// Print header with layer info
	fmt.Fprintf(o.orchestratorOut, "%s %d modules in %d layer(s)%s%s",
		capitalize(o.config.ActionVerb()), len(allMonikers), len(layers), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb(), len(allMonikers), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run when starting execution
	o.SetPhase(display.PhaseRun)

	// Execute layers sequentially
	allResults := make([]WorkResult, len(allMonikers))
	globalIndex := 0

	for layerIdx, layerMonikers := range layers {
		if len(layerMonikers) == 0 {
			continue
		}

		fmt.Fprintf(o.orchestratorOut, "Layer %d: %s%s", layerIdx+1, formatMonikerList(layerMonikers), LineEnding)

		// Create work items for this layer with local indices (0-based within layer)
		workItems := make([]WorkItem, len(layerMonikers))
		for i, moniker := range layerMonikers {
			workItems[i] = WorkItem{
				Moniker: moniker,
				Index:   i, // Local index within this layer's results
			}
		}

		// Execute this layer in parallel
		layerResults := o.executeParallel(workItems)

		// Copy layer results to all results at the correct global offset
		for i, result := range layerResults {
			result.Index = globalIndex + i // Update to global index
			allResults[globalIndex+i] = result
		}

		// Check if any module in this layer failed - stop if so
		// Note: ExitCode -1 means cached/skipped, which is NOT a failure
		for _, result := range layerResults {
			if result.ExitCode > 0 {
				// Stop display and return early with results up to this point
				if o.display != nil {
					o.display.stop()
					o.display.flushCompletedLines()
				}
				return allResults[:globalIndex+len(layerMonikers)], nil
			}
		}

		globalIndex += len(layerMonikers)
	}

	// Stop display manager and flush all collected completion lines
	if o.display != nil {
		o.display.stop()
		o.display.flushCompletedLines()
	}

	return allResults, nil
}

// formatMonikerList formats a list of monikers for display.
func formatMonikerList(monikers []string) string {
	if len(monikers) <= 5 {
		return fmt.Sprintf("%v", monikers)
	}
	return fmt.Sprintf("%v... (%d total)", monikers[:5], len(monikers))
}

// Run executes all work items in parallel and returns the results.
func (o *Orchestrator) Run(monikers []string) ([]WorkResult, error) {
	// Initialize if not already done
	if err := o.Init(); err != nil {
		return nil, err
	}

	// Initialize tracking state
	o.tuiMu.Lock()
	o.tuiTotal = len(monikers)
	o.tuiCompleted = 0
	o.tuiRunning = nil
	o.tuiMu.Unlock()

	// Start TUI console if enabled (lifecycle management)
	if o.tuiConsole != nil {
		o.tuiConsole.StartAsync(o.tuiCtx)
	}

	// Emit initial progress event to all observers
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(monikers),
	})

	// Print header (use display names for cleaner output)
	displayNames := output.PackageDisplayNames(monikers)
	fmt.Fprintf(o.orchestratorOut, "%s %d items in parallel:%s%s%s",
		capitalize(o.config.ActionVerb()), len(monikers), output.ListFormat(displayNames, 60, 5), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb(), len(monikers), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run when starting execution
	o.SetPhase(display.PhaseRun)

	// Create work items
	workItems := make([]WorkItem, len(monikers))
	for i, moniker := range monikers {
		workItems[i] = WorkItem{
			Moniker: moniker,
			Index:   i,
		}
	}

	// Execute work items in parallel
	results := o.executeParallel(workItems)

	// Stop display manager and flush all collected completion lines
	if o.display != nil {
		o.display.stop()
		o.display.flushCompletedLines()
	}

	return results, nil
}

// executeParallel runs work items in parallel.
// Note: Concurrency is controlled at the component level by UnitScheduler.
// This method launches all module work items; actual resource throttling happens
// when components acquire slots from the weighted semaphore.
func (o *Orchestrator) executeParallel(workItems []WorkItem) []WorkResult {
	results := make([]WorkResult, len(workItems))
	var wg sync.WaitGroup

	for _, item := range workItems {
		wg.Add(1)

		go func(wi WorkItem) {
			defer wg.Done()
			result := o.processWorkItem(wi)
			results[wi.Index] = result
		}(item)
	}

	wg.Wait()
	return results
}

// processWorkItem processes a single work item.
func (o *Orchestrator) processWorkItem(item WorkItem) WorkResult {
	startTime := time.Now()

	result := WorkResult{
		Moniker: item.Moniker,
		Index:   item.Index,
	}

	// Helper to mark completed in display or TUI
	markCompleted := func() {
		if o.display != nil {
			o.display.markCompleted(&result)
		}
		o.tuiMarkCompleted(item.Moniker, result.ExitCode)
	}

	// Create output directory for this module
	// Use simplified display name for cleaner paths, then sanitize for filesystem
	sanitizedMoniker := sanitizePathForFS(output.PackageDisplayName(item.Moniker))
	moduleOutputDir := filepath.Join(o.config.WorkspaceRoot, o.config.OutputBaseDir, sanitizedMoniker)
	parentDir := filepath.Dir(moduleOutputDir)

	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create parent directory %s: %v", parentDir, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName())
		result.Duration = time.Since(startTime)
		markCompleted()
		return result
	}

	// In dry-run mode, preserve existing artifacts and use separate log file
	// This allows meta-tests to mock builds without destroying real build outputs
	if o.config.DryRun {
		// Don't purge - preserve existing artifacts
	} else {
		// Purge existing output directory (best effort)
		_ = os.RemoveAll(moduleOutputDir)
	}

	if err := os.MkdirAll(moduleOutputDir, 0o755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory %s: %v", moduleOutputDir, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName())
		result.Duration = time.Since(startTime)
		markCompleted()
		return result
	}

	// Create log file (use separate file in dry-run mode to preserve real build logs)
	logFileName := o.config.LogFileName()
	if o.config.DryRun {
		logFileName = "dry-run." + logFileName
	}
	logPath := filepath.Join(moduleOutputDir, logFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create log file %s: %v", logPath, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName())
		result.Duration = time.Since(startTime)
		markCompleted()
		return result
	}

	// Mark as running in display and TUI
	if o.display != nil {
		o.display.markRunning(item.Moniker)
	}
	o.tuiMarkRunning(item.Moniker)

	// Create writer for worker - use writerFactory if available (e.g., TUIObserver)
	var workerWriter io.Writer
	if o.writerFactory != nil {
		workerWriter = o.writerFactory.NewWriter(item.Moniker, logFile)
	} else {
		workerWriter = logFile
	}

	// Execute worker function (module-level workers get background context)
	exitCode := o.worker(context.Background(), item.Moniker, workerWriter)

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
	result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName())
	result.Duration = time.Since(startTime)

	// Set component type from config if available
	if o.config.ComponentTypesDisplay != nil {
		if t, ok := o.config.ComponentTypesDisplay[item.Moniker]; ok {
			result.Type = t
		}
	}

	// Mark as completed in display (will print completion line)
	markCompleted()

	return result
}

// PrintSummary prints a summary of all results to the orchestrator output.
func (o *Orchestrator) PrintSummary(results []WorkResult) {
	totalFailed := 0
	totalWarnings := 0
	var totalDuration time.Duration
	failedModules := []string{}

	for _, result := range results {
		totalDuration += result.Duration
		if result.ExitCode != 0 {
			totalFailed++
			failedModules = append(failedModules, result.Moniker)
		}
		if len(result.Warnings) > 0 {
			totalWarnings += len(result.Warnings)
		}
	}

	nl := LineEnding
	fmt.Fprintf(o.orchestratorOut, "%s%s%s", nl, output.SectionHeader(capitalize(o.config.ActionVerb())+" Summary"), nl)
	fmt.Fprintf(o.orchestratorOut, "%s%s", output.SummaryCount("Modules", len(results), len(results)-totalFailed, totalFailed), nl)

	if totalWarnings > 0 {
		fmt.Fprintf(o.orchestratorOut, "Warnings: %d%s", totalWarnings, nl)
	}

	// Show failed modules with errors
	if len(failedModules) > 0 {
		fmt.Fprintf(o.orchestratorOut, "%s❌ Failed:%s", nl, nl)
		for _, result := range results {
			if result.ExitCode != 0 {
				fmt.Fprintf(o.orchestratorOut, "  %s%s", result.Moniker, nl)
				if len(result.Errors) > 0 {
					for _, errMsg := range result.Errors {
						if len(errMsg) > 120 {
							errMsg = errMsg[:117] + "..."
						}
						fmt.Fprintf(o.orchestratorOut, "    %s%s", errMsg, nl)
					}
				}
			}
		}
	}

	// Timing summary (only shown with --timings flag)
	if o.config.ShowTimings {
		fmt.Fprintf(o.orchestratorOut, "%s%s%s", nl, output.SectionHeader("Timing Summary"), nl)

		// Sort results by duration (longest first)
		sortedResults := make([]WorkResult, len(results))
		copy(sortedResults, results)
		sort.Slice(sortedResults, func(i, j int) bool {
			return sortedResults[i].Duration > sortedResults[j].Duration
		})

		for _, result := range sortedResults {
			fmt.Fprintf(o.orchestratorOut, "%s%s", output.TimingLine(result.Duration, result.Moniker), nl)
		}
		fmt.Fprintf(o.orchestratorOut, "%s%s", output.TimingTotal(totalDuration), nl)
	}

	// Output location
	fmt.Fprintf(o.orchestratorOut, "%sOutput: %s%s", nl, o.config.OutputBaseDir, nl)
}

// Close releases resources held by the orchestrator.
func (o *Orchestrator) Close() {
	// Stop TUI if not already stopped
	o.StopTUI()

	// Note: registry is global, don't close it here
	// Component scheduler manages its own semaphore lifecycle
}

// GetRegistry returns the lock tracking registry.
// Useful for external components that need lock visualization.
func (o *Orchestrator) GetRegistry() *locktracker.Registry {
	return o.registry
}

// GetOrchestratorOut returns the writer for orchestrator-level output.
// When TUI is enabled, this writes to the log file only.
// When TUI is not enabled, this writes to both stdout and log file.
func (o *Orchestrator) GetOrchestratorOut() io.Writer {
	return o.orchestratorOut
}

// SetWorker sets the worker function for the orchestrator.
// Must be called before Run or RunLayered.
func (o *Orchestrator) SetWorker(worker WorkerFunc) {
	o.worker = worker
}

// UpdateConfig updates the orchestrator's configuration with loaded values.
// This is used when the orchestrator is created early (before config is loaded)
// and needs to be updated with actual values once available.
// Only non-zero/non-nil fields are applied.
func (o *Orchestrator) UpdateConfig(update ConfigUpdate) {
	if update.WorkspaceRoot != "" {
		o.config.WorkspaceRoot = update.WorkspaceRoot
	}
	if update.OutputBaseDir != "" {
		o.config.OutputBaseDir = update.OutputBaseDir
	}
	if update.MaxConcurrency > 0 {
		o.config.MaxConcurrency = update.MaxConcurrency
	}
	if update.Turbo > 0 {
		o.config.Turbo = update.Turbo
	}
	if update.ComponentTypesDisplay != nil {
		o.config.ComponentTypesDisplay = update.ComponentTypesDisplay
	}
	if update.ShowTimings {
		o.config.ShowTimings = true
	}
	if update.DryRun {
		o.config.DryRun = true
	}
}

// SetComponentTypesDisplay updates the component types display map in the config.
// Useful when component types are determined after orchestrator creation.
func (o *Orchestrator) SetComponentTypesDisplay(componentTypes map[string]string) {
	o.config.ComponentTypesDisplay = componentTypes
}

// SetMaxConcurrency updates the maximum concurrency for subsequent Run calls.
// Useful for running sequential tests after parallel tests.
func (o *Orchestrator) SetMaxConcurrency(maxConcurrency int) {
	o.config.MaxConcurrency = maxConcurrency
}

// SetUnitExtras stores additional data for a component result.
// This is called by workers to pass test counts or other data that will be
// merged into the UnitResult when processing completes.
// Does nothing if no component scheduler is active.
func (o *Orchestrator) SetUnitExtras(module, component string, extras UnitExtras) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetUnitExtras(module, component, extras)
	}
}

// SetCacheTimes sets the cache times for modules that are up-to-date (cached).
// These times are passed to the TUI to display when cached artifacts were built.
// If the scheduler hasn't been created yet, times are stored and applied when it starts.
func (o *Orchestrator) SetCacheTimes(times map[string]time.Time) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetCacheTimes(times)
	} else {
		// Store for later application when scheduler is created
		o.pendingCacheTimes = times
	}
}

// SetCacheDetection configures background cache detection for early TUI feedback.
// When set, cached tabs will progressively "light up" blue as detection completes,
// and workers will short-circuit for already-detected cached items.
//
// verifier: function to check if a component is cached
// cachedModules: pre-computed set of modules known to be cached
//
// If the scheduler hasn't been created yet, config is stored and applied when it starts.
func (o *Orchestrator) SetCacheDetection(verifier CacheVerifier, cachedModules map[string]bool) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetCacheDetection(verifier, cachedModules)
	} else {
		// Store for later application when scheduler is created
		o.pendingCacheVerifier = verifier
		o.pendingCachedModules = cachedModules
	}
}

// SetSummaryBuilder sets the summary builder for incremental summary computation.
// The builder receives component results as they complete, enabling parallel
// summary computation during execution.
func (o *Orchestrator) SetSummaryBuilder(builder SummaryBuilder) {
	o.summaryBuilder = builder
}

// GetLastUnitResults returns the component-level results from the last
// RunUnitsParallel call.
// Returns nil if no component execution has occurred.
func (o *Orchestrator) GetLastUnitResults() []UnitResult {
	return o.lastUnitResults
}

// RunUnitsParallel executes all components in parallel with weighted scheduling.
// Components respect intra-module dependencies (DependsOn).
//
// Returns WorkResult aggregated at module level for compatibility with existing code.
func (o *Orchestrator) RunUnitsParallel(work []workunit.UnitSpec, worker UnitWorkerFunc) ([]WorkResult, error) {
	if len(work) == 0 {
		return []WorkResult{}, nil
	}

	// Initialize if not already done
	if err := o.Init(); err != nil {
		return nil, err
	}

	// Create component scheduler with dynamic capacity management
	scheduler := NewUnitScheduler(&o.config, o.tuiConsole, o.registry, o.emit, o.writerFactory)
	o.currentScheduler = scheduler

	// Apply pending cache times if any
	if o.pendingCacheTimes != nil {
		scheduler.SetCacheTimes(o.pendingCacheTimes)
		o.pendingCacheTimes = nil
	}

	// Apply pending cache detection if any
	if o.pendingCacheVerifier != nil {
		scheduler.SetCacheDetection(o.pendingCacheVerifier, o.pendingCachedModules)
		o.pendingCacheVerifier = nil
		o.pendingCachedModules = nil
	}

	// Apply summary builder if set
	if o.summaryBuilder != nil {
		scheduler.SetSummaryBuilder(o.summaryBuilder)
	}

	// Skip redundant UnitQueuedEvent if InitSummary already registered all tabs
	if o.initSummaryEmitted {
		scheduler.SetInitSummaryEmitted()
	}

	defer func() {
		scheduler.Close()
		o.currentScheduler = nil
	}()

	// Initialize tracking state
	o.tuiMu.Lock()
	o.tuiTotal = len(work)
	o.tuiCompleted = 0
	o.tuiRunning = nil
	o.tuiMu.Unlock()

	// Start TUI console if enabled (lifecycle management)
	if o.tuiConsole != nil {
		o.tuiConsole.StartAsync(o.tuiCtx)
	}

	// Emit initial progress event to all observers
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(work),
	})

	// Print header
	uniqueModules := countUniqueModules(work)
	fmt.Fprintf(o.orchestratorOut, "%s %d components across %d modules in parallel%s%s",
		capitalize(o.config.ActionVerb()), len(work), uniqueModules, LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb(), len(work), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run
	o.SetPhase(display.PhaseRun)

	// Set indices
	for i := range work {
		work[i].Index = i
	}

	// Initialize scheduler
	scheduler.InitializeWork(work)

	// Execute all components
	results := scheduler.RunUnits(work, worker)

	// Stop display manager
	if o.display != nil {
		o.display.stop()
		o.display.flushCompletedLines()
	}

	// Store component results for detailed reporting
	o.lastUnitResults = results

	// Aggregate component results to module results
	return AggregateToWorkResults(results, work), nil
}

// countUniqueModules counts unique modules in work units.
func countUniqueModules(work []workunit.UnitSpec) int {
	seen := make(map[string]bool)
	for _, w := range work {
		seen[w.ID.Module] = true
	}
	return len(seen)
}

// GetExitCode returns the appropriate exit code based on results.
// Returns 1 if any module failed (ExitCode > 0), 0 otherwise.
// ExitCode -1 means skipped/cached and is treated as success.
func GetExitCode(results []WorkResult) int {
	for _, result := range results {
		if result.ExitCode > 0 {
			return 1
		}
	}
	return 0
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	// Simple ASCII capitalization
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-32) + s[1:]
	}
	return s
}

// sanitizePathForFS converts a moniker to a filesystem-safe path
// Replaces : with _ (Windows doesn't allow : in paths)
// Normalizes path separators to forward slashes.
func sanitizePathForFS(path string) string {
	safe := strings.ReplaceAll(path, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}
