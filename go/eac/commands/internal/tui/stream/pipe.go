// Package stream provides output streaming utilities for the TUI console.
package stream

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console"
)

// OutputPipe captures subprocess output and streams it to a channel.
// Implements io.Writer for use with exec.Command.
type OutputPipe struct {
	lineChan chan<- console.Line
	source   string // Module moniker
	filter   *Filter

	reader *io.PipeReader
	writer *io.PipeWriter

	wg     sync.WaitGroup
	closed bool
	mu     sync.Mutex
}

// NewOutputPipe creates a new pipe that streams lines to the channel.
func NewOutputPipe(lineChan chan<- console.Line, source string) *OutputPipe {
	r, w := io.Pipe()
	op := &OutputPipe{
		lineChan: lineChan,
		source:   source,
		filter:   NewFilter(),
		reader:   r,
		writer:   w,
	}

	// Start goroutine to read lines
	op.wg.Add(1)
	go op.readLines()

	return op
}

// Writer returns an io.Writer for subprocess stdout/stderr.
func (op *OutputPipe) Writer() io.Writer {
	return op.writer
}

// Close closes the pipe and waits for the reader goroutine to finish.
func (op *OutputPipe) Close() error {
	op.mu.Lock()
	if op.closed {
		op.mu.Unlock()
		return nil
	}
	op.closed = true
	op.mu.Unlock()

	// Close writer first - this causes scanner.Scan() to return EOF
	err := op.writer.Close()

	// Wait for readLines() to exit (should be quick after writer closes)
	// Use a timeout to avoid hanging forever if something goes wrong
	done := make(chan struct{})
	go func() {
		op.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Normal exit
	case <-time.After(2 * time.Second):
		// Timeout - readLines() is stuck, but we can't block forever
		// This is a fallback safety; ideally shouldn't happen
	}

	return err
}

func (op *OutputPipe) readLines() {
	defer op.wg.Done()
	defer op.reader.Close()

	scanner := bufio.NewScanner(op.reader)

	// Handle long lines (Go test output can be verbose)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		text := scanner.Text()

		// Apply filter
		if !op.filter.ShouldShow(text) {
			continue
		}

		// Check if closed before sending (protects against send on closed channel)
		op.mu.Lock()
		closed := op.closed
		op.mu.Unlock()
		if closed {
			return
		}

		level := classifyLine(text)

		select {
		case op.lineChan <- console.Line{
			Text:      text,
			Source:    op.source,
			Level:     level,
			Timestamp: time.Now(),
		}:
		default:
			// Drop if channel full (non-blocking)
		}
	}
}

// classifyLine determines the severity level of an output line.
// Uses precise patterns to avoid false positives from summary lines like "0 errors".
func classifyLine(text string) console.Level {
	lowerText := strings.ToLower(text)

	// Skip classification for summary/stats lines that mention counts
	// These often contain words like "failed" or "error" in non-error contexts
	if isSummaryLine(lowerText) {
		return console.LevelInfo
	}

	// Check for actual Go test failures (--- FAIL: TestName)
	if strings.HasPrefix(text, "--- FAIL:") {
		return console.LevelError
	}

	// Check for package-level FAIL (FAIL<tab>package/path)
	// This is the final verdict for a package
	if strings.HasPrefix(text, "FAIL\t") {
		return console.LevelError
	}

	// Specific error indicators (more precise than substring matching)
	// Note: "error:" alone causes false positives with test names like "TestFoo/error:_bar"
	// so we only match "error:" when it appears at start of line or after whitespace/colon
	errorIndicators := []string{
		"panic:", "fatal error:", "compilation failed",
		"undefined:", "cannot find package", "build failed",
	}
	for _, indicator := range errorIndicators {
		if strings.Contains(lowerText, indicator) {
			return console.LevelError
		}
	}

	// Check for "error:" more carefully - it must appear at a word boundary
	// to avoid false positives in test names like "TestIsSummaryLine/error:_something"
	// Returns LevelWarn for test output containing errors (orange), LevelError for real errors (red)
	if level := classifyErrorIndicator(lowerText); level != console.LevelInfo {
		return level
	}

	// Check for SKIP (Go test output: --- SKIP:)
	if strings.HasPrefix(text, "--- SKIP:") {
		return console.LevelWarn
	}

	// Warnings - be more specific
	if strings.Contains(lowerText, "warning:") || strings.Contains(lowerText, "[warning]") {
		return console.LevelWarn
	}
	if strings.Contains(lowerText, "deprecated") {
		return console.LevelWarn
	}

	return console.LevelInfo
}

// classifyErrorIndicator checks if the line contains an error indicator
// and returns the appropriate level. Returns LevelInfo if no error indicator found.
// Test output containing "Error:" messages gets LevelWarn (orange) to distinguish
// from actual test failures which get LevelError (red).
func classifyErrorIndicator(lowerText string) console.Level {
	// Skip lines that are clearly test output markers (--- PASS/FAIL/SKIP, === RUN)
	if strings.Contains(lowerText, "--- pass:") ||
		strings.Contains(lowerText, "--- fail:") ||
		strings.Contains(lowerText, "--- skip:") ||
		strings.Contains(lowerText, "=== run") {
		return console.LevelInfo
	}

	// Check if this is test output (contains _test.go: filename pattern)
	// These are log messages from tests, not actual failures
	isTestOutput := strings.Contains(lowerText, "_test.go:")

	// Match "error:" at start of line (after optional whitespace)
	trimmed := strings.TrimSpace(lowerText)
	if strings.HasPrefix(trimmed, "error:") {
		if isTestOutput {
			return console.LevelWarn // Orange for test output containing errors
		}
		return console.LevelError
	}

	// Match ": error:" (typical Go compiler output like "main.go:10: error:")
	if strings.Contains(lowerText, ": error:") {
		if isTestOutput {
			return console.LevelWarn // Orange for test output containing errors
		}
		return console.LevelError
	}

	return console.LevelInfo
}

// isSummaryLine detects summary/statistics lines that should not be colored as errors.
// Examples: "Tests: 10 passed, 0 failed", "0 errors", "Build completed".
func isSummaryLine(lowerText string) bool {
	summaryPatterns := []string{
		"passed", "completed", "succeeded",
		"tests:", "packages:", "total:",
		"0 error", "0 failed", "0 failure",
		"no error", "no failure",
		"=== test", "=== run", // Go test markers
	}
	for _, pattern := range summaryPatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}
	return false
}

// MultiWriter creates an io.Writer that writes to multiple destinations.
// This is similar to io.MultiWriter but includes the OutputPipe.
type MultiWriter struct {
	writers []io.Writer
	pipe    *OutputPipe
}

// NewMultiWriter creates a MultiWriter that includes an OutputPipe.
func NewMultiWriter(lineChan chan<- console.Line, source string, writers ...io.Writer) *MultiWriter {
	pipe := NewOutputPipe(lineChan, source)
	allWriters := append([]io.Writer{pipe.Writer()}, writers...)
	return &MultiWriter{
		writers: allWriters,
		pipe:    pipe,
	}
}

// Write writes to all underlying writers.
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return n, err
		}
	}
	return len(p), nil
}

// Close closes the output pipe.
func (mw *MultiWriter) Close() error {
	return mw.pipe.Close()
}
