// Command: validate specs
// Description: Validate existing Gherkin specifications against contracts
// Short: Validate Gherkin specifications against quality contracts
// Long: The validate specs command checks existing .feature files against the specification contract,
// Long: ensuring they follow proper Gherkin syntax, and project standards.
// Long: Validation covers structure (Feature/Rule/Scenario hierarchy), tags, step formatting, and content quality.
// Long: The command can validate a single file or recursively validate all .feature files in a directory.
// Long: By default, output is in human-readable text format. Use --format json for machine-readable output.
// Long: Exit code is 0 if all validations pass, 1 if any critical errors are found.
// Flag.quiet: type=bool, shorthand=q, default=false, usage=Suppress success messages and show only validation errors and warnings
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed validation output including metadata and additional context
// Flag.format: type=string, shorthand=f, default=text, completion=text,json, usage=Output format for validation results (text for human-readable, json for machine-readable)
// Usage: validate specs <path> [--quiet] [--verbose] [--format json]
// Flags: --quiet (show only errors), --verbose (detailed output), --format (output format: text|json)
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateSpecs)
}

// ============================================================================
// Mock Support for Testing
// ============================================================================

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository

// getGitRepo returns the git repository, creating one if needed
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	return git.Open(workspaceRoot)
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the mock git repository.
func ResetGitRepo() {
	gitRepo = nil
}

// ============================================================================

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

// ValidateSpecs validates existing Gherkin specification files
func ValidateSpecs() int {
	// Parse configuration
	config, err := parseValidateConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Initialize logger (logs to out/logs/specs/)
	var logger *logging.Logger
	if config.Verbose {
		logger, err = logging.NewWithDebug("specs", config.RepositoryRoot)
	} else {
		logger, err = logging.NewDefault("specs", config.RepositoryRoot)
	}
	if err != nil {
		// Continue without logger - not fatal
		if config.Verbose {
			log.Errorf("Warning: Failed to initialize logger: %v", err)
		}
	} else {
		config.Logger = logger
		defer logger.Sync()
	}

	// Log command start
	if config.Logger != nil {
		config.Logger.Info("Starting specs validate",
			zap.String("path", config.Path),
			zap.String("format", config.Format),
			zap.Bool("quiet", config.Quiet),
			zap.Bool("verbose", config.Verbose))
	}

	// Validate path security
	if err := validatePath(config.Path, config.RepositoryRoot); err != nil {
		if config.Logger != nil {
			config.Logger.Error("Path security validation failed",
				zap.String("path", config.Path),
				zap.Error(err))
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Determine if path is file or directory
	info, err := os.Stat(config.Path)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("File or directory not found",
				zap.String("path", config.Path),
				zap.Error(err))
		}
		log.Errorf("Error: file or directory not found: %s", config.Path)
		return 1
	}

	if config.Logger != nil {
		config.Logger.Debug("Starting validation",
			zap.String("path", config.Path),
			zap.Bool("isDir", info.IsDir()))
	}

	var results []*ValidationResult
	if info.IsDir() {
		// Validate directory (recursive)
		results, err = validateDirectoryWithLogger(config.Path, config.RepositoryRoot, config.Quiet, config.CheckTags, config.Logger)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Error("Directory validation failed", zap.Error(err))
			}
			log.Errorf("Error: %v", err)
			return 1
		}
	} else {
		// Validate single file
		errors, err := validateGherkinFile(config.Path, config.RepositoryRoot, config.CheckTags)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Error("File validation failed",
					zap.String("path", config.Path),
					zap.Error(err))
			}
			log.Errorf("Error: %v", err)
			return 1
		}

		// Apply fixes if --fix flag is set
		if config.Fix && len(errors) > 0 {
			fixResult, fixErr := fixGherkinFile(config.Path, errors)
			if fixErr != nil {
				if config.Logger != nil {
					config.Logger.Error("Failed to fix file",
						zap.String("path", config.Path),
						zap.Error(fixErr))
				}
				log.Errorf("Error fixing file: %v", fixErr)
			} else if fixResult.FixCount() > 0 {
				// Display fix results
				log.Info(formatFixResult(fixResult, config.RepositoryRoot))

				if config.Logger != nil {
					config.Logger.Info("Applied fixes",
						zap.String("path", config.Path),
						zap.Int("fixCount", fixResult.FixCount()))
				}

				// Re-validate after fixes
				errors, err = validateGherkinFile(config.Path, config.RepositoryRoot, config.CheckTags)
				if err != nil {
					if config.Logger != nil {
						config.Logger.Error("Re-validation failed after fixes",
							zap.String("path", config.Path),
							zap.Error(err))
					}
					log.Errorf("Error re-validating after fixes: %v", err)
					return 1
				}
			}
		}

		criticalCount := contracts.CountCriticalErrors(errors)
		results = []*ValidationResult{
			{
				Path:   config.Path,
				Valid:  criticalCount == 0,
				Errors: errors,
			},
		}

		if config.Logger != nil {
			config.Logger.Debug("Single file validation complete",
				zap.String("path", config.Path),
				zap.Bool("valid", criticalCount == 0),
				zap.Int("errorCount", len(errors)))
		}
	}

	passed := countPassed(results)
	failed := countFailed(results)

	if config.Logger != nil {
		config.Logger.Info("Validation complete",
			zap.Int("total", len(results)),
			zap.Int("passed", passed),
			zap.Int("failed", failed))
	}

	// Format and display output
	if config.Format == "json" {
		outputJSON(results)
	} else {
		outputText(results, config.Quiet, config.Verbose)
	}

	// Return exit code based on validation results
	if failed > 0 {
		return 1
	}

	return 0
}

// ValidateConfig holds configuration for specs validate command
type ValidateConfig struct {
	Path           string
	Quiet          bool   // -q, --quiet: Show only errors
	Verbose        bool   // -v, --verbose: Detailed output
	Format         string // -f, --format: Output format (text, json)
	CheckTags      bool   // --check-tags: Enable tag validation (default: true)
	Fix            bool   // --fix: Auto-fix correctable tag issues
	RepositoryRoot string
	Logger         *logging.Logger
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
		Format:    "text", // Default format
		CheckTags: true,   // Default: tag validation enabled
	}

	args := os.Args[3:] // Skip program name, "validate", and "specs"
	var path string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-q", "--quiet":
			config.Quiet = true
		case "-v", "--verbose":
			config.Verbose = true
		case "-f", "--format":
			if i+1 < len(args) {
				config.Format = args[i+1]
				i++
			} else {
				return nil, fmt.Errorf("--format requires an argument (text|json)")
			}
		case "--check-tags":
			config.CheckTags = true
		case "--no-check-tags":
			config.CheckTags = false
		case "--fix":
			config.Fix = true
		default:
			if !strings.HasPrefix(arg, "-") {
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
		return nil, fmt.Errorf("file path or directory is required\n\nUsage: specs validate <path> [--quiet] [--verbose] [--format json]\nExample: specs validate specs/eac-commands/specs/create/specification.feature")
	}

	config.Path = path

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}

	config.RepositoryRoot = repoRoot

	// Make path absolute if relative
	// Relative paths are interpreted relative to the current working directory,
	// not the repository root
	if !filepath.IsAbs(config.Path) {
		// First, check for R2R_PWD environment variable (used in isolated tests)
		// This allows tests to run from a different directory than the command invocation
		cwd := os.Getenv("R2R_PWD")
		if cwd == "" {
			// Fall back to actual working directory
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get current working directory: %w", err)
			}
		}
		config.Path = filepath.Clean(filepath.Join(cwd, config.Path))
	} else {
		// Even if path is absolute, clean it
		config.Path = filepath.Clean(config.Path)
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
func validateGherkinFile(filePath string, repoRoot string, checkTags bool) ([]contracts.ValidationError, error) {
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

	// Load tags config for advanced tag validation (only when checkTags is enabled)
	var tagsConfig *config.TestingTagsConfig
	if checkTags {
		cfg, cfgErr := config.Load(config.DefaultLoadOptions())
		if cfgErr == nil {
			tagsConfig = cfg.TestingTags
		}
		// Config load errors are ignored - validation continues without advanced tag checks
	}

	// Create validator with tags config support
	validator := contracts.NewGherkinValidatorWithTags(contractData, tagsConfig, antiCorruptionRules)

	// Validate content
	errors := validator.Validate(string(content), nil)

	return errors, nil
}

// validateDirectory validates all .feature files in a directory (recursive)
// Deprecated: Use validateDirectoryWithLogger for better logging support
func validateDirectory(dirPath string, repoRoot string, quiet bool, checkTags bool) ([]*ValidationResult, error) {
	return validateDirectoryWithLogger(dirPath, repoRoot, quiet, checkTags, nil)
}

// validateDirectoryWithLogger validates all .feature files in a directory with logging support
func validateDirectoryWithLogger(dirPath string, repoRoot string, quiet bool, checkTags bool, logger *logging.Logger) ([]*ValidationResult, error) {
	var results []*ValidationResult

	if logger != nil {
		logger.Debug("Walking directory for .feature files", zap.String("dir", dirPath))
	}

	// Walk directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if logger != nil {
				logger.Warn("Error accessing path", zap.String("path", path), zap.Error(err))
			}
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

		if logger != nil {
			logger.Debug("Validating file", zap.String("path", relativePath(path, repoRoot)))
		}

		// Validate file
		errors, validateErr := validateGherkinFile(path, repoRoot, checkTags)
		if validateErr != nil {
			// Log error but continue processing other files
			if logger != nil {
				logger.Warn("Failed to validate file",
					zap.String("path", path),
					zap.Error(validateErr))
			}
			log.Errorf("Warning: failed to validate %s: %v", path, validateErr)
			return nil
		}

		criticalCount := contracts.CountCriticalErrors(errors)
		result := &ValidationResult{
			Path:   path,
			Valid:  criticalCount == 0,
			Errors: errors,
		}

		results = append(results, result)

		if logger != nil {
			logger.Debug("File validation result",
				zap.String("path", relativePath(path, repoRoot)),
				zap.Bool("valid", result.Valid),
				zap.Int("criticalErrors", criticalCount),
				zap.Int("totalErrors", len(errors)))
		}

		// In quiet mode, only show progress for invalid files
		if !quiet || !result.Valid {
			if result.Valid && len(result.Errors) == 0 {
				log.Infof("✅ %s", relativePath(path, repoRoot))
			} else if result.Valid {
				log.Infof("✅ %s (%d warning(s))", relativePath(path, repoRoot), len(result.Errors))
			} else {
				log.Infof("❌ %s", relativePath(path, repoRoot))
			}
		}

		return nil
	})

	if err != nil {
		if logger != nil {
			logger.Error("Directory walk failed", zap.Error(err))
		}
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	if logger != nil {
		logger.Debug("Directory walk complete",
			zap.Int("filesProcessed", len(results)))
	}

	return results, nil
}

// outputText displays validation results in text format
func outputText(results []*ValidationResult, quiet bool, verbose bool) {
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
		log.Errorf("Error encoding JSON: %v", err)
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

// normalizePath converts a path to Unix-style relative path from repository root
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

// ============================================================================
// Fix functionality for --fix flag
// ============================================================================

// FixResult holds the result of fixing a single file
type FixResult struct {
	Path    string
	Fixes   []FixedIssue
	Error   error
}

// FixedIssue represents a single fix applied
type FixedIssue struct {
	Line        int
	Code        string
	Description string
}

// FixCount returns the number of fixes applied
func (r *FixResult) FixCount() int {
	return len(r.Fixes)
}

// fixGherkinFile attempts to fix issues in a feature file
// Supports fixing:
// - MISSING_VERIFICATION_TAG: adds @ov tag before scenario
// - INVALID_FEATURE_NAMING: renames feature to <module>_<kebab-name> format
func fixGherkinFile(filePath string, errors []contracts.ValidationError) (*FixResult, error) {
	result := &FixResult{
		Path:  filePath,
		Fixes: []FixedIssue{},
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		return result, result.Error
	}

	lines := strings.Split(string(content), "\n")
	modified := false

	// First pass: fix feature naming (do this first as it doesn't change line numbers)
	for _, e := range errors {
		if e.Code == "INVALID_FEATURE_NAMING" && e.Line > 0 {
			idx := e.Line - 1
			if idx >= 0 && idx < len(lines) {
				newFeatureName := generateFeatureName(filePath, lines[idx])
				if newFeatureName != "" {
					oldLine := lines[idx]
					lines[idx] = "Feature: " + newFeatureName
					result.Fixes = append(result.Fixes, FixedIssue{
						Line:        e.Line,
						Code:        "INVALID_FEATURE_NAMING",
						Description: fmt.Sprintf("Renamed feature to '%s'", newFeatureName),
					})
					modified = true
					_ = oldLine // suppress unused warning
				}
			}
		}
	}

	// Second pass: collect lines needing @ov insertion
	linesToFix := []int{}
	for _, e := range errors {
		if e.Code == "MISSING_VERIFICATION_TAG" && e.Line > 0 {
			linesToFix = append(linesToFix, e.Line)
		}
	}

	// Sort descending to insert from bottom up (preserves line numbers)
	if len(linesToFix) > 0 {
		sort.Sort(sort.Reverse(sort.IntSlice(linesToFix)))

		for _, lineNum := range linesToFix {
			idx := lineNum - 1 // 0-based index
			if idx >= 0 && idx < len(lines) {
				// Get indentation from scenario line
				indent := getIndentation(lines[idx])
				// Insert @ov tag before scenario
				newLine := indent + "@ov"
				lines = insertLine(lines, idx, newLine)
				result.Fixes = append(result.Fixes, FixedIssue{
					Line:        lineNum,
					Code:        "MISSING_VERIFICATION_TAG",
					Description: "Added @ov tag before Scenario",
				})
				modified = true
			}
		}
	}

	// Write back if modified
	if modified {
		err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
		if err != nil {
			result.Error = fmt.Errorf("failed to write file: %w", err)
			return result, result.Error
		}
	}

	return result, nil
}

// generateFeatureName creates a valid feature name from file path and current feature line
// Format: <module>_<kebab-case-description>
func generateFeatureName(filePath string, featureLine string) string {
	// Extract module from file path (e.g., specs/eac-core/logging/spec.feature -> eac-core)
	module := extractModuleFromPath(filePath)
	if module == "" {
		return ""
	}

	// Extract current feature name/description
	currentName := strings.TrimPrefix(strings.TrimSpace(featureLine), "Feature:")
	currentName = strings.TrimSpace(currentName)

	// Convert to kebab-case
	kebabName := toKebabCase(currentName)
	if kebabName == "" {
		return ""
	}

	return module + "_" + kebabName
}

// extractModuleFromPath extracts the module name from a spec file path
// e.g., specs/eac-core/logging/specification.feature -> eac-core
// e.g., specs/eac-commands/commit/specification.feature -> eac-commands
func extractModuleFromPath(filePath string) string {
	// Normalize path separators
	normalized := filepath.ToSlash(filePath)

	// Find "specs/" in the path
	specsIdx := strings.Index(normalized, "specs/")
	if specsIdx == -1 {
		return ""
	}

	// Get the part after "specs/"
	afterSpecs := normalized[specsIdx+6:]

	// Split by "/" and get the first component (module name)
	parts := strings.Split(afterSpecs, "/")
	if len(parts) < 1 {
		return ""
	}

	return parts[0]
}

// toKebabCase converts a string to kebab-case
// e.g., "Dual-output logging with configurable routing" -> "dual-output-logging-with-configurable-routing"
func toKebabCase(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	prevHyphen := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
			prevHyphen = false
		} else if r == '-' && !prevHyphen && result.Len() > 0 {
			result.WriteRune('-')
			prevHyphen = true
		}
	}

	// Remove trailing hyphen
	resultStr := result.String()
	resultStr = strings.TrimSuffix(resultStr, "-")

	return resultStr
}

// getIndentation returns the leading whitespace from a line
func getIndentation(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

// insertLine inserts a new line at the given index
func insertLine(lines []string, idx int, newLine string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = newLine
	return lines
}

// formatFixResult formats a fix result for display
func formatFixResult(result *FixResult, repoRoot string) string {
	if result.FixCount() == 0 {
		return ""
	}

	var output strings.Builder
	relPath := relativePath(result.Path, repoRoot)
	output.WriteString(fmt.Sprintf("🔧 Fixed %d issue(s) in %s:\n", result.FixCount(), relPath))

	// Show fixes in order (feature naming first, then tags in ascending line order)
	// First show non-tag fixes
	for _, fix := range result.Fixes {
		if fix.Code != "MISSING_VERIFICATION_TAG" {
			output.WriteString(fmt.Sprintf("   - Line %d: %s\n", fix.Line, fix.Description))
		}
	}

	// Then show tag fixes in reverse order (they were added bottom-up)
	for i := len(result.Fixes) - 1; i >= 0; i-- {
		fix := result.Fixes[i]
		if fix.Code == "MISSING_VERIFICATION_TAG" {
			output.WriteString(fmt.Sprintf("   - Line %d: %s\n", fix.Line, fix.Description))
		}
	}

	return output.String()
}
