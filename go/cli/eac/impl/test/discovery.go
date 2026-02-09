// Package test provides test command discovery utilities
package test

import (
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/testrunners"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/testing"
)

// groupTestsByPackage groups tests by their package path for execution.
// Uses each test type's registered runner to determine grouping strategy:
// - BDD runners (godog, tscucumber): Uses runner.FindTestRoot + BuildPackagePath
// - Non-BDD runners (gotest, mocha): Groups by directory
func groupTestsByPackage(tests []testing.TestReference, workspaceRoot string, cfg *config.EACConfig) map[string][]testing.TestReference {
	testsByPackage := make(map[string][]testing.TestReference)

	// Track counts by type for logging
	typeCounts := make(map[string]int)
	typeSkipped := make(map[string]int)

	for i := range tests {
		test := &tests[i]
		var pkgPath string

		// Get the runner for this test type
		testRunner := testrunners.Get(test.Type)

		// Calculate relative path from workspace root
		relPath, err := filepath.Rel(workspaceRoot, test.FilePath)
		if err != nil {
			typeSkipped[test.Type]++
			continue
		}
		relPath = filepath.ToSlash(relPath)

		// Use the per-type descriptor's IsBDD (not the runner's) because a single
		// runner can handle both BDD and non-BDD types (e.g., GoTestRunner handles
		// "gotest" non-BDD and "godog" BDD).
		desc := testrunners.GetDescriptor(test.Type)
		isBDD := desc != nil && desc.IsBDD
		if testRunner != nil && isBDD {
			// BDD tests: use runner to find test root and build package path
			testRoot := testRunner.FindTestRoot(relPath, cfg)
			if testRoot == "" {
				// No test runner found - skip this test
				log.Debugf("groupTestsByPackage: skipping %s test (no test root): %s", test.Type, relPath)
				typeSkipped[test.Type]++
				continue
			}
			pkgPath = testRunner.BuildPackagePath(testRoot, relPath)
			log.Debugf("groupTestsByPackage: %s test grouped to %s (root=%s)", test.Type, pkgPath, testRoot)
		} else {
			// Non-BDD tests (gotest, mocha): group by directory
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
	log.Debugf("groupTestsByPackage: grouped %d tests into %d packages", len(tests), len(testsByPackage))
	for t, c := range typeCounts {
		skipped := typeSkipped[t]
		if skipped > 0 {
			log.Debugf("groupTestsByPackage: %s: %d grouped, %d skipped", t, c, skipped)
		} else {
			log.Debugf("groupTestsByPackage: %s: %d grouped", t, c)
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
