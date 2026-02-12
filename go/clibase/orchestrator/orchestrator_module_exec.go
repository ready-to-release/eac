package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/output"
)

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

	// Helper to finalize result with error and mark completed.
	failWith := func(errMsg string) WorkResult {
		result.ExitCode = 1
		result.Errors = []string{errMsg}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker(item.Moniker), o.config.LogFileName())
		result.Duration = time.Since(startTime)
		o.markWorkItemCompleted(item.Moniker, &result)
		return result
	}

	// Create output directory for this module
	// Use simplified display name for cleaner paths, then sanitize for filesystem
	moduleOutputDir := filepath.Join(o.config.WorkspaceRoot, o.config.OutputBaseDir, sanitizedMoniker(item.Moniker))

	if err := os.MkdirAll(filepath.Dir(moduleOutputDir), 0o755); err != nil {
		return failWith(fmt.Sprintf("Failed to create parent directory %s: %v", filepath.Dir(moduleOutputDir), err))
	}

	// In dry-run mode, preserve existing artifacts and use separate log file
	// This allows meta-tests to mock builds without destroying real build outputs
	if !o.config.DryRun {
		// Purge existing output directory (best effort)
		_ = os.RemoveAll(moduleOutputDir)
	}

	if err := os.MkdirAll(moduleOutputDir, 0o755); err != nil {
		return failWith(fmt.Sprintf("Failed to create directory %s: %v", moduleOutputDir, err))
	}

	// Create log file (use separate file in dry-run mode to preserve real build logs)
	logFileName := o.config.LogFileName()
	if o.config.DryRun {
		logFileName = "dry-run." + logFileName
	}
	logPath := filepath.Join(moduleOutputDir, logFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		return failWith(fmt.Sprintf("Failed to create log file %s: %v", logPath, err))
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
	result.LogPath = filepath.Join(o.config.OutputBaseDir, sanitizedMoniker(item.Moniker), o.config.LogFileName())
	result.Duration = time.Since(startTime)

	// Set component type from config if available
	if o.config.ComponentTypesDisplay != nil {
		if t, ok := o.config.ComponentTypesDisplay[item.Moniker]; ok {
			result.Type = t
		}
	}

	// Mark as completed in display (will print completion line)
	o.markWorkItemCompleted(item.Moniker, &result)

	return result
}

// markWorkItemCompleted updates display manager and TUI for a completed work item.
func (o *Orchestrator) markWorkItemCompleted(moniker string, result *WorkResult) {
	if o.display != nil {
		o.display.markCompleted(result)
	}
	o.tuiMarkCompleted(moniker, result.ExitCode)
}

// sanitizedMoniker returns a filesystem-safe display name for a moniker.
func sanitizedMoniker(moniker string) string {
	return sanitizePathForFS(output.PackageDisplayName(moniker))
}
