package init

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// APIKeySource indicates where the API key was found
type APIKeySource int

const (
	APIKeySourceNone APIKeySource = iota
	APIKeySourceEnvVar
	APIKeySourcePersonalConfig
)

// APIKeyResolution holds the result of API key resolution
type APIKeyResolution struct {
	Source       APIKeySource
	Key          string
	EnvVarName   string
	ConfigPath   string
}

// resolveAPIKey resolves the API key from multiple sources with priority:
// 1. Environment variable (highest)
// 2. Personal config file
// 3. ERROR (no interactive prompt - agent friendly)
func resolveAPIKey(workspaceRoot, provider string) (*APIKeyResolution, error) {
	envVarName := getEnvVarNameForProvider(provider)

	// Priority 1: Check environment variable
	if envVarName != "" {
		if key := os.Getenv(envVarName); key != "" {
			return &APIKeyResolution{
				Source:     APIKeySourceEnvVar,
				Key:        key,
				EnvVarName: envVarName,
			}, nil
		}
	}

	// Priority 2: Check personal config
	eacDir := filepath.Join(workspaceRoot, ".eac")
	personalConfigPath := filepath.Join(eacDir, "ai-provider.personal.yml")

	if fileExists(personalConfigPath) {
		key, err := loadAPIKeyFromConfig(personalConfigPath)
		if err == nil && key != "" {
			return &APIKeyResolution{
				Source:     APIKeySourcePersonalConfig,
				Key:        key,
				ConfigPath: personalConfigPath,
			}, nil
		}
	}

	// Priority 3: No key found - return error with guidance
	return nil, formatAPIKeyNotFoundError(provider, envVarName, personalConfigPath)
}

// getEnvVarNameForProvider returns the environment variable name for a provider
func getEnvVarNameForProvider(provider string) string {
	switch provider {
	case "claude-api":
		return "ANTHROPIC_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GOOGLE_API_KEY"
	default:
		return ""
	}
}

// loadAPIKeyFromConfig loads the API key from a YAML config file
func loadAPIKeyFromConfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var config struct {
		AI struct {
			APIKey string `yaml:"api_key"`
		} `yaml:"ai"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.AI.APIKey, nil
}

// formatAPIKeyNotFoundError creates a clear error message with guidance
// No interactive prompts - this must be agent/automation friendly
func formatAPIKeyNotFoundError(provider, envVarName, personalConfigPath string) error {
	return fmt.Errorf(`API key not found for provider '%s'

To fix this, choose one of the following options:

Option 1: Set environment variable (recommended for CI/CD and agents)
  export %s=your-key-here

Option 2: Create personal config file (recommended for local development)
  Create: %s

  Example config:
    ai:
      provider: %s
      api_key: your-key-here

Then re-run: eac init --ai-provider %s`,
		provider, envVarName, personalConfigPath, provider, provider)
}
