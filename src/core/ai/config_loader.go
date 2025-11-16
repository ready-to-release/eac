// File: src/core/ai/config_loader.go
package ai

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses agent configuration from agent-config.yml
//
// Intent: Load AI provider configuration from file with environment variable substitution.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single LoadConfig(path string) function with clear signature
//   - Explicit error handling at each step (file read, YAML parse, env var substitution)
//   - substituteEnvVars() is a pure function - no side effects
//   - All transformations are visible and traceable
//   - Clear error messages with recovery instructions
//
// Easy to change:
//   - Config struct is separate from loading logic (can change one without the other)
//   - Environment variable substitution is isolated in its own function
//   - File loading and parsing are separate concerns
//   - Adding new config fields doesn't require changing loader logic
//
// Hard to break:
//   - Tests cover all cases: valid config, missing file, malformed YAML, missing env vars
//   - Early validation: fail fast with clear error if config is invalid
//   - No global state - function is stateless and testable
//   - Errors include actionable instructions to run r2r agent init
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent-config.yml not found at %s\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini", path)
		}
		return nil, fmt.Errorf("failed to read config file %s: %w\n\nPlease run: r2r agent init --ai <provider>", path, err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil, fmt.Errorf("agent-config.yml is empty\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini")
	}

	// Parse YAML
	var rawConfig struct {
		Provider Config `yaml:"provider"`
	}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse agent-config.yml: %w\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini", err)
	}

	// Substitute environment variables
	config := &rawConfig.Provider
	config.APIKey = substituteEnvVars(config.APIKey)
	config.Endpoint = substituteEnvVars(config.Endpoint)

	// Validate required fields
	if config.ProviderName == "" {
		return nil, fmt.Errorf("provider name is required in agent-config.yml\n\nPlease run: r2r agent init --ai <provider>\nSupported providers: claude-cli, claude-api, openai, gemini")
	}

	return config, nil
}

// substituteEnvVars replaces ${VAR_NAME} with environment variable values.
// This is a pure function with no side effects.
//
// Intent: Replace environment variable placeholders with actual values.
//
// Design:
//   - Uses regex for reliable pattern matching
//   - Returns empty string for undefined variables (defensive)
//   - No side effects - doesn't modify environment
func substituteEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// Extract VAR_NAME from ${VAR_NAME}
		varName := match[2 : len(match)-1]
		return os.Getenv(varName)
	})
}
