// Command: init
// Short: Initialize AI provider configuration for the project
// Long: Initialize AI provider configuration for the project.
// Long:
// Long: Creates .r2r/agent-config.yml with the specified AI provider settings.
// Long: The configuration file is safe to commit as it only contains environment variable references.
// Long:
// Long: Available providers:
// Long:   - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)
// Long:   - claude-cli: Claude via CLI subscription (no API key needed)
// Long:   - openai: OpenAI via API (requires OPENAI_API_KEY)
// Long:   - gemini: Google Gemini via API (requires GOOGLE_API_KEY)
// Long:
// Long: Example:
// Long:   init --ai claude-cli
// Long:   init --ai claude-api
// Flag.ai: type=string, shorthand=a, usage=AI provider to configure, required=true, completion=claude-api,claude-cli,openai,gemini
// HasSideEffects: true
package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/ai/providers"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(Init)
}

// Intent: Initialize AI provider configuration for a project
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear sequential flow: parse args → validate → create dirs → write config
//   - Helper functions with single responsibility (configureProvider, writeAgentConfig, etc.)
//   - User feedback at each step
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

// Init initializes AI provider configuration
func Init() int {
	// Parse --ai flag
	aiProvider := ""
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--ai" && i+1 < len(os.Args) {
			aiProvider = os.Args[i+1]
			break
		}
	}

	// Validate required flag
	if aiProvider == "" {
		fmt.Fprintf(os.Stderr, "Error: --ai flag is required\n")
		fmt.Fprintf(os.Stderr, "Usage: init --ai <provider>\n")
		fmt.Fprintf(os.Stderr, "\nAvailable providers: claude-api, claude-cli, openai, gemini\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  init --ai claude-api\n")
		fmt.Fprintf(os.Stderr, "  init --ai claude-cli   # Subscription access (no API costs)\n")
		return 1
	}

	// Get workspace root (where to create .r2r)
	// In tests, this will be the isolated temp directory
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Get the source repository root (where prompt templates live)
	// In tests, this should be the REAL repository, not the isolated temp dir
	// We detect this by temporarily unsetting R2R_REPO_ROOT
	sourceRepoRoot := workspaceRoot
	if repoRoot := os.Getenv("R2R_REPO_ROOT"); repoRoot != "" {
		// We're in a test - find the real repository root for copying templates
		// Unset the override temporarily
		os.Unsetenv("R2R_REPO_ROOT")
		realRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			// If we can't find real root, fall back to workspace root
			// (tests will fail but won't crash)
			sourceRepoRoot = workspaceRoot
		} else {
			sourceRepoRoot = realRoot
		}
		// Restore the override
		os.Setenv("R2R_REPO_ROOT", repoRoot)
	}

	r2rDir := filepath.Join(workspaceRoot, ".r2r")
	configPath := filepath.Join(r2rDir, "agent-config.yml")

	// Check if already initialized
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("⚠️  Project already initialized")
		fmt.Printf("   Config exists: %s\n", configPath)
		fmt.Println("")
		fmt.Println("🔄 Reconfiguring agent configuration...")
		fmt.Println("")
	}

	fmt.Println("🤖 Initialize Agent Configuration")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")

	// Create .r2r directory structure
	fmt.Println("📁 Creating directory structure...")
	if err := createDirectoryStructure(workspaceRoot, sourceRepoRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory structure: %v\n", err)
		return 1
	}
	fmt.Println("✅ Directory structure created")

	// Configure agent using --ai flag
	config, err := configureAgent(aiProvider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during configuration: %v\n", err)
		return 1
	}

	// Write configuration
	if err := writeAgentConfig(configPath, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing configuration: %v\n", err)
		return 1
	}

	// Success message
	fmt.Println("")
	fmt.Println("✅ Configuration saved")
	fmt.Printf("   File: %s\n", configPath)
	fmt.Println("")
	fmt.Println("ℹ️  Next steps:")
	fmt.Println("   1. Commit the config file (safe - contains no secrets)")
	if config.envVarName != "" {
		fmt.Printf("   2. Set environment variable: %s\n", config.envVarName)
	}
	fmt.Println("   3. Run AI-powered commands (e.g., specs create, commit)")
	fmt.Println("")

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

	case "claude-cli":
		config.providerName = "claude-cli"
		config.envVarName = ""
		config.model = providers.DefaultClaudeCLIModel
		config.endpoint = ""

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
		return fmt.Errorf("unsupported provider: %s\nSupported: claude-api, claude-cli, openai, gemini", provider)
	}

	return nil
}

// displayProviderInfo shows information about the selected provider
func displayProviderInfo(config *agentConfig) {
	fmt.Println("")
	fmt.Printf("✓ %s selected\n", config.providerName)
	if config.envVarName != "" {
		fmt.Printf("  Environment variable: %s\n", config.envVarName)
	}

	// Provider-specific API key instructions
	switch config.providerName {
	case "claude-api":
		fmt.Println("  Get your API key at: https://claude.ai/settings/api")
		fmt.Println("  Note: Personal or workspace-owned API keys both work")
		fmt.Println("  Requires: ANTHROPIC_API_KEY environment variable")
	case "claude-cli":
		fmt.Println("  Uses Claude Pro subscription (no API key needed)")
		fmt.Println("  Requires: `claude` CLI tool installed")
		fmt.Println("  No API costs - uses subscription credits")
	case "openai":
		fmt.Println("  Get your API key at: https://platform.openai.com/api-keys")
	case "gemini":
		fmt.Println("  Get your API key at: https://makersuite.google.com/app/apikey")
	}

	fmt.Println("")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
}

// createDirectoryStructure creates the .r2r directory structure
func createDirectoryStructure(workspaceRoot string, repoRoot string) error {
	// Create .r2r directory
	r2rDir := filepath.Join(workspaceRoot, ".r2r")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r directory: %w", err)
	}

	// Copy AI contracts to .r2r/contracts for user customization
	if err := copyAIContracts(r2rDir, repoRoot); err != nil {
		return fmt.Errorf("failed to copy AI contracts: %w", err)
	}

	return nil
}

// copyAIContracts copies AI contracts from contracts/ai to .r2r/contracts for user customization
func copyAIContracts(r2rDir string, repoRoot string) error {
	// Source: contracts/ai/
	sourceAIDir := filepath.Join(repoRoot, "contracts", "ai")

	// Destination: .r2r/contracts/ai/
	destAIDir := filepath.Join(r2rDir, "contracts", "ai")

	// Copy entire contracts/ai directory structure
	if err := copyDir(sourceAIDir, destAIDir); err != nil {
		return fmt.Errorf("failed to copy contracts/ai directory: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src string, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory %s: %w", src, err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", src, err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}

	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dst, err)
	}

	return nil
}

// writeAgentConfig writes the agent configuration to .r2r/agent-config.yml
func writeAgentConfig(configPath string, config *agentConfig) error {
	var content strings.Builder

	content.WriteString("# Agent Configuration\n")
	content.WriteString("# Generated by: init command\n")
	content.WriteString("# SAFE TO COMMIT: Contains only environment variable references, not actual secrets\n")
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
