//go:build L0
// +build L0

package environments

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectRuntime(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected RuntimeEnv
	}{
		{
			name:     "default is DevBox",
			envVars:  map[string]string{},
			expected: DevBox,
		},
		{
			name:     "CI env var set",
			envVars:  map[string]string{EnvCI: "true"},
			expected: CI,
		},
		{
			name:     "GITHUB_ACTIONS set",
			envVars:  map[string]string{EnvGitHubActions: "true"},
			expected: CI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear CI-related env vars
			ciVars := []string{EnvCI, EnvGitHubActions}
			saved := make(map[string]string)
			for _, k := range ciVars {
				if v, exists := os.LookupEnv(k); exists {
					saved[k] = v
				}
				os.Unsetenv(k)
			}

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Test
			result := DetectRuntime()
			assert.Equal(t, tt.expected, result)

			// Restore
			for _, k := range ciVars {
				os.Unsetenv(k)
			}
			for k, v := range saved {
				os.Setenv(k, v)
			}
		})
	}
}

func TestIsCI(t *testing.T) {
	// Save and clear
	ciVal, ciExists := os.LookupEnv(EnvCI)
	gaVal, gaExists := os.LookupEnv(EnvGitHubActions)
	os.Unsetenv(EnvCI)
	os.Unsetenv(EnvGitHubActions)

	// Test default (DevBox)
	assert.False(t, IsCI())
	assert.True(t, IsDevBox())

	// Test CI
	os.Setenv(EnvCI, "true")
	assert.True(t, IsCI())
	assert.False(t, IsDevBox())

	// Restore
	os.Unsetenv(EnvCI)
	os.Unsetenv(EnvGitHubActions)
	if ciExists {
		os.Setenv(EnvCI, ciVal)
	}
	if gaExists {
		os.Setenv(EnvGitHubActions, gaVal)
	}
}

func TestRuntimeEnv_String(t *testing.T) {
	assert.Equal(t, "devbox", DevBox.String())
	assert.Equal(t, "ci", CI.String())
}
