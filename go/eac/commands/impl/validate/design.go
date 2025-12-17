// Command: validate design
// Description: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Short: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Long: Validates DSL files for syntax errors and structural issues using the official
// Long: Structurizr CLI running in Docker. Checks DSL syntax, element relationships, view definitions,
// Long: and ensures the workspace can be properly rendered. Supports multiple DSL files per module -
// Long: files starting with "_" are treated as fragments (for !include) and skipped.
// Long: Use --all to validate all modules, or --file to validate a specific DSL file.
// Long:
// Long: Expected Output:
// Long:   Displays validation results in console. Results saved to out/logs/design/validation-results.json.
// Long:   Shows syntax errors, structural issues, and render status. Exit code 0 if valid, 1 if errors, 2 if system errors.
// Usage: validate design <module> [--file=<name>]
// Flag.all: type=bool, shorthand=a, default=false, usage=Validate all workspace files in specs/*/.design/ directories
// Flag.file: type=string, shorthand=f, default="", usage=Validate only a specific DSL file (e.g., --file=landscape)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs and detailed logs to out/logs/design/ for debugging
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show Docker command and raw Structurizr CLI output
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	designInternal "github.com/ready-to-release/eac/go/eac/commands/impl/design/helper"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateDesign)
}

// ValidateDesign validates workspace files using Structurizr CLI
func ValidateDesign() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program, "validate", and "design"

	var module string
	var file string

	// Use shared flags package for boolean flags
	all := flags.HasFlag(args, "--all", "-a")
	verbose := flags.HasFlag(args, "--verbose", "-v")
	debug := flags.ParseDebugFlag(args) // Accepted but currently unused

	// Parse value flags and positional args
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			printDesignValidateUsage()
			return 0
		case strings.HasPrefix(arg, "--file="):
			file = strings.TrimPrefix(arg, "--file=")
		case strings.HasPrefix(arg, "-f="):
			file = strings.TrimPrefix(arg, "-f=")
		case (arg == "--file" || arg == "-f") && i+1 < len(args):
			i++
			file = args[i]
		case !strings.HasPrefix(arg, "-"):
			module = arg
		case arg == "-a" || arg == "--all" || arg == "-v" || arg == "--verbose" || arg == "-d" || arg == "--debug":
			// Already handled by shared flags package
		default:
			log.Errorf("unknown flag: %s", arg)
			printDesignValidateUsage()
			return 1
		}
	}
	_ = debug // Suppress unused variable warning

	// Initialize logger
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 2
	}


	// Create validator
	validator, err := designInternal.NewValidator()
	if err != nil {
		log.Infof("❌ Failed to initialize validator: %v", err)
		return 1
	}

	// Check Docker running
	validatorImpl, ok := validator.(*designInternal.StructurizrValidatorImpl)
	if ok && !validatorImpl.IsDockerRunning() {
		log.Info("❌ Error: Docker is not running")
		log.Info("")
		log.Info("Docker is required to run validation. Please:")
		log.Info("  1. Start Docker Desktop (Windows/Mac)")
		log.Info("  2. Or start Docker daemon: sudo systemctl start docker (Linux)")
		log.Info("  3. Verify with: docker ps")
		log.Info("")
		log.Info("Note: Docker is also required for 'design serve' command.")
		return 2
	}

	// Determine output path (use /out/logs/design directory)
	outputPath, err := getValidationOutputPath(repoRoot)
	if err != nil {
		log.Infof("❌ Failed to determine output path: %v", err)
		return 2
	}

	if all {
		// Validate all modules
		return validateAllModules(validator, outputPath, verbose)
	} else if module != "" {
		// Validate single module (optionally a specific file)
		return validateSingleModule(validator, module, file, outputPath, verbose)
	} else {
		log.Info("❌ Error: module name required or use --all flag")
		log.Info("")
		printDesignValidateUsage()
		return 2
	}
}

func validateSingleModule(validator designInternal.StructurizrValidator, module string, file string, outputPath string, verbose bool) int {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("❌ Failed to find repository root: %v", err)
		return 2
	}

	// Validate module name for security
	if err := designInternal.ValidateModuleName(module); err != nil {
		log.Infof("❌ Invalid module name: %v", err)
		return 2
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		log.Errorf("❌ Failed to load module contracts: %v", err)
		return 2
	}

	mod, exists := moduleReport.Registry.Get(module)
	if !exists {
		log.Errorf("❌ Module not found: %s\n\nAvailable modules:\n%s",
			module, formatDesignModuleList(moduleReport))
		return 2
	}

	// Use validated moniker
	module = mod.Moniker

	// Validate module (all files or specific file)
	var result *designInternal.ValidationResult
	if file != "" {
		// Validate specific file
		result, err = validator.ValidateModuleFile(module, file)
	} else {
		// Validate all files in module
		result, err = validator.ValidateModule(module)
	}
	if err != nil {
		log.Infof("❌ Validation failed: %v", err)
		return 2
	}


	// Display console output
	log.Info(designInternal.FormatValidationResult(result, verbose))

	// Write JSON file
	if err := designInternal.WriteValidationResultJSON(result, outputPath); err != nil {
		log.Infof("\n⚠️  Failed to write JSON file: %v", err)
	} else {
		log.Infof("\n📝 Results written to: %s", outputPath)
		if verbose {
			log.Infof("💡 View detailed output in JSON: %s", outputPath)
		}
	}

	// Return exit code
	if result.Valid {
		return 0
	}
	return 1
}

func validateAllModules(validator designInternal.StructurizrValidator, outputPath string, verbose bool) int {
	// Validate all modules
	summary, err := validator.ValidateAll()
	if err != nil {
		log.Infof("❌ Validation failed: %v", err)
		return 2
	}

	// Display console output
	log.Info(designInternal.FormatValidationSummary(summary, verbose))

	// Write JSON file
	if err := designInternal.WriteValidationSummaryJSON(summary, outputPath); err != nil {
		log.Infof("\n⚠️  Failed to write JSON file: %v", err)
	} else {
		log.Infof("\n📝 Results written to: %s", outputPath)
		if verbose {
			log.Infof("💡 View detailed output in JSON: %s", outputPath)
		}
	}

	// Return exit code
	if summary.FailedModules > 0 {
		return 1
	}
	return 0
}

func printDesignValidateUsage() {
	log.Info("Validate DSL files using Structurizr CLI")
	log.Info("")
	log.Info("Validates all DSL files in a module's .design folder for syntax errors")
	log.Info("and structural issues. Files starting with '_' are treated as fragments")
	log.Info("(for !include) and skipped. Runs validation in Docker using Structurizr CLI.")
	log.Info("")
	log.Info("Usage:")
	log.Info("  r2r validate design <module>              Validate all DSL files in module")
	log.Info("  r2r validate design <module> --file=NAME  Validate specific DSL file")
	log.Info("  r2r validate design --all                 Validate all modules")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --all, -a           Validate all DSL files in specs/*/.design/")
	log.Info("  --file, -f <name>   Validate only a specific DSL file (e.g., --file=landscape)")
	log.Info("  --debug, -d         Save intermediate outputs and detailed logs")
	log.Info("  --verbose, -v       Show Docker command and raw Structurizr CLI output")
	log.Info("  --help, -h          Show this help message")
	log.Info("")
	log.Info("Examples:")
	log.Info("  r2r validate design r2r-cli                    # All DSL files")
	log.Info("  r2r validate design r2r-cli --file=workspace   # Just workspace.dsl")
	log.Info("  r2r validate design r2r-cli --file=landscape   # Just landscape.dsl")
	log.Info("  r2r validate design eac-commands --verbose")
	log.Info("  r2r validate design --all")
	log.Info("")
	log.Info("Multi-DSL Support:")
	log.Info("  specs/module/.design/")
	log.Info("    workspace.dsl      # Main module design (validated)")
	log.Info("    landscape.dsl      # Cross-module view (validated)")
	log.Info("    _model.dsl         # Fragment for !include (skipped)")
	log.Info("    _styles.dsl        # Shared styles (skipped)")
	log.Info("")
	log.Info("Output:")
	log.Info("  Console: Human-readable validation summary")
	log.Info("  File:    out/logs/design/validation-results.json")
}

// getValidationOutputPath returns the absolute path to the validation output JSON file
func getValidationOutputPath(repoRoot string) (string, error) {
	// Use out/logs/design directory
	outDir := filepath.Join(paths.LogsPath(repoRoot), "design")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return filepath.Join(outDir, "validation-results.json"), nil
}

// formatDesignModuleList returns a formatted list of available modules
func formatDesignModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, mod.Files.Root))
	}
	return sb.String()
}
