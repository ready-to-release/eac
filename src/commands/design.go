// Command: design
// Description: Architecture documentation tools using Structurizr
// Usage: design new <module> --name <name> --description <description>
//        design add container <module> <name> --tech <tech> --desc <desc>
//        design add relationship <module> <source> <dest> --desc <desc> [--tech <tech>]
//        design export <module> [--output <file>]
//        design serve <module> [--no-auto-open-link] [--port <port>] [--stop]
//        design list
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"github.com/ready-to-release/eac/src/commands/design"
)

func init() {
	Register("design", Design)
	Register("design serve", DesignServe)
	Register("design list", DesignList)
	Register("design validate", DesignValidate)
	// New authoring commands are registered in their respective files:
	// - design-new.go
	// - design-add-container.go
	// - design-add-relationship.go
	// - design-export.go
}

// Design command entry point
func Design() int {
	args := os.Args[2:] // Skip "go" and "run" and "."

	if len(args) == 0 {
		printDesignUsage()
		return 1
	}

	// Check for subcommands
	switch args[0] {
	case "new", "add", "export":
		// Handled by separate registrations in respective files
		return 0
	case "serve":
		// Handled by separate registration
		return 0
	case "list":
		// Handled by separate registration
		return 0
	case "validate":
		// Handled by separate registration
		return 0
	case "--help", "-h":
		printDesignUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printDesignUsage()
		return 1
	}
}

// DesignServe starts or stops Structurizr server for a module
func DesignServe() int {
	args := os.Args[3:] // Skip "go", "run", ".", and "design" and "serve"

	var module string
	var noAutoOpenLink bool
	var port int = 8081
	var stop bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "--no-auto-open-link":
			noAutoOpenLink = true
		case "--stop":
			stop = true
		case "--port", "-p":
			if i+1 < len(args) {
				i++
				p, err := strconv.Atoi(args[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid port number: %s\n", args[i])
					return 1
				}
				port = p
			} else {
				fmt.Fprintf(os.Stderr, "Error: --port requires a value\n")
				return 1
			}
		default:
			if arg[0] != '-' {
				module = arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				return 1
			}
		}
	}

	// Handle --stop flag
	if stop {
		return handleStop(module)
	}

	// Require module argument for starting server
	if module == "" {
		fmt.Println("❌ Error: module name required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run . design serve <module>")
		fmt.Println("\nRun 'go run . design list' to see available modules")
		return 1
	}

	// Create client
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	// Validate module
	err = client.ValidateModule(module)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println("\n💡 Tip: Run 'go run . design list' to see available modules")
		return 1
	}

	// Check if already running
	running, info, err := client.IsRunning(module)
	if err != nil {
		fmt.Printf("❌ Failed to check container status: %v\n", err)
		return 1
	}

	if running && info != nil {
		fmt.Printf("ℹ️  Structurizr is already running for module: %s\n", module)
		fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

		if !noAutoOpenLink {
			err = client.OpenBrowser(info.URL)
			if err != nil {
				fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
				fmt.Printf("📖 Please open manually: %s\n", info.URL)
			}
		}
		return 0
	}

	// Start container
	fmt.Printf("🚀 Starting Structurizr Lite for module: %s\n", module)

	info, err = client.StartContainer(module, port)
	if err != nil {
		if info != nil {
			fmt.Printf("⚠️  %v\n", err)
			fmt.Printf("📖 Try accessing manually: %s\n", info.URL)
		} else {
			fmt.Printf("❌ Failed to start container: %v\n", err)
			return 1
		}
	}

	// Get module details
	moduleInfo, err := client.GetModuleInfo(module)
	if err != nil {
		// Ignore errors - not critical
	}

	// Display success
	fmt.Printf("\n✅ Structurizr Lite is running for module: %s\n", module)
	fmt.Printf("📊 Architecture documentation: %s\n", info.URL)

	if moduleInfo != nil {
		if moduleInfo.ViewCount > 0 {
			fmt.Printf("\n📈 Available views: %d\n", moduleInfo.ViewCount)
		}
		if moduleInfo.HasDocs {
			fmt.Printf("📚 Documentation sections: %d\n", moduleInfo.DocCount)
		}
		if moduleInfo.HasDecisions {
			fmt.Printf("🎯 Architecture Decisions: %d ADRs\n", moduleInfo.DecisionCount)
		}

		if !moduleInfo.HasDocs {
			fmt.Println("\n⚠️  Module has no documentation sections")
		}
		if !moduleInfo.HasDecisions {
			fmt.Println("⚠️  Module has no architecture decisions")
		}
	}

	// Open browser
	if !noAutoOpenLink {
		err = client.OpenBrowser(info.URL)
		if err != nil {
			fmt.Printf("\n⚠️  Failed to open browser: %v\n", err)
			fmt.Printf("📖 Please open manually: %s\n", info.URL)
		}
	}

	// Show tips
	fmt.Println("\n💡 Tips:")
	fmt.Println("  • Container will keep running until stopped")
	fmt.Printf("  • Stop with: go run . design serve cli --stop\n")
	fmt.Printf("  • Or: docker stop %s\n", info.Name)
	fmt.Printf("  • View logs: docker logs %s\n", info.Name)

	return 0
}

// DesignList lists available modules with architecture documentation
func DesignList() int {
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		return 1
	}

	if len(modules) == 0 {
		fmt.Println("ℹ️  No modules with architecture documentation found")
		fmt.Println("\nExpected location: docs/reference/design/<module>/workspace.dsl")
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("Available modules with architecture documentation:")
	fmt.Println()
	fmt.Fprintln(w, "MODULE\tSTATUS\tVIEWS\tDOCS\tDECISIONS\tPATH")
	fmt.Fprintln(w, "──────\t──────\t─────\t────\t─────────\t────")

	for _, module := range modules {
		docsCount := "-"
		if module.HasDocs {
			docsCount = fmt.Sprintf("%d", module.DocCount)
		}

		decisionsCount := "-"
		if module.HasDecisions {
			decisionsCount = fmt.Sprintf("%d", module.DecisionCount)
		}

		viewsCount := fmt.Sprintf("%d", module.ViewCount)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			module.Name,
			module.GetStatus(),
			viewsCount,
			docsCount,
			decisionsCount,
			module.Path,
		)
	}

	w.Flush()

	fmt.Println("\n💡 To view documentation:")
	fmt.Println("  go run . design serve <module>")
	return 0
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

func handleStop(module string) int {
	client, err := design.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize: %v\n", err)
		return 1
	}
	defer client.Close()

	// If module specified, stop that container
	if module != "" {
		err = client.StopContainer(module)
		if err != nil {
			fmt.Printf("❌ Failed to stop container: %v\n", err)
			return 1
		}
		fmt.Printf("✅ Structurizr container stopped for module: %s\n", module)
		return 0
	}

	// Otherwise, stop all running containers
	modules, err := client.ListModules()
	if err != nil {
		fmt.Printf("❌ Failed to list modules: %v\n", err)
		return 1
	}

	stoppedAny := false
	for _, mod := range modules {
		running, _, err := client.IsRunning(mod.Name)
		if err != nil {
			continue
		}

		if running {
			err = client.StopContainer(mod.Name)
			if err != nil {
				fmt.Printf("⚠️  Failed to stop container for %s: %v\n", mod.Name, err)
			} else {
				fmt.Printf("✅ Stopped container for module: %s\n", mod.Name)
				stoppedAny = true
			}
		}
	}

	if !stoppedAny {
		fmt.Println("ℹ️  No running Structurizr containers found")
	}

	return 0
}

func printDesignUsage() {
	fmt.Println("Architecture documentation tools using Structurizr")
	fmt.Println()
	fmt.Println("Usage: go run . design <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Authoring Subcommands:")
	fmt.Println("  new <module>              Create new architecture workspace")
	fmt.Println("  add container             Add container to workspace")
	fmt.Println("  add relationship          Add relationship between containers")
	fmt.Println("  export <module>           Export workspace DSL content")
	fmt.Println()
	fmt.Println("Viewing Subcommands:")
	fmt.Println("  serve <module>            Start Structurizr Lite viewer")
	fmt.Println("  list                      List available modules with documentation")
	fmt.Println()
	fmt.Println("Validation Subcommands:")
	fmt.Println("  validate <module>         Validate workspace file for one module")
	fmt.Println("  validate --all            Validate workspace files for all modules")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create workspace")
	fmt.Println("  go run . design new src-cli --name \"CLI\" --description \"CLI Architecture\"")
	fmt.Println()
	fmt.Println("  # Add containers")
	fmt.Println("  go run . design add container src-cli parser --tech \"Go\" --desc \"Parses commands\"")
	fmt.Println()
	fmt.Println("  # Add relationships")
	fmt.Println("  go run . design add relationship src-cli parser executor --desc \"sends to\"")
	fmt.Println()
	fmt.Println("  # Export workspace")
	fmt.Println("  go run . design export src-cli")
	fmt.Println()
	fmt.Println("  # View in browser")
	fmt.Println("  go run . design serve src-cli")
	fmt.Println()
	fmt.Println("  # List modules")
	fmt.Println("  go run . design list")
	fmt.Println()
	fmt.Println("  # Validate workspace")
	fmt.Println("  go run . design validate src-cli")
	fmt.Println("  go run . design validate --all")
	fmt.Println()
	fmt.Println("For detailed help on a subcommand:")
	fmt.Println("  go run . design <subcommand> --help")
}
