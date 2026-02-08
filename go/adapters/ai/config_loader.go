// File: go/eac/adapters/ai/config_loader.go
package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/ready-to-release/eac/go/core/domain/schema"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
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

// LoadConfigWithOverrides loads defaults, team config, and personal overrides.
// Merge order: defaults → team → personal (each layer overrides the previous).
// Personal config can override any field: provider, model, api_key, endpoint, git token.
// Validates configs against the ai-provider schema.
//
// Schema (all files use same structure):
//
//	ai:
//	  provider: claude-api          # or claude-cli, openai, gemini
//	  model: claude-3-haiku-20240307
//	  endpoint: https://api.anthropic.com/v1
//	  api_key: ${ANTHROPIC_API_KEY}  # or literal key
//	git:
//	  token: ${GIT_TOKEN}  # or literal token
//
// Team and personal configs only need to specify fields to override.
func LoadConfigWithOverrides(workspaceRoot, teamConfigPath, personalConfigPath string) (*Config, error) {
	// Start with defaults (optional - system still works without them)
	config, _ := loadAIProviderDefaults(workspaceRoot)
	if config == nil {
		config = &Config{}
	}

	// Load and validate team config (optional with defaults, required without)
	if teamConfigPath != "" {
		if _, statErr := os.Stat(teamConfigPath); statErr == nil {
			teamConfig, loadErr := loadConfigFileWithValidation(teamConfigPath, workspaceRoot)
			if loadErr != nil {
				return nil, loadErr
			}
			mergeConfig(config, teamConfig)
		}
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

// loadAIProviderDefaults loads AI provider defaults from contracts folder.
// Container-aware: uses R2R_CONTAINER_ROOT when running in container.
func loadAIProviderDefaults(repoRoot string) (*Config, error) {
	root := repoRoot
	if containerRoot := os.Getenv(environments.EnvR2RContainerRoot); containerRoot != "" {
		root = containerRoot
	}
	if root == "" {
		return nil, fmt.Errorf("no root available for loading defaults")
	}

	fsPath := filepath.Join(root, "contracts", "core", paths.DefaultsVersion, "schemas", "defaults", "ai-provider.yml")
	data, err := os.ReadFile(fsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("reading ai-provider defaults: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing ai-provider defaults: %w", err)
	}

	return &cfg, nil
}

// mergeConfig merges override config into base config.
func mergeConfig(base, override *Config) {
	if override == nil {
		return
	}
	if override.AI.Provider != "" {
		base.AI.Provider = override.AI.Provider
	}
	if override.AI.Model != "" {
		base.AI.Model = override.AI.Model
	}
	if override.AI.APIKey != "" {
		base.AI.APIKey = override.AI.APIKey
	}
	if override.AI.Endpoint != "" {
		base.AI.Endpoint = override.AI.Endpoint
	}
	if override.Git.Token != "" {
		base.Git.Token = override.Git.Token
	}
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
		return nil, fmt.Errorf("failed to parse .eac/ai-provider.yml: %w\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: claude-api, openai, gemini", err)
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
		return fmt.Errorf("missing environment variable(s) for API key: %v\n\nPlease set:\n  export %s=your-api-key\n\nOr use claude-cli provider (no API key needed):\n  Run: .\\importer.ps1 (creates .eac/ai-provider.personal.yml with claude-cli)",
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
