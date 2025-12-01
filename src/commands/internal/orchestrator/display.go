package orchestrator

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/output"
)

// displayManager handles all console output in a single goroutine
// This prevents interleaved output from multiple goroutines
type displayManager struct {
	mu              sync.Mutex
	logger          *log.Logger // goroutine-safe logger
	actionVerb      string
	startTime       time.Time
	running         map[string]bool // monikers currently running
	completed       int
	total           int
	updateInterval  time.Duration
	completionChan  chan *WorkResult
	statusTicker    *time.Ticker
	done            chan bool
	completedLines  []string // collected completion lines for batch output
}

// newDisplayManager creates a new display manager
func newDisplayManager(logger *log.Logger, actionVerb string, total int, updateIntervalSec int) *displayManager {
	if updateIntervalSec <= 0 {
		updateIntervalSec = 2 // default to 2 seconds
	}

	return &displayManager{
		logger:         logger,
		actionVerb:     actionVerb,
		startTime:      time.Now(),
		running:        make(map[string]bool),
		completed:      0,
		total:          total,
		updateInterval: time.Duration(updateIntervalSec) * time.Second,
		completionChan: make(chan *WorkResult, 100),
		done:           make(chan bool),
	}
}

// start begins the display loop in a separate goroutine
func (dm *displayManager) start() {
	dm.statusTicker = time.NewTicker(dm.updateInterval)
	go dm.displayLoop()
}

// stop terminates the display loop
func (dm *displayManager) stop() {
	if dm.statusTicker != nil {
		dm.statusTicker.Stop()
	}
	dm.done <- true
	close(dm.done)
	close(dm.completionChan)
}

// flushCompletedLines prints all collected completion lines
// Should be called after stop() to ensure all completions are processed
func (dm *displayManager) flushCompletedLines() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, line := range dm.completedLines {
		dm.logger.Print(line)
	}
}

// markRunning marks a moniker as currently running
func (dm *displayManager) markRunning(moniker string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.running[moniker] = true
}

// markCompleted marks a moniker as completed and queues it for display
func (dm *displayManager) markCompleted(result *WorkResult) {
	dm.completionChan <- result
}

// displayLoop is the main display goroutine - only this goroutine writes to output
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

// handleCompletion processes a completed work item and prints its status
func (dm *displayManager) handleCompletion(result *WorkResult) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	delete(dm.running, result.Moniker)
	dm.completed++

	// Format completion line with timing after the icon
	// Format: "✅ 6.2s  module-name                    type     -"
	var icon string
	var suffix string

	if result.ExitCode != 0 {
		icon = output.IconFail
	} else if len(result.Warnings) > 0 {
		icon = output.IconWarn
		suffix = fmt.Sprintf("(%d warnings)", len(result.Warnings))
	} else {
		icon = output.IconPass
	}

	// Use Type from result, fallback to "-" if not set
	typeStr := result.Type
	if typeStr == "" {
		typeStr = "-"
	}

	// Format: icon + timing + name + type + suffix (no result column for builds)
	timing := fmt.Sprintf("%5.1fs", result.Duration.Seconds())
	baseLine := output.ResultLineNoTimeWithSuffix(icon, result.Moniker, typeStr, "", suffix)
	// Insert timing after the icon (icon is first 2-3 chars including space)
	statusLine := fmt.Sprintf("%s %s %s", icon, timing, baseLine[len(icon)+1:]) + LineEndingPrefix

	// Store for later batch output instead of printing immediately
	dm.completedLines = append(dm.completedLines, statusLine)
}

// displayStatus shows periodic status updates
func (dm *displayManager) displayStatus() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if len(dm.running) == 0 {
		return // Don't print status if nothing is running
	}

	elapsed := time.Since(dm.startTime)
	runningCount := len(dm.running)

	// Get sorted list of running monikers
	names := make([]string, 0, len(dm.running))
	for name := range dm.running {
		names = append(names, name)
	}
	sort.Strings(names)

	// Format status message
	nameList := ""
	for i, name := range names {
		if i > 0 {
			nameList += ", "
		}
		nameList += name
	}

	// Note: On Windows, log.Printf only adds \n, but Windows console needs \r\n
	// LineEndingPrefix adds \r on Windows so the final output is \r\n (log adds \n automatically)
	dm.logger.Printf("Status: %s elapsed, %d/%d completed. %d running (%s)%s",
		formatDuration(elapsed), dm.completed, dm.total, runningCount, nameList, LineEndingPrefix)
}

// formatDuration formats a duration as "1m 23s" or "45s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
