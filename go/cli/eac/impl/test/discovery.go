// Package test provides test command discovery utilities
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/test/runners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/testing"
)

var discoveryLog = logging.C()

// groupTestsByPackage groups tests by their package path for execution.
// Different test types have different grouping strategies:
// - godog/tscucumber: Uses runner to find test root and build package path
// - mocha: Groups by test directory
// - gotest: Groups by directory.
func groupTestsByPackage(tests []testing.TestReference, workspaceRoot string, cfg *config.EACConfig) map[string][]testing.TestReference {
	testsByPackage := make(map[string][]testing.TestReference)

	// Track counts by type for logging
	typeCounts := make(map[string]int)
	typeSkipped := make(map[string]int)

	for i := range tests {
		test := &tests[i]
		var pkgPath string

		// Get the runner for this test type
		testRunner := runners.Get(test.Type)

		// Calculate relative path from workspace root
		relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
		if err != nil {
			typeSkipped[test.Type]++
			continue
		}
		relPath = filepath.ToSlash(relPath)

		if testRunner != nil && (test.Type == "godog" || test.Type == "tscucumber") {
			// BDD tests: use runner to find test root
			testRoot := testRunner.FindTestRoot(relPath, cfg)
			if testRoot == "" {
				// No test runner found - skip this test
				discoveryLog.Debugf("groupTestsByPackage: skipping %s test (no test root): %s", test.Type, relPath)
				typeSkipped[test.Type]++
				continue
			}
			pkgPath = testRunner.BuildPackagePath(testRoot, relPath)
			discoveryLog.Debugf("groupTestsByPackage: %s test grouped to %s (root=%s)", test.Type, pkgPath, testRoot)
		} else if test.Type == "mocha" {
			// Mocha tests: group by test directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				typeSkipped[test.Type]++
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
			discoveryLog.Debugf("groupTestsByPackage: mocha test grouped to %s", pkgPath)
		} else {
			// Go tests (gotest): group by directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				typeSkipped[test.Type]++
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
		}
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], *test)
		typeCounts[test.Type]++
	}

	// Log summary
	discoveryLog.Debugf("groupTestsByPackage: grouped %d tests into %d packages", len(tests), len(testsByPackage))
	for t, c := range typeCounts {
		skipped := typeSkipped[t]
		if skipped > 0 {
			discoveryLog.Debugf("groupTestsByPackage: %s: %d grouped, %d skipped", t, c, skipped)
		} else {
			discoveryLog.Debugf("groupTestsByPackage: %s: %d grouped", t, c)
		}
	}

	return testsByPackage
}

// sanitizePathForLog converts a package path to a safe directory name.
func sanitizePathForLog(pkgPath string) string {
	// Replace colons with slashes for proper path hierarchy
	safe := strings.ReplaceAll(pkgPath, ":", "/")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}

// findTscucumberTestRunner finds the test runner location for a TypeScript cucumber feature file.
// Feature files are in specs/<module>/..., the test runner is in the module's root directory
// where cucumber.js is located.
// Uses the provided merged config to get correct paths.
// Returns empty string if no matching module is found.
func findTscucumberTestRunner(featurePath string, cfg *config.EACConfig) string {
	// Extract module moniker from specs path
	// e.g., "specs/vscode-commit/progress-buffer/specification.feature" -> "vscode-commit"
	specsPrefix := cfg.Repository.Paths.SpecsRoot + "/"
	relPath := strings.TrimPrefix(filepath.ToSlash(featurePath), specsPrefix)
	relPath = strings.TrimPrefix(relPath, strings.ReplaceAll(specsPrefix, "/", "\\"))
	relPath = filepath.ToSlash(relPath)

	// Get module moniker (first path component)
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return ""
	}
	moniker := parts[0]

	// Look up the module by moniker
	module, ok := cfg.Repository.GetModule(moniker)
	if !ok {
		return ""
	}

	// Return the module's typescript package root where cucumber.js should be
	tsRoot := module.Components.GetComponentRoot("typescript")
	if tsRoot == "" {
		return ""
	}
	return filepath.ToSlash(tsRoot)
}
