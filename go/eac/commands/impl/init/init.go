// Command: init
// Short: Initialize EAC project configuration
// Long: Initialize EAC project configuration.
// Long:
// Long: Creates the .r2r/eac directory structure for repository configuration.
// Long: AI provider configuration is optional - you can configure it later.
// Long:
// Long: When --ai-provider is specified, creates:
// Long:   - .r2r/eac/ai-provider.yml (team config) or
// Long:   - .r2r/eac/ai-provider.personal.yml (personal config with tokens)
// Long:
// Long: Available AI providers:
// Long:   - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)
// Long:   - openai: OpenAI via API (requires OPENAI_API_KEY)
// Long:   - gemini: Google Gemini via API (requires GOOGLE_API_KEY)
// Long:
// Long: Examples:
// Long:   init                                                      # Initialize project (no AI config)
// Long:   init --ai-provider claude-api                             # Configure AI provider
// Long:   init --ai-provider claude-api --ai-token sk-ant-xxx       # Configure with actual token
// Long:   init --copy-templates                                     # Copy system config files for customization
// Long:   init --ai-provider claude-api --force                     # Overwrite existing config
// Flag.ai-provider: type=string, shorthand=a, usage=AI provider to configure (optional), required=false, completion=claude-api,openai,gemini
// Flag.ai-token: type=string, usage=AI provider API token (creates personal config if provided), required=false
// Flag.git-token: type=string, usage=Git provider API token for repository operations (supports GitHub, GitLab, etc.) (optional), required=false
// Flag.copy-templates: type=bool, default=false, usage=Copy system default configuration files to repository for customization, required=false
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing config file if it exists, required=false
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs to the 'out' directory for troubleshooting and analysis, required=false
package init

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(Init)
}

var log = logging.C()

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository
var gitMgr *git.RepositoryManager

// initGitManager initializes the git repository manager if needed.
func initGitManager() {
	if gitMgr == nil {
		gitMgr = git.NewManager(logging.C().Zap())
	}
}

// getGitRepo returns the git repository, initializing it if needed.
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	repo, err := gitMgr.Open(workspaceRoot)
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
	gitMgr = nil
}

// Init initializes EAC project configuration
func Init() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags to get all options
	aiProvider := ""
	aiToken := ""
	gitToken := ""
	copyTemplates := false
	force := false
	debug := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--ai-provider", "-a":
			if i+1 < len(os.Args) {
				aiProvider = os.Args[i+1]
				i++ // Skip the value
			}
		case "--ai-token":
			if i+1 < len(os.Args) {
				aiToken = os.Args[i+1]
				i++ // Skip the value
			}
		case "--git-token":
			if i+1 < len(os.Args) {
				gitToken = os.Args[i+1]
				i++ // Skip the value
			}
		case "--copy-templates":
			copyTemplates = true
		case "--force":
			force = true
		case "--debug", "-d":
			debug = true
		}
	}

	// Get workspace root via repository API
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Create .r2r/eac directory structure (always)
	log.Info("📁 Initializing EAC project...")
	log.Info(fmt.Sprintf("   Repository root: %s", workspaceRoot))
	log.Info("")
	if err := createDirectoryStructure(workspaceRoot); err != nil {
		log.Error(fmt.Sprintf("Error creating directory structure: %v", err))
		return 1
	}
	log.Info("✅ Directory structure created")

	// Copy system templates if requested
	if copyTemplates {
		log.Info("")
		log.Info("📄 Copying system template files...")
		if err := copySystemTemplates(workspaceRoot, force); err != nil {
			log.Error(fmt.Sprintf("Error copying templates: %v", err))
			return 1
		}
		log.Info("✅ System templates copied")
	}

	// If no AI provider specified, just initialize directory structure
	if aiProvider == "" {
		log.Info("")
		log.Info("✅ EAC project initialized")
		log.Info("")
		log.Info("ℹ️  Next steps:")
		log.Info("   To configure AI provider (optional):")
		log.Info("     eac init --ai-provider claude-api")
		log.Info("     eac init --ai-provider openai")
		log.Info("     eac init --ai-provider gemini")
		log.Info("")
		return 0
	}

	// AI provider specified - configure it
	log.Info("")
	log.Info("🤖 Configuring AI Provider")
	log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Define paths
	teamConfigPath := paths.EACConfigFilePath(workspaceRoot)
	personalConfigPath := paths.EACConfigPersonalFilePath(workspaceRoot)

	// Check if config already exists (only when configuring AI)
	if err := checkExistingConfig(teamConfigPath, personalConfigPath, force); err != nil {
		log.Error(fmt.Sprintf("%v", err))
		return 1
	}

	// Configure agent using --ai-provider flag
	config, err := configureAgent(aiProvider)
	if err != nil {
		log.Errorf("Error during configuration: %v", err)
		return 1
	}

	// Create tokens struct
	tokens := &tokenConfig{
		aiToken:  aiToken,
		gitToken: gitToken,
	}

	// Write configuration (team or personal based on token presence)
	configPath, err := writeConfig(workspaceRoot, config, tokens)
	if err != nil {
		log.Error(fmt.Sprintf("Error writing configuration: %v", err))
		return 1
	}

	// Success message
	log.Info("")
	log.Info("✅ AI provider configured")
	log.Info(fmt.Sprintf("   File: %s", configPath))
	log.Info("")

	// Provide appropriate next steps based on config type
	if aiToken != "" {
		// Personal config with tokens
		log.Info("ℹ️  Next steps:")
		log.Info("   1. Do NOT commit this file (contains actual tokens)")
		log.Info("   2. Run AI-powered commands (e.g., specs create, commit)")
	} else {
		// Team config with placeholders
		log.Info("ℹ️  Next steps:")
		log.Info("   1. Commit the config file (safe - contains no secrets)")
		if config.envVarName != "" {
			log.Info(fmt.Sprintf("   2. Set environment variable: %s", config.envVarName))
		}
		log.Info("   3. Run AI-powered commands (e.g., specs create, commit)")
	}
	log.Info("")

	return 0
}

// agentConfig holds configuration for an AI provider
type agentConfig struct {
	providerName string // "claude-api", "claude-cli", "openai", "gemini"
	envVarName   string // "ANTHROPIC_API_KEY", etc. (empty for claude-cli)
	model        string // "claude-3-haiku-20240307", etc.
	endpoint     string // API endpoint URL (empty for claude-cli)
}

// tokenConfig holds actual token values (for personal config)
type tokenConfig struct {
	aiToken  string // Actual AI API token
	gitToken string // Actual Git API token
}

// checkExistingConfig checks if config files exist and validates force flag
func checkExistingConfig(teamPath, personalPath string, force bool) error {
	teamExists := fileExists(teamPath)
	personalExists := fileExists(personalPath)

	if !teamExists && !personalExists {
		// No config exists, proceed
		return nil
	}

	// Config exists
	if !force {
		log.Warn("Configuration already exists")
		log.Info("")
		if teamExists {
			log.Info(fmt.Sprintf("  Team config: %s", teamPath))
		}
		if personalExists {
			log.Info(fmt.Sprintf("  Personal config: %s", personalPath))
		}
		log.Info("")
		log.Info("Use --force to overwrite existing config")
		return fmt.Errorf("config already exists (use --force to overwrite)")
	}

	// Force flag provided, log that we're overwriting
	log.Info("⚠️  Overwriting existing configuration (--force)")
	if teamExists {
		log.Info(fmt.Sprintf("  Will replace: %s", teamPath))
	}
	if personalExists {
		log.Info(fmt.Sprintf("  Will replace: %s", personalPath))
	}
	log.Info("")

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// configureAgent configures the AI provider based on user input
func configureAgent(aiProvider string) (*agentConfig, error) {
	config := &agentConfig{}

	// Configure provider based on --ai flag
	if err := configureProvider(config, aiProvider); err != nil {
		return nil, err
	}

	displayProviderInfo(config)
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
func displayProviderInfo(config *agentConfig) {
	log.Info("")
	log.Info(fmt.Sprintf("✓ %s selected", config.providerName))
	if config.envVarName != "" {
		log.Info(fmt.Sprintf("  Environment variable: %s", config.envVarName))
	}

	// Provider-specific API key instructions
	switch config.providerName {
	case "claude-api":
		log.Info("  Get your API key at: https://claude.ai/settings/api")
		log.Info("  Note: Personal or workspace-owned API keys both work")
		log.Info("  Requires: ANTHROPIC_API_KEY environment variable")
	case "claude-cli":
		log.Info("  Uses Claude Code CLI (no API key needed)")
		log.Info("  Note: Requires Claude Code to be installed and authenticated")
	case "openai":
		log.Info("  Get your API key at: https://platform.openai.com/api-keys")
	case "gemini":
		log.Info("  Get your API key at: https://makersuite.google.com/app/apikey")
	}

	log.Info("")
	log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Info("")
}

// createDirectoryStructure creates the .r2r/eac directory structure
func createDirectoryStructure(workspaceRoot string) error {
	// Create .r2r/eac directory
	eacDir := paths.EACConfigPath(workspaceRoot)
	log.Info(fmt.Sprintf("   Creating directory: %s", eacDir))

	if err := os.MkdirAll(eacDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	// Verify directory was created
	if _, err := os.Stat(eacDir); os.IsNotExist(err) {
		return fmt.Errorf("directory creation reported success but directory does not exist: %s", eacDir)
	}

	return nil
}

// writeConfig writes the EAC configuration (team or personal based on tokens)
func writeConfig(workspaceRoot string, config *agentConfig, tokens *tokenConfig) (string, error) {
	// Determine which file to write and whether to use env vars or direct tokens
	var configPath string
	var useEnvVars bool

	if tokens.aiToken != "" {
		// User provided AI token - write personal config with direct values
		configPath = paths.EACConfigPersonalFilePath(workspaceRoot)
		useEnvVars = false
		log.Info("📝 Creating personal configuration with actual tokens...")
	} else {
		// No tokens provided - write team config with env var placeholders
		configPath = paths.EACConfigFilePath(workspaceRoot)
		useEnvVars = true
		log.Info("📝 Creating team configuration with environment variable placeholders...")
	}

	// Build config content
	content := buildConfigContent(config, tokens, useEnvVars)

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// buildConfigContent builds the YAML config content
func buildConfigContent(config *agentConfig, tokens *tokenConfig, useEnvVars bool) string {
	var content strings.Builder

	// Header comment
	if useEnvVars {
		content.WriteString("# EAC Configuration (Team/Shared)\n")
		content.WriteString("# Generated by: eac init command\n")
		content.WriteString("# SAFE TO COMMIT: Contains environment variable placeholders\n")
		content.WriteString("#\n")
		content.WriteString("# This file is committed and shared across the team.\n")
		content.WriteString("# Used by eac CLI, CI/CD, and production workflows.\n")
		content.WriteString("#\n")
		content.WriteString("# For local development with Claude CLI (no API costs):\n")
		content.WriteString("# - Run: .\\importer.ps1\n")
		content.WriteString("# - This creates: .r2r/eac/ai-provider.personal.yml (gitignored)\n")
		content.WriteString("# - Personal config takes precedence over this team config\n")
	} else {
		content.WriteString("# EAC Configuration (Personal)\n")
		content.WriteString("# Generated by: eac init command\n")
		content.WriteString("# GITIGNORED: Contains actual token values\n")
		content.WriteString("# DO NOT COMMIT THIS FILE\n")
		content.WriteString("#\n")
		content.WriteString("# This file is specific to your machine and contains actual API tokens.\n")
		content.WriteString("# You can manually change tokens to ${ENV_VAR} format if preferred.\n")
	}
	content.WriteString("\n")

	// AI configuration
	content.WriteString("ai:\n")
	content.WriteString(fmt.Sprintf("  provider: %s\n", config.providerName))
	content.WriteString(fmt.Sprintf("  model: %s\n", config.model))

	if config.endpoint != "" {
		content.WriteString(fmt.Sprintf("  endpoint: %s\n", config.endpoint))
	}

	if config.envVarName != "" {
		if useEnvVars {
			content.WriteString(fmt.Sprintf("  api_key: ${%s}\n", config.envVarName))
		} else {
			content.WriteString(fmt.Sprintf("  api_key: %s\n", tokens.aiToken))
		}
	}

	// Git configuration
	content.WriteString("\ngit:\n")
	if useEnvVars {
		content.WriteString("  token: ${GIT_TOKEN}\n")
	} else if tokens.gitToken != "" {
		content.WriteString(fmt.Sprintf("  token: %s\n", tokens.gitToken))
	} else {
		content.WriteString("  token: \"\"\n")
	}

	return content.String()
}

// copySystemTemplates copies system default configuration files to user repository
func copySystemTemplates(workspaceRoot string, force bool) error {
	// Get system root (Docker container or local dev)
	systemRoot := os.Getenv("R2R_CONTAINER_ROOT")
	if systemRoot == "" {
		// Local dev mode - system defaults are in workspace root
		systemRoot = workspaceRoot
	}

	// List of Category B files (system defaults) to copy
	templateFiles := []string{
		"ai-config.yml",
		"module-types.yml",
		"system-dependencies.yml",
		"security-tools.yml",
		"logging.yml",
		"environments.yml",
		"testing-tags.yml",
	}

	eacDir := paths.EACConfigPath(workspaceRoot)
	systemEacDir := paths.EACConfigPath(systemRoot)

	copiedCount := 0
	skippedCount := 0

	for _, filename := range templateFiles {
		srcPath := fmt.Sprintf("%s/%s", systemEacDir, filename)
		dstPath := fmt.Sprintf("%s/%s", eacDir, filename)

		// Check if destination exists
		if fileExists(dstPath) && !force {
			log.Info(fmt.Sprintf("   ⏭️  Skipping %s (already exists, use --force to overwrite)", filename))
			skippedCount++
			continue
		}

		// Check if source exists
		if !fileExists(srcPath) {
			log.Warn(fmt.Sprintf("   ⚠️  System template not found: %s", filename))
			continue
		}

		// Read source file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcPath, err)
		}

		// Write to destination
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dstPath, err)
		}

		log.Info(fmt.Sprintf("   ✓ Copied %s", filename))
		copiedCount++
	}

	log.Info("")
	log.Info(fmt.Sprintf("   📊 Summary: %d copied, %d skipped", copiedCount, skippedCount))

	return nil
}
