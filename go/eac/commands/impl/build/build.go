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
// Long:   build eac-commands              # Build a single module
// Long:   build eac-core r2r-cli          # Build specific modules
// Long:   build --tidy-first eac-commands # Build with go mod tidy first
// Args: modules
package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofrs/flock"
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/builders"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	systemdeps "github.com/ready-to-release/eac/go/eac/core/system-deps"
)

var log = logging.C()

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

// acquireModuleBuildLock attempts to acquire an exclusive lock for building a module.
// Returns the lock handle and nil error on success.
// Returns nil and error if lock is already held (module is being built).
func acquireModuleBuildLock(moniker, workspaceRoot string) (*flock.Flock, error) {
	// Ensure out/build directory exists (parent directory for lock files)
	buildDir := filepath.Join(workspaceRoot, "out", "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	// Create lock file path in parent directory (so it survives directory purge)
	// Use module moniker as the mutex identifier
	lockPath := filepath.Join(buildDir, fmt.Sprintf(".lock-%s", moniker))

	// Create flock instance
	lock := flock.New(lockPath)

	// Try to acquire lock (non-blocking)
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return nil, fmt.Errorf("module '%s' is already being built", moniker)
	}

	return lock, nil
}

// releaseModuleBuildLock releases the lock and removes the lock file
func releaseModuleBuildLock(lock *flock.Flock) {
	if lock == nil {
		return
	}

	lockPath := lock.Path()
	lock.Unlock()

	// Clean up the lock file
	os.Remove(lockPath)
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
	skipVerification := false // Skip system dependency verification (go, docker, etc.)
	skipModuleDeps := false   // Skip including transitive module dependencies
	showTimings := false
	pdfMode := false
	pdfTheme := ""
	version := ""
	listArtifacts := false
	dryRun := false

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
		case "--skip-deps":
			skipModuleDeps = true
		case "--skip-verification":
			skipVerification = true
		case "--timings":
			showTimings = true
		case "--pdf":
			pdfMode = true
		case "--accept-warnings":
			// Flag is handled in mkdocs builder via os.Args check
			// Just accept it here so it doesn't fail as unknown flag
		case "--list-artifacts":
			listArtifacts = true
		case "--dry-run":
			dryRun = true
		case "--pdf-theme":
			if i+1 >= len(args) {
				log.Errorf("Error: --pdf-theme requires a value (dark, light, or all)")
				printBuildUsage()
				return 1
			}
			i++
			pdfTheme = args[i]
			pdfMode = true // --pdf-theme implies --pdf
		case "--version":
			if i+1 >= len(args) {
				log.Errorf("Error: --version requires a value")
				printBuildUsage()
				return 1
			}
			i++
			version = args[i]
		default:
			if strings.HasPrefix(arg, "--pdf-theme=") {
				pdfTheme = strings.TrimPrefix(arg, "--pdf-theme=")
				pdfMode = true // --pdf-theme implies --pdf
			} else if strings.HasPrefix(arg, "--version=") {
				version = strings.TrimPrefix(arg, "--version=")
			} else if strings.HasPrefix(arg, "--") {
				log.Errorf("Error: unknown flag: %s", arg)
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
		log.Errorf("Error: failed to find repository root: %v", err)
		return 1
	}

	// Load module contracts
	moduleReport, err := reports.GetModuleContracts(workspaceRoot)
	if err != nil {
		log.Errorf("Error: failed to load module contracts: %v", err)
		return 1
	}

	// If no monikers provided, default to all buildable modules (before dependency check)
	if len(monikers) == 0 {
		for _, module := range moduleReport.Registry.All() {
			monikers = append(monikers, module.Moniker)
		}
	}

	// Handle --list-artifacts flag
	if listArtifacts {
		return listModuleArtifacts(monikers, workspaceRoot, moduleReport)
	}

	// Run build (single or multiple modules) - phases are handled inside
	return buildMultipleModules(monikers, workspaceRoot, moduleReport, tidyFirst, tidyExplicitlySet, compressed, compressedUPX, version, skipVerification, skipModuleDeps, showTimings, pdfMode, pdfTheme, dryRun)
}

// listModuleArtifacts lists the artifacts that would be produced by building the specified modules
func listModuleArtifacts(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport) int {
	// Sort monikers for consistent output
	sort.Strings(monikers)

	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			log.Errorf("Error: module not found: %s", moniker)
			continue
		}

		listFunc := builders.GetListArtifactsFunc(module.Type)
		if listFunc == nil {
			// No artifacts for this module type
			continue
		}

		artifacts := listFunc(module, workspaceRoot)
		outputDir := repository.BuildOutputPath(workspaceRoot, moniker)

		for _, artifact := range artifacts {
			// Output full path relative to workspace root
			fullPath := filepath.Join(outputDir, artifact)
			relPath, err := filepath.Rel(workspaceRoot, fullPath)
			if err != nil {
				relPath = fullPath
			}
			fmt.Println(relPath)
		}
	}

	return 0
}


// buildMultipleModules builds multiple modules in parallel using the orchestrator
func buildMultipleModules(monikers []string, workspaceRoot string, moduleReport *reports.ModuleContractReport, tidyFirst bool, tidyExplicitlySet bool, compressed bool, compressedUPX bool, version string, skipVerification bool, skipModuleDeps bool, showTimings bool, pdfMode bool, pdfTheme string, dryRun bool) int {
	// Show execution context
	log.Infof("Executing build via %s. \"%s\"", logging.GetExecutionContext(), logging.GetFullCommand())
	log.Info("")

	// Phase 1: Module Discovery
	log.Info(output.PhaseHeader(1, "Module Discovery"))
	log.Infof("Requested: %d modules:%s", len(monikers), output.ListFormat(monikers, 60, 5))

	// Calculate execution order to determine dependencies early
	includeDependencies := !skipModuleDeps
	executionPlan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, includeDependencies)
	if err != nil {
		log.Errorf("Failed to calculate execution order: %v", err)
		return 1
	}

	// Show dependencies if any were added
	if includeDependencies && len(executionPlan.ExecutionOrder) > len(monikers) {
		// Find which modules were added as dependencies
		requestedSet := make(map[string]bool)
		for _, m := range monikers {
			requestedSet[m] = true
		}
		var addedDeps []string
		for _, m := range executionPlan.ExecutionOrder {
			if !requestedSet[m] {
				addedDeps = append(addedDeps, m)
			}
		}
		log.Infof("Dependencies: %d modules:%s", len(addedDeps), output.ListFormat(addedDeps, 60, 5))
		log.Infof("Total: %d modules to build", len(executionPlan.ExecutionOrder))
	}

	// Check if any requested modules are Go modules (for tidy mode logging)
	hasGoModules := false
	for _, mon := range monikers {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			if builders.IsGoModuleType(module.Type) {
				hasGoModules = true
				break
			}
		}
	}

	if hasGoModules {
		if tidyFirst {
			if tidyExplicitlySet {
				log.Info("Tidy mode: enabled (explicit flag)")
			} else {
				log.Info("Tidy mode: enabled (default for local builds)")
			}
		} else {
			if tidyExplicitlySet {
				log.Info("Tidy mode: disabled (explicit flag)")
			} else {
				log.Info("Tidy mode: disabled (CI environment detected)")
			}
		}
	}
	log.Info("")

	// Phase 2: Dependency Verification (system dependencies like go, docker, etc.)
	// Use the expanded execution order to verify deps for ALL modules that will be built
	if !skipVerification {
		if exitCode := verifyBuildDependencies(executionPlan.ExecutionOrder, moduleReport); exitCode != 0 {
			return exitCode
		}
	}

	// Phase 3: Build Execution
	log.Info(output.PhaseHeader(3, "Build Execution"))
	log.Info(output.OutputDir("out/build/"))
	log.Info("")

	// Build module type lookup for ALL modules in execution plan (including dependencies)
	moduleTypes := make(map[string]string)
	for _, mon := range executionPlan.ExecutionOrder {
		if module, exists := moduleReport.Registry.Get(mon); exists {
			moduleTypes[mon] = module.Type
		}
	}

	// Configure orchestrator
	orchConfig := orchestrator.Config{
		WorkspaceRoot:        workspaceRoot,
		OutputBaseDir:        "out/build",
		LogFileName:          "build.log",
		OrchestratorLogName:  "orchestrator.log",
		ActionVerb:           "building",
		MaxConcurrency:       0, // Use default (number of CPUs)
		StatusUpdateInterval: 2, // Update every 2 seconds
		ModuleTypes:          moduleTypes,
		ShowTimings:          showTimings,
	}

	// Create worker function that builds a single module and returns type info
	worker := func(moniker string, logWriter io.Writer) int {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(logWriter, "Error: module not found: %s\n", moniker)
			return 1
		}

		// Skip lock acquisition in dry-run mode since we're not actually building
		if !dryRun {
			// Acquire exclusive lock for this module FIRST (before any directory operations)
			lockFile, err := acquireModuleBuildLock(moniker, workspaceRoot)
			if err != nil {
				fmt.Fprintf(logWriter, "Error: module '%s' is already being built\n", moniker)
				fmt.Fprintf(logWriter, "Details: %v\n", err)
				return 1
			}
			defer releaseModuleBuildLock(lockFile)
		}

		moduleOutputDir := repository.BuildOutputPath(workspaceRoot, moniker)
		return runModuleBuild(module, workspaceRoot, moduleOutputDir, logWriter, tidyFirst, compressed, compressedUPX, version, pdfMode, pdfTheme, dryRun)
	}

	// Create and run orchestrator with layered execution
	orch := orchestrator.New(orchConfig, worker)
	results, err := orch.RunLayered(executionPlan.Layers)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Print summary and close orchestrator
	orch.PrintSummary(results)
	orch.Close()

	// Return exit code based on results
	return orchestrator.GetExitCode(results)
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer, tidyFirst bool, compressed bool, compressedUPX bool, version string, pdfMode bool, pdfTheme string, dryRun bool) int {
	// Get build function for module type
	buildFunc := builders.GetBuildFunc(module.Type)

	opts := builders.BuildOptions{
		TidyFirst:     tidyFirst,
		Compressed:    compressed,
		CompressedUPX: compressedUPX,
		Version:       version,
		PDFMode:       pdfMode,
		PDFTheme:      pdfTheme,
		DryRun:        dryRun,
	}

	// In dry-run mode, simulate a successful build
	if dryRun {
		writeln(logWriter, "Build: %s (dry-run)", module.Moniker)
		writeln(logWriter, "Type: %s", module.Type)
		writeln(logWriter, "Root: %s", module.Files.Root)
		writeln(logWriter, "")
		writeln(logWriter, "Dry-run mode: skipping actual build")
		return 0
	}

	exitCode := buildFunc(module, workspaceRoot, outputDir, logWriter, opts)
	if exitCode != 0 {
		return exitCode
	}

	// Execute post-build steps if build succeeded
	return builders.ExecutePostBuildSteps(module.Type, module.Moniker, workspaceRoot, outputDir, logWriter)
}

// verifyBuildDependencies checks that all required build dependencies are available
func verifyBuildDependencies(monikers []string, moduleReport *reports.ModuleContractReport) int {
	cfg := config.Global()
	if cfg == nil || cfg.ModuleTypes == nil {
		// Config not loaded, skip verification
		return 0
	}

	// Collect unique build dependencies from all modules
	depsMap := make(map[string]bool)
	for _, moniker := range monikers {
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			continue
		}

		deps := cfg.ModuleTypes.GetBuildDeps(module.Type)
		for _, dep := range deps {
			if dep != "" {
				depsMap[dep] = true
			}
		}
	}

	// No dependencies to verify
	if len(depsMap) == 0 {
		return 0
	}

	// Convert to sorted slice for consistent output
	deps := make([]string, 0, len(depsMap))
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	// Print phase header
	log.Info(output.PhaseHeader(2, "Dependency Verification"))
	log.Infof("Required: %s", strings.Join(deps, ", "))

	// Verify all dependencies
	results := systemdeps.VerifyAll(deps)
	var missing []string

	for _, result := range results {
		log.Info(output.DependencyLine(result.Available, result.Name, result.Version))
		if !result.Available {
			missing = append(missing, result.Moniker)
		}
	}

	log.Info("")

	if len(missing) > 0 {
		log.Errorf("❌ Error: Required build dependencies are missing: %s", strings.Join(missing, ", "))
		log.Errorf("   Use --skip-verification to bypass this check")
		return 1
	}

	return 0
}

func printBuildUsage() {
	log.Info("Build one or more modules by moniker")
	log.Info("")
	log.Info("Usage: build [flags] [module1] [module2] ...")
	log.Info("")
	log.Info("Arguments:")
	log.Info("  module1, module2, ...     Module monikers to build (builds all if none specified)")
	log.Info("")
	log.Info("Flags:")
	log.Info("  --dry-run                 Simulate build without running actual commands")
	log.Info("  --list-artifacts          List artifacts that would be produced (no build)")
	log.Info("  --tidy-first              Run 'go mod tidy' before building (default for local)")
	log.Info("  --no-tidy                 Skip 'go mod tidy' (default for CI)")
	log.Info("  --skip-deps               Only build specified modules (skip transitive dependencies)")
	log.Info("  --skip-verification       Skip system dependency verification (go, docker, etc.)")
	log.Info("  --timings                 Show detailed timing summary")
	log.Info("  --compressed              Strip debug info for smaller binaries (go-cli only)")
	log.Info("  --compressed-upx          Also apply UPX compression for maximum size reduction")
	log.Info("  --version VERSION         Inject version string into binary (go-cli only)")
	log.Info("  --pdf                     Generate PDF documentation (mkdocs modules only)")
	log.Info("  --pdf-theme THEME         PDF theme: dark, light, or all (default: dark)")
	log.Info("  --accept-warnings         Don't fail on MkDocs warnings (non-strict mode)")
	log.Info("  -h, --help                Show this help message")
	log.Info("")
	log.Info("Compression (go-cli only):")
	log.Info("  Default (dev):     Full debug info for debugging (~39 MB)")
	log.Info("  --compressed:      Strip debug info with -ldflags \"-s -w\" (~26 MB, ~30% smaller)")
	log.Info("  --compressed-upx:  Also UPX compress (~10 MB, ~70% smaller total)")
	log.Info("")
	log.Info("PDF Generation (mkdocs only):")
	log.Info("  --pdf:                Enable PDF export alongside HTML site (dark theme)")
	log.Info("  --pdf-theme=dark:     Dark PDF for digital viewing")
	log.Info("  --pdf-theme=light:    Light PDF for paper printing")
	log.Info("  --pdf-theme=all:      Build both dark and light PDFs")
	log.Info("                        Uses mkdocs-with-pdf plugin with WeasyPrint")
	log.Info("                        Output: out/build/<module>/site/pdf/ready-to-release-docs-{theme}.pdf")
	log.Info("")
	log.Info("Examples:")
	log.Info("  build                                # Build all modules (dev mode)")
	log.Info("  build r2r-cli                        # Build CLI with debug info")
	log.Info("  build r2r-cli --compressed           # Build CLI for release")
	log.Info("  build r2r-cli --compressed-upx       # Build CLI with UPX for minimal size")
	log.Info("  build r2r-cli --version 1.0.0        # Build with version injection")
	log.Info("  build --tidy-first docs              # Build with go mod tidy first")
	log.Info("  build docs --pdf                     # Build docs with dark PDF")
	log.Info("  build docs --pdf-theme=all           # Build docs with both PDF themes")
	log.Info("  build r2r-cli --list-artifacts       # List artifacts without building")
}
