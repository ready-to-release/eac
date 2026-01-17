// markdown.go - Lint handler for markdown files using markdownlint-cli2
// This uses the same engine as the VSCode markdownlint extension for consistency.
package linters

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
)

func init() {
	RegisterHandler(&MarkdownHandler{})
}

// MarkdownHandler lints markdown files using markdownlint-cli2.
// Configuration files (one source of truth):
// - .markdownlint.yml for rules
// - .markdownlint-cli2.yaml for ignores (extends .markdownlint.yml)
// Both files are used by the VSCode markdownlint extension automatically.
type MarkdownHandler struct{}

func (h *MarkdownHandler) Name() string { return "markdown" }

func (h *MarkdownHandler) Capabilities() []string { return []string{"markdown"} }

func (h *MarkdownHandler) Requirements() []string { return []string{"markdownlint-cli2"} }

func (h *MarkdownHandler) ValidateModule(moduleRoot, workspaceRoot string) error {
	// Check that module root exists
	if _, err := os.Stat(moduleRoot); os.IsNotExist(err) {
		return fmt.Errorf("module root not found at %s", moduleRoot)
	}
	return nil
}

// findMarkdownLintCLI2 locates the markdownlint-cli2 executable.
// On Windows, npm installs tools as .cmd files, which need to be located explicitly.
func findMarkdownLintCLI2() (string, error) {
	// Try finding markdownlint-cli2 directly
	if path, err := exec.LookPath("markdownlint-cli2"); err == nil {
		return path, nil
	}

	// On Windows, also look for the .cmd variant
	if runtime.GOOS == "windows" {
		if path, err := exec.LookPath("markdownlint-cli2.cmd"); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("markdownlint-cli2 not found")
}

func (h *MarkdownHandler) Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int {
	Logln(logWriter, "\n=== Linting Markdown files ===")

	// Format tables first if --fix is specified
	// This uses Go-native table formatting (MD060 compliance)
	if opts.Fix {
		formatted, err := h.formatTables(moduleRoot, workspaceRoot, logWriter)
		if err != nil {
			Logln(logWriter, "Warning: table formatting failed: %v", err)
		} else if formatted > 0 {
			Logln(logWriter, emojiCheck+" Formatted tables in %d file(s)", formatted)
		}
	}

	// Find markdownlint-cli2 executable
	// On Windows, npm installs as .cmd files which require special handling
	cmdPath, err := findMarkdownLintCLI2()
	if err != nil {
		Logln(logWriter, emojiX+" markdownlint-cli2 not found in PATH")
		Logln(logWriter, "")
		Logln(logWriter, "Install markdownlint-cli2 with:")
		Logln(logWriter, "  npm install -g markdownlint-cli2")
		Logln(logWriter, "")
		return 1
	}

	// Create output file for JSON results
	jsonOutputPath := filepath.Join(outputDir, "lint.json")

	// markdownlint-cli2 auto-discovers config files (.markdownlint-cli2.yaml, .markdownlint.yml)
	// from the current directory and parent directories. We run from moduleRoot, and the
	// workspace root config files will be found automatically via directory traversal.
	//
	// Only pass --config if explicitly specified via opts.Config
	args := []string{}

	if opts.Config != "" {
		args = append(args, "--config", opts.Config)
		Logln(logWriter, "Using explicit config: %s", opts.Config)
	}

	// Add fix flag if requested
	if opts.Fix {
		args = append(args, "--fix")
	}

	// Add the target glob (relative, since cmd.Dir is moduleRoot)
	args = append(args, "**/*.md")

	Logln(logWriter, "Running: markdownlint-cli2 %v", args)

	// Run markdownlint-cli2
	cmd := exec.Command(cmdPath, args...)
	cmd.Dir = moduleRoot
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			Logln(logWriter, "Error running markdownlint-cli2: %v", err)
			return 1
		}
	}

	// Write a summary JSON
	h.writeResultsJSON(jsonOutputPath, exitCode, logWriter)

	if exitCode == 0 {
		Logln(logWriter, "\n"+emojiCheck+" No markdown lint issues found")
	} else {
		Logln(logWriter, "\n"+emojiX+" Markdown lint issues found")
	}

	return exitCode
}

// MarkdownLintOutput represents the structured JSON output for markdown linting
type MarkdownLintOutput struct {
	Success bool `json:"success"`
	Tool    string `json:"tool"`
}

func (h *MarkdownHandler) writeResultsJSON(jsonPath string, exitCode int, logWriter io.Writer) {
	output := MarkdownLintOutput{
		Success: exitCode == 0,
		Tool:    "markdownlint-cli2",
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		Logln(logWriter, "Warning: could not marshal lint.json: %v", err)
		return
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		Logln(logWriter, "Warning: could not write lint.json: %v", err)
	}
}

// formatTables finds and formats all markdown tables in .md files for MD060 compliance.
// Uses the Go-native render.FormatMarkdownTable for aligned table formatting.
func (h *MarkdownHandler) formatTables(moduleRoot, workspaceRoot string, logWriter io.Writer) (int, error) {
	// Load ignore patterns from .markdownlint-cli2.yaml if present
	ignorePatterns := h.loadIgnorePatterns(workspaceRoot)

	var filesFormatted int

	// Walk the module directory
	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be ignored
			dirName := info.Name()
			for _, pattern := range []string{"node_modules", ".git", "out", ".vscode", ".idea", ".cache"} {
				if dirName == pattern {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Check if file matches ignore patterns
		relPath, _ := filepath.Rel(workspaceRoot, path)
		relPath = filepath.ToSlash(relPath) // Normalize path separators
		for _, pattern := range ignorePatterns {
			if matchGlob(pattern, relPath) {
				return nil
			}
		}

		// Format tables in this file
		formatted, err := h.formatTablesInFile(path)
		if err != nil {
			return nil // Skip files with errors, don't fail the whole walk
		}
		if formatted {
			filesFormatted++
		}

		return nil
	})

	return filesFormatted, err
}

// loadIgnorePatterns reads ignore patterns from .markdownlint-cli2.yaml
func (h *MarkdownHandler) loadIgnorePatterns(workspaceRoot string) []string {
	configPath := filepath.Join(workspaceRoot, ".markdownlint-cli2.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	// Simple parsing - find "ignores:" section and extract patterns
	var patterns []string
	lines := strings.Split(string(content), "\n")
	inIgnores := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "ignores:") {
			inIgnores = true
			continue
		}

		if inIgnores {
			// Check if we've hit a new top-level key
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && trimmed != "" && !strings.HasPrefix(trimmed, "-") {
				break
			}

			// Extract pattern from "  - \"pattern\"" format
			if strings.HasPrefix(trimmed, "-") {
				pattern := strings.TrimPrefix(trimmed, "-")
				pattern = strings.TrimSpace(pattern)
				pattern = strings.Trim(pattern, "\"'")
				if pattern != "" {
					patterns = append(patterns, pattern)
				}
			}
		}
	}

	return patterns
}

// matchGlob performs simple glob matching (supports * and **)
func matchGlob(pattern, path string) bool {
	// Convert glob to regex
	regexPattern := "^"
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				regexPattern += ".*"
				i++ // Skip second *
				if i+1 < len(pattern) && pattern[i+1] == '/' {
					i++ // Skip trailing /
				}
			} else {
				regexPattern += "[^/]*"
			}
		case '?':
			regexPattern += "."
		case '.', '+', '^', '$', '(', ')', '{', '}', '[', ']', '|', '\\':
			regexPattern += "\\" + string(pattern[i])
		default:
			regexPattern += string(pattern[i])
		}
	}
	regexPattern += "$"

	matched, _ := regexp.MatchString(regexPattern, path)
	return matched
}

// formatTablesInFile formats all tables in a single markdown file
func (h *MarkdownHandler) formatTablesInFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	original := string(content)
	formatted := formatMarkdownTables(original)

	// Only write if changed
	if formatted != original {
		return true, os.WriteFile(filePath, []byte(formatted), 0644)
	}

	return false, nil
}

// formatMarkdownTables finds and formats all tables in markdown content
func formatMarkdownTables(content string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	var tableLines []string
	inTable := false
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()

		// Track code blocks to avoid formatting tables inside them
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		if inCodeBlock {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Check if line is a table row (starts with |)
		trimmed := strings.TrimSpace(line)
		isTableRow := strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")

		if isTableRow {
			if !inTable {
				inTable = true
			}
			tableLines = append(tableLines, line)
		} else {
			// Not a table row - if we were in a table, format and output it
			if inTable {
				formatted := formatTable(tableLines)
				result.WriteString(formatted)
				tableLines = nil
				inTable = false
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	// Handle table at end of file
	if inTable && len(tableLines) > 0 {
		formatted := formatTable(tableLines)
		result.WriteString(formatted)
	}

	// Remove trailing newline if original didn't have one
	output := result.String()
	if !strings.HasSuffix(content, "\n") && strings.HasSuffix(output, "\n") {
		output = strings.TrimSuffix(output, "\n")
	}

	return output
}

// formatTable formats a slice of table lines using go-pretty's aligned output
func formatTable(lines []string) string {
	if len(lines) < 2 {
		// Not a valid table (need at least header + separator)
		return strings.Join(lines, "\n") + "\n"
	}

	// Use the render package's FormatMarkdownTable
	tableContent := strings.Join(lines, "\n")
	formatted := render.FormatMarkdownTable(tableContent)

	return formatted + "\n"
}
