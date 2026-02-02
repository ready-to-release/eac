//go:build L0 && ov

// Package config provides source tracking tests for configuration files.
// These tests verify the GetLoadedFiles() function which returns information
// about all config files including their paths, layers, existence, and value counts.
package config

import (
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadedFile_Types verifies the LoadedFile type structure.
func TestLoadedFile_Types(t *testing.T) {
	t.Run("LoadedFile has required fields", func(t *testing.T) {
		file := LoadedFile{
			Path:   "/path/to/config.yml",
			Layer:  LayerContract,
			Exists: true,
			Values: 10,
		}

		assert.Equal(t, "/path/to/config.yml", file.Path)
		assert.Equal(t, LayerContract, file.Layer)
		assert.True(t, file.Exists)
		assert.Equal(t, 10, file.Values)
	})

	t.Run("LoadedConfig has required fields", func(t *testing.T) {
		cfg := LoadedConfig{
			Name:  "repository",
			Files: []LoadedFile{
				{Path: "/contract/repository.yml", Layer: LayerContract, Exists: true, Values: 5},
				{Path: "/user/repository.yml", Layer: LayerUser, Exists: true, Values: 3},
			},
		}

		assert.Equal(t, "repository", cfg.Name)
		assert.Len(t, cfg.Files, 2)
	})
}

// TestConfigLayer_Constants verifies the layer constant values.
func TestConfigLayer_Constants(t *testing.T) {
	tests := []struct {
		name     string
		layer    ConfigLayer
		expected string
	}{
		{"contract layer", LayerContract, "contract"},
		{"user layer", LayerUser, "user"},
		{"personal layer", LayerPersonal, "personal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.layer))
		})
	}
}

// TestGetLoadedFiles_EnumeratesAllConfigs verifies all expected config files are enumerated.
func TestGetLoadedFiles_EnumeratesAllConfigs(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	// Verify we get all expected config names
	expectedConfigs := []string{
		"repository",
		"component-types",
		"environments",
		"testing-tags",
		"test-suites",
		"books",
		"commands",
		"lint-providers",
	}

	configNames := make(map[string]bool)
	for _, lc := range files {
		configNames[lc.Name] = true
	}

	for _, expected := range expectedConfigs {
		assert.True(t, configNames[expected], "expected config %q to be enumerated", expected)
	}

	_ = repoRoot // Used for path verification below
}

// TestGetLoadedFiles_ContractDefaultsPath verifies contract defaults path is correct.
func TestGetLoadedFiles_ContractDefaultsPath(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	// Find the repository config and verify contract path
	var repoConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "repository" {
			repoConfig = &files[i]
			break
		}
	}
	require.NotNil(t, repoConfig, "repository config should be present")

	// Find the contract layer file
	var contractFile *LoadedFile
	for i := range repoConfig.Files {
		if repoConfig.Files[i].Layer == LayerContract {
			contractFile = &repoConfig.Files[i]
			break
		}
	}
	require.NotNil(t, contractFile, "contract layer file should be present")

	expectedPath := filepath.Join(repoRoot, "contracts", "core", paths.DefaultsVersion, "defaults", "repository.yml")
	assert.Equal(t, expectedPath, contractFile.Path)
}

// TestGetLoadedFiles_UserConfigPath verifies user config path is correct.
func TestGetLoadedFiles_UserConfigPath(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	// Find the repository config
	var repoConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "repository" {
			repoConfig = &files[i]
			break
		}
	}
	require.NotNil(t, repoConfig, "repository config should be present")

	// Find the user layer file
	var userFile *LoadedFile
	for i := range repoConfig.Files {
		if repoConfig.Files[i].Layer == LayerUser {
			userFile = &repoConfig.Files[i]
			break
		}
	}
	require.NotNil(t, userFile, "user layer file should be present")

	expectedPath := filepath.Join(repoRoot, ".eac", "repository.yml")
	assert.Equal(t, expectedPath, userFile.Path)
}

// TestGetLoadedFiles_ExistenceStatus verifies existence status is reported correctly.
func TestGetLoadedFiles_ExistenceStatus(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	t.Run("repository contract defaults exist", func(t *testing.T) {
		var repoConfig *LoadedConfig
		for i := range files {
			if files[i].Name == "repository" {
				repoConfig = &files[i]
				break
			}
		}
		require.NotNil(t, repoConfig)

		var contractFile *LoadedFile
		for i := range repoConfig.Files {
			if repoConfig.Files[i].Layer == LayerContract {
				contractFile = &repoConfig.Files[i]
				break
			}
		}
		require.NotNil(t, contractFile)
		assert.True(t, contractFile.Exists, "contract defaults should exist")
	})

	t.Run("repository user config exists", func(t *testing.T) {
		var repoConfig *LoadedConfig
		for i := range files {
			if files[i].Name == "repository" {
				repoConfig = &files[i]
				break
			}
		}
		require.NotNil(t, repoConfig)

		var userFile *LoadedFile
		for i := range repoConfig.Files {
			if repoConfig.Files[i].Layer == LayerUser {
				userFile = &repoConfig.Files[i]
				break
			}
		}
		require.NotNil(t, userFile)
		assert.True(t, userFile.Exists, "user repository.yml should exist")
	})

	t.Run("non-existent user configs report false", func(t *testing.T) {
		// lint-providers.yml is typically not present in user config
		var lintConfig *LoadedConfig
		for i := range files {
			if files[i].Name == "lint-providers" {
				lintConfig = &files[i]
				break
			}
		}
		require.NotNil(t, lintConfig)

		var userFile *LoadedFile
		for i := range lintConfig.Files {
			if lintConfig.Files[i].Layer == LayerUser {
				userFile = &lintConfig.Files[i]
				break
			}
		}
		require.NotNil(t, userFile)
		assert.False(t, userFile.Exists, "user lint-providers.yml should not exist")
	})
}

// TestGetLoadedFiles_ValueCounts verifies value counts are reported.
func TestGetLoadedFiles_ValueCounts(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	t.Run("existing files have positive value counts", func(t *testing.T) {
		var repoConfig *LoadedConfig
		for i := range files {
			if files[i].Name == "repository" {
				repoConfig = &files[i]
				break
			}
		}
		require.NotNil(t, repoConfig)

		for _, file := range repoConfig.Files {
			if file.Exists {
				assert.Greater(t, file.Values, 0, "existing file %s should have values", file.Path)
			}
		}
	})

	t.Run("non-existent files have zero value counts", func(t *testing.T) {
		for _, lc := range files {
			for _, file := range lc.Files {
				if !file.Exists {
					assert.Equal(t, 0, file.Values, "non-existent file %s should have zero values", file.Path)
				}
			}
		}
	})
}

// TestGetLoadedFiles_ComponentTypes verifies component-types config is tracked.
func TestGetLoadedFiles_ComponentTypes(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var ctConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "component-types" {
			ctConfig = &files[i]
			break
		}
	}
	require.NotNil(t, ctConfig, "component-types config should be tracked")

	// Contract defaults should exist
	var contractFile *LoadedFile
	for i := range ctConfig.Files {
		if ctConfig.Files[i].Layer == LayerContract {
			contractFile = &ctConfig.Files[i]
			break
		}
	}
	require.NotNil(t, contractFile)
	assert.True(t, contractFile.Exists, "component-types contract defaults should exist")
	assert.Greater(t, contractFile.Values, 0, "component-types should have values")
}

// TestGetLoadedFiles_EnvironmentsConfig verifies environments config is tracked.
func TestGetLoadedFiles_EnvironmentsConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var envConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "environments" {
			envConfig = &files[i]
			break
		}
	}
	require.NotNil(t, envConfig, "environments config should be tracked")

	// Should have both contract and user layers
	hasContract := false
	hasUser := false
	for _, file := range envConfig.Files {
		if file.Layer == LayerContract {
			hasContract = true
		}
		if file.Layer == LayerUser {
			hasUser = true
		}
	}
	assert.True(t, hasContract, "environments should have contract layer")
	assert.True(t, hasUser, "environments should have user layer")
}

// TestGetLoadedFiles_BooksConfig verifies books config is tracked.
func TestGetLoadedFiles_BooksConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var booksConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "books" {
			booksConfig = &files[i]
			break
		}
	}
	require.NotNil(t, booksConfig, "books config should be tracked")
}

// TestGetLoadedFiles_TestSuitesConfig verifies test-suites config is tracked.
func TestGetLoadedFiles_TestSuitesConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var suitesConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "test-suites" {
			suitesConfig = &files[i]
			break
		}
	}
	require.NotNil(t, suitesConfig, "test-suites config should be tracked")
}

// TestGetLoadedFiles_TestingTagsConfig verifies testing-tags config is tracked.
func TestGetLoadedFiles_TestingTagsConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var tagsConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "testing-tags" {
			tagsConfig = &files[i]
			break
		}
	}
	require.NotNil(t, tagsConfig, "testing-tags config should be tracked")
}

// TestGetLoadedFiles_CommandsConfig verifies commands config is tracked.
func TestGetLoadedFiles_CommandsConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var cmdConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "commands" {
			cmdConfig = &files[i]
			break
		}
	}
	require.NotNil(t, cmdConfig, "commands config should be tracked")
}

// TestGetLoadedFiles_LintProvidersConfig verifies lint-providers config is tracked.
func TestGetLoadedFiles_LintProvidersConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	var lintConfig *LoadedConfig
	for i := range files {
		if files[i].Name == "lint-providers" {
			lintConfig = &files[i]
			break
		}
	}
	require.NotNil(t, lintConfig, "lint-providers config should be tracked")
}

// TestGetLoadedFiles_AbsolutePaths verifies all paths are absolute.
func TestGetLoadedFiles_AbsolutePaths(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	for _, lc := range files {
		for _, file := range lc.Files {
			assert.True(t, filepath.IsAbs(file.Path),
				"path should be absolute: %s (config: %s, layer: %s)",
				file.Path, lc.Name, file.Layer)
		}
	}
}

// TestGetLoadedFiles_LayerOrdering verifies files are ordered by layer priority.
func TestGetLoadedFiles_LayerOrdering(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	// For configs with multiple files, contract should come before user
	for _, lc := range files {
		if len(lc.Files) < 2 {
			continue
		}

		lastLayerPriority := -1
		for _, file := range lc.Files {
			priority := layerPriority(file.Layer)
			assert.GreaterOrEqual(t, priority, lastLayerPriority,
				"layers should be in priority order for config %s", lc.Name)
			lastLayerPriority = priority
		}
	}
}

// layerPriority returns the priority of a layer (lower = earlier in merge order).
func layerPriority(layer ConfigLayer) int {
	switch layer {
	case LayerContract:
		return 0
	case LayerUser:
		return 1
	case LayerPersonal:
		return 2
	default:
		return 99
	}
}
