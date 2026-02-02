package orchestrator

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/clibase/output"
)

// displayManager handles all console output in a single goroutine
// This prevents interleaved output from multiple goroutines.
type displayManager struct {
	mu               sync.Mutex
	logger           *log.Logger // goroutine-safe logger
	actionVerb       string
	startTime        time.Time
	running          map[string]bool // monikers currently running
	completed        int
	failed           int // count of failures
	total            int
	updateInterval   time.Duration
	completionChan   chan *WorkResult
	statusTicker     *time.Ticker
	done             chan bool
	completedResults []*WorkResult         // collected results for summary generation
	tuiMode          bool                  // when true, skip running list (TUI tabs show it)
	registry         *locktracker.Registry // lock tracker registry for status display
}

// newDisplayManager creates a new display manager.
// If registry is non-nil, lock info will be included in status updates.
func newDisplayManager(logger *log.Logger, actionVerb string, total, updateIntervalMs int, tuiMode bool, registry *locktracker.Registry) *displayManager {
	if updateIntervalMs <= 0 {
		updateIntervalMs = 500 // default to 500ms for responsive feedback
	}

	return &displayManager{
		logger:         logger,
		actionVerb:     actionVerb,
		startTime:      time.Now(),
		running:        make(map[string]bool),
		completed:      0,
		total:          total,
		updateInterval: time.Duration(updateIntervalMs) * time.Millisecond,
		completionChan: make(chan *WorkResult, 100),
		done:           make(chan bool),
		tuiMode:        tuiMode,
		registry:       registry,
	}
}

// start begins the display loop in a separate goroutine.
func (dm *displayManager) start() {
	dm.statusTicker = time.NewTicker(dm.updateInterval)
	go dm.displayLoop()
}

// stop terminates the display loop.
func (dm *displayManager) stop() {
	if dm.statusTicker != nil {
		dm.statusTicker.Stop()
	}
	dm.done <- true
	close(dm.done)
	close(dm.completionChan)
}

// typeSummary holds aggregated stats for a test type.
type typeSummary struct {
	Type     string
	Packages int
	Passed   int
	Failed   int
	Warnings int
}

// flushCompletedLines prints summary table for successes and individual lines for failures/warnings
// Should be called after stop() to ensure all completions are processed.
func (dm *displayManager) flushCompletedLines() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Separate successful results from failed/warning results
	var failedOrWarnResults []*WorkResult
	summaryByType := make(map[string]*typeSummary)

	for _, result := range dm.completedResults {
		typeStr := result.Type
		if typeStr == "" {
			typeStr = "-"
		}

		// Initialize summary for this type if not exists
		if _, ok := summaryByType[typeStr]; !ok {
			summaryByType[typeStr] = &typeSummary{Type: typeStr}
		}
		summaryByType[typeStr].Packages++

		if result.ExitCode != 0 {
			summaryByType[typeStr].Failed++
			failedOrWarnResults = append(failedOrWarnResults, result)
		} else if len(result.Warnings) > 0 {
			summaryByType[typeStr].Warnings++
			failedOrWarnResults = append(failedOrWarnResults, result)
		} else {
			summaryByType[typeStr].Passed++
		}
	}

	// Print summary table if there are any results
	if len(summaryByType) > 0 {
		dm.printSummaryTable(summaryByType)
	}

	// Print individual lines for failed/warning results with error context
	for _, result := range failedOrWarnResults {
		dm.printResultWithErrors(result)
	}
}

// printSummaryTable prints a markdown-style table grouped by test type.
func (dm *displayManager) printSummaryTable(summaryByType map[string]*typeSummary) {
	// Collect and sort types for consistent output
	types := make([]string, 0, len(summaryByType))
	for t := range summaryByType {
		types = append(types, t)
	}
	sort.Strings(types)

	// Print table header
	dm.logger.Printf("| %-12s | %8s | %6s | %6s | %8s |%s", "Type", "Packages", "Passed", "Failed", "Warnings", LineEndingPrefix)
	dm.logger.Printf("|%s|%s|%s|%s|%s|%s", strings.Repeat("-", 14), strings.Repeat("-", 10), strings.Repeat("-", 8), strings.Repeat("-", 8), strings.Repeat("-", 10), LineEndingPrefix)

	// Print rows
	for _, t := range types {
		s := summaryByType[t]
		dm.logger.Printf("| %-12s | %8d | %6d | %6d | %8d |%s", s.Type, s.Packages, s.Passed, s.Failed, s.Warnings, LineEndingPrefix)
	}
	dm.logger.Printf("%s", LineEndingPrefix)
}

// printResultWithErrors prints a failed/warning result with its error lines.
func (dm *displayManager) printResultWithErrors(result *WorkResult) {
	displayName := output.PackageDisplayName(result.Moniker)
	displayName = strings.ReplaceAll(displayName, "\\", "/")

	var icon string
	var suffix string

	if result.ExitCode != 0 {
		icon = output.IconFail
		if len(result.Errors) > 0 {
			suffix = fmt.Sprintf("(%d errors)", len(result.Errors))
		}
	} else if len(result.Warnings) > 0 {
		icon = output.IconWarn
		suffix = fmt.Sprintf("(%d warnings)", len(result.Warnings))
	}

	typeStr := result.Type
	if typeStr == "" {
		typeStr = "-"
	}

	// Print the status line
	timing := fmt.Sprintf("%5.1fs", result.Duration.Seconds())
	baseLine := output.ResultLineNoTimeWithSuffix(icon, displayName, typeStr, "", suffix)
	statusLine := fmt.Sprintf("%s %s %s", icon, timing, baseLine[len(icon)+1:])
	dm.logger.Printf("%s%s", statusLine, LineEndingPrefix)

	// Print error/warning details (up to 5 lines)
	var details []string
	if result.ExitCode != 0 && len(result.Errors) > 0 {
		details = result.Errors
	} else if len(result.Warnings) > 0 {
		details = result.Warnings
	}

	for _, msg := range details {
		if len(msg) > 120 {
			msg = msg[:117] + "..."
		}
		dm.logger.Printf("    %s%s", msg, LineEndingPrefix)
	}
}

// markRunning marks a moniker as currently running.
func (dm *displayManager) markRunning(moniker string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.running[moniker] = true
}

// markCompleted marks a moniker as completed and queues it for display.
func (dm *displayManager) markCompleted(result *WorkResult) {
	dm.completionChan <- result
}

// displayLoop is the main display goroutine - only this goroutine writes to output.
func (dm *displayManager) displayLoop() {
	for {
		select {
		case <-dm.done:
			return
		case <-dm.statusTicker.C:
			dm.displayStatus()
		case result := <-dm.completionChan:
			dm.handleCompletion(result)
		}
	}
}

// handleCompletion processes a completed work item and stores it for summary.
func (dm *displayManager) handleCompletion(result *WorkResult) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	delete(dm.running, result.Moniker)
	dm.completed++

	if result.ExitCode != 0 {
		dm.failed++
	}

	// Store result for later summary generation
	dm.completedResults = append(dm.completedResults, result)
}

// displayStatus shows periodic status updates.
func (dm *displayManager) displayStatus() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if len(dm.running) == 0 {
		return // Don't print status if nothing is running
	}

	elapsed := time.Since(dm.startTime)
	runningCount := len(dm.running)

	// Note: On Windows, log.Printf only adds \n, but Windows console needs \r\n
	// LineEndingPrefix adds \r on Windows so the final output is \r\n (log adds \n automatically)
	failedStr := ""
	if dm.failed > 0 {
		failedStr = fmt.Sprintf(" (%d failed)", dm.failed)
	}

	// Get lock info if registry is available
	lockStr := ""
	if dm.registry != nil {
		lockStr = formatLockInfo(dm.registry)
		if lockStr != "" {
			lockStr = " | " + lockStr
		}
	}

	// In TUI mode, skip the running list - TUI tabs already show running modules
	if dm.tuiMode {
		dm.logger.Printf("Status: %s elapsed, %d/%d completed%s. %d running%s%s",
			formatDuration(elapsed), dm.completed, dm.total, failedStr, runningCount, lockStr, LineEndingPrefix)
		return
	}

	// Non-TUI mode: include running module names for visibility
	names := make([]string, 0, len(dm.running))
	for name := range dm.running {
		displayName := output.PackageDisplayName(name)
		names = append(names, strings.ReplaceAll(displayName, "\\", "/"))
	}
	sort.Strings(names)

	nameList := ""
	for i, name := range names {
		if i > 0 {
			nameList += ", "
		}
		nameList += name
	}

	dm.logger.Printf("Status: %s elapsed, %d/%d completed%s. %d running (%s)%s%s",
		formatDuration(elapsed), dm.completed, dm.total, failedStr, runningCount, nameList, lockStr, LineEndingPrefix)
}

// formatLockInfo returns a compact string describing lock status.
// Returns empty string if no locks or registry is nil.
func formatLockInfo(registry *locktracker.Registry) string {
	if registry == nil {
		return ""
	}

	summary := registry.Summary()
	if summary.Total == 0 {
		return ""
	}

	var parts []string

	// Show semaphore/weighted usage (combined capacity)
	if summary.TotalCapacity > 0 {
		parts = append(parts, fmt.Sprintf("slots:%d/%d",
			summary.TotalUsed, summary.TotalCapacity))
	}

	// Show waiting if any
	if summary.TotalWaiting > 0 {
		parts = append(parts, fmt.Sprintf("wait:%d", summary.TotalWaiting))
	}

	// Show file lock count
	fileLocks := summary.ByType[locktracker.LockTypeFileLock]
	if fileLocks > 0 {
		parts = append(parts, fmt.Sprintf("locks:%d", fileLocks))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// formatDuration formats a duration as "1m 23s" or "45s".
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
