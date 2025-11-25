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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

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

	for _, arg := range args {
		switch arg {
		case "--tidy-first":
			tidyFirst = true
			tidyExplicitlySet = true
		case "--no-tidy":
			tidyFirst = false
			tidyExplicitlySet = true
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
				printBuildUsage()
				return 1
			}
			monikers = append(monikers, arg)
		}
	}

	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load module contracts: %v\n", err)
		return 1
	}

	// If exactly one module specified, run single module build (verbose output)
	if len(monikers) == 1 {
		return buildSingleModule(monikers[0], workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet)
	}

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		fmt.Println("ℹ️  No modules specified, building all modules...")
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Multiple modules - run parallel build with orchestrator
	return buildMultipleModules(monikers, workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet)
}

// buildSingleModule builds a single module with verbose output to console
func buildSingleModule(moniker string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool) int {
	// Get the module from registry
	module, exists := moduleReport.Registry.Get(moniker)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
		return 1
	}

	// Get build function for module type
	buildFunc, hasBuilder := buildFunctions[module.Type]
	if !hasBuilder {
		fmt.Fprintf(os.Stderr, "Error: no build function for type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "Module: %s\n", moniker)
		fmt.Fprintf(os.Stderr, "Type: %s\n", module.Type)
		fmt.Fprintf(os.Stderr, "\nAvailable build functions:\n")
		for moduleType := range buildFunctions {
			fmt.Fprintf(os.Stderr, "  - %s\n", moduleType)
		}
		return 1
	}

	// Determine output directory
	var outputDir string
	testRunID := os.Getenv("R2R_TEST_RUN_ID")
	if testRunID != "" {
		outputDir = filepath.Join(workspaceRoot, "out", "test", testRunID, "build-artifacts", moniker)
	} else {
		outputDir = filepath.Join(workspaceRoot, "out", "build", moniker)
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
	fmt.Fprintf(multiWriter, "Building module: %s (type: %s)\n", moniker, module.Type)
	fmt.Fprintf(multiWriter, "Module root: %s\n", module.Source.Root)
	fmt.Fprintf(multiWriter, "Output directory: %s\n", outputDir)
	fmt.Fprintf(multiWriter, "Build log: %s\n", logPath)

	// Log tidy behavior (only relevant for Go modules)
	if IsGoModuleType(module.Type) {
		if tidyFirst {
			if tidyExplicitlySet {
				fmt.Fprintf(multiWriter, "Tidy mode: enabled (explicit flag)\n")
			} else {
				fmt.Fprintf(multiWriter, "Tidy mode: enabled (default for local builds)\n")
			}
		} else {
			if tidyExplicitlySet {
				fmt.Fprintf(multiWriter, "Tidy mode: disabled (explicit flag)\n")
			} else {
				fmt.Fprintf(multiWriter, "Tidy mode: disabled (CI environment detected)\n")
			}
		}
	}

	// Execute build
	buildOpts := BuildOptions{
		TidyFirst: tidyFirst,
	}
	return buildFunc(module, workspaceRoot, outputDir, multiWriter, buildOpts)
}

// buildMultipleModules builds multiple modules in parallel
func buildMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool) int {
	// Create orchestrator log file
	orchestratorLogPath := filepath.Join(workspaceRoot, "out", "build", "orchestrator.log")
	if err := os.MkdirAll(filepath.Dir(orchestratorLogPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create orchestrator log directory: %v\n", err)
		return 1
	}
	orchestratorLog, err := os.Create(orchestratorLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create orchestrator log: %v\n", err)
		return 1
	}
	defer orchestratorLog.Close()

	orchestratorLogBuf := bufio.NewWriter(orchestratorLog)
	defer orchestratorLogBuf.Flush()

	orchestratorOut := io.MultiWriter(os.Stdout, orchestratorLogBuf)

	// Check if any modules are Go modules (for tidy mode message)
	hasGoModules := false
	for _, mon := range monikers {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			if IsGoModuleType(module.Type) {
				hasGoModules = true
				break
			}
		}
	}

	// Log tidy behavior (only if there are Go modules)
	if hasGoModules {
		if tidyFirst {
			if tidyExplicitlySet {
				fmt.Fprintf(orchestratorOut, "Tidy mode: enabled (explicit flag)\n")
			} else {
				fmt.Fprintf(orchestratorOut, "Tidy mode: enabled (default for local builds)\n")
			}
		} else {
			if tidyExplicitlySet {
				fmt.Fprintf(orchestratorOut, "Tidy mode: disabled (explicit flag)\n")
			} else {
				fmt.Fprintf(orchestratorOut, "Tidy mode: disabled (CI environment detected)\n")
			}
		}
	}

	fmt.Fprintf(orchestratorOut, "Building %d modules in parallel: %v\n\n", len(monikers), monikers)

	// Build each module in parallel
	var mu sync.Mutex
	var wg sync.WaitGroup
	buildResults := []BuildResult{}
	builtModules := []*modules.ModuleContract{}

	for i, moniker := range monikers {
		wg.Add(1)
		go func(idx int, mon string, orchOut io.Writer) {
			defer wg.Done()

			module, exists := moduleReport.Registry.Get(mon)
			if !exists {
				mu.Lock()
				fmt.Fprintf(orchOut, "[building] %s (Module not found) ........ Failed\n", mon)
				buildResults = append(buildResults, BuildResult{
					Moniker:  mon,
					ExitCode: 1,
					Errors:   []string{"Module not found"},
				})
				mu.Unlock()
				return
			}

			moduleOutputDir := filepath.Join(workspaceRoot, "out", "build", mon)
			parentDir := filepath.Dir(moduleOutputDir)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				mu.Lock()
				fmt.Fprintf(orchOut, "[building] %s (Failed to create parent directory) ........ Failed\n", mon)
				buildResults = append(buildResults, BuildResult{
					Moniker:  mon,
					ExitCode: 1,
					Errors:   []string{fmt.Sprintf("Failed to create parent directory %s: %v", parentDir, err)},
				})
				mu.Unlock()
				return
			}

			if err := os.RemoveAll(moduleOutputDir); err != nil {
				fmt.Fprintf(orchOut, "[building] %s (Warning: failed to remove existing directory: %v)\n", mon, err)
			}

			if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
				mu.Lock()
				fmt.Fprintf(orchOut, "[building] %s (Failed to create directory) ........ Failed\n", mon)
				buildResults = append(buildResults, BuildResult{
					Moniker:  mon,
					ExitCode: 1,
					Errors:   []string{fmt.Sprintf("Failed to create directory %s: %v", moduleOutputDir, err)},
				})
				mu.Unlock()
				return
			}

			logPath := filepath.Join(moduleOutputDir, "build.log")
			logFile, err := os.Create(logPath)
			if err != nil {
				mu.Lock()
				fmt.Fprintf(orchOut, "[building] %s (Failed to create log) ........ Failed\n", mon)
				buildResults = append(buildResults, BuildResult{
					Moniker:  mon,
					ExitCode: 1,
					Errors:   []string{fmt.Sprintf("Failed to create log file %s: %v", logPath, err)},
				})
				mu.Unlock()
				return
			}
			defer logFile.Close()

			multiWriter := io.MultiWriter(logFile)

			done := make(chan bool)
			go showProgress(orchOut, &mu, mon, done)

			exitCode := runModuleBuild(module, workspaceRoot, moduleOutputDir, multiWriter, tidyFirst)

			done <- true
			close(done)

			logFile.Close()

			warnings, errors := parseLogForIssues(logPath)

			mu.Lock()
			builtModules = append(builtModules, module)
			buildResults = append(buildResults, BuildResult{
				Moniker:  mon,
				ExitCode: exitCode,
				Warnings: warnings,
				Errors:   errors,
			})

			relLogPath := filepath.Join("out", "build", mon, "build.log")
			var statusLine string
			if exitCode != 0 {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Failed\r\n", mon, relLogPath)
			} else if len(warnings) > 0 {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Done (with %d warnings)\r\n", mon, relLogPath, len(warnings))
			} else {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Done\r\n", mon, relLogPath)
			}

			orchOut.Write([]byte(statusLine))
			os.Stdout.Sync()
			mu.Unlock()
		}(i, moniker, orchestratorOut)
	}

	wg.Wait()

	// Calculate and print summary
	totalFailed := 0
	totalWarnings := 0
	modulesWithWarnings := []string{}
	failedModules := []string{}

	for _, result := range buildResults {
		if result.ExitCode != 0 {
			totalFailed++
			failedModules = append(failedModules, result.Moniker)
		}
		if len(result.Warnings) > 0 {
			totalWarnings += len(result.Warnings)
			modulesWithWarnings = append(modulesWithWarnings, result.Moniker)
		}
	}

	fmt.Fprintf(orchestratorOut, "\n===========================================\n")
	fmt.Fprintf(orchestratorOut, "Build Run Summary\n")
	fmt.Fprintf(orchestratorOut, "===========================================\n")
	fmt.Fprintf(orchestratorOut, "Total modules: %d\n", len(monikers))
	fmt.Fprintf(orchestratorOut, "Passed: %d\n", len(monikers)-totalFailed)
	fmt.Fprintf(orchestratorOut, "Failed: %d\n", totalFailed)
	fmt.Fprintf(orchestratorOut, "Warnings: %d (in %d modules)\n", totalWarnings, len(modulesWithWarnings))

	if len(failedModules) > 0 {
		fmt.Fprintf(orchestratorOut, "\n❌ Failed modules:\n")
		for _, result := range buildResults {
			if result.ExitCode != 0 {
				fmt.Fprintf(orchestratorOut, "  - %s (exit code: %d)\n", result.Moniker, result.ExitCode)
				if len(result.Errors) > 0 {
					fmt.Fprintf(orchestratorOut, "    Errors:\n")
					for _, err := range result.Errors {
						if len(err) > 80 {
							err = err[:77] + "..."
						}
						fmt.Fprintf(orchestratorOut, "      • %s\n", err)
					}
				}
			}
		}
	}

	if len(modulesWithWarnings) > 0 {
		fmt.Fprintf(orchestratorOut, "\n⚠️  Modules with warnings:\n")
		for _, result := range buildResults {
			if len(result.Warnings) > 0 && result.ExitCode == 0 {
				fmt.Fprintf(orchestratorOut, "  - %s (%d warnings)\n", result.Moniker, len(result.Warnings))
				for _, warn := range result.Warnings {
					if len(warn) > 80 {
						warn = warn[:77] + "..."
					}
					fmt.Fprintf(orchestratorOut, "      • %s\n", warn)
				}
			}
		}
	}

	fmt.Fprintf(orchestratorOut, "\nOrchestrator log: out/build/orchestrator.log\n")
	fmt.Fprintf(orchestratorOut, "Module logs: out/build/<module>/build.log\n")

	if totalFailed > 0 {
		return 1
	}
	return 0
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, tidyFirst bool) int {
	buildFunc, hasBuilder := buildFunctions[module.Type]
	if !hasBuilder {
		fmt.Fprintf(logWriter, "Error: no build function for type: %s\n", module.Type)
		return 1
	}

	opts := BuildOptions{
		TidyFirst: tidyFirst,
	}
	return buildFunc(module, workspaceRoot, outputDir, logWriter, opts)
}

// parseLogForIssues scans a log file for warnings and errors
func parseLogForIssues(logPath string) (warnings []string, errors []string) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		lowerLine := strings.ToLower(line)

		if strings.Contains(lowerLine, "error:") ||
			strings.Contains(lowerLine, "❌") ||
			strings.Contains(lowerLine, "failed") ||
			strings.Contains(lowerLine, "fatal:") {
			errors = append(errors, strings.TrimSpace(line))
		}

		if strings.Contains(lowerLine, "warning:") && !strings.Contains(lowerLine, "error:") {
			warnings = append(warnings, strings.TrimSpace(line))
		}
	}

	return warnings, errors
}

// showProgress displays dots every 5 seconds while a module is building
func showProgress(out io.Writer, mu *sync.Mutex, moniker string, done chan bool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			mu.Lock()
			fmt.Fprintf(out, "[building] %s .......\r\n", moniker)
			os.Stdout.Sync()
			mu.Unlock()
		}
	}
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
	fmt.Println("  -h, --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  build                     # Build all modules")
	fmt.Println("  build src-commands        # Build a single module (verbose output)")
	fmt.Println("  build src-core src-cli    # Build multiple modules in parallel")
	fmt.Println("  build --tidy-first docs   # Build with go mod tidy first")
}
