// Command: specs validate
// Description: Validate existing Gherkin specifications against contracts
// Short: Validate Gherkin specifications against quality contracts
// Long: The specs validate command checks existing .feature files against the specification contract,
// Long: ensuring they follow proper Gherkin syntax, BDD patterns, and project standards.
// Long: Validation covers structure (Feature/Rule/Scenario hierarchy), tags, step formatting, and content quality.
// Long: The command can validate a single file or recursively validate all .feature files in a directory.
// Long: By default, output is in human-readable text format. Use --format json for machine-readable output.
// Long: Exit code is 0 if all validations pass, 1 if any critical errors are found.
// Flag.quiet: type=bool, shorthand=q, default=false, usage=Suppress success messages and show only validation errors and warnings
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed validation output including metadata and additional context
// Flag.format: type=string, shorthand=f, default=text, completion=text,json, usage=Output format for validation results (text for human-readable, json for machine-readable)
// Usage: specs validate <path> [--quiet] [--verbose] [--format json]
// Flags: --quiet (show only errors), --verbose (detailed output), --format (output format: text|json)
// HasSideEffects: false
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(SpecsValidate)
}

// Intent: Validate Gherkin specifications against the contract
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear separation of concerns (parse config, validate, format output)
//   - Descriptive function names reveal intent
//   - Single responsibility per function
//
// Easy to change:
//   - Configuration struct decouples CLI parsing from validation logic
//   - Validation logic reuses existing GherkinValidator
//   - Output formatting is separate from validation
//
// Hard to break:
//   - Path security validation prevents traversal attacks
//   - Comprehensive error handling with context
//   - Validation uses battle-tested contract framework
//   - Returns non-zero exit code on validation failures

// SpecsValidate validates existing Gherkin specification files
func SpecsValidate() int {
	// Parse configuration
	config, err := parseValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate path security
	if err := validatePath(config.Path, config.RepositoryRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Determine if path is file or directory
	info, err := os.Stat(config.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: file or directory not found: %s\n", config.Path)
		return 1
	}

	var results []*ValidationResult
	if info.IsDir() {
		// Validate directory (recursive)
		results, err = validateDirectory(config.Path, config.RepositoryRoot, config.Quiet)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		// Validate single file
		errors, err := validateGherkinFile(config.Path, config.RepositoryRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}

		criticalCount := contracts.CountCriticalErrors(errors)
		results = []*ValidationResult{
			{
				Path:   config.Path,
				Valid:  criticalCount == 0,
				Errors: errors,
			},
		}
	}

	// Format and display output
	if config.Format == "json" {
		outputJSON(results)
	} else {
		outputText(results, config.Quiet, config.Verbose)
	}

	// Return exit code based on validation results
	for _, result := range results {
		if !result.Valid {
			return 1
		}
	}

	return 0
}

// ValidateConfig holds configuration for specs validate command
type ValidateConfig struct {
	Path           string
	Quiet          bool
	Verbose        bool
	Format         string
	RepositoryRoot string
}

// ValidationResult holds the validation result for a single file
type ValidationResult struct {
	Path   string                       `json:"path"`
	Valid  bool                         `json:"valid"`
	Errors []contracts.ValidationError   `json:"errors"`
}

// parseValidateConfig parses command line arguments into configuration
func parseValidateConfig() (*ValidateConfig, error) {
	config := &ValidateConfig{
		Format: "text", // Default format
	}

	args := os.Args[3:] // Skip program name and "specs validate"
	var path string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--quiet":
			config.Quiet = true
		case "--verbose":
			config.Verbose = true
		case "--format":
			if i+1 < len(args) {
				config.Format = args[i+1]
				i++
			} else {
				return nil, fmt.Errorf("--format requires an argument (text|json)")
			}
		default:
			if !strings.HasPrefix(arg, "--") {
				if path == "" {
					path = arg
				} else {
					return nil, fmt.Errorf("unexpected argument: %s", arg)
				}
			}
		}
	}

	// Validate path argument
	if path == "" {
		return nil, fmt.Errorf("file path or directory is required\n\nUsage: specs validate <path> [--quiet] [--verbose] [--format json]\nExample: specs validate specs/src-commands/specs/create/specification.feature")
	}

	config.Path = path

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}

	config.RepositoryRoot = repoRoot

	// Make path absolute if relative
	if !filepath.IsAbs(config.Path) {
		config.Path = filepath.Join(repoRoot, config.Path)
	}

	// Validate format
	if config.Format != "text" && config.Format != "json" {
		return nil, fmt.Errorf("invalid format: %s (must be 'text' or 'json')", config.Format)
	}

	return config, nil
}

// validatePath ensures the path is within the repository (prevents path traversal attacks)
func validatePath(path string, repoRoot string) error {
	// Clean paths
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(repoRoot)

	// Make path absolute
	absPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		absPath = filepath.Join(cleanRoot, cleanPath)
		absPath = filepath.Clean(absPath)
	}

	absRoot := cleanRoot
	if !filepath.IsAbs(cleanRoot) {
		var err error
		absRoot, err = filepath.Abs(cleanRoot)
		if err != nil {
			return fmt.Errorf("failed to resolve repository root: %w", err)
		}
	}

	// Ensure path is within repository
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// If the relative path starts with "..", it's trying to escape the repository
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("path must be within repository (attempted: %s)", path)
	}

	return nil
}

// validateGherkinFile validates a single Gherkin specification file
func validateGherkinFile(filePath string, repoRoot string) ([]contracts.ValidationError, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Load contract and validator
	loader := contracts.NewContractLoader(repoRoot, "ai/specifications", "0.1.0")

	contractData, err := loader.LoadContract()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %w", err)
	}

	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		return nil, fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Create validator
	validator := contracts.NewGherkinValidator(contractData, antiCorruptionRules)

	// Validate content
	errors := validator.Validate(string(content), nil)

	return errors, nil
}

// validateDirectory validates all .feature files in a directory (recursive)
func validateDirectory(dirPath string, repoRoot string, quiet bool) ([]*ValidationResult, error) {
	var results []*ValidationResult

	// Walk directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .feature files
		if !strings.HasSuffix(path, ".feature") {
			return nil
		}

		// Validate file
		errors, validateErr := validateGherkinFile(path, repoRoot)
		if validateErr != nil {
			// Log error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to validate %s: %v\n", path, validateErr)
			return nil
		}

		criticalCount := contracts.CountCriticalErrors(errors)
		result := &ValidationResult{
			Path:   path,
			Valid:  criticalCount == 0,
			Errors: errors,
		}

		results = append(results, result)

		// In quiet mode, only show progress for invalid files
		if !quiet || !result.Valid {
			if result.Valid && len(result.Errors) == 0 {
				fmt.Printf("✅ %s\n", relativePath(path, repoRoot))
			} else if result.Valid {
				fmt.Printf("✅ %s (%d warning(s))\n", relativePath(path, repoRoot), len(result.Errors))
			} else {
				fmt.Printf("❌ %s\n", relativePath(path, repoRoot))
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return results, nil
}

// outputText displays validation results in text format
func outputText(results []*ValidationResult, quiet bool, verbose bool) {
	if len(results) == 0 {
		fmt.Println("No specification files found")
		return
	}

	// For single file, show detailed output
	if len(results) == 1 {
		fmt.Println(formatValidationResult(results[0]))
		return
	}

	// For multiple files, show summary
	fmt.Println()
	fmt.Println(formatValidationSummary(results))

	// Show details for failed validations (unless quiet)
	if !quiet {
		fmt.Println()
		for _, result := range results {
			if !result.Valid {
				fmt.Println(formatValidationResult(result))
				fmt.Println()
			}
		}
	}
}

// outputJSON displays validation results in JSON format
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
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

// formatValidationResult formats a single validation result for display
func formatValidationResult(result *ValidationResult) string {
	var output strings.Builder

	if result.Valid {
		output.WriteString("✅ Validation passed")
	} else {
		output.WriteString("❌ Validation failed")
	}

	output.WriteString(fmt.Sprintf(": %s\n", result.Path))

	if len(result.Errors) == 0 {
		return output.String()
	}

	// Count errors and warnings
	errorCount := 0
	warningCount := 0
	for _, e := range result.Errors {
		if e.Severity == "error" {
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
	output.WriteString(contracts.FormatValidationErrors(result.Errors))

	return output.String()
}

// formatValidationSummary formats a summary of multiple validation results
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

// countPassed counts the number of passed validations
func countPassed(results []*ValidationResult) int {
	count := 0
	for _, r := range results {
		if r.Valid {
			count++
		}
	}
	return count
}

// countFailed counts the number of failed validations
func countFailed(results []*ValidationResult) int {
	count := 0
	for _, r := range results {
		if !r.Valid {
			count++
		}
	}
	return count
}

// relativePath returns a path relative to the repository root for display
func relativePath(path string, repoRoot string) string {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return path
	}
	return rel
}
