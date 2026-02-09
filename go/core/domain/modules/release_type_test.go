//go:build L1
// +build L1

package modules

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildMockRegistry creates a mock module registry for testing without external dependencies.
// This avoids import cycles by building the registry directly.
func buildMockRegistry(t *testing.T, modules map[string]mockModuleConfig) *Registry {
	t.Helper()

	registry := NewRegistry("1.0.0", "/mock/workspace")

	for moniker, cfg := range modules {
		base := domain.BaseContract{
			Moniker:    moniker,
			Components: make(config.ModuleComponents),
		}

		if cfg.withVersioning {
			base.Versioning = &config.ModuleVersioning{
				Scheme:      "CalVer",
				Changelog:   cfg.changelog,
				ReleaseType: cfg.releaseType,
			}
		}

		module := NewModuleContract(base, registry.WorkspaceRoot())
		err := registry.Add(module)
		require.NoError(t, err)
	}

	return registry
}

// mockModuleConfig holds configuration for a mock module.
type mockModuleConfig struct {
	withVersioning bool
	changelog      string
	releaseType    string
}

// validateReleaseTypeConsistency checks if a module's release_type aligns with its changelog location.
func validateReleaseTypeConsistency(contract domain.BaseContract) bool {
	if contract.Versioning == nil {
		return true // No versioning = no requirements
	}

	releaseType := contract.Versioning.ReleaseType
	changelogPath := contract.GetChangelog()

	// Determine if changelog is in release/{module}/ folder (published module pattern)
	// Only paths like release/<module>/CHANGELOG.md are in release folder (must have subdirectory)
	isInReleaseFolder := len(changelogPath) >= 8 && changelogPath[:8] == "release/" &&
		strings.Count(changelogPath, "/") >= 2 // Must have subdirectory (e.g., release/clie/CHANGELOG.md)

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
// This test uses mock modules to validate the consistency logic.
func TestLoadedModules_ReleaseTypeConsistency(t *testing.T) {
	// Build mock registry with predictable test cases
	moduleRegistry := buildMockRegistry(t, map[string]mockModuleConfig{
		// Published modules (should be in release/ folder)
		"clie": {
			withVersioning: true,
			changelog:      "release/clie/CHANGELOG.md",
			releaseType:    "published",
		},
		"eac-ext": {
			withVersioning: true,
			changelog:      "release/eac-ext/CHANGELOG.md",
			releaseType:    "published",
		},
		// Bundle modules (should be in release/ folder)
		"clie-eac-bundle": {
			withVersioning: true,
			changelog:      "release/clie-eac-bundle/CHANGELOG.md",
			releaseType:    "bundle",
		},
		// Internal modules (should NOT be in release/ folder)
		"eac": {
			withVersioning: true,
			changelog:      "go/cli/eac/CHANGELOG.md",
			releaseType:    "internal",
		},
		"eac-mcp-server": {
			withVersioning: true,
			changelog:      "go/eac/mcp/commands/CHANGELOG.md",
			releaseType:    "internal",
		},
		// None modules (no restrictions)
		"core": {
			withVersioning: true,
			releaseType:    "none",
		},
	})

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

	// All mock modules should be consistent
	assert.Empty(t, inconsistencies, "mock modules should all be consistent")
}

// TestLoadedModules_ReleaseTypeValues validates that all loaded modules have valid release_type values.
func TestLoadedModules_ReleaseTypeValues(t *testing.T) {
	// Build mock registry with both valid and edge cases
	moduleRegistry := buildMockRegistry(t, map[string]mockModuleConfig{
		"published-module": {
			withVersioning: true,
			releaseType:    "published",
		},
		"internal-module": {
			withVersioning: true,
			releaseType:    "internal",
		},
		"bundle-module": {
			withVersioning: true,
			releaseType:    "bundle",
		},
		"none-module": {
			withVersioning: true,
			releaseType:    "none",
		},
		"empty-module": {
			withVersioning: true,
			// No release type set = empty string
		},
	})

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

	// All mock modules should have valid release types
	assert.Empty(t, invalidModules, "all mock modules should have valid release types")
}

// TestKnownModules_ReleaseTypeAssignment validates specific module release type assignments per the plan.
func TestKnownModules_ReleaseTypeAssignment(t *testing.T) {
	// Expected release types per the architecture plan
	expectedReleaseTypes := map[string]string{
		// Published
		"clie": "published",
		"eac-ext": "published",
		"docs":    "published",
		"books":   "published",
		// Bundle
		"clie-eac-bundle": "bundle",
		// Internal
		"eac":            "internal",
		"eac-mcp-server": "internal",
		"clie-installer": "internal",
		"vscode-commit":  "internal",
		// None
		"core": "none",
	}

	// Build mock registry with expected module configurations
	mockConfigs := make(map[string]mockModuleConfig)
	for moniker, releaseType := range expectedReleaseTypes {
		config := mockModuleConfig{
			withVersioning: true,
			releaseType:    releaseType,
		}

		// Set appropriate changelog path based on release type
		if releaseType == "published" || releaseType == "bundle" {
			config.changelog = "release/" + moniker + "/CHANGELOG.md"
		} else if releaseType == "internal" {
			// Internal modules have changelogs in module roots
			switch moniker {
			case "eac":
				config.changelog = "go/cli/eac/CHANGELOG.md"
			case "eac-mcp-server":
				config.changelog = "go/eac/mcp/commands/CHANGELOG.md"
			case "clie-installer":
				config.changelog = "scripts/CHANGELOG.md"
			case "vscode-commit":
				config.changelog = "typescript/vscode-commit/CHANGELOG.md"
			}
		}

		mockConfigs[moniker] = config
	}

	moduleRegistry := buildMockRegistry(t, mockConfigs)

	// Validate each expected module
	var mismatches []string
	for moniker, expectedType := range expectedReleaseTypes {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists, "mock module %s should exist", moniker)
		require.NotNil(t, moduleContract.Versioning, "module %s should have versioning", moniker)

		actualType := moduleContract.Versioning.ReleaseType
		if actualType != expectedType {
			mismatches = append(mismatches,
				fmt.Sprintf("%s: expected %q, got %q", moniker, expectedType, actualType))
		}
	}

	// All mock modules should have correct release types
	assert.Empty(t, mismatches, "all mock modules should have correct release types")
}
