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

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/adapters/tui"
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

	// Current component scheduler (set during RunUnitsLayered/RunUnitsParallel)
	currentScheduler *UnitScheduler

	// TUI console for real-time output display
	tuiConsole tui.Console
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

	// Observer pattern fields
	observers     []interfaces.ExecutionObserver
	observersMu   sync.RWMutex
	writerFactory interfaces.WriterFactory
}

// New creates a new Orchestrator with the given configuration and worker function.
func New(config *Config, worker WorkerFunc) *Orchestrator {
	// MaxConcurrency=0 means dynamic (calculated from CPU×RAM).
	// Only use MaxConcurrency as ceiling if explicitly set by user.

	// Set default TUI height
	if config.TUIHeight <= 0 {
		config.TUIHeight = tui.DefaultHeight
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

// SetConsole sets an external TUI console for the orchestrator.
// This allows the cmdframework to inject a console created via the TUI registry.
// If called with a non-nil console, it replaces any console created in New().
// The console should implement the tui.Console interface.
func (o *Orchestrator) SetConsole(console tui.Console) {
	if console != nil {
		o.tuiConsole = console
		if o.tuiCtx == nil {
			o.tuiCtx, o.tuiCancel = context.WithCancel(context.Background())
		}
	}
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
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(allMonikers),
	})

	// Print header with layer info
	fmt.Fprintf(o.orchestratorOut, "%s %d modules in %d layer(s)%s%s",
		capitalize(o.config.ActionVerb), len(allMonikers), len(layers), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(allMonikers), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run when starting execution
	o.SetPhase(tui.PhaseRun)

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
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(monikers),
	})

	// Print header (use display names for cleaner output)
	displayNames := output.PackageDisplayNames(monikers)
	fmt.Fprintf(o.orchestratorOut, "%s %d items in parallel:%s%s%s",
		capitalize(o.config.ActionVerb), len(monikers), output.ListFormat(displayNames, 60, 5), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(monikers), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run when starting execution
	o.SetPhase(tui.PhaseRun)

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
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName)
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
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName)
		result.Duration = time.Since(startTime)
		markCompleted()
		return result
	}

	// Create log file (use separate file in dry-run mode to preserve real build logs)
	logFileName := o.config.LogFileName
	if o.config.DryRun {
		logFileName = "dry-run." + logFileName
	}
	logPath := filepath.Join(moduleOutputDir, logFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create log file %s: %v", logPath, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName)
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

	// Execute worker function
	exitCode := o.worker(item.Moniker, workerWriter)

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
	result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName)
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
	fmt.Fprintf(o.orchestratorOut, "%s%s%s", nl, output.SectionHeader(capitalize(o.config.ActionVerb)+" Summary"), nl)
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

// WaitTUI waits for the TUI to exit naturally (e.g., user presses a key).
// Use this after SendSummary() to wait for user to review and exit.
func (o *Orchestrator) WaitTUI() {
	if o.tuiConsole != nil {
		o.tuiConsole.Wait()
	}
	// Restore stdout after TUI exits
	o.orchestratorOut = os.Stdout
	o.logger = log.New(o.orchestratorOut, "", 0)
}

// StopTUI stops the TUI console and restores stdout output.
// Must be called before PrintSummary when TUI is enabled.
func (o *Orchestrator) StopTUI() {
	if o.tuiConsole != nil {
		o.tuiConsole.Stop()
		o.tuiConsole = nil
	}
	if o.tuiCancel != nil {
		o.tuiCancel()
		o.tuiCancel = nil
	}
	// Restore stdout for subsequent output (like PrintSummary)
	o.orchestratorOut = os.Stdout
	o.logger = log.New(o.orchestratorOut, "", 0)
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

// AddObserver registers an observer to receive execution events.
func (o *Orchestrator) AddObserver(observer interfaces.ExecutionObserver) {
	o.observersMu.Lock()
	defer o.observersMu.Unlock()
	o.observers = append(o.observers, observer)
}

// RemoveObserver unregisters an observer.
func (o *Orchestrator) RemoveObserver(observer interfaces.ExecutionObserver) {
	o.observersMu.Lock()
	defer o.observersMu.Unlock()
	for i, obs := range o.observers {
		if obs == observer {
			o.observers = append(o.observers[:i], o.observers[i+1:]...)
			break
		}
	}
}

// SetWriterFactory sets the factory for creating output writers.
// If not set, output goes only to log files.
func (o *Orchestrator) SetWriterFactory(factory interfaces.WriterFactory) {
	o.writerFactory = factory
}

// emit sends an event to all registered observers.
func (o *Orchestrator) emit(event interfaces.ExecutionEvent) {
	o.observersMu.RLock()
	observers := make([]interfaces.ExecutionObserver, len(o.observers))
	copy(observers, o.observers)
	o.observersMu.RUnlock()

	for _, obs := range observers {
		obs.OnEvent(event)
	}
}

// tuiMarkRunning adds a module to the running list and updates TUI status.
func (o *Orchestrator) tuiMarkRunning(moniker string) {
	o.tuiMu.Lock()
	o.tuiRunning = append(o.tuiRunning, moniker)
	running := make([]string, len(o.tuiRunning))
	copy(running, o.tuiRunning)
	completed := o.tuiCompleted
	total := o.tuiTotal
	o.tuiMu.Unlock()

	// Emit events to observers
	o.emit(interfaces.UnitStartedEvent{
		Time: time.Now(),
		ID:   moniker,
	})
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// tuiMarkCompleted removes a module from running, increments completed count, and reports exit code.
func (o *Orchestrator) tuiMarkCompleted(moniker string, exitCode int) {
	o.tuiMu.Lock()
	// Remove from running list
	for i, m := range o.tuiRunning {
		if m == moniker {
			o.tuiRunning = append(o.tuiRunning[:i], o.tuiRunning[i+1:]...)
			break
		}
	}
	o.tuiCompleted++
	running := make([]string, len(o.tuiRunning))
	copy(running, o.tuiRunning)
	completed := o.tuiCompleted
	total := o.tuiTotal
	o.tuiMu.Unlock()

	// Emit events to observers
	o.emit(interfaces.UnitCompletedEvent{
		Time:     time.Now(),
		ID:       moniker,
		ExitCode: exitCode,
	})
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// SetPhase switches to a specific phase (Init, Run, End).
// This emits a PhaseStartedEvent to all registered observers.
func (o *Orchestrator) SetPhase(phase tui.Phase) {
	o.emit(interfaces.PhaseStartedEvent{
		Phase:     interfaces.ExecutionPhase(phase),
		Time:      time.Now(),
		TotalWork: o.tuiTotal,
	})
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed).
// Note: This is handled by observers via phase events.
func (o *Orchestrator) SetPhaseSummary(phase tui.Phase, summary string) {
	// Phase summary is now handled via PhaseCompletedEvent
	// This method is kept for backward compatibility
}

// CompletePhase marks a phase as complete.
// This emits a PhaseCompletedEvent to all registered observers.
func (o *Orchestrator) CompletePhase(phase tui.Phase, success bool, summary string) {
	o.emit(interfaces.PhaseCompletedEvent{
		Phase:   interfaces.ExecutionPhase(phase),
		Time:    time.Now(),
		Success: success,
		Summary: summary,
	})
}

// WriteToPhase writes a line to a specific phase's buffer.
// This emits an OutputLineEvent to all registered observers.
func (o *Orchestrator) WriteToPhase(phase tui.Phase, text string) {
	o.emit(interfaces.OutputLineEvent{
		Time:   time.Now(),
		Source: phase.String(),
		Text:   text,
		Level:  interfaces.OutputLevelInfo,
	})
}

// SendInitLine sends a line to the Init phase buffer.
func (o *Orchestrator) SendInitLine(text string) {
	o.WriteToPhase(tui.PhaseInit, text)
}

// SendEndLine sends a line to the results buffer (appears below Run pane).
// This emits an OutputLineEvent to all registered observers.
func (o *Orchestrator) SendEndLine(text string) {
	o.emit(interfaces.OutputLineEvent{
		Time:   time.Now(),
		Source: "results",
		Text:   text,
		Level:  interfaces.OutputLevelInfo,
	})
}

// IsTUIEnabled returns whether TUI mode is enabled in config.
func (o *Orchestrator) IsTUIEnabled() bool {
	return o.config.TUI
}

// IsTUIStarted returns whether StartTUI has been called.
// When TUI is enabled but not started, WriteInit should output to console.
func (o *Orchestrator) IsTUIStarted() bool {
	o.tuiMu.Lock()
	defer o.tuiMu.Unlock()
	return o.tuiStarted
}

// phaseWriter implements io.Writer and forwards all writes to a specific phase.
type phaseWriter struct {
	orch  *Orchestrator
	phase tui.Phase
}

// Write implements io.Writer by emitting OutputLineEvent for the phase.
func (w *phaseWriter) Write(p []byte) (n int, err error) {
	// Convert bytes to string, trim trailing newline (observers add their own)
	text := string(p)
	text = strings.TrimSuffix(text, "\n")
	if text != "" {
		w.orch.WriteToPhase(w.phase, text)
	}
	return len(p), nil
}

// GetTUIWriter returns an io.Writer that sends output to the specified phase.
// Returns nil if TUI mode is not enabled.
func (o *Orchestrator) GetTUIWriter(phase tui.Phase) io.Writer {
	if !o.config.TUI {
		return nil
	}
	return &phaseWriter{orch: o, phase: phase}
}

// SendSummary sends summary data to all registered observers.
// This emits a SummaryReadyEvent that observers can handle appropriately.
func (o *Orchestrator) SendSummary(data *tui.SummaryData) {
	if data == nil {
		return
	}
	o.emit(interfaces.SummaryReadyEvent{
		Time:      time.Now(),
		Success:   data.Success,
		TotalTime: data.TotalTime,
		Details:   data.Details,
		NextSteps: data.NextSteps,
	})
}

// SetInitSummary sends structured init summary data to all registered observers.
// This emits an InitSummaryEvent that observers can handle appropriately.
func (o *Orchestrator) SetInitSummary(data *tui.InitSummary) {
	if data == nil {
		return
	}

	// Convert TUI InitSummary to observer event
	modules := make([]interfaces.ModuleInfo, len(data.ExecutionTree))
	for i, mod := range data.ExecutionTree {
		units := make([]interfaces.UnitInfo, len(mod.UoWs))
		for j, uow := range mod.UoWs {
			units[j] = interfaces.UnitInfo{
				ID:          uow.ID,
				DisplayName: uow.DisplayName,
				Weight:      uow.Weight,
			}
		}
		modules[i] = interfaces.ModuleInfo{
			Name:  mod.Name,
			Units: units,
		}
	}

	// Convert PlannedTools to interface type
	plannedTools := make([]interfaces.PlannedToolInfo, len(data.PlannedTools))
	for i, tool := range data.PlannedTools {
		plannedTools[i] = interfaces.PlannedToolInfo{
			Name:        tool.Name,
			IsContainer: tool.IsContainer,
		}
	}

	o.emit(interfaces.InitSummaryEvent{
		Time:             time.Now(),
		Command:          data.Command,
		ExecutionContext: data.ExecutionContext,
		RequestedModules: data.RequestedModules,
		ResolvedModules:  data.CalculatedModules,
		TotalUnits:       data.UoWCount,
		Modules:          modules,
		Parallelism: interfaces.ParallelismInfo{
			Mode:             data.ParallelismMode,
			EffectiveWorkers: data.EffectiveWorkers,
			TurboBoost:       data.TurboBoost,
			WeightedCapacity: data.WeightedCapacity,
		},
		Flags: interfaces.FlagsInfo{
			TidyFirst:    data.Flags.TidyFirst,
			ForceRebuild: data.Flags.ForceRebuild,
			DryRun:       data.Flags.DryRun,
			UseTUI:       data.Flags.UseTUI,
			SkipDeps:     data.Flags.SkipDeps,
			SkipDepm:     data.Flags.SkipDepm,
		},
		PlannedTools: plannedTools,
	})
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

// StartTUI starts only the TUI console without running any jobs.
// Use this to enable phase output before calling Run or RunLayered.
// Returns quickly after TUI is initialized.
// Call Init() first if you need logging infrastructure.
func (o *Orchestrator) StartTUI() {
	if o.tuiConsole == nil {
		// Debug: TUI console not set - this means the type assertion failed or
		// useRegistryTUI was false. Check cmdframework init.go debug logs.
		return
	}
	o.tuiConsole.StartAsync(o.tuiCtx)
	// Set initial phase to Init
	o.SetPhase(tui.PhaseInit)
	// Mark TUI as started so WriteInit knows to skip console output
	o.tuiMu.Lock()
	o.tuiStarted = true
	o.tuiMu.Unlock()
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
// RunUnitsLayered or RunUnitsParallel call.
// Returns nil if no component execution has occurred.
func (o *Orchestrator) GetLastUnitResults() []UnitResult {
	return o.lastUnitResults
}

// RunUnitsLayered executes component builds in dependency layers.
// Within each layer, components run in parallel with weighted scheduling.
// Components respect intra-module dependencies (DependsOn) and inter-module
// dependencies (layers).
//
// Returns WorkResult aggregated at module level for compatibility with existing code.
func (o *Orchestrator) RunUnitsLayered(layers [][]workunit.UnitSpec, worker UnitWorkerFunc) ([]WorkResult, error) {
	// Flatten to get total count
	var allWork []workunit.UnitSpec
	for _, layer := range layers {
		allWork = append(allWork, layer...)
	}

	if len(allWork) == 0 {
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

	defer func() {
		scheduler.Close()
		o.currentScheduler = nil
	}()

	// Initialize tracking state
	o.tuiMu.Lock()
	o.tuiTotal = len(allWork)
	o.tuiCompleted = 0
	o.tuiRunning = nil
	o.tuiMu.Unlock()

	// Start TUI console if enabled (lifecycle management)
	if o.tuiConsole != nil {
		o.tuiConsole.StartAsync(o.tuiCtx)
	}

	// Emit initial progress event to all observers
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(allWork),
	})

	// Print header
	uniqueModules := countUniqueModules(allWork)
	fmt.Fprintf(o.orchestratorOut, "%s %d components across %d modules in %d layer(s)%s%s",
		capitalize(o.config.ActionVerb), len(allWork), uniqueModules, len(layers), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(allWork), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run
	o.SetPhase(tui.PhaseRun)

	// Execute layers sequentially
	var allResults []UnitResult

	for layerIdx, layerWork := range layers {
		if len(layerWork) == 0 {
			continue
		}

		// Format layer info
		layerModules := getLayerModules(layerWork)
		fmt.Fprintf(o.orchestratorOut, "Layer %d: %s (%d components)%s",
			layerIdx+1, formatMonikerList(layerModules), len(layerWork), LineEnding)

		// Initialize work items with layer-relative indices
		// (RunComponents creates a results array sized for this layer only)
		for i := range layerWork {
			layerWork[i].Index = i
		}

		// Initialize scheduler for this layer
		scheduler.InitializeWork(layerWork)

		// Execute layer with parallel component scheduling
		layerResults := scheduler.RunUnits(layerWork, worker)

		// Collect results
		allResults = append(allResults, layerResults...)

		// Check if any component failed - stop if so
		// Note: ExitCode -1 means cached/skipped, which is NOT a failure
		for _, result := range layerResults {
			if result.ExitCode > 0 {
				// Aggregate to module results before returning
				if o.display != nil {
					o.display.stop()
					o.display.flushCompletedLines()
				}
				// Store component results for detailed reporting
				o.lastUnitResults = allResults
				return AggregateToWorkResults(allResults, allWork), nil
			}
		}
	}

	// Stop display manager
	if o.display != nil {
		o.display.stop()
		o.display.flushCompletedLines()
	}

	// Store component results for detailed reporting
	o.lastUnitResults = allResults

	// Aggregate component results to module results
	return AggregateToWorkResults(allResults, allWork), nil
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
	o.emit(interfaces.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(work),
	})

	// Print header
	uniqueModules := countUniqueModules(work)
	fmt.Fprintf(o.orchestratorOut, "%s %d components across %d modules in parallel%s%s",
		capitalize(o.config.ActionVerb), len(work), uniqueModules, LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(work), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run
	o.SetPhase(tui.PhaseRun)

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

// getLayerModules returns unique module names from work units.
func getLayerModules(work []workunit.UnitSpec) []string {
	seen := make(map[string]bool)
	var modules []string
	for _, w := range work {
		module := w.ID.Module
		if !seen[module] {
			seen[module] = true
			modules = append(modules, module)
		}
	}
	return modules
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
