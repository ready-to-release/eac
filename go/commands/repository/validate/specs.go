// Usage: validate specs <path> [--quiet] [--verbose] [--format json]
package validate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type validateSpecsCommand struct{}

var _ core.SimpleCommandPort = (*validateSpecsCommand)(nil)

func (c *validateSpecsCommand) Name() string { return "validate specs" }

func (c *validateSpecsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-specs",
		Short:         "Validate Gherkin specifications against quality contracts",
		Long: "The validate specs command checks existing .feature files against the specification contract,\nensuring they follow proper Gherkin syntax, and project standards.\nValidation covers structure (Feature/Rule/Scenario hierarchy), tags, step formatting, and content quality.\nThe command can validate a single file or recursively validate all .feature files in a directory.\nBy default, output is in human-readable text format. Use --format json for machine-readable output.",
		Notes: "Expected Output:\n  Displays validation results for Gherkin specification structure, tags, and step formatting.\n  Shows errors and warnings with line numbers. Exit code 0 if all pass, 1 if critical errors found.",
		Flags: []core.FlagSpec{
			{Name: "quiet", Shorthand: "q", Type: "bool", DefaultValue: "false", Usage: "Suppress success messages and show only validation errors and warnings"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed validation output including metadata and additional context"},
			{Name: "format", Shorthand: "f", Type: "string", DefaultValue: "text", Usage: "Output format for validation results (text for human-readable, json for machine-readable)"},
		},
	}
}

func (c *validateSpecsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateSpecs()
}

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

		criticalCount := domain.CountCriticalErrors(errors)
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
	Path   string                   `json:"path"`
	Valid  bool                     `json:"valid"`
	Errors []domain.ValidationError `json:"errors"`
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
		return nil, fmt.Errorf("file path or directory is required\n\nUsage: specs validate <path> [--quiet] [--verbose] [--format json]\nExample: specs validate specs/eac/specs/create/specification.feature")
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
		// First, check for CLIE_PWD environment variable (used in isolated tests)
		// This allows tests to run from a different directory than the command invocation
		cwd := os.Getenv(environments.EnvCLIEPWD)
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
