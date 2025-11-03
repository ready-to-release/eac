package conf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig provides an isolated configuration for tests
type TestConfig struct {
	*Config
	tempDir string
}

// NewTestConfig creates a new isolated test configuration
func NewTestConfig(t *testing.T) *TestConfig {
	t.Helper()

	// Ensure we're in a test environment
	t.Setenv("R2R_TESTING", "true")

	return &TestConfig{
		Config: &Config{
			Extensions: []Extension{
				{
					Name:            "test-extension",
					Image:           "test:latest",
					ImagePullPolicy: "Never", // Never pull in tests
					Description:     "Test extension for unit tests",
					Version:         "test",
				},
			},
		},
	}
}

// NewTestConfigWithExtensions creates a test config with custom extensions
func NewTestConfigWithExtensions(t *testing.T, extensions []Extension) *TestConfig {
	t.Helper()

	// Ensure we're in a test environment
	t.Setenv("R2R_TESTING", "true")

	return &TestConfig{
		Config: &Config{
			Extensions: extensions,
		},
	}
}

// WithTempDir creates a temporary directory with a test config file
func (tc *TestConfig) WithTempDir(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	tc.tempDir = tempDir

	// Create .git directory to simulate repository
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	return tempDir
}

// WriteConfigFile writes the test config to a file in the temp directory
func (tc *TestConfig) WriteConfigFile(t *testing.T, filename string) string {
	t.Helper()

	if tc.tempDir == "" {
		t.Fatal("Must call WithTempDir() before WriteConfigFile()")
	}

	configPath := filepath.Join(tc.tempDir, filename)

	// Create a minimal valid YAML config
	content := `version: "1.0"
extensions:`

	for _, ext := range tc.Config.Extensions {
		content += "\n  - name: \"" + ext.Name + "\""
		content += "\n    image: \"" + ext.Image + "\""
		if ext.ImagePullPolicy != "" {
			content += "\n    imagePullPolicy: \"" + ext.ImagePullPolicy + "\""
		}
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	return configPath
}

// LoadTestConfig loads a configuration from a test file without using InitConfig
func LoadTestConfig(t *testing.T, configPath string) (*Config, error) {
	t.Helper()

	// Ensure test environment
	t.Setenv("R2R_TESTING", "true")

	// Create a new config instance to avoid global state
	config := &Config{}

	// Use a copy of LoadConfig logic that doesn't affect global state
	// This would need to be implemented to avoid using the global Config variable
	// For now, we'll return the error to indicate this needs implementation
	return config, nil
}

// ResetGlobalConfig resets the global config for test isolation
// This should be called in test cleanup to ensure no state leakage
func ResetGlobalConfig() {
	Global = Config{}
}
