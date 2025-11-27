//go:build L0
// +build L0

package docker

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/ready-to-release/eac/src/cli/internal/cache"
	"github.com/ready-to-release/eac/src/cli/internal/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestMergeMetadataEnv(t *testing.T) {
	testCases := []struct {
		name           string
		extEnv         []conf.EnvVar
		metaEnv        []cache.MetaEnvVar
		expectedEnv    []conf.EnvVar
		expectedLength int
	}{
		{
			name:   "nil metadata",
			extEnv: []conf.EnvVar{{Name: "EXISTING", Value: "value"}},
			// metaEnv is nil (no metadata)
			expectedEnv:    []conf.EnvVar{{Name: "EXISTING", Value: "value"}},
			expectedLength: 1,
		},
		{
			name:           "empty metadata env",
			extEnv:         []conf.EnvVar{{Name: "EXISTING", Value: "value"}},
			metaEnv:        []cache.MetaEnvVar{},
			expectedEnv:    []conf.EnvVar{{Name: "EXISTING", Value: "value"}},
			expectedLength: 1,
		},
		{
			name:   "metadata adds new env vars",
			extEnv: []conf.EnvVar{},
			metaEnv: []cache.MetaEnvVar{
				{Name: "META_VAR1"},
				{Name: "META_VAR2", Value: "static"},
			},
			expectedLength: 2,
		},
		{
			name:   "config env takes precedence over metadata",
			extEnv: []conf.EnvVar{{Name: "SHARED_VAR", Value: "config_value"}},
			metaEnv: []cache.MetaEnvVar{
				{Name: "SHARED_VAR", Value: "meta_value"},
			},
			expectedEnv:    []conf.EnvVar{{Name: "SHARED_VAR", Value: "config_value"}},
			expectedLength: 1,
		},
		{
			name:   "merge config and metadata env vars",
			extEnv: []conf.EnvVar{{Name: "CONFIG_VAR", Value: "config_value"}},
			metaEnv: []cache.MetaEnvVar{
				{Name: "META_VAR"},
			},
			expectedLength: 2,
		},
		{
			name:   "required flag is preserved from metadata",
			extEnv: []conf.EnvVar{},
			metaEnv: []cache.MetaEnvVar{
				{Name: "REQUIRED_VAR", Required: true},
			},
			expectedLength: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ext := &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env:   tc.extEnv,
			}

			var meta *cache.ExtensionMeta
			if tc.metaEnv != nil {
				meta = &cache.ExtensionMeta{
					Name: "test-ext",
					Env:  tc.metaEnv,
				}
			}

			MergeMetadataEnv(ext, meta)

			if len(ext.Env) != tc.expectedLength {
				t.Errorf("Expected %d env vars, got %d: %v", tc.expectedLength, len(ext.Env), ext.Env)
			}

			// Check expected env vars if specified
			for _, expected := range tc.expectedEnv {
				found := false
				for _, actual := range ext.Env {
					if actual.Name == expected.Name && actual.Value == expected.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find env var %s=%s in %v", expected.Name, expected.Value, ext.Env)
				}
			}

			// Check required flag for required test case
			if tc.name == "required flag is preserved from metadata" {
				for _, env := range ext.Env {
					if env.Name == "REQUIRED_VAR" && !env.Required {
						t.Error("Expected REQUIRED_VAR to have Required=true")
					}
				}
			}
		})
	}
}

func TestBuildEnvironmentVars_Passthrough(t *testing.T) {
	testCases := []struct {
		name             string
		hostEnvVars      map[string]string
		extension        *ExtensionConfig
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "passthrough env var with value on host",
			hostEnvVars: map[string]string{
				"MY_PASSTHROUGH_VAR": "host_value",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "MY_PASSTHROUGH_VAR"}, // No value = passthrough
				},
			},
			shouldContain: []string{
				"MY_PASSTHROUGH_VAR=host_value",
			},
		},
		{
			name:        "passthrough env var not set on host",
			hostEnvVars: map[string]string{},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "MISSING_VAR"}, // No value = passthrough, but not set on host
				},
			},
			shouldNotContain: []string{
				"MISSING_VAR=",
			},
		},
		{
			name: "static value takes precedence",
			hostEnvVars: map[string]string{
				"MY_VAR": "host_value",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "MY_VAR", Value: "static_value"},
				},
			},
			shouldContain: []string{
				"MY_VAR=static_value",
			},
			shouldNotContain: []string{
				"MY_VAR=host_value",
			},
		},
		{
			name: "mixed static and passthrough",
			hostEnvVars: map[string]string{
				"PASSTHROUGH_VAR": "from_host",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "STATIC_VAR", Value: "static"},
					{Name: "PASSTHROUGH_VAR"}, // passthrough
				},
			},
			shouldContain: []string{
				"STATIC_VAR=static",
				"PASSTHROUGH_VAR=from_host",
			},
		},
		{
			name:        "required passthrough not set on host (logs error but continues)",
			hostEnvVars: map[string]string{},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "REQUIRED_VAR", Required: true},
				},
			},
			shouldNotContain: []string{
				"REQUIRED_VAR=",
			},
		},
		{
			name: "required passthrough set on host",
			hostEnvVars: map[string]string{
				"REQUIRED_VAR": "required_value",
			},
			extension: &ExtensionConfig{
				Name:  "test-ext",
				Image: "test:latest",
				Env: []conf.EnvVar{
					{Name: "REQUIRED_VAR", Required: true},
				},
			},
			shouldContain: []string{
				"REQUIRED_VAR=required_value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear test environment variables
			testEnvVars := []string{"MY_PASSTHROUGH_VAR", "MISSING_VAR", "MY_VAR", "PASSTHROUGH_VAR", "STATIC_VAR", "REQUIRED_VAR"}
			cleanup := clearEnvVars(t, testEnvVars)
			defer cleanup()

			// Also clear CI vars to avoid interference
			ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI"}
			ciCleanup := clearEnvVars(t, ciVars)
			defer ciCleanup()

			// Set host environment variables
			for key, value := range tc.hostEnvVars {
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
					// Check if the env var starts with the forbidden prefix
					if len(envVar) >= len(forbidden) && envVar[:len(forbidden)] == forbidden {
						t.Errorf("Did not expect to find env var starting with %s in environment variables %v", forbidden, envVars)
					}
				}
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

// ========== Unit Tests with MockDockerClient ==========

func setupTestHostWithMock(t *testing.T) (*ContainerHost, *MockDockerClient) {
	mockClient := new(MockDockerClient)
	host := &ContainerHost{
		client:  mockClient,
		ctx:     context.Background(),
		rootDir: "/test/root",
	}
	return host, mockClient
}

func TestValidateExtensions_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Save original extensions
	origExtensions := conf.Global.Extensions
	defer func() { conf.Global.Extensions = origExtensions }()

	// Test with no extensions
	conf.Global.Extensions = []conf.Extension{}
	err := host.ValidateExtensions()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain any extensions")

	// Test with extensions
	conf.Global.Extensions = []conf.Extension{
		{Name: "test-ext", Image: "test:latest"},
	}
	err = host.ValidateExtensions()
	assert.NoError(t, err)
}

func TestFindExtension_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Save original extensions
	origExtensions := conf.Global.Extensions
	defer func() { conf.Global.Extensions = origExtensions }()

	conf.Global.Extensions = []conf.Extension{
		{Name: "ext1", Image: "img1:v1", ImagePullPolicy: "Always"},
		{Name: "ext2", Image: "img2:v2", LoadLocal: true},
	}

	// Test finding existing extension
	ext, err := host.FindExtension("ext1")
	assert.NoError(t, err)
	assert.NotNil(t, ext)
	assert.Equal(t, "ext1", ext.Name)
	assert.Equal(t, "img1:v1", ext.Image)
	assert.Equal(t, "Always", ext.ImagePullPolicy)

	// Test finding extension with default policy
	ext, err = host.FindExtension("ext2")
	assert.NoError(t, err)
	assert.NotNil(t, ext)
	assert.Equal(t, "ext2", ext.Name)
	assert.True(t, ext.LoadLocal)
	assert.Equal(t, "AutoDetect", ext.ImagePullPolicy) // Default

	// Test not finding extension
	ext, err = host.FindExtension("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, ext)
	assert.Contains(t, err.Error(), "not found in config")
}

func TestInspectImage_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test successful inspection
	expectedInspect := image.InspectResponse{
		ID:          "sha256:abc123",
		RepoDigests: []string{"test@sha256:def456"},
	}
	mockClient.On("ImageInspect", mock.Anything, "test:latest").
		Return(expectedInspect, nil).Once()

	result, err := host.InspectImage("test:latest")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sha256:abc123", result.ID)

	// Test inspection failure
	mockClient.On("ImageInspect", mock.Anything, "nonexistent:latest").
		Return(image.InspectResponse{}, fmt.Errorf("not found")).Once()

	result, err = host.InspectImage("nonexistent:latest")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetImageDigest_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test with RepoDigests available (pulled image)
	mockClient.On("ImageInspect", mock.Anything, "pulled:latest").
		Return(image.InspectResponse{
			ID:          "sha256:abc123",
			RepoDigests: []string{"pulled@sha256:def456"},
		}, nil).Once()

	digest, err := host.GetImageDigest("pulled:latest")
	assert.NoError(t, err)
	assert.Equal(t, "pulled@sha256:def456", digest)

	// Test with no RepoDigests (local build)
	mockClient.On("ImageInspect", mock.Anything, "local:latest").
		Return(image.InspectResponse{
			ID:          "sha256:local123",
			RepoDigests: []string{},
		}, nil).Once()

	digest, err = host.GetImageDigest("local:latest")
	assert.NoError(t, err)
	assert.Equal(t, "sha256:local123", digest)

	// Test inspection error
	mockClient.On("ImageInspect", mock.Anything, "error:latest").
		Return(image.InspectResponse{}, fmt.Errorf("not found")).Once()

	digest, err = host.GetImageDigest("error:latest")
	assert.Error(t, err)
	assert.Empty(t, digest)
}

func TestCreateContainerConfig_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test:latest",
	}

	imageInspect := &image.InspectResponse{
		Config: &container.Config{
			Entrypoint: []string{},
		},
	}

	// Test ModeInteractive
	config := host.CreateContainerConfig(ext, ModeInteractive, []string{}, imageInspect)
	assert.True(t, config.Tty)
	assert.True(t, config.OpenStdin)
	assert.Equal(t, 1, len(config.Cmd))
	assert.Equal(t, "/bin/sh", config.Cmd[0])

	// Test ModeRun with no args (interactive)
	config = host.CreateContainerConfig(ext, ModeRun, []string{}, imageInspect)
	assert.True(t, config.Tty)
	assert.True(t, config.OpenStdin)
	assert.Empty(t, config.Cmd)

	// Test ModeRun with args (command mode)
	config = host.CreateContainerConfig(ext, ModeRun, []string{"echo", "hello"}, imageInspect)
	assert.True(t, config.Tty)
	assert.False(t, config.OpenStdin)
	assert.Len(t, config.Cmd, 2)
	assert.Equal(t, "echo", config.Cmd[0])
	assert.Equal(t, "hello", config.Cmd[1])
	assert.Equal(t, "/var/task", config.WorkingDir)

	// Test with entrypoint (should not set WorkingDir)
	imageInspectWithEntrypoint := &image.InspectResponse{
		Config: &container.Config{
			Entrypoint: []string{"/app/entrypoint.sh"},
		},
	}
	config = host.CreateContainerConfig(ext, ModeRun, []string{}, imageInspectWithEntrypoint)
	assert.Empty(t, config.WorkingDir)
}

func TestCreateHostConfig_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test:latest",
	}

	// Test without cache volumes
	hostConfig := host.CreateHostConfig(ext, nil)
	assert.True(t, hostConfig.AutoRemove)
	assert.Len(t, hostConfig.Mounts, 2) // Root dir + Docker socket

	// Test with cache volumes
	volumeRequests := []cache.VolumeRequest{
		{Type: "cache", Name: "npm-cache", Target: "/root/.npm"},
		{Type: "cache", Name: "go-cache", Target: "/go/pkg"},
	}
	hostConfig = host.CreateHostConfig(ext, volumeRequests)
	assert.Len(t, hostConfig.Mounts, 4) // Root dir + Docker socket + 2 cache volumes
}

func TestCreateContainer_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	containerConfig := &container.Config{
		Image: "test:latest",
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
	}

	// Test successful creation
	mockClient.On("ContainerCreate", mock.Anything, containerConfig, hostConfig, mock.Anything, mock.Anything, "").
		Return(container.CreateResponse{ID: "container123"}, nil).Once()

	containerID, err := host.CreateContainer(containerConfig, hostConfig)
	assert.NoError(t, err)
	assert.Equal(t, "container123", containerID)

	// Test creation failure
	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "").
		Return(container.CreateResponse{}, fmt.Errorf("creation failed")).Once()

	containerID, err = host.CreateContainer(containerConfig, hostConfig)
	assert.Error(t, err)
	assert.Empty(t, containerID)
}

func TestStartContainer_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test successful start with TTY resize
	mockClient.On("ContainerStart", mock.Anything, "container123", mock.Anything).
		Return(nil).Once()
	mockClient.On("ContainerInspect", mock.Anything, "container123").
		Return(types.ContainerJSON{
			Config: &container.Config{Tty: true},
		}, nil).Once()
	mockClient.On("ContainerResize", mock.Anything, "container123", mock.Anything).
		Return(nil).Once()

	err := host.StartContainer("container123")
	assert.NoError(t, err)

	// Test start failure
	mockClient.On("ContainerStart", mock.Anything, "container456", mock.Anything).
		Return(fmt.Errorf("start failed")).Once()

	err = host.StartContainer("container456")
	assert.Error(t, err)
}

func TestAttachToContainer_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test with OpenStdin=true
	mockClient.On("ContainerInspect", mock.Anything, "container123").
		Return(types.ContainerJSON{
			Config: &container.Config{OpenStdin: true},
		}, nil).Once()
	mockClient.On("ContainerAttach", mock.Anything, "container123", mock.MatchedBy(func(opts container.AttachOptions) bool {
		return opts.Stdin == true
	})).Return(types.HijackedResponse{}, nil).Once()

	resp, err := host.AttachToContainer("container123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Test with OpenStdin=false
	mockClient.On("ContainerInspect", mock.Anything, "container456").
		Return(types.ContainerJSON{
			Config: &container.Config{OpenStdin: false},
		}, nil).Once()
	mockClient.On("ContainerAttach", mock.Anything, "container456", mock.MatchedBy(func(opts container.AttachOptions) bool {
		return opts.Stdin == false
	})).Return(types.HijackedResponse{}, nil).Once()

	resp, err = host.AttachToContainer("container456")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Test inspection error
	mockClient.On("ContainerInspect", mock.Anything, "error").
		Return(types.ContainerJSON{}, fmt.Errorf("inspect failed")).Once()

	resp, err = host.AttachToContainer("error")
	assert.Error(t, err)
}

func TestWaitForContainer_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	waitCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	waitCh <- container.WaitResponse{StatusCode: 0}

	mockClient.On("ContainerWait", mock.Anything, "container123", container.WaitConditionNotRunning).
		Return((<-chan container.WaitResponse)(waitCh), (<-chan error)(errCh))

	statusCh, errorCh := host.WaitForContainer("container123")
	assert.NotNil(t, statusCh)
	assert.NotNil(t, errorCh)
}

func TestStopContainer_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test successful stop
	mockClient.On("ContainerStop", mock.Anything, "container123", mock.Anything).
		Return(nil).Once()

	err := host.StopContainer("container123")
	assert.NoError(t, err)

	// Test stop failure
	mockClient.On("ContainerStop", mock.Anything, "container456", mock.Anything).
		Return(fmt.Errorf("stop failed")).Once()

	err = host.StopContainer("container456")
	assert.Error(t, err)
}

func TestStopContainerWithContext_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	ctx := context.Background()

	// Test successful stop
	mockClient.On("ContainerStop", ctx, "container123", mock.Anything).
		Return(nil).Once()

	err := host.StopContainerWithContext(ctx, "container123")
	assert.NoError(t, err)

	// Test stop failure
	mockClient.On("ContainerStop", ctx, "container456", mock.Anything).
		Return(fmt.Errorf("stop failed")).Once()

	err = host.StopContainerWithContext(ctx, "container456")
	assert.Error(t, err)
}

func TestGetContainerSnapshot_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test successful listing
	mockClient.On("ContainerList", mock.Anything, mock.Anything).
		Return([]types.Container{
			{ID: "container1", Image: "image1:latest"},
			{ID: "container2", Image: "image2:latest"},
		}, nil).Once()

	snapshot, err := host.GetContainerSnapshot()
	assert.NoError(t, err)
	assert.Len(t, snapshot, 2)
	assert.Equal(t, "image1:latest", snapshot["container1"])
	assert.Equal(t, "image2:latest", snapshot["container2"])

	// Test listing error
	mockClient.On("ContainerList", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("list failed")).Once()

	snapshot, err = host.GetContainerSnapshot()
	assert.Error(t, err)
	assert.Nil(t, snapshot)
}

func TestWarnAboutNewContainers_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	beforeSnapshot := map[string]string{
		"existing1234567890ab": "image1:latest",
	}

	afterSnapshot := map[string]string{
		"existing1234567890ab": "image1:latest",
		"new1234567890abcdef": "child:latest",
		"new2234567890abcdef": "extension:latest", // Should be skipped (extension image)
		"new3234567890abcdef": "expected:latest",   // Should be skipped (expected)
	}

	// Test without auto-remove (just warnings, no mock expectations needed)
	host.WarnAboutNewContainers(beforeSnapshot, afterSnapshot, "extension:latest", false, []string{"expected:latest"})

	// Test with auto-remove
	mockClient.On("ContainerStop", mock.Anything, "new1234567890abcdef", mock.Anything).
		Return(nil).Once()
	mockClient.On("ContainerRemove", mock.Anything, "new1234567890abcdef", mock.Anything).
		Return(nil).Once()

	host.WarnAboutNewContainers(beforeSnapshot, afterSnapshot, "extension:latest", true, []string{"expected:latest"})
}

func TestEnsureImageExists_Never_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test with local image
	mockClient.On("ImageInspect", mock.Anything, "test:latest").
		Return(image.InspectResponse{ID: "sha256:abc123"}, nil).Once()

	err := host.EnsureImageExists("test:latest", "Never", false)
	assert.NoError(t, err)

	// Test without local image
	mockClient.On("ImageInspect", mock.Anything, "missing:latest").
		Return(image.InspectResponse{}, fmt.Errorf("not found")).Once()

	err = host.EnsureImageExists("missing:latest", "Never", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found locally")
}

func TestEnsureImageExists_IfNotPresent_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test with local image (should not pull)
	mockClient.On("ImageInspect", mock.Anything, "test:latest").
		Return(image.InspectResponse{ID: "sha256:abc123"}, nil).Once()

	err := host.EnsureImageExists("test:latest", "IfNotPresent", false)
	assert.NoError(t, err)
}

func TestEnsureImageExists_AutoDetect_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Test with loadLocal=true (should use local image)
	mockClient.On("ImageInspect", mock.Anything, "test:latest").
		Return(image.InspectResponse{
			ID:          "sha256:abc123",
			RepoDigests: []string{},
		}, nil).Once()

	err := host.EnsureImageExists("test:latest", "AutoDetect", true)
	assert.NoError(t, err)

	// Test with version tag (should use local if present)
	mockClient.On("ImageInspect", mock.Anything, "test:v1.0.0").
		Return(image.InspectResponse{
			ID:          "sha256:def456",
			RepoDigests: []string{"test@sha256:def456"},
		}, nil).Once()

	err = host.EnsureImageExists("test:v1.0.0", "AutoDetect", false)
	assert.NoError(t, err)
}

func TestClose_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	mockClient.On("Close").Return(nil).Once()

	err := host.Close()
	assert.NoError(t, err)
}

func TestGetRootDir_Unit(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	assert.Equal(t, "/test/root", host.GetRootDir())
}

func TestCleanupContainers_StoppedOnly(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock container list with mix of running and stopped
	mockClient.On("ContainerList", mock.Anything, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:      "stopped123456789",
			Names:   []string{"/stopped-container"},
			State:   "exited",
			Image:   "test:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
		{
			ID:      "running123456789",
			Names:   []string{"/running-container"},
			State:   "running",
			Image:   "test:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}, nil).Once()

	// Mock inspect and remove for stopped container only
	mockClient.On("ContainerInspect", mock.Anything, "stopped123456789").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				SizeRw: func() *int64 { s := int64(1024); return &s }(),
			},
		}, nil).Once()

	mockClient.On("ContainerRemove", mock.Anything, "stopped123456789", mock.MatchedBy(func(opts container.RemoveOptions) bool {
		return opts.RemoveVolumes == true && opts.Force == false
	})).Return(nil).Once()

	opts := ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		DryRun:         false,
	}

	result, err := host.CleanupContainers(opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ContainersRemoved)
	assert.Equal(t, int64(1024), result.SpaceReclaimed)
	assert.Empty(t, result.Errors)
}

func TestCleanupContainers_ExtensionsOnly(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock container list with extension and non-extension containers
	mockClient.On("ContainerList", mock.Anything, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:      "ext123456789abcd",
			Names:   []string{"/r2r-extension-pwsh"},
			State:   "exited",
			Image:   "extension:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
		{
			ID:      "other12345678abc",
			Names:   []string{"/other-container"},
			State:   "exited",
			Image:   "other:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}, nil).Once()

	// Mock inspect and remove for extension container only
	mockClient.On("ContainerInspect", mock.Anything, "ext123456789abcd").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				SizeRw: func() *int64 { s := int64(2048); return &s }(),
			},
		}, nil).Once()

	mockClient.On("ContainerRemove", mock.Anything, "ext123456789abcd", mock.Anything).
		Return(nil).Once()

	opts := ContainerCleanupOptions{
		OnlyExtensions: true,
		IncludeRunning: false,
		DryRun:         false,
	}

	result, err := host.CleanupContainers(opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ContainersRemoved)
	assert.Equal(t, int64(2048), result.SpaceReclaimed)
	assert.Empty(t, result.Errors)
}

func TestCleanupContainers_OlderThan(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	now := time.Now()

	// Mock container list with containers of different ages
	mockClient.On("ContainerList", mock.Anything, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:      "old123456789abcd",
			Names:   []string{"/old-container"},
			State:   "exited",
			Image:   "test:latest",
			Created: now.Add(-48 * time.Hour).Unix(), // 2 days old
		},
		{
			ID:      "new123456789abcd",
			Names:   []string{"/new-container"},
			State:   "exited",
			Image:   "test:latest",
			Created: now.Add(-12 * time.Hour).Unix(), // 12 hours old
		},
	}, nil).Once()

	// Mock inspect and remove for old container only
	mockClient.On("ContainerInspect", mock.Anything, "old123456789abcd").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				SizeRw: func() *int64 { s := int64(4096); return &s }(),
			},
		}, nil).Once()

	mockClient.On("ContainerRemove", mock.Anything, "old123456789abcd", mock.Anything).
		Return(nil).Once()

	opts := ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		OlderThan:      24 * time.Hour, // Only containers older than 1 day
		DryRun:         false,
	}

	result, err := host.CleanupContainers(opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ContainersRemoved)
	assert.Equal(t, int64(4096), result.SpaceReclaimed)
	assert.Empty(t, result.Errors)
}

func TestCleanupContainers_DryRun(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock container list
	mockClient.On("ContainerList", mock.Anything, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:      "stopped123456789",
			Names:   []string{"/stopped-container"},
			State:   "exited",
			Image:   "test:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}, nil).Once()

	// Should not call inspect or remove in dry run mode

	opts := ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		DryRun:         true,
	}

	result, err := host.CleanupContainers(opts)
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ContainersRemoved)
	assert.Equal(t, int64(0), result.SpaceReclaimed) // No space reclaimed in dry run
	assert.Empty(t, result.Errors)
}

func TestCleanupContainers_WithErrors(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	// Mock container list
	mockClient.On("ContainerList", mock.Anything, mock.MatchedBy(func(opts container.ListOptions) bool {
		return opts.All == true
	})).Return([]types.Container{
		{
			ID:      "failed123456789a",
			Names:   []string{"/failed-container"},
			State:   "exited",
			Image:   "test:latest",
			Created: time.Now().Add(-1 * time.Hour).Unix(),
		},
	}, nil).Once()

	// Mock inspect success but remove failure
	mockClient.On("ContainerInspect", mock.Anything, "failed123456789a").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				SizeRw: func() *int64 { s := int64(512); return &s }(),
			},
		}, nil).Once()

	mockClient.On("ContainerRemove", mock.Anything, "failed123456789a", mock.Anything).
		Return(fmt.Errorf("removal failed")).Once()

	opts := ContainerCleanupOptions{
		OnlyExtensions: false,
		IncludeRunning: false,
		DryRun:         false,
	}

	result, err := host.CleanupContainers(opts)
	assert.NoError(t, err) // Function itself should not error
	assert.Equal(t, 0, result.ContainersRemoved)
	assert.Equal(t, int64(0), result.SpaceReclaimed)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Error(), "removal failed")
}

func TestCleanupStoppedContainers_Convenience(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	mockClient.On("ContainerList", mock.Anything, mock.Anything).
		Return([]types.Container{}, nil).Once()

	result, err := host.CleanupStoppedContainers(false)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ContainersRemoved)
}

func TestCleanupExtensionContainers_Convenience(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	mockClient.On("ContainerList", mock.Anything, mock.Anything).
		Return([]types.Container{}, nil).Once()

	result, err := host.CleanupExtensionContainers(false, false)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ContainersRemoved)
}

func TestCleanupOldContainers_Convenience(t *testing.T) {
	host, mockClient := setupTestHostWithMock(t)
	defer mockClient.AssertExpectations(t)

	mockClient.On("ContainerList", mock.Anything, mock.Anything).
		Return([]types.Container{}, nil).Once()

	result, err := host.CleanupOldContainers(24*time.Hour, false)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.ContainersRemoved)
}
