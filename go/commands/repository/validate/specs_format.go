package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/repository"
)

// outputText displays validation results in text format.
func outputText(results []*ValidationResult, quiet, verbose bool) {
	if len(results) == 0 {
		log.Info("No specification files found")
		return
	}

	// For single file, show detailed output
	if len(results) == 1 {
		log.Info(formatValidationResult(results[0]))
		return
	}

	// For multiple files, show summary
	log.Info("")
	log.Info(formatValidationSummary(results))

	// Show details for failed validations (unless quiet)
	if !quiet {
		log.Info("")
		for _, result := range results {
			if !result.Valid {
				log.Info(formatValidationResult(result))
				log.Info("")
			}
		}
	}
}

// outputJSON displays validation results in JSON format.
func outputJSON(results []*ValidationResult) {
	output := map[string]interface{}{
		"results": results,
		"summary": map[string]int{
			"total":  len(results),
			"passed": countPassed(results),
			"failed": countFailed(results),
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		log.Errorf("Error encoding JSON: %v", err)
	}
}

// formatValidationResult formats a single validation result for display.
func formatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if result.Valid {
		output.WriteString("✅ Validation passed")
	} else {
		output.WriteString("❌ Validation failed")
	}

	// Normalize path: relative to repo root and Unix-style separators
	displayPath := normalizePath(result.Path)
	output.WriteString(fmt.Sprintf(": %s\n", displayPath))

	if len(result.Errors) == 0 {
		return output.String()
	}

	// Count errors and warnings
	errorCount := 0
	warningCount := 0
	for _, e := range result.Errors {
		if !e.IsWarning() {
			errorCount++
		} else {
			warningCount++
		}
	}

	// Display counts
	if errorCount > 0 && warningCount > 0 {
		output.WriteString(fmt.Sprintf("\n%d error(s), %d warning(s):\n\n", errorCount, warningCount))
	} else if errorCount > 0 {
		output.WriteString(fmt.Sprintf("\n%d error(s):\n\n", errorCount))
	} else {
		output.WriteString(fmt.Sprintf("\n%d warning(s):\n\n", warningCount))
	}

	// Display each error/warning
	output.WriteString(domain.FormatValidationErrors(result.Errors))

	return output.String()
}

// formatValidationSummary formats a summary of multiple validation results.
func formatValidationSummary(results []*ValidationResult) string {
	if len(results) == 0 {
		return "No specification files found"
	}

	passed := countPassed(results)
	failed := countFailed(results)

	var output strings.Builder
	output.WriteString("═══════════════════════════════════════════════════════════\n")
	output.WriteString("  Validation Summary\n")
	output.WriteString("═══════════════════════════════════════════════════════════\n\n")
	output.WriteString(fmt.Sprintf("  Total files:  %d\n", len(results)))
	output.WriteString(fmt.Sprintf("  %d passed, %d failed\n", passed, failed))
	output.WriteString("\n═══════════════════════════════════════════════════════════")

	return output.String()
}

// countPassed counts the number of passed validations.
func countPassed(results []*ValidationResult) int {
	count := 0
	for _, r := range results {
		if r.Valid {
			count++
		}
	}
	return count
}

// countFailed counts the number of failed validations.
func countFailed(results []*ValidationResult) int {
	count := 0
	for _, r := range results {
		if !r.Valid {
			count++
		}
	}
	return count
}

// relativePath returns a path relative to the repository root for display.
func relativePath(path, repoRoot string) string {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return path
	}
	return rel
}

// normalizePath converts a path to Unix-style relative path from repository root.
func normalizePath(path string) string {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		// If we can't get repo root, just normalize slashes
		return filepath.ToSlash(path)
	}

	// Make path relative to repo root
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		// If we can't make it relative, just normalize slashes
		return filepath.ToSlash(path)
	}

	// Convert to Unix-style path separators
	return filepath.ToSlash(rel)
}
