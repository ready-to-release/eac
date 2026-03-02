//go:build L0 && ov

package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers to eliminate repeated lookup boilerplate.

func findConfig(t *testing.T, files []LoadedConfig, name string) *LoadedConfig {
	t.Helper()
	for i := range files {
		if files[i].Name == name {
			return &files[i]
		}
	}
	t.Fatalf("config %q not found", name)
	return nil
}

func findFileByLayer(t *testing.T, lc *LoadedConfig, layer ConfigLayer) *LoadedFile {
	t.Helper()
	for i := range lc.Files {
		if lc.Files[i].Layer == layer {
			return &lc.Files[i]
		}
	}
	t.Fatalf("layer %q not found in config %q", layer, lc.Name)
	return nil
}

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

func TestGetLoadedFiles_EnumeratesAllConfigs(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	expectedConfigs := []string{
		"repository", "environments", "testing-tags",
		"test-suites", "books", "commands", "lint-providers",
	}

	configNames := make(map[string]bool)
	for _, lc := range files {
		configNames[lc.Name] = true
	}

	for _, expected := range expectedConfigs {
		assert.True(t, configNames[expected], "expected config %q to be enumerated", expected)
	}
}

func TestGetLoadedFiles_ContractDefaultsPath(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()
	contractFile := findFileByLayer(t, findConfig(t, files, "repository"), LayerContract)

	assert.Equal(t, "embedded:contracts/core/defaults/repository.yml", contractFile.Path)
}

func TestGetLoadedFiles_UserConfigPath(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()
	userFile := findFileByLayer(t, findConfig(t, files, "repository"), LayerUser)

	expectedPath := filepath.Join(repoRoot, ".eac", "repository.yml")
	assert.Equal(t, expectedPath, userFile.Path)
}

func TestGetLoadedFiles_ExistenceStatus(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	t.Run("repository contract defaults exist", func(t *testing.T) {
		contractFile := findFileByLayer(t, findConfig(t, files, "repository"), LayerContract)
		assert.True(t, contractFile.Exists, "contract defaults should exist")
	})

	t.Run("repository user config exists", func(t *testing.T) {
		userFile := findFileByLayer(t, findConfig(t, files, "repository"), LayerUser)
		assert.True(t, userFile.Exists, "user repository.yml should exist")
	})

	t.Run("non-existent user configs report false", func(t *testing.T) {
		userFile := findFileByLayer(t, findConfig(t, files, "lint-providers"), LayerUser)
		assert.False(t, userFile.Exists, "user lint-providers.yml should not exist")
	})
}

func TestGetLoadedFiles_ValueCounts(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	t.Run("existing files have positive value counts", func(t *testing.T) {
		repoConfig := findConfig(t, files, "repository")
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

func TestGetLoadedFiles_EnvironmentsConfig(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	envConfig := findConfig(t, cfg.GetLoadedFiles(), "environments")

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

func TestGetLoadedFiles_ConfigTracking(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	files := cfg.GetLoadedFiles()

	for _, name := range []string{"books", "test-suites", "testing-tags", "commands", "lint-providers"} {
		t.Run(name, func(t *testing.T) {
			findConfig(t, files, name) // Fatals if not found
		})
	}
}

func TestGetLoadedFiles_PathFormats(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	for _, lc := range cfg.GetLoadedFiles() {
		for _, file := range lc.Files {
			if file.Layer == LayerContract {
				assert.True(t, strings.HasPrefix(file.Path, "embedded:"),
					"contract path should start with embedded: prefix: %s (config: %s)",
					file.Path, lc.Name)
			} else {
				assert.True(t, filepath.IsAbs(file.Path),
					"path should be absolute: %s (config: %s, layer: %s)",
					file.Path, lc.Name, file.Layer)
			}
		}
	}
}

func TestGetLoadedFiles_LayerOrdering(t *testing.T) {
	cfg, err := Load(DefaultLoadOptions())
	require.NoError(t, err)

	for _, lc := range cfg.GetLoadedFiles() {
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
