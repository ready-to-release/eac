package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// TestingMocksConfig represents the root configuration for test mocking.
type TestingMocksConfig struct {
	Mocks MocksConfig `yaml:"mocks" json:"mocks"`
}

// MocksConfig contains all mock configurations for different systems.
type MocksConfig struct {
	AI       AIMockConfig       `yaml:"ai" json:"ai"`
	Security SecurityMockConfig `yaml:"security" json:"security"`
	Docker   DockerMockConfig   `yaml:"docker" json:"docker"`
	GitHub   GitHubMockConfig   `yaml:"github" json:"github"`
}

// AIMockConfig configures AI provider mocking.
type AIMockConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	MockDir string `yaml:"mock_dir,omitempty" json:"mock_dir,omitempty"`
}

// SecurityMockConfig configures security tool mocking (Trivy, Semgrep, ZAP).
type SecurityMockConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Tools   []string `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// DockerMockConfig configures Docker client mocking.
type DockerMockConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// GitHubMockConfig configures GitHub CLI mocking.
type GitHubMockConfig struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`
	NoWorkflows bool `yaml:"no_workflows,omitempty" json:"no_workflows,omitempty"`
}

// LoadTestingMocks loads the testing mocks configuration from the specified file.
// If the file doesn't exist, it returns a default configuration based on environment variables.
func LoadTestingMocks(configFile string) (TestingMocksConfig, error) {
	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// File doesn't exist, fall back to environment variables
		return loadFromEnvironment(), nil
	}

	// Read the file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return TestingMocksConfig{}, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Check for empty file
	if len(data) == 0 || len(strings.TrimSpace(string(data))) == 0 {
		return TestingMocksConfig{}, fmt.Errorf("validation failed for %s: configuration file is empty", configFile)
	}

	// Pre-validate raw YAML to check for required fields before unmarshaling
	if err := validateRequiredFields(data); err != nil {
		return TestingMocksConfig{}, fmt.Errorf("validation failed for %s: %w", configFile, err)
	}

	// Parse YAML with strict mode to catch unknown fields
	var config TestingMocksConfig
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true) // Strict mode - reject unknown fields
	if err := decoder.Decode(&config); err != nil {
		// Check if it's an unknown field error (validation error)
		if strings.Contains(err.Error(), "field") && strings.Contains(err.Error(), "not found") {
			return TestingMocksConfig{}, fmt.Errorf("validation failed for %s: %w", configFile, err)
		}
		return TestingMocksConfig{}, fmt.Errorf("failed to parse YAML in %s: %w", configFile, err)
	}

	// Validate against schema
	if err := ValidateTestingMocks(config); err != nil {
		return TestingMocksConfig{}, fmt.Errorf("validation failed for %s: %w", configFile, err)
	}

	return config, nil
}

// LoadTestingMocksWithOverrides loads the base configuration and applies personal overrides.
// If the personal file doesn't exist, only the base configuration is loaded.
func LoadTestingMocksWithOverrides(baseFile, personalFile string) (TestingMocksConfig, error) {
	// Load base configuration
	baseData, err := os.ReadFile(baseFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, fall back to environment variables
			return loadFromEnvironment(), nil
		}
		return TestingMocksConfig{}, fmt.Errorf("failed to read base config %s: %w", baseFile, err)
	}

	// Parse base configuration with strict mode
	var baseConfig TestingMocksConfig
	decoder := yaml.NewDecoder(strings.NewReader(string(baseData)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&baseConfig); err != nil {
		return TestingMocksConfig{}, fmt.Errorf("failed to parse base config %s: %w", baseFile, err)
	}

	// Check if personal override exists
	if _, err := os.Stat(personalFile); os.IsNotExist(err) {
		// No personal override, validate and return base config
		if err := ValidateTestingMocks(baseConfig); err != nil {
			return TestingMocksConfig{}, fmt.Errorf("validation failed for %s: %w", baseFile, err)
		}
		return baseConfig, nil
	}

	// Load personal override
	personalData, err := os.ReadFile(personalFile)
	if err != nil {
		return TestingMocksConfig{}, fmt.Errorf("failed to read personal override %s: %w", personalFile, err)
	}

	// Parse personal override into base config (merging) with strict mode
	decoder = yaml.NewDecoder(strings.NewReader(string(personalData)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&baseConfig); err != nil {
		return TestingMocksConfig{}, fmt.Errorf("failed to parse personal override %s: %w", personalFile, err)
	}

	// Validate merged config
	if err := ValidateTestingMocks(baseConfig); err != nil {
		return TestingMocksConfig{}, fmt.Errorf("validation failed for merged config: %w", err)
	}

	return baseConfig, nil
}

// loadFromEnvironment creates configuration from environment variables (backward compatibility).
func loadFromEnvironment() TestingMocksConfig {
	return TestingMocksConfig{
		Mocks: MocksConfig{
			AI: AIMockConfig{
				Enabled: parseBool(os.Getenv(environments.EnvCLIEMockAI)),
				MockDir: os.Getenv(environments.EnvCLIEMockAIDir),
			},
			Security: SecurityMockConfig{
				Enabled: parseBool(os.Getenv(environments.EnvCLIEMockSecurity)),
				Tools:   parseTools(os.Getenv(environments.EnvCLIEMockSecurityTools)),
			},
			Docker: DockerMockConfig{
				Enabled: parseBool(os.Getenv(environments.EnvCLIEMockDocker)),
			},
			GitHub: GitHubMockConfig{
				// Support both CLIE_MOCK_GITHUB and CLIE_MOCK_GITHUB_CLI for backward compatibility
				Enabled:     parseBool(os.Getenv(environments.EnvCLIEMockGitHub)) || parseBool(os.Getenv(environments.EnvCLIEMockGitHubCLI)),
				NoWorkflows: parseBool(os.Getenv(environments.EnvCLIEMockNoWorkflows)) || parseBool(os.Getenv(environments.EnvCLIEMockGitHubNoWorkflows)),
			},
		},
	}
}

// parseBool parses a boolean value from a string.
// Accepts: "true", "1", "yes", "y", "on" (case-insensitive) as true.
// Empty string or any other value is false.
func parseBool(s string) bool {
	if s == "" {
		return false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes" || s == "y" || s == "on"
}

// parseTools parses a comma-separated list of tools.
func parseTools(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// validateRequiredFields checks for required fields in the raw YAML before unmarshaling.
// This catches cases where boolean fields are missing (which would be filled with false by YAML unmarshaler).
func validateRequiredFields(data []byte) error {
	// Unmarshal into a generic map to check for field presence
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Check if mocks section exists
	mocks, ok := raw["mocks"].(map[string]interface{})
	if !ok {
		// If mocks doesn't exist or isn't a map, the schema validation will catch it
		return nil
	}

	// Check each mock type for required 'enabled' field
	mockTypes := []string{"ai", "security", "docker", "github"}
	for _, mockType := range mockTypes {
		if mockSection, exists := mocks[mockType]; exists {
			mockMap, ok := mockSection.(map[string]interface{})
			if !ok {
				// If the section exists but isn't a map, schema validation will catch it
				continue
			}
			// If the section exists, 'enabled' field is required
			if _, hasEnabled := mockMap["enabled"]; !hasEnabled {
				return fmt.Errorf("missing required field 'enabled' in mocks.%s", mockType)
			}
		}
	}

	return nil
}

// ValidateTestingMocks validates the configuration against the JSON schema.
func ValidateTestingMocks(config TestingMocksConfig) error {
	// Find schema file relative to this package
	schemaPath, err := findSchemaPath()
	if err != nil {
		return fmt.Errorf("failed to locate schema: %w", err)
	}

	// Load schema
	schemaLoader := gojsonschema.NewReferenceLoader("file:///" + filepath.ToSlash(schemaPath))

	// Convert config to JSON for validation
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		// Collect all validation errors
		var errMsgs []string
		for _, desc := range result.Errors() {
			msg := desc.Description()
			// Reformat uniqueItems errors to include the keyword name
			if strings.Contains(msg, "must be unique") {
				msg = fmt.Sprintf("%s (uniqueItems)", msg)
			}
			errMsgs = append(errMsgs, fmt.Sprintf("- %s: %s", desc.Field(), msg))
		}
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errMsgs, "\n"))
	}

	// Additional validation rules not covered by JSON schema
	if err := validateAdditionalRules(config); err != nil {
		return err
	}

	return nil
}

// validateAdditionalRules performs additional validation beyond JSON schema.
func validateAdditionalRules(config TestingMocksConfig) error {
	// Validate mock_dir is not just whitespace if provided
	if config.Mocks.AI.MockDir != "" {
		if strings.TrimSpace(config.Mocks.AI.MockDir) == "" {
			return fmt.Errorf("ai.mock_dir cannot be whitespace-only")
		}
	}

	// Validate security tools are unique (uniqueItems validation)
	if config.Mocks.Security.Tools != nil {
		seen := make(map[string]bool)
		for i, tool := range config.Mocks.Security.Tools {
			if seen[tool] {
				// Find the first occurrence
				var firstIdx int
				for j, t := range config.Mocks.Security.Tools {
					if t == tool && j < i {
						firstIdx = j
						break
					}
				}
				return fmt.Errorf("configuration validation failed:\n- mocks.security.tools: uniqueItems constraint violated (duplicate at indices %d and %d)", firstIdx, i)
			}
			seen[tool] = true
		}
	}

	return nil
}

// findSchemaPath locates the testing-mocks.schema.json file.
// It searches up from the current package directory to find contracts/core/0.1.0/testing-mocks.schema.json
func findSchemaPath() (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up directory tree looking for contracts/core/0.1.0/testing-mocks.schema.json
	dir := cwd
	for {
		schemaPath := filepath.Join(dir, "contracts", "core", "0.1.0", "schemas", "testing-mocks.schema.json")
		if _, err := os.Stat(schemaPath); err == nil {
			return schemaPath, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding schema
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("testing-mocks.schema.json not found in contracts/core/0.1.0/schemas/")
}

// ToEnvironmentVariables converts the configuration to environment variables.
// This is used by the test infrastructure to set up the test environment.
func (c *TestingMocksConfig) ToEnvironmentVariables() []string {
	var env []string

	if c.Mocks.AI.Enabled {
		env = append(env, "CLIE_MOCK_AI=true")
		if c.Mocks.AI.MockDir != "" {
			env = append(env, fmt.Sprintf("CLIE_MOCK_AI_DIR=%s", c.Mocks.AI.MockDir))
		}
	}

	if c.Mocks.Security.Enabled {
		env = append(env, "CLIE_MOCK_SECURITY=true")
		if len(c.Mocks.Security.Tools) > 0 {
			env = append(env, fmt.Sprintf("CLIE_MOCK_SECURITY_TOOLS=%s", strings.Join(c.Mocks.Security.Tools, ",")))
		}
	}

	if c.Mocks.Docker.Enabled {
		env = append(env, "CLIE_MOCK_DOCKER=true")
		// Legacy support
		env = append(env, "CLIE_MOCK_STRUCTURIZR=true")
	}

	if c.Mocks.GitHub.Enabled {
		env = append(env, "CLIE_MOCK_GITHUB_CLI=true")
		if c.Mocks.GitHub.NoWorkflows {
			env = append(env, "CLIE_MOCK_NO_WORKFLOWS=true")
		}
	}

	return env
}
