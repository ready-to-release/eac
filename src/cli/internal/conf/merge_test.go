//go:build L1
// +build L1

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMergeExtensionPartialOverride tests merging partial extension configuration
func TestMergeExtensionPartialOverride(t *testing.T) {
	// Create base extension
	base := &Extension{
		Name:            "pwsh",
		Image:           "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0",
		ImagePullPolicy: "IfNotPresent",
		LoadLocal:       false,
	}

	// Create override with only LoadLocal field
	override := &Extension{
		Name:      "pwsh",
		LoadLocal: true,
		// Image and ImagePullPolicy are empty - should not override
	}

	// Merge override into base
	mergeExtension(base, override)

	// Verify fields
	assert.Equal(t, "pwsh", base.Name, "Name should remain unchanged")
	assert.Equal(t, "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0", base.Image, "Image should remain unchanged")
	assert.Equal(t, "IfNotPresent", base.ImagePullPolicy, "ImagePullPolicy should remain unchanged")
	assert.True(t, base.LoadLocal, "LoadLocal should be updated from override")
}

// TestMergeExtensionFullOverride tests merging with all fields specified
func TestMergeExtensionFullOverride(t *testing.T) {
	// Create base extension
	base := &Extension{
		Name:            "python",
		Image:           "python:3.9",
		Description:     "Python runtime",
		ImagePullPolicy: "Always",
		LoadLocal:       false,
		MemoryLimit:     "1g",
		CPULimit:        "1.0",
	}

	// Create override with all fields
	override := &Extension{
		Name:            "python",
		Image:           "python:3.11",
		Description:     "Updated Python runtime",
		ImagePullPolicy: "IfNotPresent",
		LoadLocal:       true,
		MemoryLimit:     "2g",
		CPULimit:        "2.0",
	}

	// Merge override into base
	mergeExtension(base, override)

	// Verify all fields are overridden
	assert.Equal(t, "python", base.Name)
	assert.Equal(t, "python:3.11", base.Image, "Image should be updated")
	assert.Equal(t, "Updated Python runtime", base.Description, "Description should be updated")
	assert.Equal(t, "IfNotPresent", base.ImagePullPolicy, "ImagePullPolicy should be updated")
	assert.True(t, base.LoadLocal, "LoadLocal should be updated")
	assert.Equal(t, "2g", base.MemoryLimit, "MemoryLimit should be updated")
	assert.Equal(t, "2.0", base.CPULimit, "CPULimit should be updated")
}

// TestMergeExtensionEnvironmentVariables tests merging environment variables
func TestMergeExtensionEnvironmentVariables(t *testing.T) {
	// Create base extension with env vars
	base := &Extension{
		Name:  "test",
		Image: "test:latest",
		Env: []EnvVar{
			{Name: "VAR1", Value: "value1"},
			{Name: "VAR2", Value: "value2"},
		},
	}

	// Create override that updates VAR1 and adds VAR3
	override := &Extension{
		Name: "test",
		Env: []EnvVar{
			{Name: "VAR1", Value: "updated1"},
			{Name: "VAR3", Value: "value3"},
		},
	}

	// Merge override into base
	mergeExtension(base, override)

	// Verify environment variables are merged correctly
	assert.Len(t, base.Env, 3, "Should have 3 environment variables")

	// Create a map for easier verification
	envMap := make(map[string]string)
	for _, env := range base.Env {
		envMap[env.Name] = env.Value
	}

	assert.Equal(t, "updated1", envMap["VAR1"], "VAR1 should be updated")
	assert.Equal(t, "value2", envMap["VAR2"], "VAR2 should remain")
	assert.Equal(t, "value3", envMap["VAR3"], "VAR3 should be added")
}

// TestMergeExtensionEmptyOverride tests that empty override doesn't change base
func TestMergeExtensionEmptyOverride(t *testing.T) {
	// Create base extension
	base := &Extension{
		Name:            "node",
		Image:           "node:16",
		Description:     "Node.js runtime",
		ImagePullPolicy: "Always",
		LoadLocal:       false,
	}

	// Create a copy for comparison
	originalBase := *base

	// Create empty override (only name specified)
	override := &Extension{
		Name: "node",
	}

	// Merge override into base
	mergeExtension(base, override)

	// Verify nothing changed except LoadLocal might be false
	assert.Equal(t, originalBase.Name, base.Name)
	assert.Equal(t, originalBase.Image, base.Image)
	assert.Equal(t, originalBase.Description, base.Description)
	assert.Equal(t, originalBase.ImagePullPolicy, base.ImagePullPolicy)
}

// TestMergeConfigsExtensions tests merging configurations with extensions
func TestMergeConfigsExtensions(t *testing.T) {
	// Create base config
	base := &Config{
		Extensions: []Extension{
			{
				Name:            "pwsh",
				Image:           "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0",
				ImagePullPolicy: "IfNotPresent",
			},
			{
				Name:  "python",
				Image: "python:3.9",
			},
		},
	}

	// Create override config
	override := &Config{
		Extensions: []Extension{
			{
				Name:      "pwsh",
				LoadLocal: true, // Partial override
			},
			{
				Name:  "node", // New extension
				Image: "node:16",
			},
		},
	}

	// Merge configs
	mergeConfigs(base, override)

	// Verify extensions
	assert.Len(t, base.Extensions, 3, "Should have 3 extensions (2 original + 1 new)")

	// Find extensions by name for verification
	extMap := make(map[string]*Extension)
	for i := range base.Extensions {
		extMap[base.Extensions[i].Name] = &base.Extensions[i]
	}

	// Verify pwsh was merged
	pwsh := extMap["pwsh"]
	assert.NotNil(t, pwsh)
	assert.Equal(t, "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0", pwsh.Image, "Image should remain from base")
	assert.True(t, pwsh.LoadLocal, "LoadLocal should be updated from override")

	// Verify python remained unchanged
	python := extMap["python"]
	assert.NotNil(t, python)
	assert.Equal(t, "python:3.9", python.Image)

	// Verify node was added
	node := extMap["node"]
	assert.NotNil(t, node)
	assert.Equal(t, "node:16", node.Image)
}

// TestMergeConfigFile tests the full MergeConfigFile function
func TestMergeConfigFile(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	// Create base configuration file
	baseConfig := `version: "1.0"
extensions:
  - name: "pwsh"
    image: "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0"
    image_pull_policy: "IfNotPresent"
  - name: "python"
    image: "python:3.9"
    description: "Python runtime"
`
	baseConfigPath := filepath.Join(tempDir, "r2r-cli.yml")
	err := os.WriteFile(baseConfigPath, []byte(baseConfig), 0644)
	require.NoError(t, err)

	// Load base configuration
	Global = Config{} // Reset global config
	err = LoadConfig(baseConfigPath)
	require.NoError(t, err)

	// Verify base config loaded correctly
	assert.Len(t, Global.Extensions, 2)
	assert.Equal(t, "pwsh", Global.Extensions[0].Name)
	assert.Equal(t, "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0", Global.Extensions[0].Image)
	assert.False(t, Global.Extensions[0].LoadLocal)

	// Create override configuration file
	overrideConfig := `version: "1.0"
extensions:
  - name: "pwsh"
    load_local: true
  - name: "node"
    image: "node:16"
`
	overrideConfigPath := filepath.Join(tempDir, "r2r-cli.local.yml")
	err = os.WriteFile(overrideConfigPath, []byte(overrideConfig), 0644)
	require.NoError(t, err)

	// Merge override configuration
	err = MergeConfigFile(overrideConfigPath)
	require.NoError(t, err)

	// Verify merged configuration
	assert.Len(t, Global.Extensions, 3, "Should have 3 extensions after merge")

	// Find extensions by name
	extMap := make(map[string]*Extension)
	for i := range Global.Extensions {
		extMap[Global.Extensions[i].Name] = &Global.Extensions[i]
	}

	// Verify pwsh was merged correctly
	pwsh := extMap["pwsh"]
	assert.NotNil(t, pwsh)
	assert.Equal(t, "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0", pwsh.Image, "Image should remain from base")
	assert.True(t, pwsh.LoadLocal, "LoadLocal should be updated from override")

	// Verify python remained unchanged
	python := extMap["python"]
	assert.NotNil(t, python)
	assert.Equal(t, "python:3.9", python.Image)
	assert.Equal(t, "Python runtime", python.Description)

	// Verify node was added
	node := extMap["node"]
	assert.NotNil(t, node)
	assert.Equal(t, "node:16", node.Image)
}

// TestMergeConfigsDefaults tests merging default settings
func TestMergeConfigsDefaults(t *testing.T) {
	// Create base config with defaults
	base := &Config{
		Defaults: &Defaults{
			PullPolicy:  "Always",
			Timeout:     60,
			MemoryLimit: "1g",
			CPULimit:    "1.0",
		},
	}

	// Create override config with partial defaults
	override := &Config{
		Defaults: &Defaults{
			PullPolicy:  "IfNotPresent",
			MemoryLimit: "2g",
			// Timeout and CPULimit not specified
		},
	}

	// Merge configs
	mergeConfigs(base, override)

	// Verify defaults were merged correctly
	assert.NotNil(t, base.Defaults)
	assert.Equal(t, "IfNotPresent", base.Defaults.PullPolicy, "PullPolicy should be updated")
	assert.Equal(t, 60, base.Defaults.Timeout, "Timeout should remain from base")
	assert.Equal(t, "2g", base.Defaults.MemoryLimit, "MemoryLimit should be updated")
	assert.Equal(t, "1.0", base.Defaults.CPULimit, "CPULimit should remain from base")
}

// TestMergeConfigsEnvironment tests merging environment settings
func TestMergeConfigsEnvironment(t *testing.T) {
	// Create base config with environment
	base := &Config{
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "LOG_LEVEL", Value: "info"},
				{Name: "TIMEOUT", Value: "30"},
			},
		},
	}

	// Create override config
	override := &Config{
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "LOG_LEVEL", Value: "debug"}, // Override existing
				{Name: "DEBUG", Value: "true"},      // Add new
			},
		},
	}

	// Merge configs
	mergeConfigs(base, override)

	// Verify environment was merged correctly
	assert.NotNil(t, base.Environment)
	assert.Len(t, base.Environment.Global, 3, "Should have 3 global env vars")

	// Create map for verification
	envMap := make(map[string]string)
	for _, env := range base.Environment.Global {
		envMap[env.Name] = env.Value
	}

	assert.Equal(t, "debug", envMap["LOG_LEVEL"], "LOG_LEVEL should be updated")
	assert.Equal(t, "30", envMap["TIMEOUT"], "TIMEOUT should remain")
	assert.Equal(t, "true", envMap["DEBUG"], "DEBUG should be added")
}

// TestPartialOverrideValidation tests that partial overrides are not validated independently
func TestPartialOverrideValidation(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	// Create base configuration file
	baseConfig := `version: "1.0"
extensions:
  - name: "pwsh"
    image: "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0"
`
	baseConfigPath := filepath.Join(tempDir, "r2r-cli.yml")
	err := os.WriteFile(baseConfigPath, []byte(baseConfig), 0644)
	require.NoError(t, err)

	// Load base configuration
	Global = Config{} // Reset global config
	err = LoadConfig(baseConfigPath)
	require.NoError(t, err)

	// Create partial override configuration (missing required 'image' field)
	// This should be valid as a partial override
	overrideConfig := `version: "1.0"
extensions:
  - name: "pwsh"
    load_local: true
`
	overrideConfigPath := filepath.Join(tempDir, "r2r-cli.partial.yml")
	err = os.WriteFile(overrideConfigPath, []byte(overrideConfig), 0644)
	require.NoError(t, err)

	// Merge should succeed even though override alone would fail validation
	err = MergeConfigFile(overrideConfigPath)
	assert.NoError(t, err, "Partial override should merge successfully")

	// Verify the merge worked
	assert.Len(t, Global.Extensions, 1)
	assert.Equal(t, "pwsh", Global.Extensions[0].Name)
	assert.Equal(t, "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0", Global.Extensions[0].Image)
	assert.True(t, Global.Extensions[0].LoadLocal)
}

// TestMergeExtensionBooleanFields tests boolean field merging behavior
func TestMergeExtensionBooleanFields(t *testing.T) {
	tests := []struct {
		name              string
		baseLoadLocal     bool
		overrideLoadLocal bool
		expectedLoadLocal bool
		description       string
	}{
		{
			name:              "override_true_overwrites_false",
			baseLoadLocal:     false,
			overrideLoadLocal: true,
			expectedLoadLocal: true,
			description:       "Override with true should update base false to true",
		},
		{
			name:              "override_false_does_not_overwrite_true",
			baseLoadLocal:     true,
			overrideLoadLocal: false,
			expectedLoadLocal: true,
			description:       "Override with false should not change base true",
		},
		{
			name:              "override_false_does_not_overwrite_false",
			baseLoadLocal:     false,
			overrideLoadLocal: false,
			expectedLoadLocal: false,
			description:       "Override with false should not change base false",
		},
		{
			name:              "override_true_preserves_true",
			baseLoadLocal:     true,
			overrideLoadLocal: true,
			expectedLoadLocal: true,
			description:       "Override with true should preserve base true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &Extension{
				Name:      "test",
				Image:     "test:latest",
				LoadLocal: tt.baseLoadLocal,
			}

			override := &Extension{
				Name:      "test",
				LoadLocal: tt.overrideLoadLocal,
			}

			mergeExtension(base, override)

			assert.Equal(t, tt.expectedLoadLocal, base.LoadLocal, tt.description)
		})
	}
}

// TestMergeExtensionArrayFields tests array field merging behavior
func TestMergeExtensionArrayFields(t *testing.T) {
	// Create base extension with volumes and ports
	base := &Extension{
		Name:  "test",
		Image: "test:latest",
		Volumes: []VolumeMount{
			{Host: "/host/path1", Container: "/container/path1"},
			{Host: "/host/path2", Container: "/container/path2"},
		},
		Ports: []PortMapping{
			{Host: 8080, Container: 80},
		},
	}

	// Override with new volumes and ports (complete replacement)
	override := &Extension{
		Name: "test",
		Volumes: []VolumeMount{
			{Host: "/new/path", Container: "/new/container"},
		},
		Ports: []PortMapping{
			{Host: 9090, Container: 90},
			{Host: 9091, Container: 91},
		},
	}

	mergeExtension(base, override)

	// Verify arrays are completely replaced
	assert.Len(t, base.Volumes, 1, "Volumes should be replaced")
	assert.Equal(t, "/new/path", base.Volumes[0].Host)

	assert.Len(t, base.Ports, 2, "Ports should be replaced")
	assert.Equal(t, 9090, base.Ports[0].Host)
	assert.Equal(t, 9091, base.Ports[1].Host)
}

// TestMergeExtensionCommand tests command and entrypoint merging
func TestMergeExtensionCommand(t *testing.T) {
	// Create base extension with command and entrypoint
	base := &Extension{
		Name:       "test",
		Image:      "test:latest",
		Command:    []string{"echo", "hello"},
		Entrypoint: []string{"/bin/sh"},
	}

	// Override with new command, no entrypoint
	override := &Extension{
		Name:    "test",
		Command: []string{"echo", "world"},
		// Entrypoint not specified - should remain
	}

	mergeExtension(base, override)

	// Verify command is replaced, entrypoint remains
	assert.Equal(t, []string{"echo", "world"}, base.Command, "Command should be replaced")
	assert.Equal(t, []string{"/bin/sh"}, base.Entrypoint, "Entrypoint should remain")
}
