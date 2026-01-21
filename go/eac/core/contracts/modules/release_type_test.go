//go:build L1
// +build L1

package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// validateReleaseTypeConsistency checks if a module's release_type aligns with its changelog location.
func validateReleaseTypeConsistency(contract contracts.BaseContract) bool {
	if contract.Versioning == nil {
		return true // No versioning = no requirements
	}

	releaseType := contract.Versioning.ReleaseType
	changelogPath := contract.GetChangelog()

	// Determine if changelog is in release/{module}/ folder (published module pattern)
	// Only paths like release/<module>/CHANGELOG.md are in release folder (must have subdirectory)
	isInReleaseFolder := len(changelogPath) >= 8 && changelogPath[:8] == "release/" &&
		strings.Count(changelogPath, "/") >= 2 // Must have subdirectory (e.g., release/r2r-cli/CHANGELOG.md)

	switch releaseType {
	case "published", "bundle":
		// Must be in release/ folder
		return isInReleaseFolder
	case "internal":
		// Must NOT be in release/ folder
		return !isInReleaseFolder
	case "none", "":
		// No restrictions
		return true
	default:
		// Unknown release type
		return false
	}
}

// TestLoadedModules_ReleaseTypeConsistency validates that all loaded modules have consistent release_type configuration.
// This test loads the actual repository.yml and validates real module contracts.
func TestLoadedModules_ReleaseTypeConsistency(t *testing.T) {
	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine workspace root")
	}

	// Navigate up to repository root (from go/eac/core/contracts/modules to root)
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	// Load modules from repository.yml
	moduleRegistry, err := LoadFromWorkspace(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	// Track inconsistencies
	var inconsistencies []string

	// Test each module with versioning
	allMonikers := moduleRegistry.AllMonikers()
	for _, moniker := range allMonikers {
		moduleContract, exists := moduleRegistry.Get(moniker)
		if !exists || moduleContract.Versioning == nil {
			continue
		}

		// Check consistency
		isConsistent := validateReleaseTypeConsistency(moduleContract.BaseContract)
		if !isConsistent {
			releaseType := moduleContract.Versioning.ReleaseType
			changelogPath := moduleContract.GetChangelog()
			inconsistencies = append(inconsistencies,
				fmt.Sprintf("%s: release_type=%q, changelog=%q", moniker, releaseType, changelogPath))
		}
	}

	// Report any inconsistencies
	if len(inconsistencies) > 0 {
		t.Errorf("Found %d modules with inconsistent release_type configuration:\n  - %s",
			len(inconsistencies), strings.Join(inconsistencies, "\n  - "))
	}
}

// TestLoadedModules_ReleaseTypeValues validates that all loaded modules have valid release_type values.
func TestLoadedModules_ReleaseTypeValues(t *testing.T) {
	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine workspace root")
	}

	// Navigate up to repository root
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	// Load modules
	moduleRegistry, err := LoadFromWorkspace(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	// Valid release types
	validReleaseTypes := map[string]bool{
		"published": true,
		"internal":  true,
		"bundle":    true,
		"none":      true,
		"":          true, // empty is allowed (not specified)
	}

	// Track invalid values
	var invalidModules []string

	// Test each module
	allMonikers := moduleRegistry.AllMonikers()
	for _, moniker := range allMonikers {
		moduleContract, exists := moduleRegistry.Get(moniker)
		if !exists || moduleContract.Versioning == nil {
			continue
		}

		releaseType := moduleContract.Versioning.ReleaseType
		if !validReleaseTypes[releaseType] {
			invalidModules = append(invalidModules,
				fmt.Sprintf("%s: invalid release_type=%q", moniker, releaseType))
		}
	}

	// Report any invalid values
	if len(invalidModules) > 0 {
		t.Errorf("Found %d modules with invalid release_type values:\n  - %s",
			len(invalidModules), strings.Join(invalidModules, "\n  - "))
	}
}

// TestKnownModules_ReleaseTypeAssignment validates specific module release type assignments per the plan.
func TestKnownModules_ReleaseTypeAssignment(t *testing.T) {
	// Get workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine workspace root")
	}

	// Navigate up to repository root
	for i := 0; i < 5; i++ {
		workspaceRoot = filepath.Dir(workspaceRoot)
	}

	// Load modules
	moduleRegistry, err := LoadFromWorkspace(workspaceRoot)
	if err != nil {
		t.Fatalf("failed to load modules: %v", err)
	}

	// Expected release types per the architecture plan
	expectedReleaseTypes := map[string]string{
		// Published
		"r2r-cli":        "published",
		"ext-eac":        "published",
		"docs":           "published",
		"books":          "published",
		// Bundle
		"r2r-eac-bundle": "bundle",
		// Internal
		"eac-commands":     "internal",
		"eac-mcp-commands": "internal",
		"r2r-installer":    "internal",
		"vscode-ext-commit": "internal",
		// None
		"eac-core": "none",
	}

	// Validate each expected module
	var mismatches []string
	for moniker, expectedType := range expectedReleaseTypes {
		moduleContract, exists := moduleRegistry.Get(moniker)
		if !exists {
			t.Logf("Module %s not found in registry (may have been removed)", moniker)
			continue
		}

		if moduleContract.Versioning == nil {
			t.Errorf("Module %s has no versioning configured", moniker)
			continue
		}

		actualType := moduleContract.Versioning.ReleaseType
		if actualType != expectedType {
			// Debug: show full versioning info
			t.Logf("Module %s versioning: scheme=%s, changelog=%s, release_type=%s",
				moniker, moduleContract.Versioning.Scheme,
				moduleContract.Versioning.Changelog, moduleContract.Versioning.ReleaseType)
			mismatches = append(mismatches,
				fmt.Sprintf("%s: expected %q, got %q", moniker, expectedType, actualType))
		}
	}

	// Report any mismatches
	if len(mismatches) > 0 {
		t.Errorf("Found %d modules with incorrect release_type:\n  - %s",
			len(mismatches), strings.Join(mismatches, "\n  - "))
	}
}
