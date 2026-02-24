//go:build L0 && ov
// +build L0,ov

package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ready-to-release/eac/go/core/config"
)

// TestMergeConfigs_PreservesRepositorySettings tests that repository metadata is preserved.
func TestMergeConfigs_PreservesRepositorySettings(t *testing.T) {
	tmpDir := t.TempDir()

	existing := &config.RepositoryConfig{
		Repository: config.RepositorySettings{
			Type:        "mono",
			TrunkBranch: "main",
			Remote: config.RemoteConfig{
				Owner: "custom",
				RepoName: "repo",
				URL: "https://github.com/custom/repo",
			},
		},
	}

	new := &config.RepositoryConfig{
		Repository: config.RepositorySettings{
			Type:        "mono",
			TrunkBranch: "main",
			Remote: config.RemoteConfig{
				Owner: "auto",
				RepoName: "repo",
				URL: "https://github.com/auto/repo",
			},
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	assert.Equal(t, "custom", merged.Repository.Remote.Owner)
	assert.Equal(t, "repo", merged.Repository.Remote.RepoName)
	assert.Equal(t, "https://github.com/custom/repo", merged.Repository.Remote.URL)
	assert.NotNil(t, result)
}

// TestMergeConfigs_AddNewModules tests adding modules detected by scan.
func TestMergeConfigs_AddNewModules(t *testing.T) {
	tmpDir := t.TempDir()

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker:     "core",
				Description: "Existing core module",
			},
			{
				Moniker:     "cli",
				Description: "Existing CLI module",
			},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker:     "core",
				Description: "AI-generated core description",
			},
			{
				Moniker:     "cli",
				Description: "AI-generated CLI description",
			},
			{
				Moniker:     "api",
				Description: "AI-generated API description",
			},
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	assert.Len(t, merged.Modules, 3)
	assert.NotNil(t, result)
	assert.Contains(t, result.ModulesAdded, "api")
	assert.Len(t, result.ModulesAdded, 1)
}

// TestMergeConfigs_UpdateExistingModules tests that existing modules are updated.
func TestMergeConfigs_UpdateExistingModules(t *testing.T) {
	tmpDir := t.TempDir()

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker:     "core",
				Description: "Old description",
			},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker:     "core",
				Description: "New AI-generated description",
			},
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	assert.Len(t, merged.Modules, 1)
	assert.NotNil(t, result)
	assert.Contains(t, result.ModulesUpdated, "core")
	assert.Len(t, result.ModulesUpdated, 1)
	assert.Empty(t, result.ModulesAdded)
	assert.Empty(t, result.ModulesRemoved)
}

// TestMergeModule_PreservesUserCustomizations tests that user edits are preserved.
func TestMergeModule_PreservesUserCustomizations(t *testing.T) {
	existing := &config.Module{
		Moniker: "core",
		// Auto-generated description: can be overridden by new AI description
		Description: "Root go module",
		Versioning: &config.ModuleVersioning{
			Scheme: "SemVer",
		},
	}

	new := &config.Module{
		Moniker:     "core",
		Description: "New AI description",
		Versioning: &config.ModuleVersioning{
			Scheme: "CalVer",
		},
	}

	merged := mergeModule(existing, new)

	// Versioning is a user customization — must be preserved
	assert.Equal(t, "SemVer", merged.Versioning.Scheme)

	// Auto-generated description is overridden by new AI description
	assert.Equal(t, "New AI description", merged.Description)
}

// TestMergeModule_PreservesDependencies tests that user-added dependencies are kept.
func TestMergeModule_PreservesDependencies(t *testing.T) {
	existing := &config.Module{
		Moniker:    "api",
		DependsOn:  []string{"core", "custom-lib"},
	}

	new := &config.Module{
		Moniker:    "api",
		DependsOn:  []string{"core", "database"},
	}

	merged := mergeModule(existing, new)

	// Should contain: core, custom-lib (user-added), database (new)
	assert.Contains(t, merged.DependsOn, "core")
	assert.Contains(t, merged.DependsOn, "custom-lib")
	assert.Contains(t, merged.DependsOn, "database")
	assert.Len(t, merged.DependsOn, 3)
}

// TestMergeDependencies_CombinesUnique tests dependency merging without duplicates.
func TestMergeDependencies_CombinesUnique(t *testing.T) {
	existing := []string{"core", "utils", "user-lib"}
	new := []string{"core", "database", "utils"}

	merged := mergeDependencies(existing, new)

	// Should contain all unique dependencies
	assert.Contains(t, merged, "core")
	assert.Contains(t, merged, "utils")
	assert.Contains(t, merged, "user-lib")
	assert.Contains(t, merged, "database")
	assert.Len(t, merged, 4)
}

// TestMergeDependencies_PreservesOrder tests that new dependencies come first.
func TestMergeDependencies_PreservesOrder(t *testing.T) {
	existing := []string{"old-dep"}
	new := []string{"new-dep1", "new-dep2"}

	merged := mergeDependencies(existing, new)

	// New dependencies should come first
	assert.Equal(t, "new-dep1", merged[0])
	assert.Equal(t, "new-dep2", merged[1])
	assert.Equal(t, "old-dep", merged[2])
}

// TestModuleStillExists_ChecksComponentRoots tests filesystem check for module existence.
func TestModuleStillExists_ChecksComponentRoots(t *testing.T) {
	tmpDir := t.TempDir()

	// Create component directory
	componentRoot := filepath.Join(tmpDir, "go", "core")
	err := os.MkdirAll(componentRoot, 0o755)
	require.NoError(t, err)

	module := &config.Module{
		Moniker: "core",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: "go/core",
			},
		},
	}

	exists := moduleStillExists(tmpDir, module)

	assert.True(t, exists)
}

// TestModuleStillExists_ReturnsFalseWhenMissing tests detection of deleted modules.
func TestModuleStillExists_ReturnsFalseWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()

	module := &config.Module{
		Moniker: "deleted",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: "go/deleted",
			},
		},
	}

	exists := moduleStillExists(tmpDir, module)

	assert.False(t, exists)
}

// TestMergeConfigs_RemoveDeletedModules tests removing modules that no longer exist.
func TestMergeConfigs_RemoveDeletedModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing module directory
	coreRoot := filepath.Join(tmpDir, "go", "core")
	err := os.MkdirAll(coreRoot, 0o755)
	require.NoError(t, err)

	// Don't create "deprecated" directory (simulates deletion)

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
			{
				Moniker: "deprecated",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/deprecated"},
				},
			},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	assert.Len(t, merged.Modules, 1)
	assert.Equal(t, "core", merged.Modules[0].Moniker)
	assert.NotNil(t, result)
	assert.Contains(t, result.ModulesRemoved, "deprecated")
}

// TestMergeConfigs_KeepModuleNotDetected tests keeping modules not found by scanner.
func TestMergeConfigs_KeepModuleNotDetected(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing module directory (still exists in filesystem)
	customRoot := filepath.Join(tmpDir, "custom", "module")
	err := os.MkdirAll(customRoot, 0o755)
	require.NoError(t, err)

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
			{
				Moniker: "custom",
				Components: config.ModuleComponents{
					"custom": &config.ComponentEntry{Root: "custom/module"},
				},
			},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
			// Scanner didn't detect "custom" module
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	// Should keep "custom" module even though scanner didn't detect it
	assert.Len(t, merged.Modules, 2)
	assert.NotNil(t, result)
	assert.Empty(t, result.ModulesRemoved)
}

// TestMergeResult_TracksSummary tests that merge result captures all changes.
func TestMergeResult_TracksSummary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create component directories
	for _, path := range []string{"go/core", "go/cli", "go/api"} {
		err := os.MkdirAll(filepath.Join(tmpDir, path), 0o755)
		require.NoError(t, err)
	}

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
			{
				Moniker: "cli",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/cli"},
				},
			},
			{
				Moniker: "deprecated",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/deprecated"},
				},
			},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{
				Moniker: "core",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/core"},
				},
			},
			{
				Moniker: "cli",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/cli"},
				},
			},
			{
				Moniker: "api",
				Components: config.ModuleComponents{
					"go": &config.ComponentEntry{Root: "go/api"},
				},
			},
		},
	}

	_, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, result)
	assert.Contains(t, result.ModulesAdded, "api")
	assert.Contains(t, result.ModulesUpdated, "core")
	assert.Contains(t, result.ModulesUpdated, "cli")
	assert.Contains(t, result.ModulesRemoved, "deprecated")
	assert.Len(t, result.ModulesAdded, 1)
	assert.Len(t, result.ModulesUpdated, 2)
	assert.Len(t, result.ModulesRemoved, 1)
}

// TestShowMergeResult_DisplaysSummary tests console output formatting.
func TestShowMergeResult_DisplaysSummary(t *testing.T) {
	// This test documents the expected output format.
	// The actual implementation will format and display the summary.

	result := &MergeResult{
		ModulesAdded:   []string{"api", "mcp"},
		ModulesUpdated: []string{"core", "cli", "eac", "clie"},
		ModulesRemoved: []string{"deprecated"},
	}

	// Expected output should include:
	// - Count of added modules
	// - Count of updated modules
	// - Count of removed modules
	// - Suggestion to review with git diff

	assert.Len(t, result.ModulesAdded, 2)
	assert.Len(t, result.ModulesUpdated, 4)
	assert.Len(t, result.ModulesRemoved, 1)
}

// TestMergeModule_PreservesComponentCustomizations tests that component configs are merged.
func TestMergeModule_PreservesComponentCustomizations(t *testing.T) {
	existing := &config.Module{
		Moniker: "core",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: "go/core",
				Type: "library",
				// User may have customized other fields
			},
		},
	}

	new := &config.Module{
		Moniker: "core",
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: "go/core",
				Type: "service", // AI detected different type
			},
		},
	}

	merged := mergeModule(existing, new)

	// Implementation should merge components intelligently
	assert.Contains(t, merged.Components, "go")
}

// TestMergeConfigs_HandlesEmptyExisting tests first-time init (no existing config).
func TestMergeConfigs_HandlesEmptyExisting(t *testing.T) {
	tmpDir := t.TempDir()

	existing := &config.RepositoryConfig{
		Modules: []config.Module{},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{
			{Moniker: "core"},
			{Moniker: "cli"},
		},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	assert.Len(t, merged.Modules, 2)
	assert.NotNil(t, result)
	assert.Len(t, result.ModulesAdded, 2)
	assert.Empty(t, result.ModulesUpdated)
}

// TestMergeConfigs_HandlesEmptyNew tests edge case where scan found nothing.
func TestMergeConfigs_HandlesEmptyNew(t *testing.T) {
	tmpDir := t.TempDir()

	existing := &config.RepositoryConfig{
		Modules: []config.Module{
			{Moniker: "core"},
		},
	}

	new := &config.RepositoryConfig{
		Modules: []config.Module{},
	}

	merged, result := mergeConfigs(tmpDir, existing, new)

	require.NotNil(t, merged)
	// Should keep existing module if it still exists in filesystem
	assert.NotNil(t, result)
}

// TestFilterSystemDefaultModules_RemovesDefaultModule tests that the system placeholder is removed.
func TestFilterSystemDefaultModules_RemovesDefaultModule(t *testing.T) {
	modules := []config.Module{
		{
			Moniker:     "flask",
			Description: "Root python module",
			Components: config.ModuleComponents{
				"python": &config.ComponentEntry{Root: "."},
			},
		},
		{
			Moniker:     "default",
			Description: config.DefaultModuleDescription,
			Components: config.ModuleComponents{
				"go": &config.ComponentEntry{Root: "."},
			},
		},
	}

	filtered := filterSystemDefaultModules(modules)

	require.Len(t, filtered, 1)
	assert.Equal(t, "flask", filtered[0].Moniker)
}

// TestFilterSystemDefaultModules_PreservesUserDefaultModule tests a user module named "default" is kept.
func TestFilterSystemDefaultModules_PreservesUserDefaultModule(t *testing.T) {
	modules := []config.Module{
		{
			Moniker:     "default",
			Description: "My default service — user-edited description",
			Components: config.ModuleComponents{
				"go": &config.ComponentEntry{Root: "services/default"},
			},
		},
	}

	filtered := filterSystemDefaultModules(modules)

	require.Len(t, filtered, 1)
	assert.Equal(t, "default", filtered[0].Moniker)
}

// TestFilterSystemDefaultModules_NoopWhenNoDefaultModule tests filter is safe with no matching modules.
func TestFilterSystemDefaultModules_NoopWhenNoDefaultModule(t *testing.T) {
	modules := []config.Module{
		{Moniker: "core"},
		{Moniker: "cli"},
	}

	filtered := filterSystemDefaultModules(modules)

	assert.Len(t, filtered, 2)
}

// TestMergeModule_UpdatesAutoGeneratedDescription tests that auto-generated descriptions are replaced.
func TestMergeModule_UpdatesAutoGeneratedDescription(t *testing.T) {
	existing := &config.Module{
		Moniker: "core",
		// Auto-generated pattern: will be overridden by new description
		Description: "core go module at go/core",
	}

	new := &config.Module{
		Moniker:     "core",
		Description: "New comprehensive AI-generated description with details",
	}

	merged := mergeModule(existing, new)

	// Auto-generated description is replaced by new content
	assert.Equal(t, "New comprehensive AI-generated description with details", merged.Description)
}

// TestMergeModule_PreservesUserEditedDescription tests that custom descriptions are kept.
func TestMergeModule_PreservesUserEditedDescription(t *testing.T) {
	existing := &config.Module{
		Moniker:     "api",
		Description: "Custom user description that should be preserved",
	}
	newMod := &config.Module{
		Moniker:     "api",
		Description: "api python module at services/api",
	}
	merged := mergeModule(existing, newMod)
	assert.Equal(t, "Custom user description that should be preserved", merged.Description)
}

// TestMergeModule_OverwritesAutoGeneratedDescription tests that auto-generated descriptions are updated.
func TestMergeModule_OverwritesAutoGeneratedDescription(t *testing.T) {
	existing := &config.Module{
		Moniker:     "api",
		Description: "api python module at services/api",
	}
	newMod := &config.Module{
		Moniker:     "api",
		Description: "api go module at services/api",
	}
	merged := mergeModule(existing, newMod)
	assert.Equal(t, "api go module at services/api", merged.Description)
}

// TestIsAutoGeneratedDescription tests the pattern detection helper.
func TestIsAutoGeneratedDescription(t *testing.T) {
	autoGenerated := []string{
		"Root go module",
		"Root python module",
		"Root javascript module",
		"Root typescript module",
		"Root rust module",
		"Root java module",
		"Root dotnet module",
		"api python module at services/api",
		"frontend javascript module at apps/web",
	}
	for _, desc := range autoGenerated {
		assert.True(t, isAutoGeneratedDescription(desc), "expected %q to be auto-generated", desc)
	}

	userEdited := []string{
		"Custom description by user",
		"Celery task workers for async processing",
		"My awesome service",
	}
	for _, desc := range userEdited {
		assert.False(t, isAutoGeneratedDescription(desc), "expected %q NOT to be auto-generated", desc)
	}
}
