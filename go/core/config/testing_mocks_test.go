//go:build L0 && ov

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestLoadTestingMocks_ValidConfig tests loading a valid configuration from YAML.
func TestLoadTestingMocks_ValidConfig(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     TestingMocksConfig
	}{
		{
			name: "complete configuration with all mocks enabled",
			yamlData: `mocks:
  ai:
    enabled: true
    mock_dir: "test/mocks/ai"
  security:
    enabled: true
    tools: ["trivy", "semgrep"]
  docker:
    enabled: true
  github:
    enabled: true
    no_workflows: true`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI: AIMockConfig{
						Enabled: true,
						MockDir: "test/mocks/ai",
					},
					Security: SecurityMockConfig{
						Enabled: true,
						Tools:   []string{"trivy", "semgrep"},
					},
					Docker: DockerMockConfig{
						Enabled: true,
					},
					GitHub: GitHubMockConfig{
						Enabled:     true,
						NoWorkflows: true,
					},
				},
			},
		},
		{
			name: "minimal configuration with only required fields",
			yamlData: `mocks:
  ai:
    enabled: false
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
		{
			name: "security mocks with empty tools array",
			yamlData: `mocks:
  ai:
    enabled: false
  security:
    enabled: true
    tools: []
  docker:
    enabled: false
  github:
    enabled: false`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{}},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
		{
			name: "github without no_workflows field",
			yamlData: `mocks:
  ai:
    enabled: false
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: true`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "testing-mocks.yml")
			err := os.WriteFile(configFile, []byte(tt.yamlData), 0644)
			require.NoError(t, err)

			// Load configuration
			got, err := LoadTestingMocks(configFile)
			require.NoError(t, err, "LoadTestingMocks should not error on valid YAML")

			// Verify results
			assert.Equal(t, tt.want, got, "Loaded configuration should match expected")
		})
	}
}

// TestLoadTestingMocks_PersonalOverrides tests merging personal overrides on top of base config.
func TestLoadTestingMocks_PersonalOverrides(t *testing.T) {
	tests := []struct {
		name         string
		baseYAML     string
		personalYAML string
		want         TestingMocksConfig
	}{
		{
			name: "personal overrides ai.enabled",
			baseYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "base/dir"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  ai:
    enabled: false`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false, MockDir: "base/dir"},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
		{
			name: "personal overrides ai.mock_dir",
			baseYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "base/dir"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "personal/custom/dir"`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "personal/custom/dir"},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
		{
			name: "personal overrides security.tools",
			baseYAML: `mocks:
  ai:
    enabled: true
  security:
    enabled: true
    tools: ["trivy"]
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  security:
    enabled: true
    tools: ["trivy", "semgrep", "zap"]`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy", "semgrep", "zap"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
		{
			name: "personal overrides multiple fields",
			baseYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "base/dir"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true
    no_workflows: false`,
			personalYAML: `mocks:
  ai:
    enabled: false
  docker:
    enabled: false
  github:
    no_workflows: true`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false, MockDir: "base/dir"},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: true},
				},
			},
		},
		{
			name: "personal overrides entire mock section",
			baseYAML: `mocks:
  ai:
    enabled: true
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  ai:
    enabled: false
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tmpDir := t.TempDir()
			baseFile := filepath.Join(tmpDir, "testing-mocks.yml")
			personalFile := filepath.Join(tmpDir, "testing-mocks.personal.yml")

			// Write base config
			err := os.WriteFile(baseFile, []byte(tt.baseYAML), 0644)
			require.NoError(t, err)

			// Write personal config
			err = os.WriteFile(personalFile, []byte(tt.personalYAML), 0644)
			require.NoError(t, err)

			// Load with overrides
			got, err := LoadTestingMocksWithOverrides(baseFile, personalFile)
			require.NoError(t, err, "LoadTestingMocksWithOverrides should not error")

			// Verify merged result
			assert.Equal(t, tt.want, got, "Merged configuration should match expected")
		})
	}
}

// TestLoadTestingMocks_MissingPersonalOverride tests that missing personal override file is gracefully handled.
func TestLoadTestingMocks_MissingPersonalOverride(t *testing.T) {
	baseYAML := `mocks:
  ai:
    enabled: true
    mock_dir: "base/dir"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`

	tmpDir := t.TempDir()
	baseFile := filepath.Join(tmpDir, "testing-mocks.yml")
	personalFile := filepath.Join(tmpDir, "testing-mocks.personal.yml") // Does not exist

	err := os.WriteFile(baseFile, []byte(baseYAML), 0644)
	require.NoError(t, err)

	// Load with non-existent personal file
	got, err := LoadTestingMocksWithOverrides(baseFile, personalFile)
	require.NoError(t, err, "Should not error when personal file is missing")

	// Should return base config unchanged
	want := TestingMocksConfig{
		Mocks: MocksConfig{
			AI:       AIMockConfig{Enabled: true, MockDir: "base/dir"},
			Security: SecurityMockConfig{Enabled: true},
			Docker:   DockerMockConfig{Enabled: true},
			GitHub:   GitHubMockConfig{Enabled: true},
		},
	}
	assert.Equal(t, want, got, "Should return base config when personal override is missing")
}

// TestLoadTestingMocks_EnvironmentFallback tests fallback to environment variables.
func TestLoadTestingMocks_EnvironmentFallback(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		fileYAML string // Empty string = missing file
		want     TestingMocksConfig
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"R2R_MOCK_AI":             "true",
				"R2R_MOCK_AI_DIR":         "env/mock/dir",
				"R2R_MOCK_SECURITY":       "true",
				"R2R_MOCK_SECURITY_TOOLS": "trivy,semgrep",
				"R2R_MOCK_DOCKER":         "true",
				"R2R_MOCK_GITHUB":         "true",
				"R2R_MOCK_GITHUB_NO_WORKFLOWS": "true",
			},
			fileYAML: "", // Missing file
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "env/mock/dir"},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy", "semgrep"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: true},
				},
			},
		},
		{
			name: "partial environment variables",
			envVars: map[string]string{
				"R2R_MOCK_AI":       "true",
				"R2R_MOCK_AI_DIR":   "env/ai",
				"R2R_MOCK_SECURITY": "false",
			},
			fileYAML: "", // Missing file
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "env/ai"},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
		{
			name: "file config takes precedence over environment",
			envVars: map[string]string{
				"R2R_MOCK_AI":     "false",
				"R2R_MOCK_AI_DIR": "env/dir",
			},
			fileYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "file/dir"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`,
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "file/dir"},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
		{
			name:     "no file, no environment variables = defaults",
			envVars:  map[string]string{},
			fileYAML: "", // Missing file
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
		{
			name: "boolean parsing from strings",
			envVars: map[string]string{
				"R2R_MOCK_AI":       "1",
				"R2R_MOCK_SECURITY": "yes",
				"R2R_MOCK_DOCKER":   "TRUE",
				"R2R_MOCK_GITHUB":   "on",
			},
			fileYAML: "", // Missing file
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "testing-mocks.yml")

			// Create file if YAML provided
			if tt.fileYAML != "" {
				err := os.WriteFile(configFile, []byte(tt.fileYAML), 0644)
				require.NoError(t, err)
			}

			// Load configuration (will use env vars if file doesn't exist)
			got, err := LoadTestingMocks(configFile)
			require.NoError(t, err, "LoadTestingMocks should not error")

			assert.Equal(t, tt.want, got, "Configuration should match expected")
		})
	}
}

// TestLoadTestingMocks_InvalidYAML tests handling of invalid YAML syntax.
func TestLoadTestingMocks_InvalidYAML(t *testing.T) {
	tests := []struct {
		name         string
		yamlData     string
		wantErrorMsg string
	}{
		{
			name:         "malformed YAML",
			yamlData:     `mocks:\n  ai: [unclosed bracket`,
			wantErrorMsg: "failed to parse YAML",
		},
		{
			name: "missing required field",
			yamlData: `mocks:
  ai:
    mock_dir: "dir/only"`,
			wantErrorMsg: "validation failed",
		},
		{
			name: "invalid field type",
			yamlData: `mocks:
  ai:
    enabled: "not-a-boolean"`,
			wantErrorMsg: "failed to parse YAML",
		},
		{
			name: "invalid security tool",
			yamlData: `mocks:
  ai:
    enabled: true
  security:
    enabled: true
    tools: ["invalid-tool"]
  docker:
    enabled: false
  github:
    enabled: false`,
			wantErrorMsg: "validation failed",
		},
		{
			name: "additional properties not allowed",
			yamlData: `mocks:
  ai:
    enabled: true
  unknown_field: "value"
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			wantErrorMsg: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "testing-mocks.yml")
			err := os.WriteFile(configFile, []byte(tt.yamlData), 0644)
			require.NoError(t, err)

			_, err = LoadTestingMocks(configFile)
			require.Error(t, err, "LoadTestingMocks should error on invalid YAML")
			assert.Contains(t, err.Error(), tt.wantErrorMsg, "Error message should indicate the problem")
		})
	}
}

// TestLoadTestingMocks_SchemaValidation tests validation against JSON schema.
func TestLoadTestingMocks_SchemaValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    TestingMocksConfig
		wantValid bool
		wantError string
	}{
		{
			name: "valid config",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "valid/dir"},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: false},
				},
			},
			wantValid: true,
		},
		{
			name: "invalid security tool",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"invalid"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
			wantValid: false,
			wantError: "validation failed",
		},
		{
			name: "empty mock_dir when required",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: ""},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
			wantValid: true, // mock_dir is optional even when enabled
		},
		{
			name: "duplicate security tools",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy", "trivy"}},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
			wantValid: false,
			wantError: "uniqueItems",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTestingMocks(tt.config)

			if tt.wantValid {
				assert.NoError(t, err, "ValidateTestingMocks should not error for valid config")
			} else {
				require.Error(t, err, "ValidateTestingMocks should error for invalid config")
				if tt.wantError != "" {
					assert.Contains(t, err.Error(), tt.wantError, "Error should contain expected message")
				}
			}
		})
	}
}

// TestLoadTestingMocks_EmptyConfig tests handling of empty configuration file.
func TestLoadTestingMocks_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "testing-mocks.yml")

	// Create empty file
	err := os.WriteFile(configFile, []byte(""), 0644)
	require.NoError(t, err)

	_, err = LoadTestingMocks(configFile)
	require.Error(t, err, "LoadTestingMocks should error on empty config")
	assert.Contains(t, err.Error(), "validation failed", "Should fail validation")
}

// TestLoadTestingMocks_MissingConfigFile tests handling when config file doesn't exist.
func TestLoadTestingMocks_MissingConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "non-existent.yml")

	// Load without any environment variables set (should use defaults)
	got, err := LoadTestingMocks(configFile)
	require.NoError(t, err, "LoadTestingMocks should not error when file is missing")

	// Should return default config
	want := TestingMocksConfig{
		Mocks: MocksConfig{
			AI:       AIMockConfig{Enabled: false},
			Security: SecurityMockConfig{Enabled: false},
			Docker:   DockerMockConfig{Enabled: false},
			GitHub:   GitHubMockConfig{Enabled: false},
		},
	}
	assert.Equal(t, want, got, "Should return default config when file is missing")
}

// TestTestingMocksConfig_MarshalJSON tests JSON serialization.
func TestTestingMocksConfig_MarshalJSON(t *testing.T) {
	config := TestingMocksConfig{
		Mocks: MocksConfig{
			AI:       AIMockConfig{Enabled: true, MockDir: "test/dir"},
			Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy", "semgrep"}},
			Docker:   DockerMockConfig{Enabled: true},
			GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: true},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err, "Should marshal to JSON without error")

	var unmarshaled TestingMocksConfig
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err, "Should unmarshal from JSON without error")

	assert.Equal(t, config, unmarshaled, "Round-trip JSON serialization should preserve data")
}

// TestTestingMocksConfig_MarshalYAML tests YAML serialization.
func TestTestingMocksConfig_MarshalYAML(t *testing.T) {
	config := TestingMocksConfig{
		Mocks: MocksConfig{
			AI:       AIMockConfig{Enabled: true, MockDir: "test/dir"},
			Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy"}},
			Docker:   DockerMockConfig{Enabled: true},
			GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: false},
		},
	}

	data, err := yaml.Marshal(config)
	require.NoError(t, err, "Should marshal to YAML without error")

	var unmarshaled TestingMocksConfig
	err = yaml.Unmarshal(data, &unmarshaled)
	require.NoError(t, err, "Should unmarshal from YAML without error")

	assert.Equal(t, config, unmarshaled, "Round-trip YAML serialization should preserve data")
}

// TestLoadTestingMocks_RealWorldScenarios tests realistic usage scenarios.
func TestLoadTestingMocks_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name        string
		baseYAML    string
		personalYAML string
		envVars     map[string]string
		description string
		want        TestingMocksConfig
	}{
		{
			name: "developer disables AI mocks locally for real API testing",
			baseYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "go/eac/specs/impl/eac-cli/assets"
  security:
    enabled: true
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  ai:
    enabled: false`,
			description: "Personal override allows local testing against real AI API",
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false, MockDir: "go/eac/specs/impl/eac-cli/assets"},
					Security: SecurityMockConfig{Enabled: true},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
		{
			name: "CI environment uses file config only",
			baseYAML: `mocks:
  ai:
    enabled: true
    mock_dir: "go/eac/specs/impl/eac-cli/assets"
  security:
    enabled: true
    tools: []
  docker:
    enabled: true
  github:
    enabled: true
    no_workflows: false`,
			personalYAML: "", // No personal file in CI
			envVars:     map[string]string{},
			description: "CI uses base config without personal overrides",
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true, MockDir: "go/eac/specs/impl/eac-cli/assets"},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true, NoWorkflows: false},
				},
			},
		},
		{
			name:     "legacy environment variable configuration",
			baseYAML: "", // No file
			personalYAML: "",
			envVars: map[string]string{
				"R2R_MOCK_SECURITY":       "true",
				"R2R_MOCK_SECURITY_TOOLS": "trivy",
				"R2R_MOCK_DOCKER":         "true",
			},
			description: "Backward compatibility with old environment-based config",
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
		},
		{
			name: "developer mocks only specific security tools",
			baseYAML: `mocks:
  ai:
    enabled: true
  security:
    enabled: true
    tools: []
  docker:
    enabled: true
  github:
    enabled: true`,
			personalYAML: `mocks:
  security:
    enabled: true
    tools: ["trivy"]`,
			description: "Personal override to mock only Trivy for faster local tests",
			want: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: true},
					Security: SecurityMockConfig{Enabled: true, Tools: []string{"trivy"}},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			tmpDir := t.TempDir()
			baseFile := filepath.Join(tmpDir, "testing-mocks.yml")
			personalFile := filepath.Join(tmpDir, "testing-mocks.personal.yml")

			// Write base config if provided
			if tt.baseYAML != "" {
				err := os.WriteFile(baseFile, []byte(tt.baseYAML), 0644)
				require.NoError(t, err)
			}

			// Write personal config if provided
			if tt.personalYAML != "" {
				err := os.WriteFile(personalFile, []byte(tt.personalYAML), 0644)
				require.NoError(t, err)
			}

			// Load configuration
			var got TestingMocksConfig
			var err error
			if tt.personalYAML != "" {
				got, err = LoadTestingMocksWithOverrides(baseFile, personalFile)
			} else {
				got, err = LoadTestingMocks(baseFile)
			}
			require.NoError(t, err, "Scenario '%s' should load without error: %s", tt.name, tt.description)

			assert.Equal(t, tt.want, got, "Scenario '%s': %s", tt.name, tt.description)
		})
	}
}

// TestLoadTestingMocks_EdgeCases tests edge cases and boundary conditions.
func TestLoadTestingMocks_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		yamlData    string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "empty tools array is valid",
			yamlData: `mocks:
  ai:
    enabled: false
  security:
    enabled: true
    tools: []
  docker:
    enabled: false
  github:
    enabled: false`,
			shouldError: false,
		},
		{
			name: "null tools array treated as empty",
			yamlData: `mocks:
  ai:
    enabled: false
  security:
    enabled: true
    tools: null
  docker:
    enabled: false
  github:
    enabled: false`,
			shouldError: false,
		},
		{
			name: "whitespace-only mock_dir is invalid",
			yamlData: `mocks:
  ai:
    enabled: true
    mock_dir: "   "
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			shouldError: true,
			errorMsg:    "validation failed",
		},
		{
			name: "negative field values are type errors",
			yamlData: `mocks:
  ai:
    enabled: -1
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			shouldError: true,
			errorMsg:    "failed to parse YAML",
		},
		{
			name: "nested additional properties",
			yamlData: `mocks:
  ai:
    enabled: true
    extra_field: "not-allowed"
  security:
    enabled: false
  docker:
    enabled: false
  github:
    enabled: false`,
			shouldError: true,
			errorMsg:    "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "testing-mocks.yml")
			err := os.WriteFile(configFile, []byte(tt.yamlData), 0644)
			require.NoError(t, err)

			_, err = LoadTestingMocks(configFile)

			if tt.shouldError {
				require.Error(t, err, "Should error for edge case: %s", tt.name)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error should contain expected message")
				}
			} else {
				require.NoError(t, err, "Should not error for edge case: %s", tt.name)
			}
		})
	}
}

// TestTestingMocksConfig_ToEnvironmentVariables tests conversion to environment variables.
func TestTestingMocksConfig_ToEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name string
		config TestingMocksConfig
		want []string
	}{
		{
			name: "all mocks enabled with all options",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI: AIMockConfig{
						Enabled: true,
						MockDir: "test/mocks/ai",
					},
					Security: SecurityMockConfig{
						Enabled: true,
						Tools:   []string{"trivy", "semgrep"},
					},
					Docker: DockerMockConfig{
						Enabled: true,
					},
					GitHub: GitHubMockConfig{
						Enabled:     true,
						NoWorkflows: true,
					},
				},
			},
			want: []string{
				"R2R_MOCK_AI=true",
				"R2R_MOCK_AI_DIR=test/mocks/ai",
				"R2R_MOCK_SECURITY=true",
				"R2R_MOCK_SECURITY_TOOLS=trivy,semgrep",
				"R2R_MOCK_DOCKER=true",
				"R2R_MOCK_STRUCTURIZR=true",
				"R2R_MOCK_GITHUB_CLI=true",
				"R2R_MOCK_NO_WORKFLOWS=true",
			},
		},
		{
			name: "all mocks disabled",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
			want: nil,
		},
		{
			name: "only AI mock enabled without mock_dir",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI: AIMockConfig{
						Enabled: true,
						MockDir: "",
					},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
			want: []string{
				"R2R_MOCK_AI=true",
			},
		},
		{
			name: "security mock enabled with empty tools array",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI: AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{
						Enabled: true,
						Tools:   []string{},
					},
					Docker: DockerMockConfig{Enabled: false},
					GitHub: GitHubMockConfig{Enabled: false},
				},
			},
			want: []string{
				"R2R_MOCK_SECURITY=true",
			},
		},
		{
			name: "github mock enabled without no_workflows",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: false},
					GitHub: GitHubMockConfig{
						Enabled:     true,
						NoWorkflows: false,
					},
				},
			},
			want: []string{
				"R2R_MOCK_GITHUB_CLI=true",
			},
		},
		{
			name: "docker mock provides legacy structurizr env var",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI:       AIMockConfig{Enabled: false},
					Security: SecurityMockConfig{Enabled: false},
					Docker:   DockerMockConfig{Enabled: true},
					GitHub:   GitHubMockConfig{Enabled: false},
				},
			},
			want: []string{
				"R2R_MOCK_DOCKER=true",
				"R2R_MOCK_STRUCTURIZR=true",
			},
		},
		{
			name: "partial mocks enabled",
			config: TestingMocksConfig{
				Mocks: MocksConfig{
					AI: AIMockConfig{
						Enabled: true,
						MockDir: "custom/path",
					},
					Security: SecurityMockConfig{
						Enabled: true,
						Tools:   []string{"zap"},
					},
					Docker: DockerMockConfig{Enabled: false},
					GitHub: GitHubMockConfig{Enabled: false},
				},
			},
			want: []string{
				"R2R_MOCK_AI=true",
				"R2R_MOCK_AI_DIR=custom/path",
				"R2R_MOCK_SECURITY=true",
				"R2R_MOCK_SECURITY_TOOLS=zap",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ToEnvironmentVariables()
			assert.Equal(t, tt.want, got, "Environment variables should match expected")
		})
	}
}

// TestToEnvironmentVariables_RoundTrip tests that config can be converted to env vars
// and the env vars produce equivalent behavior.
func TestToEnvironmentVariables_RoundTrip(t *testing.T) {
	originalConfig := TestingMocksConfig{
		Mocks: MocksConfig{
			AI: AIMockConfig{
				Enabled: true,
				MockDir: "test/dir",
			},
			Security: SecurityMockConfig{
				Enabled: true,
				Tools:   []string{"trivy", "semgrep"},
			},
			Docker: DockerMockConfig{
				Enabled: true,
			},
			GitHub: GitHubMockConfig{
				Enabled:     true,
				NoWorkflows: true,
			},
		},
	}

	// Convert to environment variables
	envVars := originalConfig.ToEnvironmentVariables()

	// Verify all expected env vars are present
	expectedEnvVars := []string{
		"R2R_MOCK_AI=true",
		"R2R_MOCK_AI_DIR=test/dir",
		"R2R_MOCK_SECURITY=true",
		"R2R_MOCK_SECURITY_TOOLS=trivy,semgrep",
		"R2R_MOCK_DOCKER=true",
		"R2R_MOCK_STRUCTURIZR=true",
		"R2R_MOCK_GITHUB_CLI=true",
		"R2R_MOCK_NO_WORKFLOWS=true",
	}

	assert.ElementsMatch(t, expectedEnvVars, envVars, "Should generate all expected env vars")

	// Set the environment variables
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		require.Len(t, parts, 2, "Env var should be in KEY=VALUE format")
		t.Setenv(parts[0], parts[1])
	}

	// Load config from environment (simulating fallback behavior)
	configFromEnv := loadFromEnvironment()

	// Verify the environment-loaded config matches the original
	assert.Equal(t, originalConfig.Mocks.AI.Enabled, configFromEnv.Mocks.AI.Enabled, "AI enabled should match")
	assert.Equal(t, originalConfig.Mocks.AI.MockDir, configFromEnv.Mocks.AI.MockDir, "AI mock_dir should match")
	assert.Equal(t, originalConfig.Mocks.Security.Enabled, configFromEnv.Mocks.Security.Enabled, "Security enabled should match")
	assert.ElementsMatch(t, originalConfig.Mocks.Security.Tools, configFromEnv.Mocks.Security.Tools, "Security tools should match")
	assert.Equal(t, originalConfig.Mocks.Docker.Enabled, configFromEnv.Mocks.Docker.Enabled, "Docker enabled should match")
	assert.Equal(t, originalConfig.Mocks.GitHub.Enabled, configFromEnv.Mocks.GitHub.Enabled, "GitHub enabled should match")
	assert.Equal(t, originalConfig.Mocks.GitHub.NoWorkflows, configFromEnv.Mocks.GitHub.NoWorkflows, "GitHub no_workflows should match")
}
