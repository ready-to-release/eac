// go.go - Lint handler for Go modules using golangci-lint
package linters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	RegisterHandler(&GoHandler{})
}

// GoHandler lints Go modules using golangci-lint.
type GoHandler struct{}

func (h *GoHandler) Name() string { return "go" }

func (h *GoHandler) Capabilities() []string { return []string{"go_module"} }

func (h *GoHandler) Requirements() []string { return []string{"golangci-lint"} }

func (h *GoHandler) ValidateModule(moduleRoot, workspaceRoot string) error {
	goMod := filepath.Join(moduleRoot, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found at %s", goMod)
	}
	return nil
}

func (h *GoHandler) Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int {
	Logln(logWriter, "\n=== Linting Go module ===")

	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		Logln(logWriter, emojiX+" golangci-lint not found in PATH")
		Logln(logWriter, "")
		Logln(logWriter, "Install golangci-lint with:")
		Logln(logWriter, "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest")
		Logln(logWriter, "")
		return 1
	}

	// Create output file for JSON results
	jsonOutputPath := filepath.Join(outputDir, "lint.json")

	// Build golangci-lint arguments
	args := []string{
		"run",
		// Add JSON output to file (v2 syntax)
		"--output.json.path", jsonOutputPath,
		// Disable text output to stdout (we only want JSON)
		"--output.text.path", "",
	}

	// Add fix flag if requested
	if opts.Fix {
		args = append(args, "--fix")
	}

	// Determine config file
	configPath := opts.Config
	if configPath == "" {
		// Look for config in workspace root first, then module root
		workspaceConfig := filepath.Join(workspaceRoot, ".golangci.yml")
		moduleConfig := filepath.Join(moduleRoot, ".golangci.yml")

		if FileExists(workspaceConfig) {
			configPath = workspaceConfig
		} else if FileExists(moduleConfig) {
			configPath = moduleConfig
		}
	}

	if configPath != "" {
		args = append(args, "--config", configPath)
		Logln(logWriter, "Using config: %s", configPath)
	}

	// Add the target
	args = append(args, "./...")

	Logln(logWriter, "Running: golangci-lint %v", args)

	// Run golangci-lint
	cmd := exec.Command("golangci-lint", args...)
	cmd.Dir = moduleRoot
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	exitCode := 0
	if err := cmd.Run(); err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			Logln(logWriter, "Error running golangci-lint: %v", err)
			return 1
		}
	}

	// Parse and summarize results
	if err := h.summarizeResults(jsonOutputPath, logWriter); err != nil {
		Logln(logWriter, "Warning: could not parse lint results: %v", err)
	}

	if exitCode == 0 {
		Logln(logWriter, "\n"+emojiCheck+" No lint issues found")
	} else {
		Logln(logWriter, "\n"+emojiX+" Lint issues found (see lint.json for details)")
	}

	return exitCode
}

// golangciLintOutput represents the JSON output from golangci-lint.
type golangciLintOutput struct {
	Issues []golangciLintIssue `json:"Issues"`
}

type golangciLintIssue struct {
	FromLinter  string   `json:"FromLinter"`
	Text        string   `json:"Text"`
	Severity    string   `json:"Severity"`
	SourceLines []string `json:"SourceLines"`
	Pos         struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
}

func (h *GoHandler) summarizeResults(jsonPath string, logWriter io.Writer) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	// Handle empty file (no issues)
	if len(data) == 0 {
		return nil
	}

	var output golangciLintOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return err
	}

	if len(output.Issues) == 0 {
		return nil
	}

	// Count issues by linter
	linterCounts := make(map[string]int)
	for _, issue := range output.Issues {
		linterCounts[issue.FromLinter]++
	}

	Logln(logWriter, "\nIssue summary:")
	for linter, count := range linterCounts {
		Logln(logWriter, "  %s: %d issue(s)", linter, count)
	}
	Logln(logWriter, "  Total: %d issue(s)", len(output.Issues))

	// Show first few issues as preview
	maxPreview := 5
	if len(output.Issues) > 0 {
		Logln(logWriter, "\nFirst %d issues:", min(maxPreview, len(output.Issues)))
		for i, issue := range output.Issues {
			if i >= maxPreview {
				Logln(logWriter, "  ... and %d more", len(output.Issues)-maxPreview)
				break
			}
			Logln(logWriter, "  %s:%d: %s (%s)", issue.Pos.Filename, issue.Pos.Line, issue.Text, issue.FromLinter)
		}
	}

	return nil
}

// Emoji constants for output.
const (
	emojiCheck = "\u2705" // check mark
	emojiX     = "\u274C" // X mark
)
