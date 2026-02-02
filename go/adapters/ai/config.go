// File: go/eac/adapters/ai/config.go
package ai

import "errors"

// ErrAIProviderNotConfigured is returned when AI provider configuration is missing.
// Commands that depend on AI should check for this error and display the message.
var ErrAIProviderNotConfigured = errors.New("ai provider not configured, please run eac init --ai-provider to initialize it")

// Config represents AI and Git configuration loaded from .r2r/eac/ai-provider.yml.
type Config struct {
	AI  AIConfig  `yaml:"ai"`  // AI provider configuration
	Git GitConfig `yaml:"git"` // Git provider configuration
}

// AIConfig represents AI provider configuration.
type AIConfig struct {
	Provider string `yaml:"provider"` // Provider identifier (claude-api, claude-cli, openai, gemini)
	Model    string `yaml:"model"`    // AI model to use
	Endpoint string `yaml:"endpoint"` // API endpoint URL (empty for claude-cli)
	APIKey   string `yaml:"api_key"`  // API key with ${VAR} substitution (empty for claude-cli)
}

// GitConfig represents Git provider configuration.
type GitConfig struct {
	Token string `yaml:"token"` // Generic git provider token (GitHub, GitLab, Bitbucket, etc.) with ${VAR} substitution
}
