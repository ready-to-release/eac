// File: go/eac/commands/internal/ai/config.go
package ai

// Config represents AI and Git configuration loaded from .r2r/eac/eac-config.yml
//
// Intent: Hold provider configuration with environment variable references.
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Self-documenting field names
//   - YAML tags match config file structure
//   - Nested structure separates AI and Git configuration
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
	AI  AIConfig  `yaml:"ai"`  // AI provider configuration
	Git GitConfig `yaml:"git"` // Git provider configuration
}

// AIConfig represents AI provider configuration
type AIConfig struct {
	Provider string `yaml:"provider"` // Provider identifier (claude-api, claude-cli, openai, gemini)
	Model    string `yaml:"model"`    // AI model to use
	Endpoint string `yaml:"endpoint"` // API endpoint URL (empty for claude-cli)
	APIKey   string `yaml:"api_key"`  // API key with ${VAR} substitution (empty for claude-cli)
}

// GitConfig represents Git provider configuration
type GitConfig struct {
	Token string `yaml:"token"` // Generic git provider token (GitHub, GitLab, Bitbucket, etc.) with ${VAR} substitution
}
