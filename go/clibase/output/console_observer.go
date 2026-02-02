package output

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

// ConsoleObserver implements ExecutionObserver for plain-text console output.
// Used when TUI is disabled (--no-tui or non-interactive terminal).
type ConsoleObserver struct {
	out io.Writer
	mu  sync.Mutex

	// State for progress display
	running   []string
	completed int
	total     int
}

// NewConsoleObserver creates a new console observer writing to stdout.
func NewConsoleObserver() *ConsoleObserver {
	return &ConsoleObserver{out: os.Stdout}
}

// NewConsoleObserverWithWriter creates a console observer with custom output.
func NewConsoleObserverWithWriter(out io.Writer) *ConsoleObserver {
	return &ConsoleObserver{out: out}
}

// OnEvent handles execution events and prints to console.
func (o *ConsoleObserver) OnEvent(event interfaces.ExecutionEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()

	switch e := event.(type) {
	case interfaces.PhaseStartedEvent:
		fmt.Fprintf(o.out, "\n=== %s ===\n", e.Phase)
	case interfaces.PhaseCompletedEvent:
		status := "OK"
		if !e.Success {
			status = "FAILED"
		}
		fmt.Fprintf(o.out, "=== %s %s (%v) ===\n", e.Phase, status, e.Duration.Round(time.Millisecond))
	case interfaces.UnitQueuedEvent:
		// Silent - just track for total count
		o.total++
	case interfaces.UnitStartedEvent:
		o.running = append(o.running, e.ID)
		fmt.Fprintf(o.out, "  [START] %s\n", e.ID)
	case interfaces.UnitCompletedEvent:
		// Remove from running
		for i, id := range o.running {
			if id == e.ID {
				o.running = append(o.running[:i], o.running[i+1:]...)
				break
			}
		}
		o.completed++

		status := "OK"
		if e.ExitCode > 0 {
			status = "FAIL"
		} else if e.ExitCode < 0 {
			status = "CACHED"
		}
		fmt.Fprintf(o.out, "  [%s] %s (%v)\n", status, e.ID, e.Duration.Round(time.Millisecond))
	case interfaces.ProgressUpdateEvent:
		o.total = e.Total
		// Optionally print periodic progress (currently silent)
	case interfaces.SummaryReadyEvent:
		fmt.Fprintf(o.out, "\n=== Summary ===\n")
		fmt.Fprintf(o.out, "Total time: %v\n", e.TotalTime.Round(time.Millisecond))
		fmt.Fprintf(o.out, "Passed: %d, Failed: %d, Cached: %d\n", e.Passed, e.Failed, e.Cached)
		if e.NextSteps != "" {
			fmt.Fprintf(o.out, "\nNext: %s\n", e.NextSteps)
		}
	}
}

// NewWriter returns a writer that only writes to the log file.
// Console observer does not intercept output lines for display.
func (o *ConsoleObserver) NewWriter(unitID string, logWriter io.Writer) io.WriteCloser {
	return &nopCloser{logWriter}
}

type nopCloser struct {
	io.Writer
}

func (n *nopCloser) Close() error { return nil }
