//go:build L0
// +build L0

package docker

import (
	"testing"

	"github.com/ready-to-release/eac/go/cli/clie/internal/cache"
	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseExtensionMetadata_ValidYAML verifies parsing valid YAML metadata
func TestParseExtensionMetadata_ValidYAML(t *testing.T) {
	yamlData := `name: eac
version: 0.0.2
description: Extension for EAC CLI
env:
  - name: EAC_DEBUG
    value: "false"
    required: false
  - name: EAC_LOG_LEVEL
    required: true
volumes:
  - name: cache
    target: /var/cache/eac
`

	// Act
	meta, err := parseExtensionMetadata(yamlData)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Equal(t, "eac", meta.Name)
	assert.Equal(t, "0.0.2", meta.Version)
	assert.Equal(t, "Extension for EAC CLI", meta.Description)
	assert.Len(t, meta.Env, 2)
	assert.Equal(t, "EAC_DEBUG", meta.Env[0].Name)
	assert.Equal(t, "false", meta.Env[0].Value)
	assert.False(t, meta.Env[0].Required)
	assert.Equal(t, "EAC_LOG_LEVEL", meta.Env[1].Name)
	assert.True(t, meta.Env[1].Required)
	assert.Len(t, meta.Volumes, 1)
	assert.Equal(t, "cache", meta.Volumes[0].Name)
	assert.Equal(t, "/var/cache/eac", meta.Volumes[0].Target)
}

// TestParseExtensionMetadata_MinimalYAML verifies parsing minimal YAML
func TestParseExtensionMetadata_MinimalYAML(t *testing.T) {
	yamlData := `name: minimal
version: 1.0.0
`

	// Act
	meta, err := parseExtensionMetadata(yamlData)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Equal(t, "minimal", meta.Name)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Empty(t, meta.Description)
	assert.Empty(t, meta.Env)
	assert.Empty(t, meta.Volumes)
}

// TestParseExtensionMetadata_InvalidYAML verifies error handling for invalid YAML
func TestParseExtensionMetadata_InvalidYAML(t *testing.T) {
	yamlData := `name: broken
version: [invalid
  - malformed
`

	// Act
	meta, err := parseExtensionMetadata(yamlData)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, meta)
	assert.Contains(t, err.Error(), "failed to unmarshal metadata YAML")
}

// TestParseExtensionMetadata_EmptyYAML verifies handling of empty input
func TestParseExtensionMetadata_EmptyYAML(t *testing.T) {
	yamlData := ""

	// Act
	meta, err := parseExtensionMetadata(yamlData)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Empty(t, meta.Name)
	assert.Empty(t, meta.Version)
}

// TestMergeMetadataEnv_AddsNewEnvVars verifies new env vars from metadata are added
func TestMergeMetadataEnv_AddsNewEnvVars(t *testing.T) {
	// Arrange
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env: []conf.EnvVar{
			{Name: "EXISTING_VAR", Value: "existing_value"},
		},
	}

	meta := &cache.ExtensionMeta{
		Env: []cache.MetaEnvVar{
			{Name: "NEW_VAR_1", Value: "value1", Required: false},
			{Name: "NEW_VAR_2", Value: "value2", Required: true},
		},
	}

	// Act
	MergeMetadataEnv(ext, meta)

	// Assert
	assert.Len(t, ext.Env, 3, "Should have 1 existing + 2 new env vars")
	assert.Equal(t, "EXISTING_VAR", ext.Env[0].Name)
	assert.Equal(t, "NEW_VAR_1", ext.Env[1].Name)
	assert.Equal(t, "value1", ext.Env[1].Value)
	assert.False(t, ext.Env[1].Required)
	assert.Equal(t, "NEW_VAR_2", ext.Env[2].Name)
	assert.Equal(t, "value2", ext.Env[2].Value)
	assert.True(t, ext.Env[2].Required)
}

// TestMergeMetadataEnv_SkipsDuplicates verifies config env vars take precedence
func TestMergeMetadataEnv_SkipsDuplicates(t *testing.T) {
	// Arrange
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env: []conf.EnvVar{
			{Name: "VAR_1", Value: "config_value"},
			{Name: "VAR_2", Value: "another_config_value"},
		},
	}

	meta := &cache.ExtensionMeta{
		Env: []cache.MetaEnvVar{
			{Name: "VAR_1", Value: "metadata_value"}, // Duplicate, should be skipped
			{Name: "VAR_3", Value: "new_value"},      // New, should be added
		},
	}

	// Act
	MergeMetadataEnv(ext, meta)

	// Assert
	assert.Len(t, ext.Env, 3, "Should have 2 original + 1 new env var")
	// First two should be unchanged
	assert.Equal(t, "VAR_1", ext.Env[0].Name)
	assert.Equal(t, "config_value", ext.Env[0].Value, "Config value should be preserved")
	assert.Equal(t, "VAR_2", ext.Env[1].Name)
	// Third should be new from metadata
	assert.Equal(t, "VAR_3", ext.Env[2].Name)
	assert.Equal(t, "new_value", ext.Env[2].Value)
}

// TestMergeMetadataEnv_NilMetadata verifies no-op when metadata is nil
func TestMergeMetadataEnv_NilMetadata(t *testing.T) {
	// Arrange
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env: []conf.EnvVar{
			{Name: "EXISTING", Value: "value"},
		},
	}
	originalEnvLen := len(ext.Env)

	// Act
	MergeMetadataEnv(ext, nil)

	// Assert
	assert.Len(t, ext.Env, originalEnvLen, "Should not modify env vars when metadata is nil")
	assert.Equal(t, "EXISTING", ext.Env[0].Name)
}

// TestMergeMetadataEnv_EmptyMetadataEnv verifies no-op when metadata has no env vars
func TestMergeMetadataEnv_EmptyMetadataEnv(t *testing.T) {
	// Arrange
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env: []conf.EnvVar{
			{Name: "EXISTING", Value: "value"},
		},
	}

	meta := &cache.ExtensionMeta{
		Env: []cache.MetaEnvVar{}, // Empty env vars
	}

	originalEnvLen := len(ext.Env)

	// Act
	MergeMetadataEnv(ext, meta)

	// Assert
	assert.Len(t, ext.Env, originalEnvLen, "Should not modify env vars when metadata env is empty")
	assert.Equal(t, "EXISTING", ext.Env[0].Name)
}

// TestMergeMetadataEnv_EmptyConfig verifies metadata env vars added to empty config
func TestMergeMetadataEnv_EmptyConfig(t *testing.T) {
	// Arrange
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env:   []conf.EnvVar{}, // Empty config env vars
	}

	meta := &cache.ExtensionMeta{
		Env: []cache.MetaEnvVar{
			{Name: "META_VAR", Value: "meta_value", Required: true},
		},
	}

	// Act
	MergeMetadataEnv(ext, meta)

	// Assert
	assert.Len(t, ext.Env, 1, "Should add metadata env var to empty config")
	assert.Equal(t, "META_VAR", ext.Env[0].Name)
	assert.Equal(t, "meta_value", ext.Env[0].Value)
	assert.True(t, ext.Env[0].Required)
}
