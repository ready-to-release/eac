// Command: build modules
// Description: Build multiple modules in sequence and collect results in a build run directory
// Usage: build modules [moniker1] [moniker2] ...
// Default: Builds all modules if no monikers specified
package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/modules"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register("build modules", BuildModules)
}

// BuildModules builds multiple modules in sequence (defaults to all modules)
func BuildModules() int {
	// Parse module monikers (no flags for build yet)
	var monikers []string

	// Parse arguments starting from index 3 (skip "binary", "build", "modules")
	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(os.Stderr, "Error: unknown flag: %s\n", arg)
			fmt.Fprintf(os.Stderr, "Usage: build modules [moniker1] [moniker2] ...\n")
			return 1
		} else {
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

	// Create build-run-id directory
	buildRunID := time.Now().Format("2006-01-02-150405")
	buildRunDir := filepath.Join(workspaceRoot, "out", "build-results", buildRunID)
	if err := os.MkdirAll(buildRunDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create build run directory: %v\n", err)
		return 1
	}

	fmt.Printf("Build Run ID: %s\n", buildRunID)
	fmt.Printf("Build Run Directory: %s\n", buildRunDir)
	fmt.Printf("Building %d modules: %v\n\n", len(monikers), monikers)

	// Build each module in sequence
	failedModules := []string{}
	builtModules := []*modules.ModuleContract{}
	for i, moniker := range monikers {
		fmt.Printf("=== [%d/%d] Building module: %s ===\n", i+1, len(monikers), moniker)

		// Get module from registry
		module, exists := moduleReport.Registry.Get(moniker)
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: module not found: %s\n", moniker)
			failedModules = append(failedModules, moniker+" (not found)")
			continue
		}

		// Create module output directory within build run
		moduleOutputDir := filepath.Join(buildRunDir, moniker)
		if err := os.MkdirAll(moduleOutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create module output directory: %v\n", err)
			failedModules = append(failedModules, moniker+" (dir error)")
			continue
		}

		// Create build log file
		logPath := filepath.Join(moduleOutputDir, "build.log")
		logFile, err := os.Create(logPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create log file: %v\n", err)
			failedModules = append(failedModules, moniker+" (log error)")
			continue
		}

		// Create multi-writer to log to both console and file
		multiWriter := io.MultiWriter(os.Stdout, logFile)

		// Run build for this module
		exitCode := runModuleBuild(module, workspaceRoot, moduleOutputDir, multiWriter)

		logFile.Close()

		// Track built modules
		builtModules = append(builtModules, module)

		if exitCode != 0 {
			failedModules = append(failedModules, moniker)
			fmt.Printf("❌ Module %s failed with exit code %d\n\n", moniker, exitCode)
		} else {
			fmt.Printf("✅ Module %s built successfully\n\n", moniker)
		}
	}

	// Print summary
	fmt.Println("===========================================")
	fmt.Printf("Build Run Summary (ID: %s)\n", buildRunID)
	fmt.Println("===========================================")
	fmt.Printf("Total modules: %d\n", len(monikers))
	fmt.Printf("Passed: %d\n", len(monikers)-len(failedModules))
	fmt.Printf("Failed: %d\n", len(failedModules))
	if len(failedModules) > 0 {
		fmt.Println("\nFailed modules:")
		for _, m := range failedModules {
			fmt.Printf("  - %s\n", m)
		}
	}
	fmt.Printf("\nResults directory: %s\n", buildRunDir)

	if len(failedModules) > 0 {
		return 1
	}
	return 0
}

// runModuleBuild runs build for a single module
func runModuleBuild(module *modules.ModuleContract, workspaceRoot string, outputDir string, logWriter io.Writer) int {
	// Get build function for module type
	buildFunc, hasBuilder := buildFunctions[module.Type]
	if !hasBuilder {
		fmt.Fprintf(logWriter, "Error: no build function for type: %s\n", module.Type)
		return 1
	}

	// Execute the build function
	return buildFunc(module, workspaceRoot, outputDir, logWriter)
}
