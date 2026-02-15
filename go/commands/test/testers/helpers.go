// helpers.go - Shared utilities for test functions
package testers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/commands/test/internal/cucumber"
	"github.com/ready-to-release/eac/go/core/platform"
	"github.com/ready-to-release/eac/go/core/tool"
)

// Writeln writes a formatted string with platform-specific line ending to the writer.
func Writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// RunTestCommand executes a test command in the specified directory.
// Output is written to both console and log file via the provided writer.
// Returns exit code (0 = success, non-zero = failure).
func RunTestCommand(dir string, logWriter io.Writer, name string, args ...string) int {
	return RunTestCommandWithEnv(dir, logWriter, nil, name, args...)
}

// RunTestCommandWithCapture executes a test command and captures output.
// Output is written to both console and log file, and also captured for summary generation.
// Returns exit code and captured output.
func RunTestCommandWithCapture(dir string, logWriter io.Writer, name string, args ...string) (int, string) {
	var outputBuffer strings.Builder

	// Create multi-writer to capture output while also writing to log
	captureWriter := io.MultiWriter(logWriter, &outputBuffer)

	// Create multi-writer for stderr to capture errors in log
	stderrWriter := io.MultiWriter(os.Stderr, captureWriter)

	toolDef := tool.GlobalRegistry().GetOrAdhoc(name)
	execCtx := &tool.ExecutionContext{
		ModuleRoot:    dir,
		StdoutWriter:  captureWriter,
		StderrWriter:  stderrWriter,
		ArgsOverrides: args,
	}

	exitCode := 0
	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		Writeln(stderrWriter, "\nError: failed to execute test command: %v", err)
		exitCode = 1
	} else if result.ExitCode != 0 {
		exitCode = result.ExitCode
		Writeln(logWriter, "\n❌ Tests exited with code %d", exitCode)
	} else {
		Writeln(logWriter, "\n✅ Tests passed")
	}

	return exitCode, outputBuffer.String()
}

// RunTestCommandWithEnv executes a test command with custom environment variables.
// Output is written to both console and log file via the provided writer.
// Returns exit code (0 = success, non-zero = failure).
func RunTestCommandWithEnv(dir string, logWriter io.Writer, env map[string]string, name string, args ...string) int {
	toolDef := tool.GlobalRegistry().GetOrAdhoc(name)

	// Build full environment
	fullEnv := os.Environ()
	fullEnv = append(fullEnv, "CLIE_TEST_LOGGING_ACTIVE=true")
	if env != nil {
		for key, value := range env {
			fullEnv = append(fullEnv, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create multi-writer for stderr to capture errors in log
	stderrWriter := io.MultiWriter(os.Stderr, logWriter)

	execCtx := &tool.ExecutionContext{
		ModuleRoot:    dir,
		StdoutWriter:  logWriter,
		StderrWriter:  stderrWriter,
		FullEnv:       fullEnv,
		ArgsOverrides: args,
	}

	result, err := tool.GlobalExecutor().Execute(context.Background(), toolDef, execCtx)
	if err != nil {
		Writeln(stderrWriter, "\nError: failed to execute test command: %v", err)
		return 1
	}
	if result.ExitCode != 0 {
		return result.ExitCode
	}

	Writeln(logWriter, "\n✅ Tests passed")
	return 0
}

// GenerateGherkinSummaryMarkdown generates summary_acceptance.md from cucumber.json.
func GenerateGherkinSummaryMarkdown(moniker, workspaceRoot, outputDir string, logWriter io.Writer) {
	cucumberPath := filepath.Join(outputDir, "cucumber.json")
	summaryPath := filepath.Join(outputDir, "summary_acceptance.md")
	appendixPath := filepath.Join(outputDir, "appendix_a.md")

	// Parse cucumber.json
	report, err := cucumber.ParseFile(cucumberPath)
	if err != nil {
		Writeln(logWriter, "Warning: failed to parse cucumber.json: %v", err)
		return
	}

	Writeln(logWriter, "Found %d features", len(report))

	// Generate summary markdown without Appendix A (fragment starting at level 2)
	var summary string
	summary += "## Acceptance Test Summary\n\n"
	summary += cucumber.RenderAllFeatures(report, nil)

	// Write summary_acceptance.md
	if err := os.WriteFile(summaryPath, []byte(summary), 0o644); err != nil {
		Writeln(logWriter, "Warning: failed to write summary_acceptance.md: %v", err)
		return
	}

	Writeln(logWriter, "✅ Generated: %s", summaryPath)

	// Generate Appendix A as separate file (fragment starting at level 2)
	var appendix string
	appendix += "## Appendix A: Specifications and Test Results\n\n"
	appendix += cucumber.RenderAppendixA(report, workspaceRoot)

	// Write appendix_a.md
	if err := os.WriteFile(appendixPath, []byte(appendix), 0o644); err != nil {
		Writeln(logWriter, "Warning: failed to write appendix_a.md: %v", err)
		return
	}

	Writeln(logWriter, "✅ Generated: %s", appendixPath)
}

// FindModulesWithResults finds all subdirectories containing cucumber.json.
func FindModulesWithResults(testRunDir string) ([]string, error) {
	entries, err := os.ReadDir(testRunDir)
	if err != nil {
		return nil, err
	}

	var foundModules []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this directory has cucumber.json
		cucumberPath := filepath.Join(testRunDir, entry.Name(), "cucumber.json")
		if _, err := os.Stat(cucumberPath); err == nil {
			foundModules = append(foundModules, entry.Name())
		}
	}

	return foundModules, nil
}

// FormatDuration formats a duration as "1m 23s" or "45s".
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
