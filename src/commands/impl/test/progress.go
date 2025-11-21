package test

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Global progress tracker instance
var globalTracker *ProgressTracker
var trackerMu sync.Mutex

// ProgressTracker tracks the status of all running tests and displays aggregate progress
type ProgressTracker struct {
	mu          sync.Mutex
	out         io.Writer
	startTime   time.Time
	runningNames map[string]bool // Track names of running tests
	completed   int
	total       int // Total number of packages/tests
	ticker      *time.Ticker
	done        chan bool
}

// StartGlobalTracker initializes and starts the global progress tracker
func StartGlobalTracker(out io.Writer, total int) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	if globalTracker != nil {
		return // Already started
	}

	globalTracker = &ProgressTracker{
		out:          out,
		startTime:    time.Now(),
		runningNames: make(map[string]bool),
		completed:    0,
		total:        total,
		done:         make(chan bool),
	}

	globalTracker.ticker = time.NewTicker(10 * time.Second)
	go globalTracker.displayLoop()
}

// StopGlobalTracker stops the global progress tracker
func StopGlobalTracker() {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	if globalTracker == nil {
		return
	}

	if globalTracker.ticker != nil {
		globalTracker.ticker.Stop()
	}
	if globalTracker.done != nil {
		globalTracker.done <- true
		close(globalTracker.done)
	}
	globalTracker = nil
}

// TrackTestStart adds a test to the running set
func TrackTestStart(name string) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	if globalTracker != nil {
		globalTracker.mu.Lock()
		globalTracker.runningNames[name] = true
		globalTracker.mu.Unlock()
	}
}

// TrackTestComplete removes a test from running and increments completed
func TrackTestComplete(name string) {
	trackerMu.Lock()
	defer trackerMu.Unlock()

	if globalTracker != nil {
		globalTracker.mu.Lock()
		delete(globalTracker.runningNames, name)
		globalTracker.completed++
		globalTracker.mu.Unlock()
	}
}

// displayLoop shows progress updates every 10 seconds
func (pt *ProgressTracker) displayLoop() {
	for {
		select {
		case <-pt.done:
			return
		case <-pt.ticker.C:
			pt.displayStatus()
		}
	}
}

// displayStatus shows the current aggregate status
func (pt *ProgressTracker) displayStatus() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	elapsed := time.Since(pt.startTime)
	runningCount := len(pt.runningNames)

	// Extract names into sorted slice
	names := make([]string, 0, len(pt.runningNames))
	for name := range pt.runningNames {
		names = append(names, name)
	}

	// Format output with running test names if any
	if runningCount > 0 {
		// Sort names for consistent output
		sortedNames := make([]string, len(names))
		copy(sortedNames, names)
		// Simple alphabetical sort
		for i := 0; i < len(sortedNames)-1; i++ {
			for j := i + 1; j < len(sortedNames); j++ {
				if sortedNames[i] > sortedNames[j] {
					sortedNames[i], sortedNames[j] = sortedNames[j], sortedNames[i]
				}
			}
		}

		// Join names with comma separator
		nameList := ""
		for i, name := range sortedNames {
			if i > 0 {
				nameList += ", "
			}
			nameList += name
		}

		fmt.Fprintf(pt.out, "Status: %s elapsed, %d/%d completed. %d running (%s)\n",
			formatDuration(elapsed), pt.completed, pt.total, runningCount, nameList)
	} else {
		fmt.Fprintf(pt.out, "Status: %s elapsed, %d/%d completed. %d running\n",
			formatDuration(elapsed), pt.completed, pt.total, runningCount)
	}
	os.Stdout.Sync()
}
