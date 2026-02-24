package ai

// ProviderConfig represents the complete AI provider configuration
// loaded from .eac/ai-provider.yml.
type ProviderConfig struct {
	AI  ConnectionConfig `yaml:"ai"`  // AI provider configuration
	Git GitConfig        `yaml:"git"` // Git provider configuration
}

// ConnectionConfig holds AI provider connection parameters.
// Renamed from AIConfig to avoid confusion with go/core/ai/config.AIConfig.
type ConnectionConfig struct {
	Provider string `yaml:"provider"` // Provider identifier (claude-api, claude-cli, openai, gemini)
	Model    string `yaml:"model"`    // AI model to use
	Endpoint string `yaml:"endpoint"` // API endpoint URL (empty for claude-cli)
	APIKey   string `yaml:"api_key"`  // API key with ${VAR} substitution (empty for claude-cli)
}

// GitConfig holds Git provider settings.
type GitConfig struct {
	Token string `yaml:"token"` // Generic git provider token with ${VAR} substitution
}
