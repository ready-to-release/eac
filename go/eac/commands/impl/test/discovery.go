// Package test provides test command discovery utilities
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/test/runners"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

// groupTestsByPackage groups tests by their package path for execution.
// Different test types have different grouping strategies:
// - godog/tscucumber: Uses runner to find test root and build package path
// - mocha: Groups by test directory
// - gotest: Groups by directory.
func groupTestsByPackage(tests []testing.TestReference, workspaceRoot string, cfg *config.EACConfig) map[string][]testing.TestReference {
	testsByPackage := make(map[string][]testing.TestReference)

	for _, test := range tests {
		var pkgPath string

		// Get the runner for this test type
		testRunner := runners.Get(test.Type)

		// Calculate relative path from workspace root
		relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
		if err != nil {
			continue
		}
		relPath = filepath.ToSlash(relPath)

		if testRunner != nil && (test.Type == "godog" || test.Type == "tscucumber") {
			// BDD tests: use runner to find test root
			testRoot := testRunner.FindTestRoot(relPath, cfg)
			if testRoot == "" {
				// No test runner found - skip this test
				continue
			}
			pkgPath = testRunner.BuildPackagePath(testRoot, relPath)
		} else if test.Type == "mocha" {
			// Mocha tests: group by test directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
		} else {
			// Go tests (gotest): group by directory
			absDir := filepath.Dir(test.FilePath)
			relDir, err := filepath.Rel(workspaceRoot, absDir)
			if err != nil {
				continue
			}
			pkgPath = filepath.ToSlash(relDir)
		}
		testsByPackage[pkgPath] = append(testsByPackage[pkgPath], test)
	}

	return testsByPackage
}

// sanitizePathForLog converts a package path to a safe directory name.
func sanitizePathForLog(pkgPath string) string {
	// Replace colons and other special chars
	safe := strings.ReplaceAll(pkgPath, ":", "_")
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
	// e.g., "specs/vscode-ext-commit/progress-buffer/specification.feature" -> "vscode-ext-commit"
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

	// Return the module's root directory where cucumber.js should be
	return filepath.ToSlash(module.Files.Root)
}
