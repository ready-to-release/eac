// Command: build
// Short: Build one or more modules by moniker
// Long: Build one or more modules by moniker.
// Long:
// Long: This command builds modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are built.
// Long:
// Long: Build results are collected in 'out/build/' with per-module logs and
// Long: a summary orchestrator log. Failed builds are clearly marked but do not
// Long: stop the execution of remaining modules.
// Long:
// Long: Example:
// Long:   build                           # Build all modules
// Long:   build src-commands              # Build a single module
// Long:   build src-core src-cli          # Build specific modules
// Long:   build --tidy-first src-commands # Build with go mod tidy first
// HasSideEffects: true
// Args: modules
package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/platform"
	"github.com/ready-to-release/eac/src/core/repository"
)

// writeln writes a formatted string with platform-specific line ending to the writer
func writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

func init() {
	registry.Register(Build)
}

// BuildResult captures the outcome of a module build
type BuildResult struct {
	Moniker  string
	ExitCode int
	Warnings []string
	Errors   []string
}


// Build command entry point - builds one or more modules
func Build() int {
	args := os.Args[2:] // Skip program name and "build"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printBuildUsage()
		return 0
	}

	// Detect CI environment
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""

	// Parse module monikers and flags
	var monikers []string
	tidyFirst := !isCI // Default: true for local, false for CI
	tidyExplicitlySet := false
	compressed := false
	compressedUPX := false
	version := ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--tidy-first":
			tidyFirst = true
			tidyExplicitlySet = true
		case "--no-tidy":
			tidyFirst = false
			tidyExplicitlySet = true
		case "--compressed":
			compressed = true
		case "--compressed-upx":
			compressedUPX = true
			compressed = true // UPX implies stripped
		case "--version":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --version requires a value\n")
				printBuildUsage()
				return 1
			}
			i++
			version = args[i]
		default:
			if strings.HasPrefix(arg, "--version=") {
				version = strings.TrimPrefix(arg, "--version=")
			} else if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printBuildUsage()
				return 1
			} else {
				monikers = append(monikers, arg)
			}
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// If exactly one module specified, run single module build (verbose output)
	if len(monikers) == 1 {
		return buildSingleModule(monikers[0], workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet, compressed, compressedUPX, version)
	}

	// If no monikers provided, default to all buildable modules
	if len(monikers) == 0 {
		fmt.Println("ℹ️  No modules specified, building all modules...")
		for _, module := range moduleReport.Registry.All() {
			// Include all modules - GetBuildFunc will return a handler for any type
			monikers = append(monikers, module.Moniker)
		}
	}

	// Multiple modules - run parallel build with orchestrator
	return buildMultipleModules(monikers, workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet, compressed, compressedUPX, version)
}

// buildSingleModule builds a single module with verbose output to console
func buildSingleModule(moniker string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool, compressed bool, compressedUPX bool, version string) int {
	// Get the module from registry
	module, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
		return 1
	}

	// Get build function for module type using the dispatch helper
	buildFunc := GetBuildFunc(module.Type)

	// Determine output directory
	var outputDir string
	testRunID := os.Getenv("R2R_TEST_RUN_ID")
	if testRunID != "" {
		outputDir = filepath.Join(workspaceRoot, repository.OutDir, "test", testRunID, "build-artifacts", moniker)
	} else {
		outputDir = repository.BuildOutputPath(workspaceRoot, moniker)
	}

	// Purge and create output directory
	if err := os.RemoveAll(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to purge output directory: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create output directory: %v\n", err)
		return 1
	}

	// Create build log file
	logPath := filepath.Join(outputDir, "build.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
		return 1
	}
	defer logFile.Close()

	// Create multi-writer to log to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Print header
	writeln(multiWriter, "Building module: %s (type: %s)", moniker, module.Type)
	writeln(multiWriter, "Module root: %s", module.Files.Root)
	writeln(multiWriter, "Output directory: %s", outputDir)
	writeln(multiWriter, "Build log: %s", logPath)

	// Log tidy behavior (only relevant for Go modules)
	if IsGoModuleType(module.Type) {
		if tidyFirst {
			if tidyExplicitlySet {
				writeln(multiWriter, "Tidy mode: enabled (explicit flag)")
			} else {
				writeln(multiWriter, "Tidy mode: enabled (default for local builds)")
			}
		} else {
			if tidyExplicitlySet {
				writeln(multiWriter, "Tidy mode: disabled (explicit flag)")
			} else {
				writeln(multiWriter, "Tidy mode: disabled (CI environment detected)")
			}
		}
	}

	// Execute build
	buildOpts := BuildOptions{
		TidyFirst:     tidyFirst,
		Compressed:    compressed,
		CompressedUPX: compressedUPX,
		Version:       version,
	}
	return buildFunc(module, workspaceRoot, outputDir, multiWriter, buildOpts)
}

// buildMultipleModules builds multiple modules in parallel using the orchestrator
func buildMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool, compressed bool, compressedUPX bool, version string) int {
	// Print tidy mode info before starting orchestrator
	hasGoModules := false
	for _, mon := range monikers {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			if IsGoModuleType(module.Type) {
				hasGoModules = true
				break
			}
		}
	}

	if hasGoModules {
		if tidyFirst {
			if tidyExplicitlySet {
				fmt.Println("Tidy mode: enabled (explicit flag)")
			} else {
				fmt.Println("Tidy mode: enabled (default for local builds)")
			}
		} else {
			if tidyExplicitlySet {
				fmt.Println("Tidy mode: disabled (explicit flag)")
			} else {
				fmt.Println("Tidy mode: disabled (CI environment detected)")
			}
		}
	}

	// Configure orchestrator
	config := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        "out/build",
		LogFileName:          "build.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "building",
		MaxConcurrency:       0, // Use default (number of CPUs)
		StatusUpdateInterval: 2, // Update every 2 seconds
	}

	// Create worker function that builds a single module
	worker := func(moniker string, logWriter io.Writer) int {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
			return 1
		}

		moduleOutputDir := repository.BuildOutputPath(workspaceRoot, moniker)
		return runModuleBuild(module, workspaceRoot, moduleOutputDir, logWriter, tidyFirst, compressed, compressedUPX, version)
	}

	// Create and run orchestrator
	orch := orchestrator.New(config, worker)
	results, err := orch.Run(monikers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Print summary and close orchestrator
	orch.PrintSummary(results)
	orch.Close()

	// Return exit code based on results
	return orchestrator.GetExitCode(results)
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, tidyFirst bool, compressed bool, compressedUPX bool, version string) int {
	// Get build function for module type using the dispatch helper
	buildFunc := GetBuildFunc(module.Type)

	opts := BuildOptions{
		TidyFirst:     tidyFirst,
		Compressed:    compressed,
		CompressedUPX: compressedUPX,
		Version:       version,
	}
	return buildFunc(module, workspaceRoot, outputDir, logWriter, opts)
}

func printBuildUsage() {
	fmt.Println("Build one or more modules by moniker")
	fmt.Println()
	fmt.Println("Usage: build [flags] [module1] [module2] ...")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  module1, module2, ...     Module monikers to build (builds all if none specified)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --tidy-first              Run 'go mod tidy' before building (default for local)")
	fmt.Println("  --no-tidy                 Skip 'go mod tidy' (default for CI)")
	fmt.Println("  --compressed              Strip debug info for smaller binaries (go-cli only)")
	fmt.Println("  --compressed-upx          Also apply UPX compression for maximum size reduction")
	fmt.Println("  --version VERSION         Inject version string into binary (go-cli only)")
	fmt.Println("  -h, --help                Show this help message")
	fmt.Println()
	fmt.Println("Compression (go-cli only):")
	fmt.Println("  Default (dev):     Full debug info for debugging (~39 MB)")
	fmt.Println("  --compressed:      Strip debug info with -ldflags \"-s -w\" (~26 MB, ~30% smaller)")
	fmt.Println("  --compressed-upx:  Also UPX compress (~10 MB, ~70% smaller total)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  build                                # Build all modules (dev mode)")
	fmt.Println("  build src-cli                        # Build CLI with debug info")
	fmt.Println("  build src-cli --compressed           # Build CLI for release")
	fmt.Println("  build src-cli --compressed-upx       # Build CLI with UPX for minimal size")
	fmt.Println("  build src-cli --version 1.0.0        # Build with version injection")
	fmt.Println("  build --tidy-first docs              # Build with go mod tidy first")
}
