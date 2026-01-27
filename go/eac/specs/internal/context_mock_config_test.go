//go:build L1 && ov

package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildMockingEnvironment_LoadsConfigFromFile tests that buildMockingEnvironment
// correctly loads mock configuration from .r2r/eac/testing-mocks.yml.
func TestBuildMockingEnvironment_LoadsConfigFromFile(t *testing.T) {
	// Get the repository root (walk up from test file location)
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "Should find repository root")

	// Create test context
	ctx := NewTestContext()
	ctx.OriginalRepoRoot = repoRoot

	// Build mocking environment
	env := ctx.buildMockingEnvironment([]string{})

	// Verify environment variables are set
	hasContainerRoot := false
	hasMockSecurity := false
	hasMockDocker := false
	hasMockGitHub := false

	for _, e := range env {
		if strings.HasPrefix(e, "R2R_CONTAINER_ROOT=") {
			hasContainerRoot = true
		}
		if strings.HasPrefix(e, "R2R_MOCK_SECURITY=") {
			hasMockSecurity = true
		}
		if strings.HasPrefix(e, "R2R_MOCK_DOCKER=") || strings.HasPrefix(e, "R2R_MOCK_STRUCTURIZR=") {
			hasMockDocker = true
		}
		if strings.HasPrefix(e, "R2R_MOCK_GITHUB_CLI=") {
			hasMockGitHub = true
		}
	}

	assert.True(t, hasContainerRoot, "Should set R2R_CONTAINER_ROOT")
	assert.True(t, hasMockSecurity, "Should set R2R_MOCK_SECURITY from config")
	assert.True(t, hasMockDocker, "Should set R2R_MOCK_DOCKER from config")
	assert.True(t, hasMockGitHub, "Should set R2R_MOCK_GITHUB_CLI from config")
}

// TestBuildMockingEnvironment_FallbackToEnvVars tests that buildMockingEnvironment
// falls back to environment variables when config file is missing.
func TestBuildMockingEnvironment_FallbackToEnvVars(t *testing.T) {
	// Create temporary directory as fake repo root
	tempDir := t.TempDir()

	// Create test context pointing to temp dir (no config file)
	ctx := NewTestContext()
	ctx.OriginalRepoRoot = tempDir

	// Build mocking environment (should not error even with missing config)
	env := ctx.buildMockingEnvironment([]string{})

	// Should still have container root set
	hasContainerRoot := false
	for _, e := range env {
		if strings.HasPrefix(e, "R2R_CONTAINER_ROOT=") {
			hasContainerRoot = true
		}
	}

	assert.True(t, hasContainerRoot, "Should set R2R_CONTAINER_ROOT even without config file")
}

// TestBuildMockingEnvironment_PreservesExistingEnv tests that buildMockingEnvironment
// preserves existing environment variables.
func TestBuildMockingEnvironment_PreservesExistingEnv(t *testing.T) {
	// Get the repository root
	repoRoot, err := findRepoRoot()
	require.NoError(t, err, "Should find repository root")

	// Create test context
	ctx := NewTestContext()
	ctx.OriginalRepoRoot = repoRoot

	// Pass existing environment variables
	existingEnv := []string{
		"PATH=/usr/bin",
		"HOME=/home/user",
		"CUSTOM_VAR=value",
	}

	// Build mocking environment
	env := ctx.buildMockingEnvironment(existingEnv)

	// Verify existing vars are preserved
	assert.Contains(t, env, "PATH=/usr/bin", "Should preserve existing PATH")
	assert.Contains(t, env, "HOME=/home/user", "Should preserve existing HOME")
	assert.Contains(t, env, "CUSTOM_VAR=value", "Should preserve custom vars")

	// Verify new mock vars are added
	hasMockVars := false
	for _, e := range env {
		if strings.HasPrefix(e, "R2R_MOCK_") {
			hasMockVars = true
			break
		}
	}
	assert.True(t, hasMockVars, "Should add mock environment variables")
}

// findRepoRoot walks up from the current directory to find the repository root.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check if .r2r directory exists
		r2rDir := filepath.Join(dir, ".r2r")
		if _, err := os.Stat(r2rDir); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .r2r
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
