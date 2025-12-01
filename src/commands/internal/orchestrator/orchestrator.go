package orchestrator

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/output"
)

// Orchestrator manages parallel execution of work items
type Orchestrator struct {
	config          Config
	worker          WorkerFunc
	display         *displayManager
	orchestratorOut io.Writer
	logger          *log.Logger // goroutine-safe logger
	logFile         *os.File
}

// New creates a new Orchestrator with the given configuration and worker function
func New(config Config, worker WorkerFunc) *Orchestrator {
	// Set default max concurrency to number of CPUs
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = runtime.NumCPU()
	}

	return &Orchestrator{
		config: config,
		worker: worker,
	}
}

// Run executes all work items in parallel and returns the results
func (o *Orchestrator) Run(monikers []string) ([]WorkResult, error) {
	// Create orchestrator log file
	orchestratorLogPath := filepath.Join(o.config.WorkspaceRoot, o.config.OutputBaseDir, o.config.OrchestratorLogName)
	if err := os.MkdirAll(filepath.Dir(orchestratorLogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create orchestrator log directory: %w", err)
	}

	orchestratorLog, err := os.Create(orchestratorLogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator log: %w", err)
	}
	o.logFile = orchestratorLog

	// Create a goroutine-safe logger that writes to both console and log file
	// log.Logger is documented as safe for concurrent use from multiple goroutines
	multiWriter := io.MultiWriter(os.Stdout, orchestratorLog)
	o.logger = log.New(multiWriter, "", 0) // No prefix, no flags (we format our own output)
	o.orchestratorOut = multiWriter

	// Print header
	fmt.Fprintf(o.orchestratorOut, "%s %d modules in parallel:%s%s%s",
		capitalize(o.config.ActionVerb), len(monikers), output.ListFormat(monikers, 60, 5), LineEnding, LineEnding)

	// Create and start display manager
	o.display = newDisplayManager(o.logger, o.config.ActionVerb, len(monikers), o.config.StatusUpdateInterval)
	o.display.start()

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
	o.display.stop()
	o.display.flushCompletedLines()

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

	// Create output directory for this module
	moduleOutputDir := filepath.Join(o.config.WorkspaceRoot, o.config.OutputBaseDir, item.Moniker)
	parentDir := filepath.Dir(moduleOutputDir)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create parent directory %s: %v", parentDir, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, item.Moniker, o.config.LogFileName)
		result.Duration = time.Since(startTime)
		o.display.markCompleted(&result)
		return result
	}

	// Purge existing output directory (best effort)
	_ = os.RemoveAll(moduleOutputDir)

	if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create directory %s: %v", moduleOutputDir, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, item.Moniker, o.config.LogFileName)
		result.Duration = time.Since(startTime)
		o.display.markCompleted(&result)
		return result
	}

	// Create log file
	logPath := filepath.Join(moduleOutputDir, o.config.LogFileName)
	logFile, err := os.Create(logPath)
	if err != nil {
		result.ExitCode = 1
		result.Errors = []string{fmt.Sprintf("Failed to create log file %s: %v", logPath, err)}
		result.LogPath = filepath.Join(o.config.OutputBaseDir, item.Moniker, o.config.LogFileName)
		result.Duration = time.Since(startTime)
		o.display.markCompleted(&result)
		return result
	}

	// Mark as running in display
	o.display.markRunning(item.Moniker)

	// Execute worker function (all output goes to log file)
	exitCode := o.worker(item.Moniker, logFile)
	logFile.Close()

	// Parse log for warnings/errors
	warnings, errors := parseLogForIssues(logPath)

	result.ExitCode = exitCode
	result.Warnings = warnings
	result.Errors = errors
	result.LogPath = filepath.Join(o.config.OutputBaseDir, item.Moniker, o.config.LogFileName)
	result.Duration = time.Since(startTime)

	// Set module type from config if available
	if o.config.ModuleTypes != nil {
		if t, ok := o.config.ModuleTypes[item.Moniker]; ok {
			result.Type = t
		}
	}

	// Mark as completed in display (will print completion line)
	o.display.markCompleted(&result)

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
						if len(errMsg) > 80 {
							errMsg = errMsg[:77] + "..."
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

// Close releases resources held by the orchestrator
// Must be called after PrintSummary
func (o *Orchestrator) Close() {
	if o.logFile != nil {
		o.logFile.Close()
	}
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
