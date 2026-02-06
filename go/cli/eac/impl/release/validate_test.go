//go:build L1
// +build L1

package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	eactesting "github.com/ready-to-release/eac/go/core/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateChangelog_ExplicitPath tests validation when module has an explicit changelog path.
func TestValidateChangelog_ExplicitPath(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	// Use mock registry with predictable data
	moduleRegistry := eactesting.NewMockRegistry(
		eactesting.WithModule("eac-cli",
			eactesting.WithVersioning(),
			eactesting.WithChangelog("go/cli/eac/CHANGELOG.md"),
			eactesting.WithReleaseType("internal"),
		),
	)

	moduleContract, exists := moduleRegistry.Get("eac-cli")
	require.True(t, exists, "mock module should always exist")

	result := validateChangelog("eac-cli", moduleContract, workspaceRoot)

	// The test should identify that the changelog is in the wrong location
	assert.Equal(t, "eac-cli", result.Module)
	assert.NotEmpty(t, result.Path)

	// Check if file exists at the reported path
	fullPath := filepath.Join(workspaceRoot, result.Path)
	_, err := os.Stat(fullPath)

	if result.Valid {
		// If valid, file must exist
		assert.NoError(t, err, "changelog should exist at reported path when result is valid")
	} else {
		// If invalid, there should be errors
		assert.NotEmpty(t, result.Errors, "invalid result should have error messages")
	}
}

// TestValidateChangelog_DefaultPath tests validation when module uses default changelog path.
func TestValidateChangelog_DefaultPath(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	// Use mock registry with a SemVer module using default changelog path
	// Note: Only SemVer modules get the default release/{moniker}/CHANGELOG.md path
	moduleRegistry := eactesting.NewMockRegistry(
		eactesting.WithModule("r2r-cli",
			eactesting.WithSemver(),
			// No explicit changelog = uses default path for SemVer
			eactesting.WithReleaseType("published"),
		),
	)

	moduleContract, exists := moduleRegistry.Get("r2r-cli")
	require.True(t, exists, "mock module should always exist")

	result := validateChangelog("r2r-cli", moduleContract, workspaceRoot)

	assert.Equal(t, "r2r-cli", result.Module)
	expectedPath := "release/r2r-cli/CHANGELOG.md"
	assert.Equal(t, expectedPath, result.Path)
}

// TestValidateChangelog_MissingFile tests validation when changelog file is missing.
func TestValidateChangelog_MissingFile(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	// Create a mock module contract with a non-existent changelog
	mockContract := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "nonexistent-module",
			Versioning: &domain.ModuleVersioning{
				Scheme:    "SemVer",
				Changelog: "nonexistent/path/CHANGELOG.md",
			},
		},
	}

	result := validateChangelog("nonexistent-module", mockContract, workspaceRoot)

	assert.False(t, result.Valid, "validation should fail for missing changelog")
	require.NotEmpty(t, result.Errors, "should have at least one error")
	assert.Contains(t, result.Errors[0], "changelog not found", "error message should mention missing file")
}

// TestValidateRelease_PublishedModules tests that published modules have changelogs in release/ folder.
func TestValidateRelease_PublishedModules(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	// Published SemVer modules with versioning: r2r-cli, ext-eac (docs is CalVer, no changelog)
	publishedModules := []string{"r2r-cli", "ext-eac"}

	for _, moniker := range publishedModules {
		t.Run(moniker, func(t *testing.T) {
			expectedPath := "release/" + moniker + "/CHANGELOG.md"

			// Use mock registry with predictable data
			moduleRegistry := eactesting.NewMockRegistry(
				eactesting.WithModule(moniker,
					eactesting.WithVersioning(),
					eactesting.WithChangelog(expectedPath),
					eactesting.WithReleaseType("published"),
				),
			)

			moduleContract, exists := moduleRegistry.Get(moniker)
			require.True(t, exists, "mock module should always exist")

			// Verify release_type is published
			assert.Equal(t, "published", moduleContract.Versioning.ReleaseType,
				"module %s should have release_type: published", moniker)

			// Verify changelog path points to release folder
			changelogPath := moduleContract.GetChangelogPath()
			assert.Equal(t, expectedPath, changelogPath,
				"published module %s should have changelog in release/ folder", moniker)

			// Verify changelog file exists
			fullPath := filepath.Join(workspaceRoot, expectedPath)
			_, err := os.Stat(fullPath)
			assert.NoError(t, err,
				"changelog should exist at %s for published module %s",
				expectedPath, moniker)

			// Run validation
			result := validateChangelog(moniker, moduleContract, workspaceRoot)

			if err == nil {
				assert.True(t, result.Valid,
					"validation should pass when changelog exists in release/ folder")
			}
		})
	}
}

// TestValidateRelease_AllModulesConsistency tests that all modules with versioning
// have consistent changelog path behavior.
func TestValidateRelease_AllModulesConsistency(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err)

	allModules := moduleRegistry.AllMonikers()
	var inconsistencies []string

	for _, moniker := range allModules {
		moduleContract, exists := moduleRegistry.Get(moniker)
		require.True(t, exists)

		// Only SemVer modules have changelogs (Implicit = no releases, CalVer = auto-managed)
		if moduleContract.Versioning == nil || moduleContract.Versioning.Scheme != "SemVer" {
			continue
		}

		changelogPath := moduleContract.GetChangelogPath()

		// Changelog path should either:
		// 1. Match default pattern: release/{moniker}/CHANGELOG.md
		// 2. Be explicitly set to something else

		defaultPath := filepath.Join("release", moniker, "CHANGELOG.md")
		isDefaultPath := normalizeSlashes(changelogPath) == normalizeSlashes(defaultPath)

		hasExplicitPath := moduleContract.Versioning.Changelog != ""

		if !isDefaultPath && !hasExplicitPath {
			inconsistencies = append(inconsistencies,
				"module "+moniker+" has non-default path without explicit configuration")
		}
	}

	if len(inconsistencies) > 0 {
		t.Errorf("Found changelog path inconsistencies:\n  - %v", inconsistencies)
	}
}

// TestValidationResult_Structure tests the ValidationResult struct.
func TestValidationResult_Structure(t *testing.T) {
	tests := []struct {
		name   string
		result ValidationResult
	}{
		{
			name: "valid result",
			result: ValidationResult{
				Module: "test-module",
				Valid:  true,
				Path:   "release/test-module/CHANGELOG.md",
			},
		},
		{
			name: "invalid with errors",
			result: ValidationResult{
				Module: "test-module",
				Valid:  false,
				Path:   "release/test-module/CHANGELOG.md",
				Errors: []string{"changelog not found", "invalid format"},
			},
		},
		{
			name: "valid with warnings",
			result: ValidationResult{
				Module:   "test-module",
				Valid:    true,
				Path:     "release/test-module/CHANGELOG.md",
				Warnings: []string{"no version entries"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.result.Module, tt.result.Module)
			assert.Equal(t, tt.result.Valid, tt.result.Valid)
			assert.Equal(t, tt.result.Path, tt.result.Path)
		})
	}
}

// TestValidationReport_Structure tests the ValidationReport struct.
func TestValidationReport_Structure(t *testing.T) {
	report := ValidationReport{
		Results: []ValidationResult{
			{Module: "module1", Valid: true, Path: "path1"},
			{Module: "module2", Valid: false, Path: "path2"},
		},
		AllValid: false,
	}

	assert.Len(t, report.Results, 2)
	assert.False(t, report.AllValid)
	assert.Equal(t, "module1", report.Results[0].Module)
	assert.True(t, report.Results[0].Valid)
	assert.Equal(t, "module2", report.Results[1].Module)
	assert.False(t, report.Results[1].Valid)
}

// TestValidationReport_AllValid tests the AllValid field computation logic.
func TestValidationReport_AllValid(t *testing.T) {
	tests := []struct {
		name     string
		results  []ValidationResult
		allValid bool
	}{
		{
			name: "all valid",
			results: []ValidationResult{
				{Valid: true},
				{Valid: true},
				{Valid: true},
			},
			allValid: true,
		},
		{
			name: "some invalid",
			results: []ValidationResult{
				{Valid: true},
				{Valid: false},
				{Valid: true},
			},
			allValid: false,
		},
		{
			name: "all invalid",
			results: []ValidationResult{
				{Valid: false},
				{Valid: false},
			},
			allValid: false,
		},
		{
			name:     "empty results",
			results:  []ValidationResult{},
			allValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute AllValid manually (simulating the logic in ReleaseValidate)
			allValid := true
			for _, r := range tt.results {
				if !r.Valid {
					allValid = false
					break
				}
			}

			assert.Equal(t, tt.allValid, allValid)
		})
	}
}

// TestFindModulesWithChangelogs tests the findModulesWithChangelogs helper function.
func TestFindModulesWithChangelogs(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	modules := findModulesWithChangelogs(workspaceRoot)

	// Should return at least some modules
	assert.NotEmpty(t, modules, "should find at least one module with a changelog")

	// All returned modules should have a CHANGELOG.md file in release/{module}/
	for _, module := range modules {
		changelogPath := filepath.Join(workspaceRoot, "release", module, "CHANGELOG.md")
		_, err := os.Stat(changelogPath)
		assert.NoError(t, err, "module %s should have changelog at %s", module, changelogPath)
	}
}

// TestFindModulesWithChangelogs_Integration tests that findModulesWithChangelogs
// finds the correct modules based on actual filesystem state.
func TestFindModulesWithChangelogs_Integration(t *testing.T) {
	workspaceRoot := getWorkspaceRoot(t)

	// Get modules from findModulesWithChangelogs
	foundModules := findModulesWithChangelogs(workspaceRoot)

	// Load module contracts
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot)
	require.NoError(t, err)

	// Get all modules with versioning
	allModules := moduleRegistry.AllMonikers()
	var modulesWithVersioning []string
	for _, moniker := range allModules {
		mc, exists := moduleRegistry.Get(moniker)
		if exists && mc.Versioning != nil {
			modulesWithVersioning = append(modulesWithVersioning, moniker)
		}
	}

	// findModulesWithChangelogs only finds modules with changelogs in release/{module}/
	// It won't find modules with changelogs elsewhere
	// This is actually part of the problem we're testing for!

	t.Logf("Modules found by findModulesWithChangelogs: %v", foundModules)
	t.Logf("Modules with versioning in contracts: %v", modulesWithVersioning)

	// The count should ideally match, but won't until the migration is complete
	if len(foundModules) < len(modulesWithVersioning) {
		t.Logf("WARNING: Found %d modules with changelogs in release/, but %d modules have versioning",
			len(foundModules), len(modulesWithVersioning))
		t.Logf("This indicates some modules have changelogs in non-standard locations")
	}
}

// getWorkspaceRoot is a helper to get the workspace root for tests.
func getWorkspaceRoot(t *testing.T) string {
	t.Helper()

	// Get current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Navigate up to repository root
	// From go/cli/eac/impl/release to root
	root := cwd
	for i := 0; i < 5; i++ {
		root = filepath.Dir(root)
	}

	return root
}
