package testrunners

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TestEvent represents a single event from `go test -json` output.
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`  // "run", "output", "pass", "fail", "skip", "pause", "cont"
	Package string  `json:"Package"` // Package name
	Test    string  `json:"Test"`    // Test name (empty for package-level)
	Output  string  `json:"Output"`  // Test output line
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// TestResult holds the results of a test execution.
type TestResult struct {
	Package       string
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	TestsTotal    int
	PackageFailed bool
	Duration      time.Duration
	Events        []TestEvent
}

// StreamingRunner executes go test with real-time output parsing.
// It sends human-readable output to the TUI while capturing full JSON for reporting.
type StreamingRunner struct {
	// tuiWriter receives human-readable test output for TUI display
	tuiWriter io.Writer
	// logWriter receives full JSON output for log files
	logWriter io.Writer
	// events collects all parsed events for result calculation
	events []TestEvent
	mu     sync.Mutex
}

// NewStreamingRunner creates a new streaming test runner.
// tuiWriter: receives parsed human-readable output (for TUI display)
// logWriter: receives full JSON lines (for log files and reporting).
func NewStreamingRunner(tuiWriter, logWriter io.Writer) *StreamingRunner {
	return &StreamingRunner{
		tuiWriter: tuiWriter,
		logWriter: logWriter,
		events:    make([]TestEvent, 0),
	}
}

// Run executes the command and streams output in real-time.
// Returns the test result with pass/fail counts.
func (r *StreamingRunner) Run(cmd *exec.Cmd) (TestResult, error) {
	start := time.Now()
	result := TestResult{}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return result, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return result, fmt.Errorf("failed to start command: %w", err)
	}

	// Process stdout (JSON events) in a goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r.processJSONOutput(stdout)
	}()

	// Process stderr (error messages) in a goroutine
	go func() {
		defer wg.Done()
		r.processStderr(stderr)
	}()

	// Wait for output processing to complete
	wg.Wait()

	// Wait for command to finish
	cmdErr := cmd.Wait()

	result.Duration = time.Since(start)
	result.Events = r.events
	result.PackageFailed = cmdErr != nil

	// Calculate test counts from events
	r.mu.Lock()
	for _, event := range r.events {
		if event.Package != "" && result.Package == "" {
			result.Package = event.Package
		}
		// Only count test-level events (not package-level)
		if event.Test != "" {
			switch event.Action {
			case "pass":
				result.TestsPassed++
				result.TestsTotal++
			case "fail":
				result.TestsFailed++
				result.TestsTotal++
			case "skip":
				result.TestsSkipped++
				result.TestsTotal++
			}
		}
	}
	r.mu.Unlock()

	return result, nil
}

// processJSONOutput parses JSON lines from go test and streams human-readable output.
// ANSI stripping is handled at the orchestrator level via StrippingWriter.
func (r *StreamingRunner) processJSONOutput(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse JSON event
		var event TestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// Not valid JSON, write raw line
			// Write errors are intentionally ignored - streaming is best-effort
			if r.tuiWriter != nil {
				_, _ = r.tuiWriter.Write(line)         //nolint:errcheck // best-effort streaming
				_, _ = r.tuiWriter.Write([]byte("\n")) //nolint:errcheck // best-effort streaming
			}
			continue
		}

		// Store event for result calculation
		r.mu.Lock()
		r.events = append(r.events, event)
		r.mu.Unlock()

		// Write human-readable output
		if event.Action == "output" && event.Output != "" {
			output := event.Output
			trimmed := strings.TrimSpace(output)

			// Skip empty lines
			if trimmed == "" {
				continue
			}

			// Write errors are intentionally ignored - streaming is best-effort
			if r.tuiWriter != nil {
				_, _ = r.tuiWriter.Write([]byte(output)) //nolint:errcheck // best-effort streaming
			}
		}
	}
}

// processStderr handles stderr output (usually build errors).
// ANSI stripping is handled at the orchestrator level via StrippingWriter.
func (r *StreamingRunner) processStderr(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()

		// Write errors are intentionally ignored - streaming is best-effort
		if r.tuiWriter != nil {
			_, _ = r.tuiWriter.Write(line)         //nolint:errcheck // best-effort streaming
			_, _ = r.tuiWriter.Write([]byte("\n")) //nolint:errcheck // best-effort streaming
		}
	}
}

// CountResults returns pass/fail/skip counts from collected events.
func (r *StreamingRunner) CountResults() (passed, failed, skipped int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, event := range r.events {
		if event.Test != "" {
			switch event.Action {
			case "pass":
				passed++
			case "fail":
				failed++
			case "skip":
				skipped++
			}
		}
	}
	return passed, failed, skipped
}

// GetEvents returns all collected test events.
func (r *StreamingRunner) GetEvents() []TestEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]TestEvent{}, r.events...)
}
