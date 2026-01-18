// Command: validate specs
// Short: Validate Gherkin specifications against quality contracts
// Long: The validate specs command checks existing .feature files against the specification contract,
// Long: ensuring they follow proper Gherkin syntax, and project standards.
// Long: Validation covers structure (Feature/Rule/Scenario hierarchy), tags, step formatting, and content quality.
// Long: The command can validate a single file or recursively validate all .feature files in a directory.
// Long: By default, output is in human-readable text format. Use --format json for machine-readable output.
// Long:
// Long: Expected Output:
// Long:   Displays validation results for Gherkin specification structure, tags, and step formatting.
// Long:   Shows errors and warnings with line numbers. Exit code 0 if all pass, 1 if critical errors found.
// Flag.quiet: type=bool, shorthand=q, default=false, usage=Suppress success messages and show only validation errors and warnings
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed validation output including metadata and additional context
// Flag.format: type=string, shorthand=f, default=text, completion=text,json, usage=Output format for validation results (text for human-readable, json for machine-readable)
// Usage: validate specs <path> [--quiet] [--verbose] [--format json]
package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/validation/formats/gherkin"
)

func init() {
	registry.Register(ValidateSpecs)
}

// ============================================================================
// Mock Support for Testing
// ============================================================================

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var (
	gitRepo git.GitRepository
	gitMgr  *git.RepositoryManager
)

// initGitManager initializes the git repository manager if needed.
func initGitManager() {
	if gitMgr == nil {
		gitMgr = git.NewManager(logging.C().Zap())
	}
}

// getGitRepo returns the git repository, creating one if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	return gitMgr.Open(workspaceRoot)
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the mock git repository.
func ResetGitRepo() {
	gitRepo = nil
	gitMgr = nil
}

// ============================================================================

// ValidateSpecs validates existing Gherkin specification files.
func ValidateSpecs() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse configuration
	cfg, err := parseValidateConfig()
	if cfg == nil && err == nil {
		// --help was requested, framework handles it
		return 0
	}
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Initialize logger (logs to out/commands.log)
	// File logging is separate from stdout, so JSON output is not corrupted
	if err := logging.ConfigureLoggingSimple(cfg.RepositoryRoot, "commands", nil, cfg.Verbose); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Log command start
	log.Debugf("Starting specs validate: path=%s, format=%s, quiet=%v, verbose=%v", cfg.Path, cfg.Format, cfg.Quiet, cfg.Verbose)

	// Validate path security
	if err := validatePath(cfg.Path, cfg.RepositoryRoot); err != nil {
		log.Debugf("Path security validation failed: path=%s, error=%v", cfg.Path, err)
		log.Errorf("Error: %v", err)
		return 1
	}

	// Determine if path is file or directory
	info, err := os.Stat(cfg.Path)
	if err != nil {
		log.Debugf("File or directory not found: path=%s, error=%v", cfg.Path, err)
		log.Errorf("Error: file or directory not found: %s", cfg.Path)
		return 1
	}

	log.Debugf("Starting validation: path=%s, isDir=%v", cfg.Path, info.IsDir())

	var results []*ValidationResult
	if info.IsDir() {
		// Validate directory (recursive)
		results, err = validateDirectory(cfg.Path, cfg.RepositoryRoot, cfg.Quiet, cfg.CheckTags, cfg.Format)
		if err != nil {
			log.Debugf("Directory validation failed: error=%v", err)
			log.Errorf("Error: %v", err)
			return 1
		}
	} else {
		// Validate single file
		errors, err := validateGherkinFile(cfg.Path, cfg.RepositoryRoot, cfg.CheckTags)
		if err != nil {
			log.Debugf("File validation failed: path=%s, error=%v", cfg.Path, err)
			log.Errorf("Error: %v", err)
			return 1
		}

		// Apply fixes if --fix flag is set
		if cfg.Fix && len(errors) > 0 {
			fixResult, fixErr := fixGherkinFile(cfg.Path, errors)
			if fixErr != nil {
				log.Debugf("Failed to fix file: path=%s, error=%v", cfg.Path, fixErr)
				log.Errorf("Error fixing file: %v", fixErr)
			} else if fixResult.FixCount() > 0 {
				// Display fix results
				log.Info(formatFixResult(fixResult, cfg.RepositoryRoot))

				log.Debugf("Applied fixes: path=%s, fixCount=%d", cfg.Path, fixResult.FixCount())

				// Re-validate after fixes
				errors, err = validateGherkinFile(cfg.Path, cfg.RepositoryRoot, cfg.CheckTags)
				if err != nil {
					log.Debugf("Re-validation failed after fixes: path=%s, error=%v", cfg.Path, err)
					log.Errorf("Error re-validating after fixes: %v", err)
					return 1
				}
			}
		}

		criticalCount := contracts.CountCriticalErrors(errors)
		results = []*ValidationResult{
			{
				Path:   cfg.Path,
				Valid:  criticalCount == 0,
				Errors: errors,
			},
		}

		log.Debugf("Single file validation complete: path=%s, valid=%v, errorCount=%d", cfg.Path, criticalCount == 0, len(errors))
	}

	passed := countPassed(results)
	failed := countFailed(results)

	log.Debugf("Validation complete: total=%d, passed=%d, failed=%d", len(results), passed, failed)

	// Format and display output
	if cfg.Format == "json" {
		outputJSON(results)
	} else {
		outputText(results, cfg.Quiet, cfg.Verbose)
	}

	// Return exit code based on validation results
	if failed > 0 {
		return 1
	}

	return 0
}

// ValidateConfig holds configuration for specs validate command.
type ValidateConfig struct {
	Path           string
	Quiet          bool   // -q, --quiet: Show only errors
	Verbose        bool   // -v, --verbose: Detailed output
	Format         string // -f, --format: Output format (text, json)
	CheckTags      bool   // --check-tags: Enable tag validation (default: true)
	Fix            bool   // --fix: Auto-fix correctable tag issues
	RepositoryRoot string
}

// ValidationResult holds the validation result for a single file.
type ValidationResult struct {
	Path   string                      `json:"path"`
	Valid  bool                        `json:"valid"`
	Errors []contracts.ValidationError `json:"errors"`
}

// parseValidateConfig parses command line arguments into configuration.
func parseValidateConfig() (*ValidateConfig, error) {
	cfg := &ValidateConfig{
		Format:    "text", // Default format
		CheckTags: true,   // Default: tag validation enabled
	}

	args := os.Args[3:] // Skip program name, "validate", and "specs"
	var path string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			// Help is handled by the framework via command comments
			return nil, nil
		case "-q", "--quiet":
			cfg.Quiet = true
		case "-v", "--verbose":
			cfg.Verbose = true
		case "-f", "--format":
			if i+1 < len(args) {
				cfg.Format = args[i+1]
				i++
			} else {
				return nil, fmt.Errorf("--format requires an argument (text|json)")
			}
		case "--check-tags":
			cfg.CheckTags = true
		case "--no-check-tags":
			cfg.CheckTags = false
		case "--fix":
			cfg.Fix = true
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

	cfg.Path = path

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}

	cfg.RepositoryRoot = repoRoot

	// Make path absolute if relative
	// Relative paths are interpreted relative to the current working directory,
	// not the repository root
	if !filepath.IsAbs(cfg.Path) {
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
		cfg.Path = filepath.Clean(filepath.Join(cwd, cfg.Path))
	} else {
		// Even if path is absolute, clean it
		cfg.Path = filepath.Clean(cfg.Path)
	}

	// Validate format
	if cfg.Format != "text" && cfg.Format != "json" {
		return nil, fmt.Errorf("invalid format: %s (must be 'text' or 'json')", cfg.Format)
	}

	return cfg, nil
}

// validatePath ensures the path is within the repository (prevents path traversal attacks).
func validatePath(path, repoRoot string) error {
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

// validateGherkinFile validates a single Gherkin specification file.
func validateGherkinFile(filePath, repoRoot string, checkTags bool) ([]contracts.ValidationError, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Load contract and validator
	loader := ai.NewContractLoader(repoRoot, ai.TypeSpecs, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		return nil, fmt.Errorf("failed to load contract: %w", err)
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
	validator := gherkin.NewValidator(contractData, tagsConfig)

	// Validate content
	errors := validator.Validate(string(content), nil)

	return errors, nil
}

// validateDirectory validates all .feature files in a directory (recursive).
func validateDirectory(dirPath, repoRoot string, quiet, checkTags bool, format string) ([]*ValidationResult, error) {
	var results []*ValidationResult

	log.Debugf("Walking directory for .feature files: dir=%s", dirPath)

	// Walk directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debugf("Error accessing path: path=%s, error=%v", path, err)
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

		log.Debugf("Validating file: path=%s", relativePath(path, repoRoot))

		// Validate file
		errors, validateErr := validateGherkinFile(path, repoRoot, checkTags)
		if validateErr != nil {
			// Log error but continue processing other files
			log.Debugf("Failed to validate file: path=%s, error=%v", path, validateErr)
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

		log.Debugf("File validation result: path=%s, valid=%v, criticalErrors=%d, totalErrors=%d", relativePath(path, repoRoot), result.Valid, criticalCount, len(errors))

		// In quiet mode, only show progress for invalid files
		// Skip progress output entirely when format is JSON to avoid corrupting JSON output
		if format != "json" && (!quiet || !result.Valid) {
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
		log.Debugf("Directory walk failed: error=%v", err)
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	log.Debugf("Directory walk complete: filesProcessed=%d", len(results))

	return results, nil
}

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

// ============================================================================
// Fix functionality for --fix flag
// ============================================================================

// FixResult holds the result of fixing a single file.
type FixResult struct {
	Path  string
	Fixes []FixedIssue
	Error error
}

// FixedIssue represents a single fix applied.
type FixedIssue struct {
	Line        int
	Code        string
	Description string
}

// FixCount returns the number of fixes applied.
func (r *FixResult) FixCount() int {
	return len(r.Fixes)
}

// fixGherkinFile attempts to fix issues in a feature file
// Supports fixing:
// - MISSING_VERIFICATION_TAG: adds @ov tag before scenario
// - INVALID_FEATURE_NAMING: renames feature to <module>_<kebab-name> format.
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
		if e.GetCode() == "INVALID_FEATURE_NAMING" && e.Line > 0 {
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
		if e.GetCode() == "MISSING_VERIFICATION_TAG" && e.Line > 0 {
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
		err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
		if err != nil {
			result.Error = fmt.Errorf("failed to write file: %w", err)
			return result, result.Error
		}
	}

	return result, nil
}

// generateFeatureName creates a valid feature name from file path and current feature line
// Format: <module>_<kebab-case-description>.
func generateFeatureName(filePath, featureLine string) string {
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
// e.g., specs/eac-commands/commit/specification.feature -> eac-commands.
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
// e.g., "Dual-output logging with configurable routing" -> "dual-output-logging-with-configurable-routing".
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

// getIndentation returns the leading whitespace from a line.
func getIndentation(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	return line[:len(line)-len(trimmed)]
}

// insertLine inserts a new line at the given index.
func insertLine(lines []string, idx int, newLine string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = newLine
	return lines
}

// formatFixResult formats a fix result for display.
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
