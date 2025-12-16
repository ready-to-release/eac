//go:build L1 && ov
// +build L1,ov

package orchestrator

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

func TestOrchestrator_Basic(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Configure orchestrator
	config := Config{
		WorkspaceRoot:        tmpDir,
		OutputBaseDir:        paths.OutTestRelPath,
		LogFileName:          "test.log",
		ActionVerb:           "processing",
		MaxConcurrency:       2,
		StatusUpdateInterval: 100, // 100ms for fast test feedback
	}

	// Track which modules were processed
	var mu sync.Mutex
	processed := make(map[string]bool)

	// Create worker function
	worker := func(moniker string, logWriter io.Writer) int {
		mu.Lock()
		processed[moniker] = true
		mu.Unlock()

		fmt.Fprintf(logWriter, "Processing %s\n", moniker)
		time.Sleep(100 * time.Millisecond) // Simulate work
		return 0
	}

	// Run orchestrator
	monikers := []string{"module1", "module2", "module3"}
	orch := New(config, worker)
	results, err := orch.Run(monikers)
	orch.Close()

	// Verify results
	if err != nil {
		t.Fatalf("Orchestrator.Run() error = %v", err)
	}

	if len(results) != len(monikers) {
		t.Errorf("Expected %d results, got %d", len(monikers), len(results))
	}

	for _, result := range results {
		if result.ExitCode != 0 {
			t.Errorf("Module %s failed with exit code %d", result.Moniker, result.ExitCode)
		}
		if !processed[result.Moniker] {
			t.Errorf("Module %s was not processed", result.Moniker)
		}
	}
}

func TestOrchestrator_WithFailures(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Configure orchestrator
	config := Config{
		WorkspaceRoot:        tmpDir,
		OutputBaseDir:        paths.OutTestRelPath,
		LogFileName:          "test.log",
		ActionVerb:           "processing",
		MaxConcurrency:       2,
		StatusUpdateInterval: 100, // 100ms for fast test feedback
	}

	// Create worker function that fails for specific modules
	worker := func(moniker string, logWriter io.Writer) int {
		if moniker == "fail-me" {
			fmt.Fprintf(logWriter, "Error: intentional failure\n")
			return 1
		}
		fmt.Fprintf(logWriter, "Success: %s\n", moniker)
		return 0
	}

	// Run orchestrator
	monikers := []string{"module1", "fail-me", "module2"}
	orch := New(config, worker)
	results, err := orch.Run(monikers)
	orch.Close() // Must close to release file handles

	// Verify results
	if err != nil {
		t.Fatalf("Orchestrator.Run() error = %v", err)
	}

	// Check that GetExitCode returns 1 when there are failures
	exitCode := GetExitCode(results)
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	// Verify individual results
	failureCount := 0
	for _, result := range results {
		if result.Moniker == "fail-me" {
			if result.ExitCode != 1 {
				t.Errorf("Expected fail-me to have exit code 1, got %d", result.ExitCode)
			}
			failureCount++
		} else {
			if result.ExitCode != 0 {
				t.Errorf("Expected %s to succeed, got exit code %d", result.Moniker, result.ExitCode)
			}
		}
	}

	if failureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", failureCount)
	}
}

func TestOrchestrator_LogParsing(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create a test log file
	logDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(logDir, "test.log")
	logContent := `This is a normal line
Warning: something might be wrong
Error: something is definitely wrong
This is another normal line
Fatal: critical failure
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse log
	warnings, errors := parseLogForIssues(logPath)

	// Verify warnings
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	} else if !strings.Contains(warnings[0], "Warning:") {
		t.Errorf("Expected warning to contain 'Warning:', got %s", warnings[0])
	}

	// Verify errors
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

func TestOrchestrator_LogParsing_MkDocs(t *testing.T) {
	// Test that MkDocs-style warnings are captured as errors
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "mkdocs.log")

	// Simulate MkDocs strict mode output
	logContent := `📚 Building MkDocs site using Docker
INFO - Building documentation...
WARNING -  Doc file 'intro.md' contains a link 'nonexistent.md', but target not found
WARNING -  Doc file 'guide.md' contains unrecognized extension
INFO -  Documentation built in 1.23 seconds
Aborted with 2 warnings in strict mode!
❌ MkDocs build failed
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	warnings, errors := parseLogForIssues(logPath)

	// MkDocs "WARNING -" lines should be treated as errors, not warnings
	if len(warnings) != 0 {
		t.Errorf("Expected 0 warnings (MkDocs WARNING - should be errors), got %d: %v", len(warnings), warnings)
	}

	// Should capture: 2 WARNING - lines + 1 Aborted + 1 ❌ line = 4 errors
	if len(errors) != 4 {
		t.Errorf("Expected 4 errors, got %d: %v", len(errors), errors)
	}

	// Verify specific error messages are captured
	hasWarning := false
	hasAborted := false
	hasEmoji := false
	for _, e := range errors {
		if strings.HasPrefix(e, "WARNING -") {
			hasWarning = true
		}
		if strings.Contains(e, "Aborted with") {
			hasAborted = true
		}
		if strings.Contains(e, "❌") {
			hasEmoji = true
		}
	}
	if !hasWarning {
		t.Error("Expected to capture MkDocs 'WARNING -' lines as errors")
	}
	if !hasAborted {
		t.Error("Expected to capture 'Aborted with' message as error")
	}
	if !hasEmoji {
		t.Error("Expected to capture ❌ emoji as error")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{30 * time.Second, "30s"},
		{65 * time.Second, "1m 5s"},
		{125 * time.Second, "2m 5s"},
		{3661 * time.Second, "61m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"", ""},
		{"Already", "Already"},
		{"a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDisplayManager(t *testing.T) {
	// Create display manager with logger writing to buffer
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	dm := newDisplayManager(logger, "testing", 3, 100, false) // 100ms for fast test

	// Start display manager
	dm.start()

	// Mark items as running
	dm.markRunning("module1")
	dm.markRunning("module2")

	// Wait a bit for status to potentially display
	time.Sleep(100 * time.Millisecond)

	// Mark one as completed
	result := &WorkResult{
		Moniker:  "module1",
		ExitCode: 0,
		LogPath:  "out/test/module1/test.log",
	}
	dm.markCompleted(result)

	// Wait for completion to be processed
	time.Sleep(100 * time.Millisecond)

	// Stop and flush completion lines (this is how orchestrator uses it)
	dm.stop()
	dm.flushCompletedLines()

	// Check that output was written
	output := buf.String()
	if output == "" {
		t.Error("Expected display manager to produce output")
	}

	// Should contain completion line
	if !strings.Contains(output, "module1") {
		t.Errorf("Expected output to contain module1, got: %q", output)
	}
}

func TestOrchestrator_PrintSummary(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Configure orchestrator
	config := Config{
		WorkspaceRoot:        tmpDir,
		OutputBaseDir:        paths.OutTestRelPath,
		LogFileName:          "test.log",
		ActionVerb:           "testing",
		MaxConcurrency:       2,
		StatusUpdateInterval: 100, // 100ms for fast test feedback
	}

	// Create orchestrator
	worker := func(moniker string, logWriter io.Writer) int { return 0 }
	orch := New(config, worker)

	// Create mock results
	results := []WorkResult{
		{Moniker: "module1", ExitCode: 0, Warnings: []string{"warning1"}},
		{Moniker: "module2", ExitCode: 1, Errors: []string{"error1"}},
		{Moniker: "module3", ExitCode: 0},
	}

	// Create file to capture summary output
	logPath := filepath.Join(tmpDir, "summary_output.txt")
	logFile, _ := os.Create(logPath)
	defer logFile.Close()

	// Set orchestratorOut for summary
	orch.orchestratorOut = logFile

	// Print summary
	orch.PrintSummary(results)

	// Verify output contains expected content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	summary := string(content)
	if !strings.Contains(summary, "Modules: 3 total") {
		t.Errorf("Summary should contain total module count, got: %s", summary)
	}
	if !strings.Contains(summary, "1 failed") {
		t.Errorf("Summary should contain failure count, got: %s", summary)
	}
	if !strings.Contains(summary, "2 passed") {
		t.Errorf("Summary should contain pass count, got: %s", summary)
	}
}
