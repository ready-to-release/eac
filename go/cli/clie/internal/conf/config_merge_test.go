//go:build L2
// +build L2

package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMergeEnvVars_NewVariables verifies new env vars are added
func TestMergeEnvVars_NewVariables(t *testing.T) {
	base := []EnvVar{
		{Name: "VAR1", Value: "value1"},
		{Name: "VAR2", Value: "value2"},
	}
	override := []EnvVar{
		{Name: "VAR3", Value: "value3"},
		{Name: "VAR4", Value: "value4"},
	}

	result := mergeEnvVars(base, override)

	assert.Len(t, result, 4, "Should have all 4 variables")
	// Check that all variables are present
	varMap := make(map[string]string)
	for _, v := range result {
		varMap[v.Name] = v.Value
	}
	assert.Equal(t, "value1", varMap["VAR1"])
	assert.Equal(t, "value2", varMap["VAR2"])
	assert.Equal(t, "value3", varMap["VAR3"])
	assert.Equal(t, "value4", varMap["VAR4"])
}

// TestMergeEnvVars_OverrideExisting verifies override replaces base values
func TestMergeEnvVars_OverrideExisting(t *testing.T) {
	base := []EnvVar{
		{Name: "VAR1", Value: "old_value"},
		{Name: "VAR2", Value: "keep_value"},
	}
	override := []EnvVar{
		{Name: "VAR1", Value: "new_value"},
	}

	result := mergeEnvVars(base, override)

	assert.Len(t, result, 2, "Should have 2 variables")
	varMap := make(map[string]string)
	for _, v := range result {
		varMap[v.Name] = v.Value
	}
	assert.Equal(t, "new_value", varMap["VAR1"], "Should override VAR1")
	assert.Equal(t, "keep_value", varMap["VAR2"], "Should keep VAR2")
}

// TestMergeEnvVars_EmptyBase verifies override with empty base
func TestMergeEnvVars_EmptyBase(t *testing.T) {
	base := []EnvVar{}
	override := []EnvVar{
		{Name: "VAR1", Value: "value1"},
	}

	result := mergeEnvVars(base, override)

	assert.Len(t, result, 1)
	assert.Equal(t, "VAR1", result[0].Name)
	assert.Equal(t, "value1", result[0].Value)
}

// TestMergeEnvVars_EmptyOverride verifies base remains unchanged with empty override
func TestMergeEnvVars_EmptyOverride(t *testing.T) {
	base := []EnvVar{
		{Name: "VAR1", Value: "value1"},
	}
	override := []EnvVar{}

	result := mergeEnvVars(base, override)

	assert.Len(t, result, 1)
	assert.Equal(t, "VAR1", result[0].Name)
	assert.Equal(t, "value1", result[0].Value)
}

// TestMergeSecretVars_NewSecrets verifies new secret vars are added
func TestMergeSecretVars_NewSecrets(t *testing.T) {
	base := []SecretVar{
		{Name: "SECRET1", Env: "ENV1"},
	}
	override := []SecretVar{
		{Name: "SECRET2", Env: "ENV2"},
	}

	result := mergeSecretVars(base, override)

	assert.Len(t, result, 2)
	secretMap := make(map[string]string)
	for _, s := range result {
		secretMap[s.Name] = s.Env
	}
	assert.Equal(t, "ENV1", secretMap["SECRET1"])
	assert.Equal(t, "ENV2", secretMap["SECRET2"])
}

// TestMergeSecretVars_OverrideExisting verifies override replaces base secrets
func TestMergeSecretVars_OverrideExisting(t *testing.T) {
	base := []SecretVar{
		{Name: "SECRET1", Env: "OLD_ENV"},
	}
	override := []SecretVar{
		{Name: "SECRET1", Env: "NEW_ENV"},
	}

	result := mergeSecretVars(base, override)

	assert.Len(t, result, 1)
	assert.Equal(t, "SECRET1", result[0].Name)
	assert.Equal(t, "NEW_ENV", result[0].Env)
}

// TestMergeExtension_ImageOverride verifies image field override
func TestMergeExtension_ImageOverride(t *testing.T) {
	base := &Extension{
		Name:  "test-ext",
		Image: "old/image:v1.0.0",
	}
	override := &Extension{
		Name:  "test-ext",
		Image: "new/image:v2.0.0",
	}

	mergeExtension(base, override)

	assert.Equal(t, "new/image:v2.0.0", base.Image, "Image should be overridden")
}

// TestMergeExtension_LoadLocalOverride verifies LoadLocal boolean override
func TestMergeExtension_LoadLocalOverride(t *testing.T) {
	base := &Extension{
		Name:      "test-ext",
		LoadLocal: false,
	}
	override := &Extension{
		Name:      "test-ext",
		LoadLocal: true,
	}

	mergeExtension(base, override)

	assert.True(t, base.LoadLocal, "LoadLocal should be set to true")
}

// TestMergeExtension_LoadLocalNotForced verifies LoadLocal false doesn't override true
func TestMergeExtension_LoadLocalNotForced(t *testing.T) {
	base := &Extension{
		Name:      "test-ext",
		LoadLocal: true,
	}
	override := &Extension{
		Name:      "test-ext",
		LoadLocal: false, // This should NOT override true to false
	}

	mergeExtension(base, override)

	assert.True(t, base.LoadLocal, "LoadLocal should remain true")
}

// TestMergeExtension_EnvVarsMerge verifies environment variables are merged
func TestMergeExtension_EnvVarsMerge(t *testing.T) {
	base := &Extension{
		Name: "test-ext",
		Env: []EnvVar{
			{Name: "VAR1", Value: "old_value"},
		},
	}
	override := &Extension{
		Name: "test-ext",
		Env: []EnvVar{
			{Name: "VAR1", Value: "new_value"},
			{Name: "VAR2", Value: "value2"},
		},
	}

	mergeExtension(base, override)

	assert.Len(t, base.Env, 2)
	envMap := make(map[string]string)
	for _, v := range base.Env {
		envMap[v.Name] = v.Value
	}
	assert.Equal(t, "new_value", envMap["VAR1"])
	assert.Equal(t, "value2", envMap["VAR2"])
}

// TestMergeExtension_ResourceLimitsOverride verifies resource limits override
func TestMergeExtension_ResourceLimitsOverride(t *testing.T) {
	base := &Extension{
		Name:        "test-ext",
		MemoryLimit: "512m",
		CPULimit:    "1.0",
	}
	override := &Extension{
		Name:        "test-ext",
		MemoryLimit: "1g",
		CPULimit:    "2.0",
	}

	mergeExtension(base, override)

	assert.Equal(t, "1g", base.MemoryLimit)
	assert.Equal(t, "2.0", base.CPULimit)
}

// TestMergeExtension_VolumesOverride verifies volumes are completely replaced
func TestMergeExtension_VolumesOverride(t *testing.T) {
	base := &Extension{
		Name: "test-ext",
		Volumes: []VolumeMount{
			{Host: "/old/path", Container: "/data"},
		},
	}
	override := &Extension{
		Name: "test-ext",
		Volumes: []VolumeMount{
			{Host: "/new/path", Container: "/data"},
		},
	}

	mergeExtension(base, override)

	assert.Len(t, base.Volumes, 1)
	assert.Equal(t, "/new/path", base.Volumes[0].Host)
}

// TestMergeExtension_PortsOverride verifies ports are completely replaced
func TestMergeExtension_PortsOverride(t *testing.T) {
	base := &Extension{
		Name: "test-ext",
		Ports: []PortMapping{
			{Host: 8080, Container: 80},
		},
	}
	override := &Extension{
		Name: "test-ext",
		Ports: []PortMapping{
			{Host: 9090, Container: 80},
		},
	}

	mergeExtension(base, override)

	assert.Len(t, base.Ports, 1)
	assert.Equal(t, 9090, base.Ports[0].Host)
}

// TestMergeExtension_NetworkModeOverride verifies network mode override
func TestMergeExtension_NetworkModeOverride(t *testing.T) {
	base := &Extension{
		Name:        "test-ext",
		NetworkMode: "bridge",
	}
	override := &Extension{
		Name:        "test-ext",
		NetworkMode: "host",
	}

	mergeExtension(base, override)

	assert.Equal(t, "host", base.NetworkMode)
}

// TestMergeExtension_CommandOverride verifies command override
func TestMergeExtension_CommandOverride(t *testing.T) {
	base := &Extension{
		Name:    "test-ext",
		Command: []string{"old", "cmd"},
	}
	override := &Extension{
		Name:    "test-ext",
		Command: []string{"new", "cmd"},
	}

	mergeExtension(base, override)

	assert.Equal(t, []string{"new", "cmd"}, base.Command)
}

// TestMergeExtension_PartialOverride verifies only specified fields are overridden
func TestMergeExtension_PartialOverride(t *testing.T) {
	base := &Extension{
		Name:        "test-ext",
		Image:       "old/image:v1",
		Description: "Original description",
		MemoryLimit: "512m",
	}
	override := &Extension{
		Name:  "test-ext",
		Image: "new/image:v2",
		// Description not set, should remain unchanged
		// MemoryLimit not set, should remain unchanged
	}

	mergeExtension(base, override)

	assert.Equal(t, "new/image:v2", base.Image, "Image should be overridden")
	assert.Equal(t, "Original description", base.Description, "Description should remain unchanged")
	assert.Equal(t, "512m", base.MemoryLimit, "MemoryLimit should remain unchanged")
}

// TestMergeConfigs_RegistryMerge verifies registry settings merge
func TestMergeConfigs_RegistryMerge(t *testing.T) {
	base := &Config{
		Registry: &Registry{
			Default:       "ghcr.io",
			Timeout:       30,
			RetryAttempts: 3,
		},
	}
	override := &Config{
		Registry: &Registry{
			Default: "docker.io",
			Timeout: 60, // Override timeout
			// RetryAttempts not set, should keep base value
		},
	}

	mergeConfigs(base, override)

	require.NotNil(t, base.Registry)
	assert.Equal(t, "docker.io", base.Registry.Default)
	assert.Equal(t, 60, base.Registry.Timeout)
	assert.Equal(t, 3, base.Registry.RetryAttempts, "Should keep base value")
}

// TestMergeConfigs_RegistryNilBase verifies registry merge with nil base
func TestMergeConfigs_RegistryNilBase(t *testing.T) {
	base := &Config{
		Registry: nil,
	}
	override := &Config{
		Registry: &Registry{
			Default: "ghcr.io",
			Timeout: 30,
		},
	}

	mergeConfigs(base, override)

	require.NotNil(t, base.Registry)
	assert.Equal(t, "ghcr.io", base.Registry.Default)
	assert.Equal(t, 30, base.Registry.Timeout)
}

// TestMergeConfigs_DefaultsMerge verifies defaults settings merge
func TestMergeConfigs_DefaultsMerge(t *testing.T) {
	base := &Config{
		Defaults: &Defaults{
			PullPolicy:  "IfNotPresent",
			Timeout:     120,
			MemoryLimit: "512m",
		},
	}
	override := &Config{
		Defaults: &Defaults{
			PullPolicy: "Always",
			CPULimit:   "2.0",
			// Timeout and MemoryLimit not set
		},
	}

	mergeConfigs(base, override)

	require.NotNil(t, base.Defaults)
	assert.Equal(t, "Always", base.Defaults.PullPolicy)
	assert.Equal(t, "2.0", base.Defaults.CPULimit)
	assert.Equal(t, 120, base.Defaults.Timeout, "Should keep base value")
	assert.Equal(t, "512m", base.Defaults.MemoryLimit, "Should keep base value")
}

// TestMergeConfigs_EnvironmentMerge verifies environment settings merge
func TestMergeConfigs_EnvironmentMerge(t *testing.T) {
	base := &Config{
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "GLOBAL1", Value: "value1"},
			},
			Secrets: []SecretVar{
				{Name: "SECRET1", Env: "ENV1"},
			},
		},
	}
	override := &Config{
		Environment: &Environment{
			Global: []EnvVar{
				{Name: "GLOBAL2", Value: "value2"},
			},
			Secrets: []SecretVar{
				{Name: "SECRET2", Env: "ENV2"},
			},
		},
	}

	mergeConfigs(base, override)

	require.NotNil(t, base.Environment)
	assert.Len(t, base.Environment.Global, 2)
	assert.Len(t, base.Environment.Secrets, 2)
}

// TestMergeConfigs_ExtensionsMerge verifies extensions merge
func TestMergeConfigs_ExtensionsMerge(t *testing.T) {
	base := &Config{
		Extensions: []Extension{
			{
				Name:  "ext1",
				Image: "old/ext1:v1",
			},
			{
				Name:  "ext2",
				Image: "ext2:v1",
			},
		},
	}
	override := &Config{
		Extensions: []Extension{
			{
				Name:  "ext1",
				Image: "new/ext1:v2", // Override ext1
			},
			{
				Name:  "ext3",
				Image: "ext3:v1", // Add new ext3
			},
		},
	}

	mergeConfigs(base, override)

	assert.Len(t, base.Extensions, 3, "Should have 3 extensions")

	// Find extensions by name
	extMap := make(map[string]string)
	for _, ext := range base.Extensions {
		extMap[ext.Name] = ext.Image
	}

	assert.Equal(t, "new/ext1:v2", extMap["ext1"], "ext1 should be overridden")
	assert.Equal(t, "ext2:v1", extMap["ext2"], "ext2 should remain unchanged")
	assert.Equal(t, "ext3:v1", extMap["ext3"], "ext3 should be added")
}

// TestMergeConfigs_CompleteOverride verifies full config merge
func TestMergeConfigs_CompleteOverride(t *testing.T) {
	base := &Config{
		Registry: &Registry{
			Default: "ghcr.io",
		},
		Defaults: &Defaults{
			PullPolicy: "IfNotPresent",
		},
		Environment: &Environment{
			Global: []EnvVar{{Name: "VAR1", Value: "val1"}},
		},
		Extensions: []Extension{
			{Name: "ext1", Image: "ext1:v1"},
		},
	}
	override := &Config{
		Registry: &Registry{
			Default: "docker.io",
		},
		Defaults: &Defaults{
			PullPolicy: "Always",
		},
		Environment: &Environment{
			Global: []EnvVar{{Name: "VAR2", Value: "val2"}},
		},
		Extensions: []Extension{
			{Name: "ext2", Image: "ext2:v1"},
		},
	}

	mergeConfigs(base, override)

	assert.Equal(t, "docker.io", base.Registry.Default)
	assert.Equal(t, "Always", base.Defaults.PullPolicy)
	assert.Len(t, base.Environment.Global, 2)
	assert.Len(t, base.Extensions, 2)
}
