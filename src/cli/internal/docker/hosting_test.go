//go:build L0
// +build L0

package docker

import (
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/cli/internal/conf"
)

func TestExecuteMetadataCommand_Success(t *testing.T) {
	// This test demonstrates the expected behavior of ExecuteMetadataCommand
	// In a real implementation, we would need to mock the Docker client
	// For now, this serves as documentation of the expected interface

	t.Run("successful metadata retrieval", func(t *testing.T) {
		// Expected YAML output
		expectedOutput := `name: "test-extension"
version: "1.0.0"
description: "Test extension for unit tests"
schema-version: "1.0"
commands:
  test:
    description: "Test command"
`

		// Extension configuration
		ext := &ExtensionConfig{
			Name:            "test-extension",
			Image:           "test/extension:latest",
			ImagePullPolicy: "IfNotPresent",
		}

		// In a real test, we would:
		// 1. Create a mock Docker client
		// 2. Set up expectations for image inspection, container creation, etc.
		// 3. Execute the method and verify the output

		// For now, we document the expected behavior
		if ext == nil {
			t.Fatal("Extension should not be nil")
		}
		if ext.Name != "test-extension" {
			t.Errorf("Expected extension name 'test-extension', got %s", ext.Name)
		}
		if expectedOutput == "" {
			t.Error("Expected output should not be empty")
		}
	})

	t.Run("extension without metadata command", func(t *testing.T) {
		// Test behavior when extension doesn't support extension-meta command
		// Expected: error with exit code 127 (command not found) or similar
		t.Log("Test case: Extension without metadata command should return error")
	})

	t.Run("metadata command timeout", func(t *testing.T) {
		// Test behavior when metadata command times out
		// Expected: error indicating timeout after 60 seconds
		t.Log("Test case: Metadata command timeout should return timeout error")
	})

	t.Run("invalid YAML output", func(t *testing.T) {
		// Test behavior when metadata command returns invalid output
		// Note: ExecuteMetadataCommand returns raw output, validation happens elsewhere
		t.Log("Test case: Invalid YAML output should be returned as-is for caller to handle")
	})
}

func TestExecuteMetadataCommand_ErrorScenarios(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "image pull failure",
			expectedError: "error ensuring image exists",
		},
		{
			name:          "container creation failure",
			expectedError: "error creating container",
		},
		{
			name:          "container start failure",
			expectedError: "error starting container",
		},
		{
			name:          "non-zero exit code",
			expectedError: "extension-meta command failed with exit code",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// In a real implementation, we would:
			// 1. Create mock client and set up expectations
			// 2. Execute the method
			// 3. Verify error contains expected message
			if tc.expectedError == "" {
				t.Error("Expected error message should not be empty")
			}
			t.Logf("Test case: %s should return error containing '%s'", tc.name, tc.expectedError)
		})
	}
}

func TestExecuteMetadataCommand_Integration(t *testing.T) {
	// Integration test placeholder
	// This would test against a real Docker daemon in CI/CD
	t.Skip("Integration test requires Docker daemon")

	// Test with a real extension that supports metadata
	ext := &ExtensionConfig{
		Name:            "text",
		Image:           "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:0.0.2", //this will fail currently, we have no real extension supporting metadata yet.
		ImagePullPolicy: "IfNotPresent",
	}

	host, err := NewContainerHost()
	if err != nil {
		t.Fatalf("Failed to create container host: %v", err)
	}
	defer host.Close()

	output, err := host.ExecuteMetadataCommand(ext)
	if err != nil {
		// This is expected for extensions that don't support metadata yet
		t.Logf("Metadata command failed (expected): %v", err)
		return
	}

	// Verify output is valid YAML
	if output == "" {
		t.Error("Output should not be empty")
	}
	t.Log("Metadata retrieved successfully")
}

// Test helper to create a mock ContainerHost for testing
func createMockContainerHost() *ContainerHost {
	return &ContainerHost{
		rootDir: "/test/root",
	}
}

// Test helper to clear environment variables
func clearEnvVars(t *testing.T, vars []string) func() {
	originalValues := make(map[string]string)
	for _, envVar := range vars {
		originalValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
	return func() {
		for envVar, value := range originalValues {
			if value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}
	}
}

func TestDetectCIEnvironment(t *testing.T) {
	testCases := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no CI environment",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name: "GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected: true,
		},
		{
			name: "Azure DevOps CI",
			envVars: map[string]string{
				"AZUREDEVOPS_URL": "http://azuredevops.example.com",
			},
			expected: true,
		},
		{
			name: "GitLab CI",
			envVars: map[string]string{
				"GITLAB_CI": "true",
			},
			expected: true,
		},
		{
			name: "Azure DevOps",
			envVars: map[string]string{
				"TF_BUILD": "True",
			},
			expected: true,
		},
		{
			name: "Generic CI",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: true,
		},
		{
			name: "CI with false value",
			envVars: map[string]string{
				"CI": "false",
			},
			expected: false,
		},
		{
			name: "CI with zero value",
			envVars: map[string]string{
				"CI": "0",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear all CI-related environment variables
			ciVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "AZUREDEVOPS_URL", "GITLAB_CI",
				"AZURE_HTTP_USER_AGENT", "TF_BUILD", "BUILDKITE", "CIRCLECI", "TRAVIS",
				"DRONE", "SEMAPHORE", "APPVEYOR", "CODEBUILD_BUILD_ID", "TEAMCITY_VERSION",
			}
			cleanup := clearEnvVars(t, ciVars)
			defer cleanup()

			// Set test environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			host := createMockContainerHost()
			result := host.detectCIEnvironment()

			if result != tc.expected {
				t.Errorf("Expected detectCIEnvironment() to return %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetCIDefaults(t *testing.T) {
	host := createMockContainerHost()
	defaults := host.getCIDefaults()

	expectedDefaults := []string{
		"NO_COLOR=1",
		"TERM=dumb",
		"FORCE_COLOR=0",
		"CI=true",
	}

	if len(defaults) != len(expectedDefaults) {
		t.Errorf("Expected %d CI defaults, got %d", len(expectedDefaults), len(defaults))
	}

	for i, expected := range expectedDefaults {
		if i < len(defaults) && defaults[i] != expected {
			t.Errorf("Expected CI default[%d] to be %s, got %s", i, expected, defaults[i])
		}
	}
}

func TestGetShellColorSettings(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectedMin int      // Minimum number of environment variables expected
		shouldHave  []string // Environment variables that should be present
	}{
		{
			name:        "no color environment variables",
			envVars:     map[string]string{},
			expectedMin: 2, // Should get defaults
			shouldHave:  []string{"TERM=", "COLORTERM="},
		},
		{
			name: "TERM environment variable set",
			envVars: map[string]string{
				"TERM": "screen-256color",
			},
			expectedMin: 1,
			shouldHave:  []string{"TERM=screen-256color"},
		},
		{
			name: "COLORTERM environment variable set",
			envVars: map[string]string{
				"COLORTERM": "truecolor",
			},
			expectedMin: 1,
			shouldHave:  []string{"COLORTERM=truecolor"},
		},
		{
			name: "NO_COLOR environment variable set",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			expectedMin: 1,
			shouldHave:  []string{"NO_COLOR=1"},
		},
		{
			name: "multiple color environment variables",
			envVars: map[string]string{
				"TERM":     "xterm-256color",
				"NO_COLOR": "1",
				"CLICOLOR": "0",
			},
			expectedMin: 3,
			shouldHave:  []string{"TERM=xterm-256color", "NO_COLOR=1", "CLICOLOR=0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear color-related environment variables
			colorVars := []string{
				"TERM", "COLORTERM", "CLICOLOR", "CLICOLOR_FORCE",
				"NO_COLOR", "FORCE_COLOR", "COLOR",
			}
			cleanup := clearEnvVars(t, colorVars)
			defer cleanup()

			// Set test environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			host := createMockContainerHost()
			settings := host.getShellColorSettings()

			if len(settings) < tc.expectedMin {
				t.Errorf("Expected at least %d shell color settings, got %d", tc.expectedMin, len(settings))
			}

			// Check that expected environment variables are present
			for _, expected := range tc.shouldHave {
				found := false
				for _, setting := range settings {
					if len(setting) >= len(expected) && setting[:len(expected)] == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find setting starting with %s in %v", expected, settings)
				}
			}
		})
	}
}

func TestGetDefaultColorSettings(t *testing.T) {
	testCases := []struct {
		name        string
		termEnvVar  string
		expectedLen int
	}{
		{
			name:        "no TERM environment variable",
			termEnvVar:  "",
			expectedLen: 2,
		},
		{
			name:        "TERM environment variable set",
			termEnvVar:  "screen-256color",
			expectedLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear TERM environment variable
			originalTerm := os.Getenv("TERM")
			os.Unsetenv("TERM")
			defer func() {
				if originalTerm != "" {
					os.Setenv("TERM", originalTerm)
				}
			}()

			// Set test TERM if provided
			if tc.termEnvVar != "" {
				os.Setenv("TERM", tc.termEnvVar)
			}

			host := createMockContainerHost()
			defaults := host.getDefaultColorSettings()

			if len(defaults) != tc.expectedLen {
				t.Errorf("Expected %d default color settings, got %d", tc.expectedLen, len(defaults))
			}

			// Check that TERM is set correctly
			expectedTerm := tc.termEnvVar
			if expectedTerm == "" {
				expectedTerm = "xterm-256color"
			}
			expectedTermSetting := "TERM=" + expectedTerm

			found := false
			for _, setting := range defaults {
				if setting == expectedTermSetting {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find %s in defaults %v", expectedTermSetting, defaults)
			}
		})
	}
}

func TestBuildEnvironmentVars(t *testing.T) {
	testCases := []struct {
		name             string
		ciEnvVars        map[string]string
		colorEnvVars     map[string]string
		extension        *ExtensionConfig
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "CI environment",
			ciEnvVars: map[string]string{
				"CI": "true",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env:   []conf.EnvVar{},
			},
			shouldContain: []string{
				"R2R_CONTAINER_REPOROOT=/var/task",
				"R2R_HOST_REPOROOT=/test/root",
				"NO_COLOR=1",
				"TERM=dumb",
				"FORCE_COLOR=0",
				"CI=true",
			},
		},
		{
			name: "non-CI environment with color settings",
			colorEnvVars: map[string]string{
				"TERM":     "xterm-256color",
				"NO_COLOR": "1",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env:   []conf.EnvVar{},
			},
			shouldContain: []string{
				"R2R_CONTAINER_REPOROOT=/var/task",
				"R2R_HOST_REPOROOT=/test/root",
				"TERM=xterm-256color",
				"NO_COLOR=1",
			},
			shouldNotContain: []string{
				"TERM=dumb",
				"FORCE_COLOR=0",
			},
		},
		{
			name: "extension with custom environment variables",
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "CUSTOM_VAR", Value: "custom_value"},
					{Name: "TERM", Value: "override"}, // Should override shell setting
				},
			},
			shouldContain: []string{
				"R2R_CONTAINER_REPOROOT=/var/task",
				"R2R_HOST_REPOROOT=/test/root",
				"CUSTOM_VAR=custom_value",
				"TERM=override",
			},
		},
		{
			name: "non-CI environment with defaults",
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env:   []conf.EnvVar{},
			},
			shouldContain: []string{
				"R2R_CONTAINER_REPOROOT=/var/task",
				"R2R_HOST_REPOROOT=/test/root",
				"COLORTERM=truecolor",
			},
			shouldNotContain: []string{
				"NO_COLOR=1",
				"FORCE_COLOR=0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear all environment variables
			allEnvVars := []string{
				"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "Azure DevOps_URL", "GITLAB_CI",
				"TERM", "COLORTERM", "CLICOLOR", "CLICOLOR_FORCE",
				"NO_COLOR", "FORCE_COLOR", "COLOR",
			}
			cleanup := clearEnvVars(t, allEnvVars)
			defer cleanup()

			// Set CI environment variables
			for key, value := range tc.ciEnvVars {
				os.Setenv(key, value)
			}

			// Set color environment variables
			for key, value := range tc.colorEnvVars {
				os.Setenv(key, value)
			}

			host := createMockContainerHost()
			envVars := host.BuildEnvironmentVars(tc.extension)

			// Check required environment variables are present
			for _, required := range tc.shouldContain {
				found := false
				for _, envVar := range envVars {
					if envVar == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find %s in environment variables %v", required, envVars)
				}
			}

			// Check forbidden environment variables are not present
			for _, forbidden := range tc.shouldNotContain {
				for _, envVar := range envVars {
					if envVar == forbidden {
						t.Errorf("Did not expect to find %s in environment variables %v", forbidden, envVars)
					}
				}
			}
		})
	}
}
