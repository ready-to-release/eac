package testrunners

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

	// Create pipes manually instead of cmd.StdoutPipe/StderrPipe.
	// StdoutPipe registers the read end in cmd's closeAfterWait list, so
	// cmd.Wait() closes it immediately after the process exits — before
	// reader goroutines can drain remaining buffered data. By creating our
	// own pipes, cmd.Wait() leaves our read ends open, and readers can
	// finish consuming all output before we close them.
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return result, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		stdoutR.Close()
		stdoutW.Close()
		return result, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	// Start the command
	if err := cmd.Start(); err != nil {
		stdoutR.Close()
		stdoutW.Close()
		stderrR.Close()
		stderrW.Close()
		return result, fmt.Errorf("failed to start command: %w", err)
	}

	// Close write ends in the parent process. The child has its own copies
	// from handle inheritance. When the child exits, its copies close too,
	// giving readers EOF — unless a grandchild also inherited them.
	stdoutW.Close()
	stderrW.Close()

	// Process stdout (JSON events) in a goroutine
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r.processJSONOutput(stdoutR)
	}()

	// Process stderr (error messages) in a goroutine
	go func() {
		defer wg.Done()
		r.processStderr(stderrR)
	}()

	// Wait for the child process to exit. Since we created our own pipes
	// (not StdoutPipe/StderrPipe), cmd.Wait() does NOT close our read
	// ends, so readers can keep draining buffered data.
	cmdErr := cmd.Wait()

	// Wait for reader goroutines to finish. Normally they get EOF once all
	// write-end copies are closed (child exited, no grandchild holds them).
	// On Windows, grandchild processes (e.g., PowerShell spawned by tests)
	// can inherit pipe handles, keeping the write end open indefinitely.
	// Use a timeout to prevent an indefinite hang in that case.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Readers finished normally
	case <-time.After(10 * time.Second):
		// Force-close read ends to unblock stuck readers
		stdoutR.Close()
		stderrR.Close()
		<-done
	}

	// Clean up read ends (no-op if already closed by timeout path)
	stdoutR.Close()
	stderrR.Close()

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
