// File: src/core/ai/config.go
package ai

// Config represents AI provider configuration loaded from .r2r/agent-config.yml
//
// Intent: Hold provider configuration with environment variable references.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Self-documenting field names
//   - YAML tags match config file structure
//   - No complex nested structures
//
// Easy to change:
//   - Can add new fields without breaking existing configs
//   - Separate from loading logic
//   - No behavior - just data
//
// Hard to break:
//   - Required fields validated during loading
//   - Immutable after creation (no setters)
//   - Clear separation between config and runtime values
type Config struct {
	ProviderName string `yaml:"name"`     // Provider identifier (claude-api, claude-cli, openai, gemini)
	Model        string `yaml:"model"`    // AI model to use
	Endpoint     string `yaml:"endpoint"` // API endpoint URL (empty for claude-cli)
	APIKey       string `yaml:"api_key"`  // API key with ${VAR} substitution (empty for claude-cli)
}
