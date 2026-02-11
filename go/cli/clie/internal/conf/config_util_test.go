//go:build L2
// +build L2

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResetConfigLoaded verifies ResetConfigLoaded resets flag
func TestResetConfigLoaded(t *testing.T) {
	// Set configLoaded to true
	configLoaded = true
	assert.True(t, configLoaded)

	// Reset it
	ResetConfigLoaded()
	assert.False(t, configLoaded, "configLoaded should be reset to false")
}

// TestValidationError_Error verifies error message formatting
func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Errors: []string{
			"error 1",
			"error 2",
			"error 3",
		},
	}

	errorMsg := ve.Error()
	assert.Contains(t, errorMsg, "configuration validation failed")
	assert.Contains(t, errorMsg, "error 1")
	assert.Contains(t, errorMsg, "error 2")
	assert.Contains(t, errorMsg, "error 3")
}

// TestValidationError_Add verifies adding errors
func TestValidationError_Add(t *testing.T) {
	ve := &ValidationError{}
	assert.Empty(t, ve.Errors)

	ve.Add("first error")
	assert.Len(t, ve.Errors, 1)
	assert.Equal(t, "first error", ve.Errors[0])

	ve.Add("second error")
	assert.Len(t, ve.Errors, 2)
	assert.Equal(t, "second error", ve.Errors[1])
}

// TestValidationError_HasErrors verifies error detection
func TestValidationError_HasErrors(t *testing.T) {
	ve := &ValidationError{}
	assert.False(t, ve.HasErrors(), "Should have no errors initially")

	ve.Add("an error")
	assert.True(t, ve.HasErrors(), "Should have errors after adding one")
}

// TestMergeConfigFile_BasicMerge verifies config file merging
func TestMergeConfigFile_BasicMerge(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()

	// Create base config file
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `registry:
  default: ghcr.io
extensions:
  - name: test-ext
    image: test/image:v1.0.0
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	// Load base config
	ResetConfigLoaded()
	Global = Config{} // Reset global
	require.NoError(t, LoadConfig(baseConfig))

	// Verify base config loaded
	assert.Equal(t, "ghcr.io", Global.Registry.Default)
	assert.Len(t, Global.Extensions, 1)
	assert.Equal(t, "test/image:v1.0.0", Global.Extensions[0].Image)

	// Create override config file
	overrideConfig := filepath.Join(tmpDir, "override.yml")
	overrideYAML := `registry:
  default: docker.io
extensions:
  - name: test-ext
    image: test/image:v2.0.0
    load_local: true
`
	require.NoError(t, os.WriteFile(overrideConfig, []byte(overrideYAML), 0644))

	// Merge override config
	require.NoError(t, MergeConfigFile(overrideConfig))

	// Verify merged config
	assert.Equal(t, "docker.io", Global.Registry.Default, "Registry should be overridden")
	assert.Len(t, Global.Extensions, 1)
	assert.Equal(t, "test/image:v2.0.0", Global.Extensions[0].Image, "Image should be overridden")
	assert.True(t, Global.Extensions[0].LoadLocal, "LoadLocal should be set")
}

// TestMergeConfigFile_AddNewExtension verifies adding new extension via override
func TestMergeConfigFile_AddNewExtension(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()

	// Create base config file
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `extensions:
  - name: ext1
    image: ext1:v1
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	// Load base config
	ResetConfigLoaded()
	Global = Config{}
	require.NoError(t, LoadConfig(baseConfig))
	assert.Len(t, Global.Extensions, 1)

	// Create override config with new extension
	overrideConfig := filepath.Join(tmpDir, "override.yml")
	overrideYAML := `extensions:
  - name: ext2
    image: ext2:v1
`
	require.NoError(t, os.WriteFile(overrideConfig, []byte(overrideYAML), 0644))

	// Merge override config
	require.NoError(t, MergeConfigFile(overrideConfig))

	// Verify both extensions present
	assert.Len(t, Global.Extensions, 2)
	extMap := make(map[string]string)
	for _, ext := range Global.Extensions {
		extMap[ext.Name] = ext.Image
	}
	assert.Equal(t, "ext1:v1", extMap["ext1"])
	assert.Equal(t, "ext2:v1", extMap["ext2"])
}

// TestMergeConfigFile_PartialOverride verifies partial config override
func TestMergeConfigFile_PartialOverride(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()

	// Create base config with multiple settings
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `registry:
  default: ghcr.io
  timeout: 30
defaults:
  pull_policy: IfNotPresent
  timeout: 120
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	// Load base config
	ResetConfigLoaded()
	Global = Config{}
	require.NoError(t, LoadConfig(baseConfig))

	// Create override config with only partial settings
	overrideConfig := filepath.Join(tmpDir, "override.yml")
	overrideYAML := `defaults:
  pull_policy: Always
  # timeout not specified, should keep base value
`
	require.NoError(t, os.WriteFile(overrideConfig, []byte(overrideYAML), 0644))

	// Merge override config
	require.NoError(t, MergeConfigFile(overrideConfig))

	// Verify selective override
	assert.Equal(t, "ghcr.io", Global.Registry.Default, "Registry should remain")
	assert.Equal(t, 30, Global.Registry.Timeout, "Registry timeout should remain")
	assert.Equal(t, "Always", Global.Defaults.PullPolicy, "Pull policy should be overridden")
	assert.Equal(t, 120, Global.Defaults.Timeout, "Defaults timeout should remain")
}

// TestMergeConfigFile_InvalidFile verifies error on invalid file
func TestMergeConfigFile_InvalidFile(t *testing.T) {
	// Setup base config
	tmpDir := t.TempDir()
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `extensions:
  - name: test
    image: test:v1
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	ResetConfigLoaded()
	Global = Config{}
	require.NoError(t, LoadConfig(baseConfig))

	// Try to merge non-existent file
	err := MergeConfigFile(filepath.Join(tmpDir, "nonexistent.yml"))
	assert.Error(t, err)
}

// TestMergeConfigFile_InvalidYAML verifies error on malformed YAML
func TestMergeConfigFile_InvalidYAML(t *testing.T) {
	// Setup base config
	tmpDir := t.TempDir()
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `extensions:
  - name: test
    image: test:v1
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	ResetConfigLoaded()
	Global = Config{}
	require.NoError(t, LoadConfig(baseConfig))

	// Create invalid YAML override
	overrideConfig := filepath.Join(tmpDir, "invalid.yml")
	invalidYAML := `extensions:
  - name: [invalid
      malformed: yaml
`
	require.NoError(t, os.WriteFile(overrideConfig, []byte(invalidYAML), 0644))

	// Try to merge invalid file
	err := MergeConfigFile(overrideConfig)
	assert.Error(t, err)
}

// TestMergeConfigFile_ValidationFailure verifies validation after merge
func TestMergeConfigFile_ValidationFailure(t *testing.T) {
	// Setup valid base config
	tmpDir := t.TempDir()
	baseConfig := filepath.Join(tmpDir, "base.yml")
	baseYAML := `extensions:
  - name: test-ext
    image: test/image:v1
`
	require.NoError(t, os.WriteFile(baseConfig, []byte(baseYAML), 0644))

	ResetConfigLoaded()
	Global = Config{}
	require.NoError(t, LoadConfig(baseConfig))

	// Create override that would make merged config invalid
	overrideConfig := filepath.Join(tmpDir, "bad-override.yml")
	badYAML := `extensions:
  - name: ""
    image: ""
`
	require.NoError(t, os.WriteFile(overrideConfig, []byte(badYAML), 0644))

	// Try to merge - should fail validation
	err := MergeConfigFile(overrideConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestLoadConfig_ValidConfig verifies loading valid config
func TestLoadConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test.yml")
	configYAML := `registry:
  default: ghcr.io
extensions:
  - name: test-ext
    image: test/image:v1.0.0
`
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0644))

	ResetConfigLoaded()
	Global = Config{}
	err := LoadConfig(configFile)

	require.NoError(t, err)
	assert.Equal(t, "ghcr.io", Global.Registry.Default)
	assert.Len(t, Global.Extensions, 1)
}

// TestLoadConfig_FileNotFound verifies error on missing file
func TestLoadConfig_FileNotFound(t *testing.T) {
	ResetConfigLoaded()
	Global = Config{}
	err := LoadConfig("/nonexistent/path/config.yml")
	assert.Error(t, err)
}

// TestLoadConfig_InvalidYAML verifies error on malformed YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "bad.yml")
	badYAML := `registry:
  default: [malformed
    invalid yaml
`
	require.NoError(t, os.WriteFile(configFile, []byte(badYAML), 0644))

	ResetConfigLoaded()
	Global = Config{}
	err := LoadConfig(configFile)
	assert.Error(t, err)
}
