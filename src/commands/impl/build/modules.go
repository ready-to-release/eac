// Command: build modules
// Short: Build multiple modules in sequence and collect results in a build run directory
// Long: Build multiple modules in sequence and collect results in a build run directory.
// Long:
// Long: This command builds multiple modules respecting their dependency order.
// Long: If no monikers are specified, all modules in the repository are built.
// Long:
// Long: Build results are collected in a timestamped directory under 'out/build-runs/'
// Long: containing logs and summary reports for each module. Failed builds are clearly
// Long: marked and do not stop the execution of remaining modules.
// Long:
// Long: Example:
// Long:   build modules                    # Build all modules
// Long:   build modules src-core src-cli   # Build specific modules
// HasSideEffects: false
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
	registry.Register(BuildModules)
}

// BuildResult captures the outcome of a module build
type BuildResult struct {
	Moniker  string
	ExitCode int
	Warnings []string
	Errors   []string
}

// BuildModules builds multiple modules in sequence (defaults to all modules)
func BuildModules() int {
	// Detect CI environment
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("GITLAB_CI") != ""

	// Parse module monikers and flags
	var monikers []string
	tidyFirst := !isCI // Default: true for local, false for CI
	tidyExplicitlySet := false

	// Parse arguments starting from index 3 (skip "binary", "build", "modules")
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
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
				fmt.Fprintf(os.Stderr, "Usage: build modules [moniker1] [moniker2] ... [--tidy-first|--no-tidy]\n")
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

	// If no monikers provided, default to all modules
	if len(monikers) == 0 {
		fmt.Println("ℹ️  No modules specified, building all modules...")
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

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

	// Use buffered writer for orchestrator log
	orchestratorLogBuf := bufio.NewWriter(orchestratorLog)
	defer orchestratorLogBuf.Flush()

	// Create multi-writer for orchestrator output (console + log file)
	orchestratorOut := io.MultiWriter(os.Stdout, orchestratorLogBuf)

	// Log tidy behavior
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

			// Get module from registry
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

			// Purge and create module output directory
			moduleOutputDir := filepath.Join(workspaceRoot, "out", "build", mon)

			// Ensure parent directory exists first (out/build)
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

			// Remove existing module output directory
			if err := os.RemoveAll(moduleOutputDir); err != nil {
				// Log warning but continue - not fatal if removal fails
				fmt.Fprintf(orchOut, "[building] %s (Warning: failed to remove existing directory: %v)\n", mon, err)
			}

			// Create module output directory
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

			// Create build log file
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

			// Create multi-writer: log to file only (not console for parallel builds)
			// Each module has its own log file in out/build/<moniker>/build.log
			multiWriter := io.MultiWriter(logFile)

			// Start progress indicator (dots every 5 seconds)
			done := make(chan bool)
			go showProgress(orchOut, &mu, mon, done)

			// Run build for this module
			exitCode := runModuleBuild(module, workspaceRoot, moduleOutputDir, multiWriter, tidyFirst)

			// Stop progress indicator
			done <- true
			close(done)

			// Close log file before parsing
			logFile.Close()

			// Parse log file for warnings and errors
			warnings, errors := parseLogForIssues(logPath)

			// Track results and print complete status line (thread-safe, atomic)
			mu.Lock()
			builtModules = append(builtModules, module)
			buildResults = append(buildResults, BuildResult{
				Moniker:  mon,
				ExitCode: exitCode,
				Warnings: warnings,
				Errors:   errors,
			})

			// Build complete status line as a single string for atomic write
			relLogPath := filepath.Join("out", "build", mon, "build.log")
			var statusLine string
			if exitCode != 0 {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Failed\r\n", mon, relLogPath)
			} else if len(warnings) > 0 {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Done (with %d warnings)\r\n", mon, relLogPath, len(warnings))
			} else {
				statusLine = fmt.Sprintf("[building] %s (See %s for details) ........ Done\r\n", mon, relLogPath)
			}

			// Write as single operation (orchOut writes to both stdout and log)
			orchOut.Write([]byte(statusLine))

			// Flush stdout explicitly for Windows console
			os.Stdout.Sync()
			mu.Unlock()
		}(i, moniker, orchestratorOut)
	}

	// Wait for all builds to complete
	wg.Wait()

	// Calculate summary statistics
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

	// Print summary
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
	// Get build function for module type
	buildFunc, hasBuilder := buildFunctions[module.Type]
	if !hasBuilder {
		fmt.Fprintf(logWriter, "Error: no build function for type: %s\n", module.Type)
		return 1
	}

	// Execute the build function with options (build all platforms, optionally tidy first)
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

		// Detect errors
		if strings.Contains(lowerLine, "error:") ||
		   strings.Contains(lowerLine, "❌") ||
		   strings.Contains(lowerLine, "failed") ||
		   strings.Contains(lowerLine, "fatal:") {
			errors = append(errors, strings.TrimSpace(line))
		}

		// Detect warnings (but not if it's already an error)
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
