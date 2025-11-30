// Command: validate design
// Description: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Short: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Long: Validates workspace.dsl files for syntax errors and structural issues using the official
// Long: Structurizr CLI running in Docker. Checks DSL syntax, element relationships, view definitions,
// Long: and ensures the workspace can be properly rendered. Validation results are displayed in the
// Long: console with human-readable output and saved to out/logs/design/validation-results.json for
// Long: detailed inspection. Use --all to validate all workspace files in specs/*/.design/ directories.
// Usage: validate design <module>
// Flag.all: type=bool, shorthand=a, default=false, usage=Validate all workspace files in specs/*/.design/ directories
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs and detailed logs to out/logs/design/ for debugging
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show Docker command and raw Structurizr CLI output
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	designInternal "github.com/ready-to-release/eac/src/commands/impl/design/helper"
	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(ValidateDesign)
}

// ValidateDesign validates workspace files using Structurizr CLI
func ValidateDesign() int {
	args := os.Args[3:] // Skip program, "validate", and "design"

	var module string
	var all bool
	var verbose bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--all", "-a":
			all = true
		case "--debug", "-d":
			// debug flag accepted but ignored (no logger)
		case "--verbose", "-v":
			verbose = true
		case "--help", "-h":
			printDesignValidateUsage()
			return 0
		default:
			if arg[0] != '-' {
				module = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printDesignValidateUsage()
				return 1
			}
		}
	}

	// Initialize logger
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 2
	}


	// Create validator
	validator, err := designInternal.NewValidator()
	if err != nil {
		fmt.Printf("❌ Failed to initialize validator: %v\n", err)
		return 1
	}

	// Check Docker running
	validatorImpl, ok := validator.(*designInternal.StructurizrValidatorImpl)
	if ok && !validatorImpl.IsDockerRunning() {
		fmt.Println("❌ Error: Docker is not running")
		fmt.Println()
		fmt.Println("Docker is required to run validation. Please:")
		fmt.Println("  1. Start Docker Desktop (Windows/Mac)")
		fmt.Println("  2. Or start Docker daemon: sudo systemctl start docker (Linux)")
		fmt.Println("  3. Verify with: docker ps")
		fmt.Println()
		fmt.Println("Note: Docker is also required for 'design serve' command.")
		return 2
	}

	// Determine output path (use /out/logs/design directory)
	outputPath, err := getValidationOutputPath(repoRoot)
	if err != nil {
		fmt.Printf("❌ Failed to determine output path: %v\n", err)
		return 2
	}

	if all {
		// Validate all modules
		return validateAllModules(validator, outputPath, verbose)
	} else if module != "" {
		// Validate single module
		return validateSingleModule(validator, module, outputPath, verbose)
	} else {
		fmt.Println("❌ Error: module name required or use --all flag")
		fmt.Println()
		printDesignValidateUsage()
		return 2
	}
}

func validateSingleModule(validator designInternal.StructurizrValidator, module string, outputPath string, verbose bool) int {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to find repository root: %v\n", err)
		return 2
	}

	// Validate module name for security
	if err := designInternal.ValidateModuleName(module); err != nil {
		fmt.Printf("❌ Invalid module name: %v\n", err)
		return 2
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load module contracts: %v\n", err)
		return 2
	}

	mod, exists := moduleReport.Registry.Get(module)
	if !exists {
		fmt.Fprintf(os.Stderr, "❌ Module not found: %s\n\nAvailable modules:\n%s\n",
			module, formatDesignModuleList(moduleReport))
		return 2
	}

	// Use validated moniker
	module = mod.Moniker

	// Validate module
	result, err := validator.ValidateModule(module)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return 2
	}


	// Display console output
	fmt.Print(designInternal.FormatValidationResult(result, verbose))

	// Write JSON file
	if err := designInternal.WriteValidationResultJSON(result, outputPath); err != nil {
		fmt.Printf("\n⚠️  Failed to write JSON file: %v\n", err)
	} else {
		fmt.Printf("\n📝 Results written to: %s\n", outputPath)
		if verbose {
			fmt.Printf("💡 View detailed output in JSON: %s\n", outputPath)
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
		fmt.Printf("❌ Validation failed: %v\n", err)
		return 2
	}

	// Display console output
	fmt.Print(designInternal.FormatValidationSummary(summary, verbose))

	// Write JSON file
	if err := designInternal.WriteValidationSummaryJSON(summary, outputPath); err != nil {
		fmt.Printf("\n⚠️  Failed to write JSON file: %v\n", err)
	} else {
		fmt.Printf("\n📝 Results written to: %s\n", outputPath)
		if verbose {
			fmt.Printf("💡 View detailed output in JSON: %s\n", outputPath)
		}
	}

	// Return exit code
	if summary.FailedModules > 0 {
		return 1
	}
	return 0
}

func printDesignValidateUsage() {
	fmt.Println("Validate workspace.dsl syntax using Structurizr CLI")
	fmt.Println()
	fmt.Println("Checks workspace.dsl files for syntax errors and structural issues.")
	fmt.Println("Runs validation in Docker using the official Structurizr CLI image.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  r2r design validate <module>   Validate one module")
	fmt.Println("  r2r design validate --all      Validate all modules")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --all, -a        Validate all workspace files in specs/*/.design/")
	fmt.Println("  --debug, -d      Save intermediate outputs and detailed logs to out/logs/design/")
	fmt.Println("  --verbose, -v    Show Docker command and raw Structurizr CLI output")
	fmt.Println("  --help, -h       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r design validate src-cli")
	fmt.Println("  r2r design validate src-commands --verbose")
	fmt.Println("  r2r design validate --all --debug")
	fmt.Println()
	fmt.Println("Module Locations:")
	fmt.Println("  src-cli        → specs/src-cli/.design/workspace.dsl")
	fmt.Println("  src-commands   → specs/src-commands/.design/workspace.dsl")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Console: Human-readable validation summary")
	fmt.Println("  File:    out/logs/design/validation-results.json (detailed results)")
	fmt.Println("  Logs:    out/logs/design/designInternal.log (when --debug is set)")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  Module argument must be a valid module moniker (e.g., src-commands).")
	fmt.Println("  Use 'show modules' to see all available modules.")
}

// getValidationOutputPath returns the absolute path to the validation output JSON file
func getValidationOutputPath(repoRoot string) (string, error) {
	// Use out/logs/design directory
	outDir := filepath.Join(repository.LogsPath(repoRoot), "design")

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
