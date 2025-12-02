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
			envVars:  map[string]string{"CI": "true"},
			expected: CI,
		},
		{
			name:     "GITHUB_ACTIONS set",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: CI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear CI-related env vars
			ciVars := []string{"CI", "GITHUB_ACTIONS"}
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
	ciVal, ciExists := os.LookupEnv("CI")
	gaVal, gaExists := os.LookupEnv("GITHUB_ACTIONS")
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")

	// Test default (DevBox)
	assert.False(t, IsCI())
	assert.True(t, IsDevBox())

	// Test CI
	os.Setenv("CI", "true")
	assert.True(t, IsCI())
	assert.False(t, IsDevBox())

	// Restore
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")
	if ciExists {
		os.Setenv("CI", ciVal)
	}
	if gaExists {
		os.Setenv("GITHUB_ACTIONS", gaVal)
	}
}

func TestRuntimeEnv_String(t *testing.T) {
	assert.Equal(t, "devbox", DevBox.String())
	assert.Equal(t, "ci", CI.String())
}
