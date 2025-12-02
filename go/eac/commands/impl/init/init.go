// Command: init
// Short: Initialize AI provider configuration for the project
// Long: Initialize AI provider configuration for the project.
// Long:
// Long: Creates .r2r/eac/agent-config.yml with the specified AI provider settings.
// Long: The configuration file is safe to commit as it only contains environment variable references.
// Long:
// Long: Available providers:
// Long:   - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)
// Long:   - openai: OpenAI via API (requires OPENAI_API_KEY)
// Long:   - gemini: Google Gemini via API (requires GOOGLE_API_KEY)
// Long:
// Long: Example:
// Long:   init --ai claude-api
// Long:   init --ai openai
// Long:   init --ai claude-api --debug
// Flag.ai: type=string, shorthand=a, usage=AI provider to configure, required=true, completion=claude-api,openai,gemini
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs to the 'out' directory for troubleshooting and analysis
package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(Init)
}

var log = logging.C()

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := git.Open(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the repository for test cleanup.
func ResetGitRepo() {
	gitRepo = nil
}

// Intent: Initialize AI provider configuration for a project
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear sequential flow: parse args → validate → create dirs → write config
//   - Helper functions with single responsibility (configureProvider, writeAgentConfig, etc.)
//   - User feedback at each step via logger
//   - Error messages indicate what went wrong and how to fix it
//
// Easy to change:
//   - Provider configuration isolated in configureProvider()
//   - Directory creation isolated in createDirectoryStructure()
//   - Config writing isolated in writeAgentConfig()
//   - Adding new providers only requires updating configureProvider()
//
// Hard to break:
//   - Tests cover all providers and error cases
//   - Validation happens early (--ai flag required, provider supported)
//   - Creates directories safely (no errors if already exist)
//   - Config file is YAML - human readable and safe to commit
//   - Logger integration for consistent output and debugging

// Init initializes AI provider configuration
func Init() int {
	// Parse flags early to get debug mode and AI provider
	aiProvider := ""
	debug := false
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--ai" && i+1 < len(os.Args) {
			aiProvider = os.Args[i+1]
			i++ // Skip the value
		} else if os.Args[i] == "--debug" || os.Args[i] == "-d" {
			debug = true
		}
	}

	// Validate required flag
	if aiProvider == "" {
		log.Errorf("--ai flag is required")
		log.Info("Usage: init --ai <provider>")
		log.Info("")
		log.Info("Available providers: claude-api, openai, gemini")
		log.Info("")
		log.Info("Example:")
		log.Info("  init --ai claude-api")
		log.Info("  init --ai openai")
		log.Info("  init --ai claude-api --debug   # With debug logging")
		return 1
	}

	// Get workspace root via repository API
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Initialize logger early so all code paths can use it
	var logger *logging.Logger
	if debug {
		logger, err = logging.NewWithDebug("init", workspaceRoot)
	} else {
		logger, err = logging.NewDefault("init", workspaceRoot)
	}
	if err != nil {
		log.Errorf("Error initializing logger: %v", err)
		return 1
	}
	defer logger.Sync()

	// Define paths
	eacDir := filepath.Join(workspaceRoot, ".r2r", "eac")
	configPath := filepath.Join(eacDir, "agent-config.yml")

	// Check if already initialized
	if _, err := os.Stat(configPath); err == nil {
		logger.Warn("Project already initialized")
		logger.Info(fmt.Sprintf("Config exists: %s", configPath))
		logger.Info("")
		logger.Info("Reconfiguring agent configuration...")
		logger.Info("")
	}

	logger.Info("🤖 Initialize Agent Configuration")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("")

	// Create .r2r/eac directory structure
	logger.Info("📁 Creating directory structure...")
	if err := createDirectoryStructure(workspaceRoot); err != nil {
		logger.Error(fmt.Sprintf("Error creating directory structure: %v", err))
		return 1
	}
	logger.Info("✅ Directory structure created")

	// Configure agent using --ai flag
	config, err := configureAgent(aiProvider, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error during configuration: %v", err))
		return 1
	}

	// Write configuration
	if err := writeAgentConfig(configPath, config); err != nil {
		logger.Error(fmt.Sprintf("Error writing configuration: %v", err))
		return 1
	}

	// Success message
	logger.Info("")
	logger.Info("✅ Configuration saved")
	logger.Info(fmt.Sprintf("   File: %s", configPath))
	logger.Info("")
	logger.Info("ℹ️  Next steps:")
	logger.Info("   1. Commit the config file (safe - contains no secrets)")
	if config.envVarName != "" {
		logger.Info(fmt.Sprintf("   2. Set environment variable: %s", config.envVarName))
	}
	logger.Info("   3. Run AI-powered commands (e.g., specs create, commit)")
	logger.Info("")

	return 0
}

// agentConfig holds configuration for an AI provider
type agentConfig struct {
	providerName string // "claude-api", "claude-cli", "openai", "gemini"
	envVarName   string // "ANTHROPIC_API_KEY", etc. (empty for claude-cli)
	model        string // "claude-3-haiku-20240307", etc.
	endpoint     string // API endpoint URL (empty for claude-cli)
}

// configureAgent configures the AI provider based on user input
func configureAgent(aiProvider string, logger *logging.Logger) (*agentConfig, error) {
	config := &agentConfig{}

	// Configure provider based on --ai flag
	if err := configureProvider(config, aiProvider); err != nil {
		return nil, err
	}

	displayProviderInfo(config, logger)
	return config, nil
}

// configureProvider sets up the config based on the provider key
func configureProvider(config *agentConfig, provider string) error {
	switch strings.ToLower(provider) {
	case "claude-api":
		config.providerName = "claude-api"
		config.envVarName = "ANTHROPIC_API_KEY"
		config.model = providers.DefaultClaudeAPIModel
		config.endpoint = "https://api.anthropic.com/v1"

	case "openai":
		config.providerName = "openai"
		config.envVarName = "OPENAI_API_KEY"
		config.model = providers.DefaultOpenAIModel
		config.endpoint = "https://api.openai.com/v1"

	case "gemini":
		config.providerName = "gemini"
		config.envVarName = "GOOGLE_API_KEY"
		config.model = providers.DefaultGeminiModel
		config.endpoint = "https://generativelanguage.googleapis.com"

	default:
		return fmt.Errorf("unsupported provider: %s\nSupported: claude-api, openai, gemini", provider)
	}

	return nil
}

// displayProviderInfo shows information about the selected provider
func displayProviderInfo(config *agentConfig, logger *logging.Logger) {
	logger.Info("")
	logger.Info(fmt.Sprintf("✓ %s selected", config.providerName))
	if config.envVarName != "" {
		logger.Info(fmt.Sprintf("  Environment variable: %s", config.envVarName))
	}

	// Provider-specific API key instructions
	switch config.providerName {
	case "claude-api":
		logger.Info("  Get your API key at: https://claude.ai/settings/api")
		logger.Info("  Note: Personal or workspace-owned API keys both work")
		logger.Info("  Requires: ANTHROPIC_API_KEY environment variable")
	case "openai":
		logger.Info("  Get your API key at: https://platform.openai.com/api-keys")
	case "gemini":
		logger.Info("  Get your API key at: https://makersuite.google.com/app/apikey")
	}

	logger.Info("")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Info("")
}

// createDirectoryStructure creates the .r2r/eac directory structure
func createDirectoryStructure(workspaceRoot string) error {
	// Create .r2r/eac directory
	eacDir := filepath.Join(workspaceRoot, ".r2r", "eac")
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	return nil
}

// writeAgentConfig writes the agent configuration to .r2r/eac/agent-config.yml
func writeAgentConfig(configPath string, config *agentConfig) error {
	var content strings.Builder

	content.WriteString("# Team AI Configuration (Production/Shared)\n")
	content.WriteString("# Generated by: init command\n")
	content.WriteString("# SAFE TO COMMIT: Contains only environment variable references, not actual secrets\n")
	content.WriteString("#\n")
	content.WriteString("# This file is committed and shared across the team.\n")
	content.WriteString("# Used by r2r CLI, CI/CD, and production workflows.\n")
	content.WriteString("#\n")
	content.WriteString("# For local development with Claude CLI (no API costs):\n")
	content.WriteString("# - Run: .\\importer.ps1\n")
	content.WriteString("# - This creates: .r2r/eac/agent-config.personal.yml (gitignored)\n")
	content.WriteString("# - Personal config takes precedence over this team config\n")
	content.WriteString("\n")
	content.WriteString("provider:\n")
	content.WriteString(fmt.Sprintf("  name: %s\n", config.providerName))
	content.WriteString(fmt.Sprintf("  model: %s\n", config.model))

	// Only add endpoint if not empty
	if config.endpoint != "" {
		content.WriteString(fmt.Sprintf("  endpoint: %s\n", config.endpoint))
	}

	// Only add api_key if envVarName is not empty
	if config.envVarName != "" {
		content.WriteString(fmt.Sprintf("  api_key: ${%s}\n", config.envVarName))
	}

	// Write to file
	if err := os.WriteFile(configPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
