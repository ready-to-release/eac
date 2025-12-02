// Package design provides formatting for Structurizr CLI validation results
package design

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// FormatValidationResult formats a single validation result for console output
func FormatValidationResult(result *ValidationResult, verbose bool) string {
	output := ""

	output += fmt.Sprintf("🔍 Validating module: %s\n", result.Module)
	output += "🐳 Using Docker: structurizr/cli:latest\n"

	if verbose {
		output += fmt.Sprintf("⏱️  Started at: %s\n", result.Timestamp.Format("15:04:05"))
	}

	// Multi-file mode
	if len(result.Files) > 1 {
		output += fmt.Sprintf("📄 Design files: %d found\n", len(result.Files))
		output += "\n"

		for i, fileResult := range result.Files {
			output += fmt.Sprintf("  [%d/%d] %s\n", i+1, len(result.Files), fileResult.FileName)
			if fileResult.Valid {
				output += fmt.Sprintf("  ✅ Valid (%d errors, %d warnings)\n", len(fileResult.Errors), len(fileResult.Warnings))
			} else {
				output += fmt.Sprintf("  ❌ Failed (%d errors, %d warnings)\n", len(fileResult.Errors), len(fileResult.Warnings))
				// Show errors for failed files
				for _, err := range fileResult.Errors {
					if err.Line > 0 {
						output += fmt.Sprintf("     Line %d: %s\n", err.Line, err.Message)
					} else {
						output += fmt.Sprintf("     %s\n", err.Message)
					}
				}
			}
			output += "\n"
		}
	} else {
		// Single file mode (backward compatible)
		output += fmt.Sprintf("📄 Workspace: %s\n", result.WorkspacePath)
		output += "\n"

		if result.Valid {
			output += "✅ Workspace is valid\n"
		} else {
			output += "❌ Workspace validation failed\n"
		}

		// Show errors
		if len(result.Errors) > 0 {
			output += "\nErrors:\n"
			for _, err := range result.Errors {
				if err.Line > 0 {
					output += fmt.Sprintf("  - Line %d: %s\n", err.Line, err.Message)
				} else {
					output += fmt.Sprintf("  - %s\n", err.Message)
				}
			}
		}

		// Show warnings
		if len(result.Warnings) > 0 {
			output += "\nWarnings:\n"
			for _, warn := range result.Warnings {
				if warn.Line > 0 {
					output += fmt.Sprintf("  - Line %d: %s\n", warn.Line, warn.Message)
				} else {
					output += fmt.Sprintf("  - %s\n", warn.Message)
				}
			}
		}
	}

	// Show summary
	output += "📊 Summary:\n"
	if len(result.Files) > 1 {
		output += fmt.Sprintf("  Files: %d validated\n", len(result.Files))
	}
	output += fmt.Sprintf("  Errors: %d\n", len(result.Errors))
	output += fmt.Sprintf("  Warnings: %d\n", len(result.Warnings))
	output += fmt.Sprintf("  Execution time: %s\n", result.ExecutionTime)

	// Show verbose details
	if verbose && len(result.Files) <= 1 {
		output += "\n🔍 Verbose Details:\n"
		output += "  Docker Command Executed:\n"

		// Extract directory from workspace path
		workspaceDir := ""
		workspaceFile := "workspace.dsl"
		if strings.Contains(result.WorkspacePath, "\\") || strings.Contains(result.WorkspacePath, "/") {
			lastSep := strings.LastIndexAny(result.WorkspacePath, "\\/")
			if lastSep > 0 {
				workspaceDir = result.WorkspacePath[:lastSep]
				workspaceFile = result.WorkspacePath[lastSep+1:]
			}
		}

		output += fmt.Sprintf("    docker run --rm \\\n")
		output += fmt.Sprintf("      -v \"%s:/workspace\" \\\n", workspaceDir)
		output += fmt.Sprintf("      structurizr/cli:latest \\\n")
		output += fmt.Sprintf("      validate -workspace /workspace/%s\n", workspaceFile)

		output += "\n  Raw Structurizr CLI Output:\n"
		if result.RawOutput == "" {
			output += "  (No output captured - this may indicate an issue with Docker execution)\n"
		} else {
			output += "  " + strings.ReplaceAll(result.RawOutput, "\n", "\n  ") + "\n"
		}
	}

	return output
}

// FormatValidationSummary formats a validation summary for console output
func FormatValidationSummary(summary *ValidationSummary, verbose bool) string {
	output := ""

	output += "🔍 Validating all modules...\n"
	output += "🐳 Using Docker: structurizr/cli:latest\n"

	if verbose {
		output += fmt.Sprintf("⏱️  Started at: %s\n", summary.Timestamp.Format("15:04:05"))
		output += fmt.Sprintf("📦 Total modules to validate: %d\n", summary.TotalModules)
	}

	output += "\n"

	// Count total files across all modules
	totalFiles := 0
	for _, result := range summary.Results {
		if result.TotalFiles > 0 {
			totalFiles += result.TotalFiles
		} else {
			totalFiles++ // Backward compat: count as 1 file
		}
	}

	// Show individual results
	for _, result := range summary.Results {
		if result.Valid {
			output += fmt.Sprintf("Module: %s", result.Module)
			if result.TotalFiles > 1 {
				output += fmt.Sprintf(" (%d files)", result.TotalFiles)
			}
			output += "\n"
			if verbose && result.TotalFiles > 1 {
				for _, f := range result.Files {
					output += fmt.Sprintf("  📄 %s: ✅\n", f.FileName)
				}
			} else if verbose {
				output += fmt.Sprintf("  📄 Workspace: %s\n", result.WorkspacePath)
			}
			output += fmt.Sprintf("  ✅ Valid (%s)\n", result.ExecutionTime)
			if len(result.Warnings) > 0 {
				output += "  Warnings:\n"
				for _, warn := range result.Warnings {
					if warn.Line > 0 {
						output += fmt.Sprintf("    - Line %d: %s\n", warn.Line, warn.Message)
					} else {
						output += fmt.Sprintf("    - %s\n", warn.Message)
					}
				}
			}
		} else {
			output += fmt.Sprintf("Module: %s", result.Module)
			if result.TotalFiles > 1 {
				output += fmt.Sprintf(" (%d files)", result.TotalFiles)
			}
			output += "\n"
			if verbose && result.TotalFiles > 1 {
				for _, f := range result.Files {
					if f.Valid {
						output += fmt.Sprintf("  📄 %s: ✅\n", f.FileName)
					} else {
						output += fmt.Sprintf("  📄 %s: ❌\n", f.FileName)
					}
				}
			} else if verbose {
				output += fmt.Sprintf("  📄 Workspace: %s\n", result.WorkspacePath)
			}
			output += fmt.Sprintf("  ❌ Failed (%s)\n", result.ExecutionTime)
			if len(result.Errors) > 0 {
				output += "  Errors:\n"
				for _, err := range result.Errors {
					if err.Line > 0 {
						output += fmt.Sprintf("    - Line %d: %s\n", err.Line, err.Message)
					} else {
						output += fmt.Sprintf("    - %s\n", err.Message)
					}
				}
			}
		}
		output += "\n"
	}

	// Show overall summary
	output += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	output += "📊 Summary:\n"
	output += fmt.Sprintf("  Total modules: %d\n", summary.TotalModules)
	if totalFiles > summary.TotalModules {
		output += fmt.Sprintf("  Total files: %d\n", totalFiles)
	}
	output += fmt.Sprintf("  Passed: %d\n", summary.PassedModules)
	output += fmt.Sprintf("  Failed: %d\n", summary.FailedModules)
	output += fmt.Sprintf("  Total errors: %d\n", summary.TotalErrors)
	output += fmt.Sprintf("  Total warnings: %d\n", summary.TotalWarnings)
	output += fmt.Sprintf("  Execution time: %s\n", summary.ExecutionTime)

	return output
}

// WriteValidationResultJSON writes a validation result to JSON file
func WriteValidationResultJSON(result *ValidationResult, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := "out"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// WriteValidationSummaryJSON writes a validation summary to JSON file
func WriteValidationSummaryJSON(summary *ValidationSummary, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := "out"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}
