package orchestrator

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
)

// Orchestrator manages parallel execution of work items
type Orchestrator struct {
	config          Config
	worker          WorkerFunc
	display         *displayManager
	orchestratorOut io.Writer
	logger          *log.Logger // goroutine-safe logger

	// TUI console for real-time output display
	tuiConsole *tui.Console
	tuiCtx     context.Context
	tuiCancel  context.CancelFunc

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int
}

// New creates a new Orchestrator with the given configuration and worker function
func New(config Config, worker WorkerFunc) *Orchestrator {
	// Set default max concurrency to number of CPUs
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = runtime.NumCPU()
	}

	// Set default TUI height
	if config.TUIHeight <= 0 {
		config.TUIHeight = tui.DefaultHeight
	}

	o := &Orchestrator{
		config: config,
		worker: worker,
	}

	// Initialize TUI console if enabled
	if config.TUI {
		o.tuiConsole = tui.New(tui.Config{
			Height:       config.TUIHeight,
			ShowHeader:   true,
			BufferSize:   1000,
			RunPhaseName: config.ActionVerb,
		})
		o.tuiCtx, o.tuiCancel = context.WithCancel(context.Background())
	}

	return o
}

// RunLayered executes modules in dependency layers - layers run sequentially, modules within a layer run in parallel
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

	// Start TUI console if enabled (and not already started)
	if o.tuiConsole != nil {
		// Initialize TUI status tracking
		o.tuiTotal = len(allMonikers)
		o.tuiCompleted = 0
		o.tuiRunning = nil

		// StartAsync is idempotent - will not restart if already running
		o.tuiConsole.StartAsync(o.tuiCtx)
		// Send initial status
		o.tuiConsole.UpdateStatus(tui.Status{
			Phase:     capitalize(o.config.ActionVerb),
			Running:   nil,
			Completed: 0,
			Total:     len(allMonikers),
		})
	}

	// Print header with layer info
	fmt.Fprintf(o.orchestratorOut, "%s %d modules in %d layer(s)%s%s",
		capitalize(o.config.ActionVerb), len(allMonikers), len(layers), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(allMonikers), o.config.StatusUpdateInterval)
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
		for _, result := range layerResults {
			if result.ExitCode != 0 {
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

// formatMonikerList formats a list of monikers for display
func formatMonikerList(monikers []string) string {
	if len(monikers) <= 5 {
		return fmt.Sprintf("%v", monikers)
	}
	return fmt.Sprintf("%v... (%d total)", monikers[:5], len(monikers))
}

// Run executes all work items in parallel and returns the results
func (o *Orchestrator) Run(monikers []string) ([]WorkResult, error) {
	// Initialize if not already done
	if err := o.Init(); err != nil {
		return nil, err
	}

	// Start TUI console if enabled (and not already started)
	if o.tuiConsole != nil {
		// Initialize TUI status tracking
		o.tuiTotal = len(monikers)
		o.tuiCompleted = 0
		o.tuiRunning = nil

		// StartAsync is idempotent - will not restart if already running
		o.tuiConsole.StartAsync(o.tuiCtx)
		// Send initial status
		o.tuiConsole.UpdateStatus(tui.Status{
			Phase:     capitalize(o.config.ActionVerb),
			Running:   nil,
			Completed: 0,
			Total:     len(monikers),
		})
	}

	// Print header (use display names for cleaner output)
	displayNames := output.PackageDisplayNames(monikers)
	fmt.Fprintf(o.orchestratorOut, "%s %d items in parallel:%s%s%s",
		capitalize(o.config.ActionVerb), len(monikers), output.ListFormat(displayNames, 60, 5), LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(monikers), o.config.StatusUpdateInterval)
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

// executeParallel runs work items in parallel with controlled concurrency
func (o *Orchestrator) executeParallel(workItems []WorkItem) []WorkResult {
	results := make([]WorkResult, len(workItems))
	var wg sync.WaitGroup

	// Create semaphore to limit concurrency
	sem := make(chan struct{}, o.config.MaxConcurrency)

	for _, item := range workItems {
		wg.Add(1)

		go func(wi WorkItem) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := o.processWorkItem(wi)
			results[wi.Index] = result
		}(item)
	}

	wg.Wait()
	return results
}

// processWorkItem processes a single work item
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
		o.tuiMarkCompleted(item.Moniker)
	}

	// Create output directory for this module
	// Sanitize moniker for filesystem (replace : with _ for Windows compatibility)
	sanitizedMoniker := sanitizePathForFS(item.Moniker)
	moduleOutputDir := filepath.Join(o.config.WorkspaceRoot, o.config.OutputBaseDir, sanitizedMoniker)
	parentDir := filepath.Dir(moduleOutputDir)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
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

	if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
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

	// Create writer for worker - use TUI multiwriter if TUI is enabled
	var workerWriter io.Writer
	if o.tuiConsole != nil {
		workerWriter = o.tuiConsole.NewWriter(item.Moniker, logFile)
	} else {
		workerWriter = logFile
	}

	// Execute worker function
	exitCode := o.worker(item.Moniker, workerWriter)
	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker, o.config.LogFileName)
	result.Duration = time.Since(startTime)

	// Set module type from config if available
	if o.config.ModuleTypes != nil {
		if t, ok := o.config.ModuleTypes[item.Moniker]; ok {
			result.Type = t
		}
	}

	// Mark as completed in display (will print completion line)
	markCompleted()

	return result
}

// PrintSummary prints a summary of all results to the orchestrator output
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

// Close releases resources held by the orchestrator
func (o *Orchestrator) Close() {
	// Stop TUI if not already stopped
	o.StopTUI()
}

// tuiMarkRunning adds a module to the running list and updates TUI status
func (o *Orchestrator) tuiMarkRunning(moniker string) {
	if o.tuiConsole == nil {
		return
	}
	o.tuiMu.Lock()
	o.tuiRunning = append(o.tuiRunning, moniker)
	running := make([]string, len(o.tuiRunning))
	copy(running, o.tuiRunning)
	completed := o.tuiCompleted
	total := o.tuiTotal
	o.tuiMu.Unlock()

	o.tuiConsole.UpdateStatus(tui.Status{
		Phase:     capitalize(o.config.ActionVerb),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// tuiMarkCompleted removes a module from running and increments completed count
func (o *Orchestrator) tuiMarkCompleted(moniker string) {
	if o.tuiConsole == nil {
		return
	}
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

	o.tuiConsole.UpdateStatus(tui.Status{
		Phase:     capitalize(o.config.ActionVerb),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// SetPhase switches the TUI to a specific phase (Init, Run, End)
func (o *Orchestrator) SetPhase(phase tui.Phase) {
	if o.tuiConsole != nil {
		o.tuiConsole.SetPhase(phase)
	}
}

// SetPhaseSummary sets the summary text for a phase (shown when collapsed)
func (o *Orchestrator) SetPhaseSummary(phase tui.Phase, summary string) {
	if o.tuiConsole != nil {
		o.tuiConsole.SetPhaseSummary(phase, summary)
	}
}

// CompletePhase marks a phase as complete
func (o *Orchestrator) CompletePhase(phase tui.Phase, success bool, summary string) {
	if o.tuiConsole != nil {
		o.tuiConsole.CompletePhase(phase, success, summary)
	}
}

// WriteToPhase writes a line to a specific phase's buffer
func (o *Orchestrator) WriteToPhase(phase tui.Phase, text string) {
	if o.tuiConsole != nil {
		o.tuiConsole.WriteToPhase(phase, text)
	}
}

// SendInitLine sends a line to the Init phase buffer
func (o *Orchestrator) SendInitLine(text string) {
	o.WriteToPhase(tui.PhaseInit, text)
}

// SendEndLine sends a line to the results buffer (appears below Run pane)
func (o *Orchestrator) SendEndLine(text string) {
	if o.tuiConsole != nil {
		o.tuiConsole.WriteResult(text)
	}
}

// IsTUIEnabled returns whether TUI is enabled
func (o *Orchestrator) IsTUIEnabled() bool {
	return o.tuiConsole != nil
}

// tuiWriter implements io.Writer and forwards all writes to the TUI Init phase
type tuiWriter struct {
	orch  *Orchestrator
	phase tui.Phase
}

// Write implements io.Writer by forwarding to the appropriate TUI pane
func (w *tuiWriter) Write(p []byte) (n int, err error) {
	if w.orch.tuiConsole != nil {
		// Convert bytes to string, trim trailing newline (TUI adds its own)
		text := string(p)
		text = strings.TrimSuffix(text, "\n")
		if text != "" {
			w.orch.WriteToPhase(w.phase, text)
		}
	}
	return len(p), nil
}

// GetTUIWriter returns an io.Writer that sends output to the specified TUI phase.
// Returns nil if TUI is not enabled.
func (o *Orchestrator) GetTUIWriter(phase tui.Phase) io.Writer {
	if o.tuiConsole == nil {
		return nil
	}
	return &tuiWriter{orch: o, phase: phase}
}

// SendSummary sends summary data to activate the TUI Summary pane.
// Should be called after work completes but before StopTUI.
func (o *Orchestrator) SendSummary(data *tui.SummaryData) {
	if o.tuiConsole == nil {
		return
	}
	o.tuiConsole.SendSummary(data)
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
		return
	}
	o.tuiConsole.StartAsync(o.tuiCtx)
	// Set initial phase to Init
	o.SetPhase(tui.PhaseInit)
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

// SetModuleTypes updates the module types map in the config.
// Useful when module types are determined after orchestrator creation.
func (o *Orchestrator) SetModuleTypes(moduleTypes map[string]string) {
	o.config.ModuleTypes = moduleTypes
}

// SetMaxConcurrency updates the maximum concurrency for subsequent Run calls.
// Useful for running sequential tests after parallel tests.
func (o *Orchestrator) SetMaxConcurrency(maxConcurrency int) {
	o.config.MaxConcurrency = maxConcurrency
}

// GetExitCode returns the appropriate exit code based on results
// Returns 1 if any module failed, 0 otherwise
func GetExitCode(results []WorkResult) int {
	for _, result := range results {
		if result.ExitCode != 0 {
			return 1
		}
	}
	return 0
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
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
// Normalizes path separators to forward slashes
func sanitizePathForFS(path string) string {
	safe := strings.ReplaceAll(path, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}
