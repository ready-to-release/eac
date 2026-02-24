package aiproviders

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/schema"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/paths"
	"gopkg.in/yaml.v3"
)

var reEnvVarSubstitution = regexp.MustCompile(`\$\{([^}]+)\}`)

// PersonalConfig represents the personal override file structure.
type PersonalConfig struct {
	AI  *ai.ConnectionConfig `yaml:"ai"`
	Git *ai.GitConfig        `yaml:"git"`
}

// LoadConfig loads and parses agent configuration from a single file.
func LoadConfig(path string) (*ai.ProviderConfig, error) {
	config, err := loadConfigFile(path)
	if err != nil {
		return nil, err
	}

	if err := applyEnvVarSubstitution(config); err != nil {
		return nil, err
	}

	if config.AI.Provider == "" {
		return nil, ai.ErrProviderNotConfigured
	}

	return config, nil
}

// LoadConfigWithOverrides loads defaults, team config, and personal overrides.
// Merge order: defaults -> team -> personal (each layer overrides the previous).
func LoadConfigWithOverrides(workspaceRoot, teamConfigPath, personalConfigPath string) (*ai.ProviderConfig, error) {
	config, _ := loadAIProviderDefaults(workspaceRoot)
	if config == nil {
		config = &ai.ProviderConfig{}
	}

	if teamConfigPath != "" {
		if _, statErr := os.Stat(teamConfigPath); statErr == nil {
			teamConfig, loadErr := loadConfigFileWithValidation(teamConfigPath, workspaceRoot)
			if loadErr != nil {
				return nil, loadErr
			}
			mergeConfig(config, teamConfig)
		}
	}

	if personalConfigPath != "" {
		if _, statErr := os.Stat(personalConfigPath); statErr == nil {
			personalConfig, loadErr := loadPersonalConfigWithValidation(personalConfigPath, workspaceRoot)
			if loadErr != nil {
				return nil, fmt.Errorf("failed to load personal config: %w", loadErr)
			}
			mergePersonalConfig(config, personalConfig)
		}
	}

	if err := applyEnvVarSubstitution(config); err != nil {
		return nil, err
	}

	if config.AI.Provider == "" {
		return nil, ai.ErrProviderNotConfigured
	}

	return config, nil
}

func loadAIProviderDefaults(repoRoot string) (*ai.ProviderConfig, error) {
	root := repoRoot
	if containerRoot := os.Getenv(environments.EnvCLIEContainerRoot); containerRoot != "" {
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

	var cfg ai.ProviderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing ai-provider defaults: %w", err)
	}

	return &cfg, nil
}

func mergeConfig(base, override *ai.ProviderConfig) {
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

func loadConfigFile(path string) (*ai.ProviderConfig, error) {
	return loadConfigFileWithValidation(path, "")
}

func loadConfigFileWithValidation(path, workspaceRoot string) (*ai.ProviderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ai.ErrProviderNotConfigured
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if len(data) == 0 {
		return nil, ai.ErrProviderNotConfigured
	}

	if workspaceRoot != "" {
		if err := validateAgentConfigSchema(workspaceRoot, data); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	var config ai.ProviderConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse .eac/ai-provider.yml: %w\n\nPlease run: eac init --ai-provider <provider>\nSupported providers: claude-api, openai, gemini", err)
	}

	return &config, nil
}

func validateAgentConfigSchema(workspaceRoot string, data []byte) error {
	v, err := schema.NewValidator(workspaceRoot)
	if err != nil {
		return nil
	}
	return v.ValidateYAML(schema.SchemaAIProvider, data)
}

func loadPersonalConfigWithValidation(path, workspaceRoot string) (*PersonalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return &PersonalConfig{}, nil
	}

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

func mergePersonalConfig(base *ai.ProviderConfig, personal *PersonalConfig) {
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

	if personal.Git != nil {
		if personal.Git.Token != "" {
			base.Git.Token = personal.Git.Token
		}
	}
}

func applyEnvVarSubstitution(config *ai.ProviderConfig) error {
	var missingVars []string

	config.AI.APIKey, missingVars = substituteEnvVars(config.AI.APIKey)
	if len(missingVars) > 0 && config.AI.Provider != "claude-cli" {
		return fmt.Errorf("missing environment variable(s) for API key: %v\n\nPlease set:\n  export %s=your-api-key\n\nOr use claude-cli provider (no API key needed):\n  Run: .\\importer.ps1 (creates .eac/ai-provider.personal.yml with claude-cli)",
			missingVars, missingVars[0])
	}

	config.AI.Endpoint, _ = substituteEnvVars(config.AI.Endpoint)
	config.Git.Token, _ = substituteEnvVars(config.Git.Token)

	return nil
}

func substituteEnvVars(s string) (string, []string) {
	var missing []string
	re := reEnvVarSubstitution
	result := re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1]
		value := os.Getenv(varName)
		if value == "" {
			missing = append(missing, varName)
		}
		return value
	})
	return result, missing
}
