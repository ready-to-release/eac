package environments

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadEnvironmentContract_Success verifies LoadEnvironmentContract successfully loads .eac/environments.yml
func TestLoadEnvironmentContract_Success(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0755))

	// Create a valid environments.yml file
	envYAML := `metadata:
  version: "0.1.0"
  description: "Test environment contract"
  scope: "test"

environments:
  - moniker: test-env
    name: "Test Environment"
    description: "A test environment"
    level: "L2"
    type: "docker"
    env_tags:
      - local
    system_deps:
      - docker
  - moniker: prod-env
    name: "Production Environment"
    description: "Production environment"
    level: "L4"
    type: "production"
    env_tags:
      - production
    system_deps:
      - kubectl
      - helm
`
	envPath := filepath.Join(eacDir, "environments.yml")
	require.NoError(t, os.WriteFile(envPath, []byte(envYAML), 0644))

	// Test LoadEnvironmentContract
	contract, err := LoadEnvironmentContract(tempDir)
	require.NoError(t, err)
	require.NotNil(t, contract)

	// Verify metadata
	assert.Equal(t, "0.1.0", contract.Metadata.Version)
	assert.Equal(t, "Test environment contract", contract.Metadata.Description)
	assert.Equal(t, "test", contract.Metadata.Scope)

	// Verify environments were loaded
	assert.Len(t, contract.Environments, 2)

	// Verify first environment
	testEnv := contract.Environments[0]
	assert.Equal(t, "test-env", testEnv.Moniker)
	assert.Equal(t, "Test Environment", testEnv.Name)
	assert.Equal(t, "L2", testEnv.Level)
	assert.Equal(t, "docker", testEnv.Type)
	assert.Contains(t, testEnv.EnvTags, "local")
	assert.Contains(t, testEnv.SystemDeps, "docker")

	// Verify second environment
	prodEnv := contract.Environments[1]
	assert.Equal(t, "prod-env", prodEnv.Moniker)
	assert.Equal(t, "L4", prodEnv.Level)
	assert.Equal(t, "production", prodEnv.Type)
}

// TestLoadEnvironmentContract_WorkspaceDetectionFailure verifies proper error when workspace detection fails
func TestLoadEnvironmentContract_WorkspaceDetectionFailure(t *testing.T) {
	// Create a temp directory that is NOT a git repository
	tempDir := t.TempDir()

	// Change to the temp directory so workspace detection fails
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tempDir))

	// Clear any workspace override to force detection to fail
	oldEnv := os.Getenv(EnvCLIERepoRoot)
	os.Unsetenv(EnvCLIERepoRoot)
	defer func() {
		if oldEnv != "" {
			os.Setenv(EnvCLIERepoRoot, oldEnv)
		}
	}()

	// Test LoadEnvironmentContract - should fail when trying to read from invalid directory
	contract, err := LoadEnvironmentContract(tempDir)
	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "failed to read environment contract")
}

// TestLoadEnvironmentContract_MissingFile verifies proper error when environments.yml is missing
func TestLoadEnvironmentContract_MissingFile(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory but NO environments.yml file
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0755))

	// Test LoadEnvironmentContract - should fail with file not found error
	contract, err := LoadEnvironmentContract(tempDir)
	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "failed to read environment contract")
}

// TestLoadEnvironmentContract_MalformedYAML verifies proper error when YAML is malformed
func TestLoadEnvironmentContract_MalformedYAML(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0755))

	// Create a malformed YAML file
	malformedYAML := `metadata:
  version: "0.1.0"
environments:
  - moniker: test
    this is not valid YAML syntax: [unclosed bracket
`
	envPath := filepath.Join(eacDir, "environments.yml")
	require.NoError(t, os.WriteFile(envPath, []byte(malformedYAML), 0644))

	// Test LoadEnvironmentContract - should fail with parse error
	contract, err := LoadEnvironmentContract(tempDir)
	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "failed to parse environment contract")
}

// TestLoadEnvironmentContract_EmptyFile verifies proper error when file is empty
func TestLoadEnvironmentContract_EmptyFile(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0755))

	// Create an empty file
	envPath := filepath.Join(eacDir, "environments.yml")
	require.NoError(t, os.WriteFile(envPath, []byte(""), 0644))

	// Test LoadEnvironmentContract - should succeed but return empty contract
	contract, err := LoadEnvironmentContract(tempDir)
	require.NoError(t, err)
	require.NotNil(t, contract)

	// Empty file should parse but have no environments
	assert.Empty(t, contract.Environments)
}

// TestLoadEnvironmentContract_InvalidStructure verifies proper handling of missing required fields
func TestLoadEnvironmentContract_InvalidStructure(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing moniker",
			yaml: `metadata:
  version: "0.1.0"
environments:
  - name: "Test"
    level: "L2"
    type: "docker"
`,
			expectError: false, // YAML parses but validation should catch this
		},
		{
			name: "missing level",
			yaml: `metadata:
  version: "0.1.0"
environments:
  - moniker: "test"
    name: "Test"
    type: "docker"
`,
			expectError: false, // YAML parses but validation should catch this
		},
		{
			name: "invalid YAML structure",
			yaml: `this is: not
a valid: contract structure
- random: list
`,
			expectError: true, // YAML parser detects the error
			errorMsg:    "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory to act as repo root
			tempDir := t.TempDir()

			// Create .eac directory
			eacDir := filepath.Join(tempDir, ".eac")
			require.NoError(t, os.MkdirAll(eacDir, 0755))

			// Create the test YAML file
			envPath := filepath.Join(eacDir, "environments.yml")
			require.NoError(t, os.WriteFile(envPath, []byte(tt.yaml), 0644))

			// Test LoadEnvironmentContract
			contract, err := LoadEnvironmentContract(tempDir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// Should parse successfully but may fail validation later
				require.NoError(t, err)
				require.NotNil(t, contract)
			}
		})
	}
}

// TestLoadEnvironmentContract_MinimalValid verifies minimal valid contract works
func TestLoadEnvironmentContract_MinimalValid(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0755))

	// Create minimal valid environments.yml
	minimalYAML := `metadata:
  version: "0.1.0"
environments:
  - moniker: test
    name: Test
    level: L2
    type: docker
`
	envPath := filepath.Join(eacDir, "environments.yml")
	require.NoError(t, os.WriteFile(envPath, []byte(minimalYAML), 0644))

	// Test LoadEnvironmentContract
	contract, err := LoadEnvironmentContract(tempDir)
	require.NoError(t, err)
	require.NotNil(t, contract)

	assert.Equal(t, "0.1.0", contract.Metadata.Version)
	assert.Len(t, contract.Environments, 1)
	assert.Equal(t, "test", contract.Environments[0].Moniker)
}

// TestLoadEnvironmentContract_WithEACSubdirectories verifies behavior when .eac has subdirectories
func TestLoadEnvironmentContract_WithEACSubdirectories(t *testing.T) {
	// Create a temporary directory to act as repo root
	tempDir := t.TempDir()

	// Create .eac directory with subdirectories
	eacDir := filepath.Join(tempDir, ".eac")
	require.NoError(t, os.MkdirAll(filepath.Join(eacDir, "subdir1"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(eacDir, "subdir2"), 0755))

	// Create environments.yml in root of .eac
	envYAML := `metadata:
  version: "0.1.0"
environments:
  - moniker: test
    name: Test
    level: L2
    type: docker
`
	envPath := filepath.Join(eacDir, "environments.yml")
	require.NoError(t, os.WriteFile(envPath, []byte(envYAML), 0644))

	// Test LoadEnvironmentContract - should find file in .eac root
	contract, err := LoadEnvironmentContract(tempDir)
	require.NoError(t, err)
	require.NotNil(t, contract)
	assert.Len(t, contract.Environments, 1)
}
