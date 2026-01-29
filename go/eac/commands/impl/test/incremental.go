// Package test provides incremental test detection utilities
package test

import (
	"path/filepath"
	"sort"
	"strings"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/testing"
	"github.com/ready-to-release/eac/go/eac/core/teststate"
)

// buildModuleTestInfo builds test file information for each module from the test packages.
// This is used for incremental test detection - tracking which source and test files
// belong to each module so we can detect when they change.
// Returns a map of module moniker to test files info, and a sorted list of module monikers.
func buildModuleTestInfo(
	testsByPackage map[string][]testing.TestReference,
	moduleRegistry *modules.Registry,
	eacCfg *config.EACConfig,
	workspaceRoot string,
) (map[string]teststate.ModuleTestFiles, []string) {
	moduleInfo := make(map[string]teststate.ModuleTestFiles)
	uniqueModules := make(map[string]bool)

	// Create a module mapper to find which module owns each package
	moduleMapper := NewModuleMapper(moduleRegistry, workspaceRoot)

	// Collect all test package paths per module
	testPackagesByModule := make(map[string][]string)
	for pkgPath := range testsByPackage {
		moduleMoniker := moduleMapper.GetModuleForPackagePath(pkgPath)
		if moduleMoniker != "" {
			uniqueModules[moduleMoniker] = true
			testPackagesByModule[moduleMoniker] = append(testPackagesByModule[moduleMoniker], pkgPath)
		}
	}

	// Build ModuleTestFiles for each module
	for moniker := range uniqueModules {
		module, exists := moduleRegistry.Get(moniker)
		if !exists {
			continue
		}

		info := teststate.ModuleTestFiles{
			Dependencies: module.GetDependencies(),
		}

		// Load build manifest to get BuildID (links tests to specific build)
		// This ensures `build --skip-cache` triggers retesting
		moduleBuildDir := eacCfg.Repository.BuildOutputPathAbs(workspaceRoot, moniker)
		if manifest, err := implinternal.LoadModuleManifest(moduleBuildDir); err == nil {
			info.BuildID = manifest.BuildID
		}

		// Get source files from module definition
		sourcePatterns := module.GetGlobPatterns()
		sourceFiles, err := teststate.ExpandGlobPatterns(workspaceRoot, sourcePatterns)
		if err == nil {
			// Filter to only include actual source files (not test files)
			for _, f := range sourceFiles {
				if !isTestFile(f) {
					info.SourceFiles = append(info.SourceFiles, f)
				}
			}
		}

		// Get test files from the test packages
		for _, pkgPath := range testPackagesByModule[moniker] {
			// Extract actual path from godog-style paths
			actualPath := pkgPath
			if idx := strings.Index(pkgPath, ":"); idx >= 0 {
				parts := strings.SplitN(pkgPath, ":", 3)
				if len(parts) >= 2 {
					actualPath = parts[1] // testRoot for godog paths
				} else {
					actualPath = parts[0]
				}
			}

			// Find test files in this package
			testGlobs := []string{
				filepath.Join(actualPath, "*_test.go"),
				filepath.Join(actualPath, "*.feature"),
			}
			testFiles, err := teststate.ExpandGlobPatterns(workspaceRoot, testGlobs)
			if err == nil {
				info.TestFiles = append(info.TestFiles, testFiles...)
			}
		}

		moduleInfo[moniker] = info
	}

	// Convert uniqueModules map to slice
	moduleList := make([]string, 0, len(uniqueModules))
	for m := range uniqueModules {
		moduleList = append(moduleList, m)
	}
	sort.Strings(moduleList)

	return moduleInfo, moduleList
}

// isTestFile returns true if the file path looks like a test file.
func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go") ||
		strings.HasSuffix(path, ".feature") ||
		strings.HasSuffix(path, ".test.ts") ||
		strings.HasSuffix(path, ".spec.ts")
}
