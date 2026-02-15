package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeyResolution_EnvVarPriority tests that environment variables have highest priority.
func TestAPIKeyResolution_EnvVarPriority(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		envVarName     string
		envVarValue    string
		expectedSource APIKeySource
	}{
		{
			name:           "claude-api with ANTHROPIC_API_KEY",
			provider:       "claude-api",
			envVarName:     "ANTHROPIC_API_KEY",
			envVarValue:    "sk-ant-test-key",
			expectedSource: APIKeySourceEnvVar,
		},
		{
			name:           "openai with OPENAI_API_KEY",
			provider:       "openai",
			envVarName:     "OPENAI_API_KEY",
			envVarValue:    "sk-test-key",
			expectedSource: APIKeySourceEnvVar,
		},
		{
			name:           "gemini with GOOGLE_API_KEY",
			provider:       "gemini",
			envVarName:     "GOOGLE_API_KEY",
			envVarValue:    "test-key",
			expectedSource: APIKeySourceEnvVar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv(tt.envVarName, tt.envVarValue)
			require.NoError(t, err)
			defer os.Unsetenv(tt.envVarName)

			// Create temporary workspace
			tmpDir := t.TempDir()

			// Resolve API key
			resolution, err := resolveAPIKey(tmpDir, tt.provider)

			// Assertions
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSource, resolution.Source)
			assert.Equal(t, tt.envVarValue, resolution.Key)
			assert.Equal(t, tt.envVarName, resolution.EnvVarName)
			assert.Empty(t, resolution.ConfigPath)
		})
	}
}

// TestAPIKeyResolution_PersonalConfigFallback tests that personal config is used when env var missing.
func TestAPIKeyResolution_PersonalConfigFallback(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create personal config file
	personalConfigPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	personalConfig := `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  api_key: sk-ant-personal-key
`
	err = os.WriteFile(personalConfigPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	// Ensure environment variable is NOT set
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Resolve API key
	resolution, err := resolveAPIKey(tmpDir, "claude-api")

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, APIKeySourcePersonalConfig, resolution.Source)
	assert.Equal(t, "sk-ant-personal-key", resolution.Key)
	assert.Equal(t, personalConfigPath, resolution.ConfigPath)
}

// TestAPIKeyResolution_ErrorWhenMissing tests agent-friendly error when no API key found.
func TestAPIKeyResolution_ErrorWhenMissing(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		envVarName   string
		expectedMsg  string
	}{
		{
			name:         "claude-api missing key",
			provider:     "claude-api",
			envVarName:   "ANTHROPIC_API_KEY",
			expectedMsg:  "API key not found for provider 'claude-api'",
		},
		{
			name:         "openai missing key",
			provider:     "openai",
			envVarName:   "OPENAI_API_KEY",
			expectedMsg:  "API key not found for provider 'openai'",
		},
		{
			name:         "gemini missing key",
			provider:     "gemini",
			envVarName:   "GOOGLE_API_KEY",
			expectedMsg:  "API key not found for provider 'gemini'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary workspace
			tmpDir := t.TempDir()

			// Ensure environment variable is NOT set
			os.Unsetenv(tt.envVarName)

			// Resolve API key
			resolution, err := resolveAPIKey(tmpDir, tt.provider)

			// Assertions
			require.Error(t, err)
			assert.Nil(t, resolution)
			assert.Contains(t, err.Error(), tt.expectedMsg)
			assert.Contains(t, err.Error(), tt.envVarName)
			assert.Contains(t, err.Error(), "export "+tt.envVarName)
			assert.Contains(t, err.Error(), "ai-provider.personal.yml")
		})
	}
}

// TestAPIKeyResolution_ErrorMessageFormat tests that error messages are clear and actionable.
func TestAPIKeyResolution_ErrorMessageFormat(t *testing.T) {
	tmpDir := t.TempDir()
	os.Unsetenv("ANTHROPIC_API_KEY")

	_, err := resolveAPIKey(tmpDir, "claude-api")

	require.Error(t, err)
	errMsg := err.Error()

	// Verify error message contains required elements
	assert.Contains(t, errMsg, "API key not found")
	assert.Contains(t, errMsg, "claude-api")
	assert.Contains(t, errMsg, "ANTHROPIC_API_KEY")
	assert.Contains(t, errMsg, "export ANTHROPIC_API_KEY")
	assert.Contains(t, errMsg, "ai-provider.personal.yml") // Platform-agnostic (no path separator)
	assert.Contains(t, errMsg, "ai:")
	assert.Contains(t, errMsg, "provider: claude-api")
	assert.Contains(t, errMsg, "api_key: your-key-here")
}

// TestAPIKeyResolution_EnvVarOverridesPersonalConfig tests priority when both exist.
func TestAPIKeyResolution_EnvVarOverridesPersonalConfig(t *testing.T) {
	// Create temporary workspace
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create personal config file with different key
	personalConfigPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	personalConfig := `ai:
  provider: claude-api
  api_key: sk-ant-personal-key
`
	err = os.WriteFile(personalConfigPath, []byte(personalConfig), 0o600)
	require.NoError(t, err)

	// Set environment variable (should override personal config)
	err = os.Setenv("ANTHROPIC_API_KEY", "sk-ant-env-key")
	require.NoError(t, err)
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Resolve API key
	resolution, err := resolveAPIKey(tmpDir, "claude-api")

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, APIKeySourceEnvVar, resolution.Source)
	assert.Equal(t, "sk-ant-env-key", resolution.Key)
	assert.NotEqual(t, "sk-ant-personal-key", resolution.Key)
}

// TestGetEnvVarNameForProvider tests environment variable name mapping.
func TestGetEnvVarNameForProvider(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"claude-api", "ANTHROPIC_API_KEY"},
		{"openai", "OPENAI_API_KEY"},
		{"gemini", "GOOGLE_API_KEY"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := getEnvVarNameForProvider(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAPIKeyResolution_NoInteractivePrompt ensures no prompts are triggered.
func TestAPIKeyResolution_NoInteractivePrompt(t *testing.T) {
	// This test documents the design decision:
	// The command MUST NOT prompt for API keys interactively.
	// It must fail with a clear error message instead.
	// This is critical for agent/automation usage.

	tmpDir := t.TempDir()
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Attempting to resolve should return error immediately
	resolution, err := resolveAPIKey(tmpDir, "claude-api")

	// Must fail with error (not prompt)
	require.Error(t, err)
	assert.Nil(t, resolution)

	// Error must be informative
	assert.Contains(t, err.Error(), "API key not found")
}

// TestAPIKeyResolution_PersonalConfigInvalidYAML tests error handling for malformed config.
func TestAPIKeyResolution_PersonalConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create invalid YAML file
	personalConfigPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	invalidYAML := `ai:
  provider: claude-api
  api_key: sk-ant-key
  invalid yaml here!!!
`
	err = os.WriteFile(personalConfigPath, []byte(invalidYAML), 0o600)
	require.NoError(t, err)

	// Unset env var to force reading config
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Resolve should still fail gracefully (fallback to error)
	_, err = resolveAPIKey(tmpDir, "claude-api")

	// Should return error about missing key (not panic)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key not found")
}

// TestAPIKeyResolution_PersonalConfigMissingAPIKey tests when config exists but has no key.
func TestAPIKeyResolution_PersonalConfigMissingAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	err := os.MkdirAll(eacDir, 0o755)
	require.NoError(t, err)

	// Create config without api_key field
	personalConfigPath := filepath.Join(eacDir, "ai-provider.personal.yml")
	configWithoutKey := `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
`
	err = os.WriteFile(personalConfigPath, []byte(configWithoutKey), 0o600)
	require.NoError(t, err)

	// Unset env var
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Resolve should fail with clear error
	_, err = resolveAPIKey(tmpDir, "claude-api")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key not found")
}

// TestAPIKeyResolution_EmptyEnvVar tests that empty env var is treated as missing.
func TestAPIKeyResolution_EmptyEnvVar(t *testing.T) {
	tmpDir := t.TempDir()

	// Set empty environment variable
	err := os.Setenv("ANTHROPIC_API_KEY", "")
	require.NoError(t, err)
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Resolve should fail (empty env var = no key)
	_, err = resolveAPIKey(tmpDir, "claude-api")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key not found")
}
