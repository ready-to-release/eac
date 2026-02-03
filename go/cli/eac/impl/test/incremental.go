// Package test provides incremental test detection utilities
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/hash"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/testing"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// buildModuleTestInfo builds test file information for each module from the test packages.
// This is used for incremental test detection - tracking which source and test files
// belong to each module so we can detect when they change.
// Returns a map of module moniker to test files info, and a loader for dependency BuildIDs.
func buildModuleTestInfo(
	testsByPackage map[string][]testing.TestReference,
	moduleRegistry *modules.Registry,
	eacCfg *config.EACConfig,
	workspaceRoot string,
) (map[string]workunit.TestModuleInfo, workunit.DependencyBuildIDLoader) {
	moduleInfo := make(map[string]workunit.TestModuleInfo)
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

	// Build TestModuleInfo for each module
	for moniker := range uniqueModules {
		module, exists := moduleRegistry.Get(moniker)
		if !exists {
			continue
		}

		info := workunit.TestModuleInfo{
			Dependencies: module.GetDependencies(),
		}

		// Get BuildID from UoW manifests (links tests to specific build)
		// This ensures `build --skip-cache` triggers retesting
		reader := coreoutput.NewReader(workspaceRoot)
		info.BuildID = reader.GetBuildID(workunit.ContextBuild, moniker)

		// Get source files from module definition
		sourcePatterns := module.GetGlobPatterns()
		sourceFiles, err := hash.ExpandGlobPatterns(workspaceRoot, sourcePatterns)
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
			testFiles, err := hash.ExpandGlobPatterns(workspaceRoot, testGlobs)
			if err == nil {
				info.TestFiles = append(info.TestFiles, testFiles...)
			}
		}

		moduleInfo[moniker] = info
	}

	// Create a loader for dependency BuildIDs
	depBuildIDLoader := func(moniker string) string {
		reader := coreoutput.NewReader(workspaceRoot)
		return reader.GetBuildID(workunit.ContextBuild, moniker)
	}

	return moduleInfo, depBuildIDLoader
}

// isTestFile returns true if the file path looks like a test file.
func isTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go") ||
		strings.HasSuffix(path, ".feature") ||
		strings.HasSuffix(path, ".test.ts") ||
		strings.HasSuffix(path, ".spec.ts")
}
