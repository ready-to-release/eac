// File: go/eac/commands/internal/ai/config_loader.go
package ai

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/schema"
	"gopkg.in/yaml.v3"
)

// Schema validator (lazy initialized).
var (
	schemaValidator     *schema.Validator
	schemaValidatorOnce sync.Once
	schemaValidatorErr  error
)

// LoadConfig loads and parses agent configuration from a single file.
// For full team+personal merge behavior, use LoadConfigWithOverrides.
func LoadConfig(path string) (*Config, error) {
	config, err := loadConfigFile(path)
	if err != nil {
		return nil, err
	}

	// Apply environment variable substitution
	if err := applyEnvVarSubstitution(config); err != nil {
		return nil, err
	}

	// Validate required fields
	if config.AI.Provider == "" {
		return nil, ErrAIProviderNotConfigured
	}

	return config, nil
}

// LoadConfigWithOverrides loads team config and merges with personal overrides.
// Personal config can override any field: provider, model, api_key, endpoint, git token.
// Validates both configs against the ai-provider schema.
//
// Schema (both files use same structure):
//
//	ai:
//	  provider: claude-api          # or claude-cli, openai, gemini
//	  model: claude-3-haiku-20240307
//	  endpoint: https://api.anthropic.com/v1
//	  api_key: ${ANTHROPIC_API_KEY}  # or literal key
//	git:
//	  token: ${GIT_TOKEN}  # or literal token
//
// Personal config only needs to specify fields to override.
func LoadConfigWithOverrides(workspaceRoot, teamConfigPath, personalConfigPath string) (*Config, error) {
	// Load and validate team config (required)
	config, err := loadConfigFileWithValidation(teamConfigPath, workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Load and validate personal config (optional) and merge
	if personalConfigPath != "" {
		if _, statErr := os.Stat(personalConfigPath); statErr == nil {
			personalConfig, loadErr := loadPersonalConfigWithValidation(personalConfigPath, workspaceRoot)
			if loadErr != nil {
				return nil, fmt.Errorf("failed to load personal config: %w", loadErr)
			}
			mergePersonalConfig(config, personalConfig)
		}
	}

	// Apply environment variable substitution
	if err := applyEnvVarSubstitution(config); err != nil {
		return nil, err
	}

	// Validate required fields
	if config.AI.Provider == "" {
		return nil, ErrAIProviderNotConfigured
	}

	return config, nil
}

// PersonalConfig represents the personal override file structure.
// Uses same nested structure as team config for consistency.
type PersonalConfig struct {
	AI  *AIConfig  `yaml:"ai"`
	Git *GitConfig `yaml:"git"`
}

// loadConfigFile loads and parses a config file without env var substitution.
func loadConfigFile(path string) (*Config, error) {
	return loadConfigFileWithValidation(path, "")
}

// loadConfigFileWithValidation loads, validates, and parses a config file.
func loadConfigFileWithValidation(path, workspaceRoot string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAIProviderNotConfigured
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if len(data) == 0 {
		return nil, ErrAIProviderNotConfigured
	}

	// Validate against schema if workspaceRoot is provided
	if workspaceRoot != "" {
		if err := validateAgentConfigSchema(workspaceRoot, data); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse .r2r/eac/ai-provider.yml: %w\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: claude-api, openai, gemini", err)
	}

	return &config, nil
}

// validateAgentConfigSchema validates config data against the ai-provider schema.
func validateAgentConfigSchema(workspaceRoot string, data []byte) error {
	schemaValidatorOnce.Do(func() {
		schemaValidator, schemaValidatorErr = schema.NewValidator(workspaceRoot)
	})

	if schemaValidatorErr != nil {
		// Schema validation is optional - don't fail if schemas can't be loaded
		return nil
	}

	return schemaValidator.ValidateYAML(schema.SchemaEACConfig, data)
}

// loadPersonalConfig loads the personal override file.
func loadPersonalConfig(path string) (*PersonalConfig, error) {
	return loadPersonalConfigWithValidation(path, "")
}

// loadPersonalConfigWithValidation loads and validates the personal override file.
func loadPersonalConfigWithValidation(path, workspaceRoot string) (*PersonalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return &PersonalConfig{}, nil
	}

	// Validate against schema if workspaceRoot is provided
	if workspaceRoot != "" {
		if err := validateAgentConfigSchema(workspaceRoot, data); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	var personal PersonalConfig
	if err := yaml.Unmarshal(data, &personal); err != nil {
		return nil, fmt.Errorf("failed to parse personal config: %w", err)
	}

	return &personal, nil
}

// mergePersonalConfig applies personal overrides to the base config.
func mergePersonalConfig(base *Config, personal *PersonalConfig) {
	// Merge AI config
	if personal.AI != nil {
		if personal.AI.Provider != "" {
			base.AI.Provider = personal.AI.Provider
		}
		if personal.AI.Model != "" {
			base.AI.Model = personal.AI.Model
		}
		if personal.AI.APIKey != "" {
			base.AI.APIKey = personal.AI.APIKey
		}
		if personal.AI.Endpoint != "" {
			base.AI.Endpoint = personal.AI.Endpoint
		}
	}

	// Merge Git config
	if personal.Git != nil {
		if personal.Git.Token != "" {
			base.Git.Token = personal.Git.Token
		}
	}
}

// applyEnvVarSubstitution substitutes environment variables in config values.
func applyEnvVarSubstitution(config *Config) error {
	var missingVars []string

	// Substitute AI config variables
	config.AI.APIKey, missingVars = substituteEnvVars(config.AI.APIKey)
	// Only error on missing env vars if provider requires an API key
	// claude-cli doesn't need an API key (uses local Claude installation)
	if len(missingVars) > 0 && config.AI.Provider != "claude-cli" {
		return fmt.Errorf("missing environment variable(s) for API key: %v\n\nPlease set:\n  export %s=your-api-key\n\nOr use claude-cli provider (no API key needed):\n  Run: .\\importer.ps1 (creates .r2r/eac/ai-provider.personal.yml with claude-cli)",
			missingVars, missingVars[0])
	}

	config.AI.Endpoint, _ = substituteEnvVars(config.AI.Endpoint)

	// Substitute Git config variables
	config.Git.Token, _ = substituteEnvVars(config.Git.Token)

	return nil
}

// substituteEnvVars replaces ${VAR_NAME} with environment variable values.
// Returns the substituted string and any variable names that were not found.
//
// Intent: Replace environment variable placeholders with actual values.
//
// Design:
//   - Uses regex for reliable pattern matching
//   - Returns empty string for undefined variables (defensive)
//   - Reports missing variables for error handling
//   - No side effects - doesn't modify environment
func substituteEnvVars(s string) (string, []string) {
	var missing []string
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract VAR_NAME from ${VAR_NAME}
		varName := match[2 : len(match)-1]
		value := os.Getenv(varName)
		if value == "" {
			missing = append(missing, varName)
		}
		return value
	})
	return result, missing
}
