//go:build L0
// +build L0

package docker

import (
	"os"
	"testing"

	"github.com/ready-to-release/eac/go/cli/clie/internal/conf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildEnvironmentVars_BaseVars verifies base environment variables
func TestBuildEnvironmentVars_BaseVars(t *testing.T) {
	// Set CLIE_HOST_REPOROOT to test context - this isolates the test from host environment
	t.Setenv("CLIE_HOST_REPOROOT", "/test/root")

	// Arrange
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert
	require.NotEmpty(t, envVars)
	assert.Contains(t, envVars, "CLIE_DOCKER_MODE=true")
	assert.Contains(t, envVars, "CLIE_CONTAINER_REPOROOT=/var/task")
	assert.Contains(t, envVars, "CLIE_HOST_REPOROOT=/test/root")

	// Should have HOST_GOOS and HOST_GOARCH
	hasGOOS := false
	hasGOARCH := false
	for _, env := range envVars {
		if len(env) >= 15 && env[:15] == "CLIE_HOST_GOOS=" {
			hasGOOS = true
		}
		if len(env) >= 17 && env[:17] == "CLIE_HOST_GOARCH=" {
			hasGOARCH = true
		}
	}
	assert.True(t, hasGOOS, "Should include CLIE_HOST_GOOS")
	assert.True(t, hasGOARCH, "Should include CLIE_HOST_GOARCH")
}

// TestBuildEnvironmentVars_TerminalDimensions verifies terminal size handling
func TestBuildEnvironmentVars_TerminalDimensions(t *testing.T) {
	// Arrange
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Clear any existing terminal env vars
	oldCols := os.Getenv("COLUMNS")
	oldLines := os.Getenv("LINES")
	defer func() {
		os.Setenv("COLUMNS", oldCols)
		os.Setenv("LINES", oldLines)
	}()

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert - should have terminal dimensions (either detected or default)
	hasCols := false
	hasLines := false
	hasDetection := false
	for _, env := range envVars {
		if len(env) >= 8 && env[:8] == "COLUMNS=" {
			hasCols = true
		}
		if len(env) >= 6 && env[:6] == "LINES=" {
			hasLines = true
		}
		if len(env) >= 24 && env[:24] == "CLIE_TERMINAL_DETECTION=" {
			hasDetection = true
		}
	}
	assert.True(t, hasCols, "Should include COLUMNS")
	assert.True(t, hasLines, "Should include LINES")
	assert.True(t, hasDetection, "Should include CLIE_TERMINAL_DETECTION")
}

// TestBuildEnvironmentVars_CIDetection verifies CI environment detection
func TestBuildEnvironmentVars_CIDetection(t *testing.T) {
	// Save original env
	originalCI := os.Getenv("CI")
	defer os.Setenv("CI", originalCI)

	// Arrange
	os.Setenv("CI", "true")
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert - should have CI defaults
	assert.Contains(t, envVars, "NO_COLOR=1")
	assert.Contains(t, envVars, "TERM=dumb")
	assert.Contains(t, envVars, "FORCE_COLOR=0")
	assert.Contains(t, envVars, "CI=true")
}

// TestBuildEnvironmentVars_NonCIShellColors verifies shell color inheritance
func TestBuildEnvironmentVars_NonCIShellColors(t *testing.T) {
	// Clear all CI indicator environment variables to simulate a non-CI environment
	// t.Setenv automatically restores the original value after the test
	ciIndicators := []string{
		"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "AZUREDEVOPS_URL", "GITLAB_CI",
		"AZURE_HTTP_USER_AGENT", "TF_BUILD", "BUILDKITE", "CIRCLECI", "TRAVIS",
		"DRONE", "SEMAPHORE", "APPVEYOR", "CODEBUILD_BUILD_ID", "TEAMCITY_VERSION",
	}
	for _, key := range ciIndicators {
		t.Setenv(key, "") // Setting to empty string effectively disables CI detection
	}

	// Arrange
	t.Setenv("TERM", "xterm-256color")
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert - should NOT have CI defaults, should inherit TERM
	assert.NotContains(t, envVars, "NO_COLOR=1")
	assert.Contains(t, envVars, "TERM=xterm-256color")
}

// TestBuildEnvironmentVars_GitHubCredentials verifies GitHub credentials passing
func TestBuildEnvironmentVars_GitHubCredentials(t *testing.T) {
	// Save original env
	originalUsername := os.Getenv("GITHUB_USERNAME")
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("GITHUB_USERNAME", originalUsername)
		os.Setenv("GITHUB_TOKEN", originalToken)
	}()

	// Arrange
	os.Setenv("GITHUB_USERNAME", "testuser")
	os.Setenv("GITHUB_TOKEN", "test-token-123")
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert
	assert.Contains(t, envVars, "GITHUB_USERNAME=testuser")
	assert.Contains(t, envVars, "GITHUB_TOKEN=test-token-123")
}

// TestBuildEnvironmentVars_ExtensionEnvVars verifies extension-specific env vars
func TestBuildEnvironmentVars_ExtensionEnvVars(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("PASSTHROUGH_VAR")
	defer os.Setenv("PASSTHROUGH_VAR", originalEnv)

	// Arrange
	os.Setenv("PASSTHROUGH_VAR", "passthrough-value")
	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
		Env: []conf.EnvVar{
			{Name: "STATIC_VAR", Value: "static-value"},
			{Name: "PASSTHROUGH_VAR"}, // No value = passthrough from host
		},
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert
	assert.Contains(t, envVars, "STATIC_VAR=static-value")
	assert.Contains(t, envVars, "PASSTHROUGH_VAR=passthrough-value")
}

// TestBuildEnvironmentVars_GlobalEnvironment verifies global config env vars
func TestBuildEnvironmentVars_GlobalEnvironment(t *testing.T) {
	// Save original config
	originalConfig := conf.Global.Environment
	defer func() { conf.Global.Environment = originalConfig }()

	// Arrange
	conf.Global.Environment = &conf.Environment{
		Global: []conf.EnvVar{
			{Name: "GLOBAL_VAR", Value: "global-value"},
		},
	}

	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert
	assert.Contains(t, envVars, "GLOBAL_VAR=global-value")
}

// TestBuildEnvironmentVars_SecretPassthrough verifies secret environment variable passing
func TestBuildEnvironmentVars_SecretPassthrough(t *testing.T) {
	// Save original config and env
	originalConfig := conf.Global.Environment
	originalSecret := os.Getenv("SECRET_VAR")
	defer func() {
		conf.Global.Environment = originalConfig
		os.Setenv("SECRET_VAR", originalSecret)
	}()

	// Arrange
	os.Setenv("SECRET_VAR", "secret-value")
	conf.Global.Environment = &conf.Environment{
		Secrets: []conf.SecretVar{
			{Name: "MY_SECRET", Env: "SECRET_VAR"},
		},
	}

	ch := &ContainerHost{
		rootDir: "/test/root",
	}
	ext := &ExtensionConfig{
		Name:  "test-ext",
		Image: "test/image:latest",
	}

	// Act
	envVars := ch.BuildEnvironmentVars(ext)

	// Assert
	assert.Contains(t, envVars, "MY_SECRET=secret-value")
}
