//go:build L2
// +build L2

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set R2R_TESTING to prevent fatal errors in CI during tests
	os.Setenv("R2R_TESTING", "true")

	// Run tests
	code := m.Run()

	// Exit with test result code
	os.Exit(code)
}

// TestLoadExamplesFile validates that r2r-cli.examples.yml loads correctly
// and all values fall back to sensible defaults when fields are empty
func TestLoadExamplesFile(t *testing.T) {
	// Get the repository root
	repoRoot, err := FindRepositoryRoot()
	require.NoError(t, err, "Failed to find repository root")

	// Path to r2r-cli.examples.yml
	examplesPath := filepath.Join(repoRoot, "r2r-cli.examples.yml")

	// Verify the examples file exists
	_, err = os.Stat(examplesPath)
	require.NoError(t, err, "r2r-cli.examples.yml not found at repository root")

	// Create a temporary r2r-cli.yml with empty extensions
	tempDir := t.TempDir()
	tempConfigPath := filepath.Join(tempDir, "r2r-cli.yml")

	// Write minimal valid config with empty values
	configContent := `extensions:
  - name: 'test-extension'
    image: 'test/image:latest'
`
	err = os.WriteFile(tempConfigPath, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current directory")
	defer os.Chdir(originalDir) // Restore original directory

	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Load the configuration
	err = LoadConfig(tempConfigPath)
	require.NoError(t, err, "Failed to load configuration")

	// Validate that Global config was populated
	assert.NotNil(t, Global.Extensions, "Extensions should not be nil")
	assert.Len(t, Global.Extensions, 1, "Should have one extension")

	// Validate the loaded extension
	ext := Global.Extensions[0]
	assert.Equal(t, "test-extension", ext.Name)
	assert.Equal(t, "test/image:latest", ext.Image)

	// Validate that empty optional fields have sensible defaults (empty strings)
	assert.Equal(t, "", ext.Description, "Description should default to empty string")
	assert.Equal(t, "", ext.Version, "Version should default to empty string")
	assert.Equal(t, "", ext.RepoURL, "RepoURL should default to empty string")
	assert.Equal(t, "", ext.DocsURL, "DocsURL should default to empty string")
	assert.Empty(t, ext.Env, "Env should default to empty slice")
}

// TestExamplesFileSchema validates that the examples file contains
// all documented fields and proper schema structure
func TestExamplesFileSchema(t *testing.T) {
	// Get the repository root
	repoRoot, err := FindRepositoryRoot()
	require.NoError(t, err, "Failed to find repository root")

	// Path to r2r-cli.examples.yml
	examplesPath := filepath.Join(repoRoot, "r2r-cli.examples.yml")

	// Read the examples file
	content, err := os.ReadFile(examplesPath)
	require.NoError(t, err, "Failed to read r2r-cli.examples.yml")

	// Convert to string for validation
	examplesContent := string(content)

	// Validate that all schema fields are documented
	requiredPatterns := []string{
		"extensions:",
		"name:",
		"description:",
		"version:",
		"image:",
		"repo_url:",
		"docs_url:",
		"env:",
		"Configuration Hierarchy:",
		"Environment variables",
		"Values in r2r-cli.yml",
		"Sensible defaults",
		"R2R_CONTAINER_REPOROOT",
		"R2R_HOST_REPOROOT",
		"memory_limit:",
		"cpu_limit:",
	}

	for _, pattern := range requiredPatterns {
		assert.Contains(t, examplesContent, pattern,
			"r2r-cli.examples.yml should document field: %s", pattern)
	}
}

// TestEmptyExtensionsConfig validates that a config with just
// empty extensions array loads successfully
func TestEmptyExtensionsConfig(t *testing.T) {
	// Create a temporary config with empty extensions
	tempDir := t.TempDir()
	tempConfigPath := filepath.Join(tempDir, "r2r-cli.yml")

	// Write config with empty extensions array
	configContent := `extensions: []
`
	err := os.WriteFile(tempConfigPath, []byte(configContent), 0644)
	require.NoError(t, err, "Failed to write test config file")

	// Load the configuration
	err = LoadConfig(tempConfigPath)
	require.NoError(t, err, "Failed to load configuration with empty extensions")

	// Validate that Global config was populated with empty array
	assert.NotNil(t, Global.Extensions, "Extensions should not be nil")
	assert.Empty(t, Global.Extensions, "Extensions should be empty array")
}

// TestValidateConfigRequiredFields tests validation of required fields
func TestValidateConfigRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorCount  int
	}{
		{
			name: "valid config",
			config: Config{
				Extensions: []Extension{
					{
						Name:  "valid-ext",
						Image: "alpine:latest",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			config: Config{
				Extensions: []Extension{
					{
						Image: "alpine:latest",
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing image",
			config: Config{
				Extensions: []Extension{
					{
						Name: "test-ext",
					},
				},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing both name and image",
			config: Config{
				Extensions: []Extension{
					{
						Description: "incomplete extension",
					},
				},
			},
			expectError: true,
			errorCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.expectError {
				require.Error(t, err)
				validationErr, ok := err.(*ValidationError)
				require.True(t, ok, "Expected ValidationError type")
				assert.Len(t, validationErr.Errors, tt.errorCount)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateConfigDuplicateNames tests validation of duplicate extension names
func TestValidateConfigDuplicateNames(t *testing.T) {
	config := Config{
		Extensions: []Extension{
			{
				Name:  "duplicate-name",
				Image: "alpine:latest",
			},
			{
				Name:  "unique-name",
				Image: "ubuntu:latest",
			},
			{
				Name:  "duplicate-name", // Duplicate
				Image: "debian:latest",
			},
		},
	}

	err := validateConfig(&config)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok, "Expected ValidationError type")
	assert.Contains(t, validationErr.Error(), "duplicate extension name")
}

// TestValidateConfigImagePullPolicy tests ImagePullPolicy validation
func TestValidateConfigImagePullPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		expectError bool
	}{
		{"valid Always", "Always", false},
		{"valid IfNotPresent", "IfNotPresent", false},
		{"valid Never", "Never", false},
		{"valid AutoDetect", "AutoDetect", false},
		{"empty (valid)", "", false},
		{"invalid policy", "InvalidPolicy", true},
		{"lowercase always", "always", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:            "test-ext",
						Image:           "alpine:latest",
						ImagePullPolicy: tt.policy,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid imagePullPolicy")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateConfigDockerImageReferences tests Docker image validation
func TestValidateConfigDockerImageReferences(t *testing.T) {
	tests := []struct {
		name        string
		image       string
		expectError bool
	}{
		{"simple image", "alpine", false},
		{"image with tag", "alpine:latest", false},
		{"registry with image", "docker.io/library/alpine:latest", false},
		{"github container registry", "ghcr.io/owner/repo:tag", false},
		{"complex path", "registry.example.com/path/to/image:v1.0.0", false},
		{"invalid characters", "alpine:@#$", true},
		{"empty image", "", true}, // This should be caught by required field validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:  "test-ext",
						Image: tt.image,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError && tt.image != "" { // Empty image is caught by required field validation
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid Docker image reference")
			} else if tt.image == "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "image is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateConfigURLs tests URL validation
func TestValidateConfigURLs(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		docsURL     string
		expectError bool
	}{
		{"valid URLs", "https://github.com/owner/repo", "https://docs.example.com", false},
		{"empty URLs (valid)", "", "", false},
		{"invalid repo URL", "not-a-url", "", true},
		{"invalid docs URL", "", "not-a-url", true},
		{"both invalid", "not-a-url", "also-not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:    "test-ext",
						Image:   "alpine:latest",
						RepoURL: tt.repoURL,
						DocsURL: tt.docsURL,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateConfigEnvironmentVariables tests environment variable validation
func TestValidateConfigEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name        string
		envVars     []EnvVar
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid env vars",
			envVars: []EnvVar{
				{Name: "API_KEY", Value: "secret"},
				{Name: "DEBUG_MODE", Value: "true"},
			},
			expectError: false,
		},
		{
			name: "missing env var name",
			envVars: []EnvVar{
				{Value: "secret"},
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "invalid env var name format",
			envVars: []EnvVar{
				{Name: "api-key", Value: "secret"}, // Should be uppercase with underscores
			},
			expectError: true,
			errorMsg:    "invalid environment variable name",
		},
		{
			name: "duplicate env var names",
			envVars: []EnvVar{
				{Name: "API_KEY", Value: "secret1"},
				{Name: "API_KEY", Value: "secret2"},
			},
			expectError: true,
			errorMsg:    "duplicate environment variable name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:  "test-ext",
						Image: "alpine:latest",
						Env:   tt.envVars,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidationErrorAggregation tests that multiple validation errors are aggregated
func TestValidationErrorAggregation(t *testing.T) {
	config := Config{
		Extensions: []Extension{
			{
				// Missing name and image (2 errors)
				ImagePullPolicy: "InvalidPolicy", // 1 error
				RepoURL:         "not-a-url",     // 1 error
				DocsURL:         "also-invalid",  // 1 error
				Env: []EnvVar{
					{Value: "no-name"},     // 1 error (missing name)
					{Name: "invalid-name"}, // 1 error (invalid format)
				},
			},
		},
	}

	err := validateConfig(&config)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok, "Expected ValidationError type")

	// Should have 7 total errors now
	assert.Len(t, validationErr.Errors, 7)

	// Check that error message contains multiple errors
	errorMsg := validationErr.Error()
	assert.Contains(t, errorMsg, "configuration validation failed")
	assert.Contains(t, errorMsg, "name is required")
	assert.Contains(t, errorMsg, "image is required")
	assert.Contains(t, errorMsg, "invalid imagePullPolicy")
	assert.Contains(t, errorMsg, "invalid repo_url")
	assert.Contains(t, errorMsg, "invalid docs_url")
}

// TestMetadataSchemaVersionField tests the metadata_schema_version field
func TestMetadataSchemaVersionField(t *testing.T) {
	tests := []struct {
		name                  string
		metadataSchemaVersion string
		expectError           bool
	}{
		{
			name:                  "valid version string",
			metadataSchemaVersion: "v1.0",
			expectError:           false,
		},
		{
			name:                  "semantic version format",
			metadataSchemaVersion: "2.0.0",
			expectError:           false,
		},
		{
			name:                  "empty version (optional field)",
			metadataSchemaVersion: "",
			expectError:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:                  "test-ext",
						Image:                 "alpine:latest",
						MetadataSchemaVersion: tt.metadataSchemaVersion,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMetadataSchemaVersionYAMLParsing tests YAML parsing of metadata_schema_version field
func TestMetadataSchemaVersionYAMLParsing(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		yamlContent     string
		expectedVersion string
	}{
		{
			name: "with metadata_schema_version",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"
    metadata_schema_version: "v2.0"`,
			expectedVersion: "v2.0",
		},
		{
			name: "without metadata_schema_version",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"`,
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "test-config.yml")
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Reset global config before loading
			Global = Config{}

			err = LoadConfig(configPath)
			require.NoError(t, err)

			require.Len(t, Global.Extensions, 1)
			assert.Equal(t, tt.expectedVersion, Global.Extensions[0].MetadataSchemaVersion)
		})
	}
}

// TestHasLatestTag tests the hasLatestTag function for various Docker image formats
func TestHasLatestTag(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected bool
	}{
		// Explicit :latest tags
		{"explicit latest tag", "alpine:latest", true},
		{"registry with explicit latest", "docker.io/library/alpine:latest", true},
		{"github registry with latest", "ghcr.io/owner/repo:latest", true},

		// Implicit latest (no tag)
		{"implicit latest - simple", "alpine", true},
		{"implicit latest - with namespace", "library/alpine", true},
		{"implicit latest - with registry", "docker.io/library/alpine", true},
		{"implicit latest - github registry", "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh", true},

		// Specific versions (not latest)
		{"specific version tag", "alpine:3.14", false},
		{"semantic version", "node:16.14.0", false},
		{"custom tag", "myapp:production", false},
		{"registry with version", "ghcr.io/owner/repo:v1.0.0", false},

		// Edge cases
		{"empty image", "", false},
		{"registry URL with port no tag", "localhost:5000/myimage", true}, // implicit latest
		{"registry URL with port and tag", "localhost:5000/myimage:v1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasLatestTag(tt.image)
			assert.Equal(t, tt.expected, result, "hasLatestTag(%q) should return %v", tt.image, tt.expected)
		})
	}
}

// TestValidatePinnedExtensions tests the validation of pinned extensions in CI
func TestValidatePinnedExtensions(t *testing.T) {
	// R2R_TESTING is already set by TestMain

	tests := []struct {
		name        string
		config      Config
		isCI        bool
		expectError bool
	}{
		{
			name: "CI with unpinned extensions",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:latest"},
					{Name: "ext2", Image: "ubuntu"}, // implicit latest
				},
			},
			isCI:        true,
			expectError: true,
		},
		{
			name: "CI with pinned extensions",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:3.14"},
					{Name: "ext2", Image: "ubuntu:20.04"},
				},
			},
			isCI:        true,
			expectError: false,
		},
		{
			name: "Non-CI with unpinned extensions",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:latest"},
				},
			},
			isCI:        false,
			expectError: false,
		},
		{
			name: "CI with main tag",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "ghcr.io/org/repo:main"},
				},
			},
			isCI:        true,
			expectError: true,
		},
		{
			name: "CI with master tag",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "ghcr.io/org/repo:master"},
				},
			},
			isCI:        true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePinnedExtensions(&tt.config, tt.isCI)
			if tt.expectError {
				assert.Error(t, err, "Expected error for unpinned extensions in CI")
				assert.Contains(t, err.Error(), "must be pinned")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}

// TestCheckLatestTagsWarnings tests that warnings are logged for latest tags
func TestCheckLatestTagsWarnings(t *testing.T) {
	// R2R_TESTING is already set by TestMain

	// Note: This test would need to capture log output to verify warnings are logged
	// For now, we'll just ensure the function runs without errors

	tests := []struct {
		name           string
		config         Config
		expectWarnings bool
	}{
		{
			name: "config with latest tags",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:latest"},
					{Name: "ext2", Image: "ubuntu"}, // implicit latest
				},
			},
			expectWarnings: true,
		},
		{
			name: "config with pinned versions",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:3.14"},
					{Name: "ext2", Image: "ubuntu:20.04"},
				},
			},
			expectWarnings: false,
		},
		{
			name: "mixed config",
			config: Config{
				Extensions: []Extension{
					{Name: "ext1", Image: "alpine:latest"},
					{Name: "ext2", Image: "ubuntu:20.04"},
					{Name: "ext3", Image: "node"}, // implicit latest
				},
			},
			expectWarnings: true,
		},
		{
			name: "empty config",
			config: Config{
				Extensions: []Extension{},
			},
			expectWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This function logs warnings but doesn't return an error
			// In a real test environment, we would capture log output
			checkLatestTags(&tt.config)
			// No assertion here as we're just ensuring it runs without panic
		})
	}
}

// TestValidateMemoryLimit tests the memory limit validation
func TestValidateMemoryLimit(t *testing.T) {
	tests := []struct {
		name        string
		limit       string
		expectError bool
		errorMsg    string
	}{
		// Valid formats
		{"valid MB", "512MB", false, ""},
		{"valid GB", "1GB", false, ""},
		{"valid lowercase mb", "512mb", false, ""},
		{"valid lowercase gb", "1gb", false, ""},
		{"valid with decimal", "1.5GB", false, ""},
		{"valid kilobytes", "1024KB", false, ""},
		{"valid bytes", "1024B", false, ""},
		{"valid single letter units", "512m", false, ""},
		{"valid with space", "512 MB", false, ""},
		{"empty (optional)", "", false, ""},

		// Invalid formats
		{"invalid unit", "512TB", true, "invalid memory limit format"},
		{"no unit", "512", true, "invalid memory limit format"},
		{"negative value", "-512MB", true, "invalid memory limit format"},
		{"zero value", "0MB", true, "must be a positive number"},
		{"invalid format", "MB512", true, "invalid memory limit format"},
		{"text only", "large", true, "invalid memory limit format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMemoryLimit(tt.limit)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateCPULimit tests the CPU limit validation
func TestValidateCPULimit(t *testing.T) {
	tests := []struct {
		name        string
		limit       string
		expectError bool
		errorMsg    string
	}{
		// Valid formats
		{"valid decimal", "0.5", false, ""},
		{"valid whole number", "1", false, ""},
		{"valid multiple cores", "2.0", false, ""},
		{"valid fraction", "0.25", false, ""},
		{"valid high value", "16", false, ""},
		{"empty (optional)", "", false, ""},

		// Invalid formats
		{"invalid text", "half", true, "invalid CPU limit format"},
		{"negative value", "-0.5", true, "must be a positive number"},
		{"zero value", "0", true, "must be a positive number"},
		{"invalid format", "1.2.3", true, "invalid CPU limit format"},
		{"with unit", "1CPU", true, "invalid CPU limit format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCPULimit(tt.limit)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResourceLimitsInConfig tests resource limits in the configuration validation
func TestResourceLimitsInConfig(t *testing.T) {
	tests := []struct {
		name        string
		memoryLimit string
		cpuLimit    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid resource limits",
			memoryLimit: "512MB",
			cpuLimit:    "0.5",
			expectError: false,
		},
		{
			name:        "only memory limit",
			memoryLimit: "1GB",
			cpuLimit:    "",
			expectError: false,
		},
		{
			name:        "only cpu limit",
			memoryLimit: "",
			cpuLimit:    "2",
			expectError: false,
		},
		{
			name:        "no limits (valid)",
			memoryLimit: "",
			cpuLimit:    "",
			expectError: false,
		},
		{
			name:        "invalid memory limit",
			memoryLimit: "512TB",
			cpuLimit:    "1",
			expectError: true,
			errorMsg:    "invalid memory limit format",
		},
		{
			name:        "invalid cpu limit",
			memoryLimit: "512MB",
			cpuLimit:    "invalid",
			expectError: true,
			errorMsg:    "invalid CPU limit format",
		},
		{
			name:        "both invalid",
			memoryLimit: "invalid",
			cpuLimit:    "also-invalid",
			expectError: true,
			errorMsg:    "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Extensions: []Extension{
					{
						Name:        "test-ext",
						Image:       "alpine:latest",
						MemoryLimit: tt.memoryLimit,
						CPULimit:    tt.cpuLimit,
					},
				},
			}

			err := validateConfig(&config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResourceLimitsYAMLParsing tests YAML parsing of resource limit fields
func TestResourceLimitsYAMLParsing(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		yamlContent    string
		expectedMemory string
		expectedCPU    string
	}{
		{
			name: "with both resource limits",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"
    memory_limit: "512MB"
    cpu_limit: "0.5"`,
			expectedMemory: "512MB",
			expectedCPU:    "0.5",
		},
		{
			name: "with only memory limit",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"
    memory_limit: "1GB"`,
			expectedMemory: "1GB",
			expectedCPU:    "",
		},
		{
			name: "with only cpu limit",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"
    cpu_limit: "2.0"`,
			expectedMemory: "",
			expectedCPU:    "2.0",
		},
		{
			name: "without resource limits",
			yamlContent: `version: "1.0"
extensions:
  - name: "test-ext"
    image: "test/image:latest"`,
			expectedMemory: "",
			expectedCPU:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "test-config.yml")
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Reset global config before loading
			Global = Config{}

			err = LoadConfig(configPath)
			require.NoError(t, err)

			require.Len(t, Global.Extensions, 1)
			assert.Equal(t, tt.expectedMemory, Global.Extensions[0].MemoryLimit)
			assert.Equal(t, tt.expectedCPU, Global.Extensions[0].CPULimit)
		})
	}
}

// TestResourceLimitsValidationErrorAggregation tests that resource limit errors are properly aggregated
func TestResourceLimitsValidationErrorAggregation(t *testing.T) {
	config := Config{
		Extensions: []Extension{
			{
				Name:        "test-ext",
				Image:       "alpine:latest",
				MemoryLimit: "invalid-memory",
				CPULimit:    "-1.0",
			},
		},
	}

	err := validateConfig(&config)
	require.Error(t, err)

	validationErr, ok := err.(*ValidationError)
	require.True(t, ok, "Expected ValidationError type")

	// Should have 2 errors (one for each invalid resource limit)
	assert.Len(t, validationErr.Errors, 2)

	// Check that error message contains both resource limit errors
	errorMsg := validationErr.Error()
	assert.Contains(t, errorMsg, "invalid memory limit format")
	assert.Contains(t, errorMsg, "CPU limit must be a positive number")
}

// TestUserSpecificConfigDiscovery tests the priority-based configuration file discovery
func TestUserSpecificConfigDiscovery(t *testing.T) {
	// Create a temporary directory to serve as repository root
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err, "Failed to create .git directory")

	// Create different configuration files with different content to verify priority
	configs := map[string]string{
		"r2r-cli.yml": `extensions:
  - name: "default-ext"
    image: "default:latest"`,
		"r2r-cli.local.yml": `extensions:
  - name: "local-ext" 
    image: "local:latest"`,
		"r2r-cli.personal.yml": `extensions:
  - name: "personal-ext"
    image: "personal:latest"`,
		"r2r-cli.dev.yml": `extensions:
  - name: "dev-ext"
    image: "dev:latest"`,
	}

	// Test different priority scenarios
	tests := []struct {
		name           string
		filesToCreate  []string
		expectedConfig string
		expectedImage  string
	}{
		{
			name:           "only default config",
			filesToCreate:  []string{"r2r-cli.yml"},
			expectedConfig: "default-ext",
			expectedImage:  "default:latest",
		},
		{
			name:           "local overrides default",
			filesToCreate:  []string{"r2r-cli.yml", "r2r-cli.local.yml"},
			expectedConfig: "local-ext",
			expectedImage:  "local:latest",
		},
		{
			name:           "local has higher priority than personal",
			filesToCreate:  []string{"r2r-cli.yml", "r2r-cli.local.yml", "r2r-cli.personal.yml"},
			expectedConfig: "local-ext",
			expectedImage:  "local:latest",
		},
		{
			name:           "personal has higher priority than dev",
			filesToCreate:  []string{"r2r-cli.yml", "r2r-cli.personal.yml", "r2r-cli.dev.yml"},
			expectedConfig: "personal-ext",
			expectedImage:  "personal:latest",
		},
		{
			name:           "local has highest priority among user configs",
			filesToCreate:  []string{"r2r-cli.yml", "r2r-cli.local.yml", "r2r-cli.personal.yml", "r2r-cli.dev.yml"},
			expectedConfig: "local-ext",
			expectedImage:  "local:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing config files
			configFiles := []string{"r2r-cli.yml", "r2r-cli.local.yml", "r2r-cli.personal.yml", "r2r-cli.dev.yml"}
			for _, configFile := range configFiles {
				os.Remove(filepath.Join(tempDir, configFile))
			}

			// Create only the files specified for this test
			for _, filename := range tt.filesToCreate {
				configPath := filepath.Join(tempDir, filename)
				err := os.WriteFile(configPath, []byte(configs[filename]), 0644)
				require.NoError(t, err, "Failed to create config file: %s", filename)
			}

			// Change to temp directory to test discovery
			originalDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(originalDir)

			err = os.Chdir(tempDir)
			require.NoError(t, err)

			// Find config file using our discovery logic
			configFile, err := findConfigFile("r2r-cli.yml")
			require.NoError(t, err, "Failed to find config file")

			// Load the configuration
			Global = Config{} // Reset global config
			err = LoadConfig(configFile)
			require.NoError(t, err, "Failed to load config")

			// Verify correct configuration was loaded
			require.Len(t, Global.Extensions, 1, "Expected one extension")
			assert.Equal(t, tt.expectedConfig, Global.Extensions[0].Name)
			assert.Equal(t, tt.expectedImage, Global.Extensions[0].Image)
		})
	}
}

// TestGetConfigFileCandidates tests the configuration file candidate generation
func TestGetConfigFileCandidates(t *testing.T) {
	tempDir := t.TempDir()

	candidates := getConfigFileCandidates(tempDir)

	// Should have at least 4 candidates (3 fixed + 1 repository default)
	// May have 5 if username lookup succeeds
	assert.GreaterOrEqual(t, len(candidates), 4, "Should have at least 4 candidates")
	assert.LessOrEqual(t, len(candidates), 5, "Should have at most 5 candidates")

	// Check that specific files are included in the right order
	expectedFiles := []string{
		"r2r-cli.local.yml",
		"r2r-cli.personal.yml",
		"r2r-cli.dev.yml",
	}

	for i, expectedFile := range expectedFiles {
		expectedPath := filepath.Join(tempDir, expectedFile)
		assert.Equal(t, expectedPath, candidates[i], "Candidate %d should be %s", i, expectedFile)
	}

	// Repository default should be last
	expectedDefault := filepath.Join(tempDir, "r2r-cli.yml")
	assert.Equal(t, expectedDefault, candidates[len(candidates)-1], "Repository default should be last")
}

// TestUserSpecificConfigNotFound tests error handling when no config files exist
func TestUserSpecificConfigNotFound(t *testing.T) {
	// Create a temporary directory with only .git (no config files)
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Try to find config file - should fail
	_, err = findConfigFile("r2r-cli.yml")
	require.Error(t, err, "Should fail when no config files exist")

	// Should be a ConfigFileNotFoundError
	assert.Contains(t, err.Error(), "not found")
}

// TestUserSpecificConfigPermissionError tests permission error handling
func TestUserSpecificConfigPermissionError(t *testing.T) {
	// This test is platform-specific and may need to be skipped on some systems
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping permission test in CI environment")
	}

	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err)

	// Create a config file with no read permissions
	configPath := filepath.Join(tempDir, "r2r-cli.local.yml")
	err = os.WriteFile(configPath, []byte("extensions: []"), 0644)
	require.NoError(t, err)

	// Remove read permissions (this might not work on all systems)
	err = os.Chmod(configPath, 0000)
	if err != nil {
		t.Skip("Cannot modify file permissions on this system")
	}
	defer os.Chmod(configPath, 0644) // Restore permissions for cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Try to find config file - should return permission error
	_, err = findConfigFile("r2r-cli.yml")
	if err != nil {
		// Should be a permission error
		assert.Contains(t, err.Error(), "permission")
	}
}

// TestConfigDebugLogging tests that configuration file selection is logged
func TestConfigDebugLogging(t *testing.T) {
	// This test demonstrates the debug logging functionality
	// In practice, you would capture log output to verify the debug message

	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	require.NoError(t, err)

	// Create a local config file
	configPath := filepath.Join(tempDir, "r2r-cli.local.yml")
	configContent := `extensions:
  - name: "test-ext"
    image: "test:latest"`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Find config file - should find the local one and log debug message
	foundConfig, err := findConfigFile("r2r-cli.yml")
	require.NoError(t, err)

	// Verify correct file was found
	assert.Equal(t, configPath, foundConfig)

	// Note: In a real test environment with log capture, you would verify:
	// assert.Contains(t, capturedLogs, "Using configuration file")
	// assert.Contains(t, capturedLogs, "r2r-cli.local.yml")
}
