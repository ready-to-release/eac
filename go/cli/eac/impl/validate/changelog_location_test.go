//go:build L1
// +build L1

package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChangelogLocationValidation validates that all modules have changelogs at expected locations.
// This test is designed to FAIL initially as part of TDD to identify modules with changelogs in wrong places.
func TestChangelogLocationValidation(t *testing.T) {
	// Get workspace root
	workspaceRoot, err := os.Getwd()
	require.NoError(t, err)
	// Navigate up to repository root (from go/cli/eac/impl/validate to root)
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	// Load all modules from repository.yml
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err, "should load module contracts successfully")
	require.NotNil(t, moduleRegistry, "module registry should not be nil")

	// Get all modules
	allModules := moduleRegistry.AllMonikers()
	require.NotEmpty(t, allModules, "should have at least one module")

	var misalignedChangelogs []MisalignedChangelog

	// Test each module that has a versioning scheme
	for _, moniker := range allModules {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists, "module %s should exist in registry", moniker)

		// Skip modules without versioning, Implicit versioning (no explicit releases), or CalVer (no changelogs)
		if moduleContract.Versioning == nil || moduleContract.Versioning.Scheme == "Implicit" || moduleContract.Versioning.Scheme == "CalVer" {
			continue
		}

		// Get expected changelog path from contract
		expectedPath := moduleContract.GetChangelogPath()
		require.NotEmpty(t, expectedPath, "module %s should have a changelog path", moniker)

		// Check if changelog exists at expected location
		fullPath := filepath.Join(workspaceRoot, expectedPath)
		_, err := os.Stat(fullPath)

		if os.IsNotExist(err) {
			// Changelog doesn't exist at expected location - check if it exists elsewhere
			actualPath := findActualChangelogLocation(workspaceRoot, moniker, moduleContract)

			misalignedChangelogs = append(misalignedChangelogs, MisalignedChangelog{
				Module:       moniker,
				ExpectedPath: expectedPath,
				ActualPath:   actualPath,
				Exists:       actualPath != "",
			})
		} else {
			require.NoError(t, err, "should be able to stat changelog for module %s at %s", moniker, expectedPath)
		}
	}

	// Report misalignments
	if len(misalignedChangelogs) > 0 {
		t.Logf("Found %d modules with misaligned changelogs:", len(misalignedChangelogs))
		for _, misaligned := range misalignedChangelogs {
			if misaligned.Exists {
				t.Logf("  - %s:", misaligned.Module)
				t.Logf("      Expected: %s", misaligned.ExpectedPath)
				t.Logf("      Actual:   %s", misaligned.ActualPath)
			} else {
				t.Logf("  - %s: changelog not found (expected at %s)", misaligned.Module, misaligned.ExpectedPath)
			}
		}

		t.Errorf("Changelog location validation failed: %d module(s) have changelogs in wrong locations", len(misalignedChangelogs))
	}
}

// TestSpecificModuleChangelogLocations validates that modules with SemVer versioning have changelogs.
// Uses the core config system to determine which modules require changelogs.
func TestSpecificModuleChangelogLocations(t *testing.T) {
	// Get workspace root
	workspaceRoot, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	// Load modules from the real config
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err, "should load module contracts successfully")

	// Test each module with SemVer versioning
	for _, moniker := range moduleRegistry.AllMonikers() {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists, "module %s should exist in registry", moniker)

		// Skip modules without versioning or non-SemVer schemes
		if moduleContract.Versioning == nil || moduleContract.Versioning.Scheme != "SemVer" {
			continue
		}

		t.Run(moniker, func(t *testing.T) {
			// Get expected changelog path from contract
			expectedPath := moduleContract.GetChangelogPath()
			require.NotEmpty(t, expectedPath,
				"module %s with SemVer versioning should define a changelog path", moniker)

			// Check if changelog exists at expected location
			expectedFullPath := filepath.Join(workspaceRoot, expectedPath)
			_, err := os.Stat(expectedFullPath)

			assert.NoError(t, err,
				"changelog should exist at %s for SemVer module %s",
				expectedPath, moniker)
		})
	}
}

// TestAllModulesWithVersioningHaveChangelogs ensures every module with versioning has a changelog somewhere.
func TestAllModulesWithVersioningHaveChangelogs(t *testing.T) {
	workspaceRoot, err := os.Getwd()
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err)

	allModules := moduleRegistry.AllMonikers()
	var modulesWithoutChangelogs []string

	for _, moniker := range allModules {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists)

		// Skip modules without versioning, Implicit versioning (no explicit releases), or CalVer (no changelogs)
		if moduleContract.Versioning == nil || moduleContract.Versioning.Scheme == "Implicit" || moduleContract.Versioning.Scheme == "CalVer" {
			continue
		}

		// Check if changelog exists anywhere
		expectedPath := moduleContract.GetChangelogPath()
		expectedFullPath := filepath.Join(workspaceRoot, expectedPath)

		_, err := os.Stat(expectedFullPath)
		if os.IsNotExist(err) {
			// Try to find it in common locations
			actualPath := findActualChangelogLocation(workspaceRoot, moniker, moduleContract)
			if actualPath == "" {
				modulesWithoutChangelogs = append(modulesWithoutChangelogs, moniker)
			}
		}
	}

	if len(modulesWithoutChangelogs) > 0 {
		t.Errorf("Modules with versioning but no changelog found: %v", modulesWithoutChangelogs)
	}
}

// MisalignedChangelog represents a module with a changelog in the wrong location.
type MisalignedChangelog struct {
	Module       string
	ExpectedPath string
	ActualPath   string
	Exists       bool
}

// findActualChangelogLocation searches for a changelog in common locations.
// Returns empty string if not found.
func findActualChangelogLocation(workspaceRoot, moniker string, moduleContract *modules.ModuleContract) string {
	// Try common locations where changelogs might be
	possibleLocations := []string{
		// Expected location
		moduleContract.GetChangelogPath(),

		// Component roots + CHANGELOG.md
		filepath.Join(moduleContract.GetComponentRoot("go"), "CHANGELOG.md"),
		filepath.Join(moduleContract.GetComponentRoot("typescript"), "CHANGELOG.md"),
		filepath.Join(moduleContract.GetComponentRoot("bash"), "CHANGELOG.md"),
		filepath.Join(moduleContract.GetComponentRoot("pwsh"), "CHANGELOG.md"),

		// Root level
		"CHANGELOG.md",

		// Scripts directory (for installer modules)
		"scripts/CHANGELOG.md",

		// Containers directory (for Docker modules)
		filepath.Join("containers", moniker, "CHANGELOG.md"),
	}

	for _, location := range possibleLocations {
		if location == "" {
			continue
		}
		fullPath := filepath.Join(workspaceRoot, location)
		if _, err := os.Stat(fullPath); err == nil {
			return location
		}
	}

	return ""
}
