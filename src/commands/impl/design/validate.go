// Command: design validate
// Description: Validate workspace files using Structurizr CLI
// HasSideEffects: false
package design

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/design/internal"
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
	fmt.Println("Validate Structurizr workspace files using Structurizr CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . design validate <module>    Validate one module")
	fmt.Println("  go run . design validate --all       Validate all modules")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --all, -a        Validate all modules with workspace files")
	fmt.Println("  --verbose, -v    Show detailed validation output (Docker command, raw CLI output)")
	fmt.Println("  --help, -h       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate single module")
	fmt.Println("  go run . design validate src-cli")
	fmt.Println()
	fmt.Println("  # Validate with verbose output")
	fmt.Println("  go run . design validate src-cli --verbose")
	fmt.Println()
	fmt.Println("  # Validate all modules")
	fmt.Println("  go run . design validate --all")
	fmt.Println()
	fmt.Println("  # Validate all with verbose output")
	fmt.Println("  go run . design validate --all --verbose")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  - Console: Human-readable validation results")
	fmt.Println("  - JSON file: out/design-validation-results.json (contains raw_output field)")
	fmt.Println()
	fmt.Println("Verbose Mode Details:")
	fmt.Println("  - Shows Docker command executed")
	fmt.Println("  - Shows raw Structurizr CLI output")
	fmt.Println("  - Shows workspace file paths")
	fmt.Println("  - Shows execution timestamps")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - Docker must be running")
	fmt.Println("  - Structurizr CLI image will be pulled automatically if not present")
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
