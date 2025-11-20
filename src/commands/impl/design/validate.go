// Command: design validate
// Description: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Short: Check workspace.dsl syntax using Structurizr CLI (requires Docker)
// Long: Validates workspace.dsl files for syntax errors and structural issues using the official
// Long: Structurizr CLI running in Docker. Checks DSL syntax, element relationships, view definitions,
// Long: and ensures the workspace can be properly rendered. Validation results are displayed in the
// Long: console with human-readable output and saved to out/design-validation-results.json for detailed
// Long: inspection. Use --all to validate all workspace files in specs/*/design/ directories.
// Usage: design validate <module>
// Flag.all: type=bool, shorthand=a, default=false, usage=Validate all workspace files in specs/*/design/ directories
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show Docker command and raw Structurizr CLI output
// HasSideEffects: false
package design

import (
	"fmt"
	"os"
	"path/filepath"

	design "github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(DesignValidate)
}

// DesignValidate validates workspace files using Structurizr CLI
func DesignValidate() int {
	args := os.Args[3:] // Skip "go", "run", ".", "design", and "validate"

	var module string
	var all bool
	var verbose bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--all", "-a":
			all = true
		case "--verbose", "-v":
			verbose = true
		case "--help", "-h":
			printValidateUsage()
			return 0
		default:
			if arg[0] != '-' {
				module = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printValidateUsage()
				return 1
			}
		}
	}

	// Create validator
	validator, err := design.NewValidator()
	if err != nil {
		fmt.Printf("❌ Failed to initialize validator: %v\n", err)
		return 1
	}

	// Check Docker running
	validatorImpl, ok := validator.(*design.StructurizrValidatorImpl)
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

	// Determine output path (use root /out directory)
	outputPath, err := getValidationOutputPath()
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
		printValidateUsage()
		return 2
	}
}

func validateSingleModule(validator design.StructurizrValidator, module string, outputPath string, verbose bool) int {
	// Clean up module path - remove specs/ prefix and /design suffix if present
	module = design.CleanModuleName(module)

	// Validate module name
	if err := design.ValidateModuleName(module); err != nil {
		fmt.Printf("❌ Invalid module name: %v\n", err)
		return 2
	}

	// Validate module
	result, err := validator.ValidateModule(module)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return 2
	}

	// Display console output
	fmt.Print(design.FormatValidationResult(result, verbose))

	// Write JSON file
	if err := design.WriteValidationResultJSON(result, outputPath); err != nil {
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

func validateAllModules(validator design.StructurizrValidator, outputPath string, verbose bool) int {
	// Validate all modules
	summary, err := validator.ValidateAll()
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return 2
	}

	// Display console output
	fmt.Print(design.FormatValidationSummary(summary, verbose))

	// Write JSON file
	if err := design.WriteValidationSummaryJSON(summary, outputPath); err != nil {
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

func printValidateUsage() {
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
	fmt.Println("  --all, -a        Validate all workspace files in specs/*/design/")
	fmt.Println("  --verbose, -v    Show Docker command and raw Structurizr CLI output")
	fmt.Println("  --help, -h       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  r2r design validate src-cli")
	fmt.Println("  r2r design validate contracts --verbose")
	fmt.Println("  r2r design validate specs/src-cli/design     (auto-cleaned)")
	fmt.Println("  r2r design validate --all")
	fmt.Println()
	fmt.Println("Module Locations:")
	fmt.Println("  src-cli     → specs/src-cli/design/workspace.dsl")
	fmt.Println("  contracts   → specs/contracts/design/workspace.dsl")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Console: Human-readable validation summary")
	fmt.Println("  File:    out/design-validation-results.json (detailed results)")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  Accepts module name with or without 'specs/' prefix and '/design' suffix.")
}

// getValidationOutputPath returns the absolute path to the validation output JSON file
func getValidationOutputPath() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find repository root (directory containing "src")
	dir := cwd
	for {
		srcPath := filepath.Join(dir, "src")
		if stat, err := os.Stat(srcPath); err == nil && stat.IsDir() {
			// Found repository root
			outDir := filepath.Join(dir, "out")

			// Create out directory if it doesn't exist
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create output directory: %w", err)
			}

			return filepath.Join(outDir, "design-validation-results.json"), nil
		}

		// Check if we're in a src subdirectory
		base := filepath.Base(dir)
		parent := filepath.Dir(dir)

		if base == "design" || base == "commands" || base == "cli" {
			grandparent := filepath.Dir(parent)
			if filepath.Base(parent) == "src" {
				outDir := filepath.Join(grandparent, "out")
				if err := os.MkdirAll(outDir, 0755); err != nil {
					return "", fmt.Errorf("failed to create output directory: %w", err)
				}
				return filepath.Join(outDir, "design-validation-results.json"), nil
			}
		}

		if base == "src" {
			outDir := filepath.Join(parent, "out")
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create output directory: %w", err)
			}
			return filepath.Join(outDir, "design-validation-results.json"), nil
		}

		// Move up one directory
		nextDir := filepath.Dir(dir)
		if nextDir == dir {
			// Reached filesystem root, use current directory
			outDir := filepath.Join(cwd, "out")
			if err := os.MkdirAll(outDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create output directory: %w", err)
			}
			return filepath.Join(outDir, "design-validation-results.json"), nil
		}
		dir = nextDir
	}
}
