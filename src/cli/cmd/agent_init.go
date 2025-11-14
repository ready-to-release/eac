package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/spf13/cobra"
)

func init() {
	AgentCmd.AddCommand(AgentInitCmd)

	// Add --ai flag (required)
	AgentInitCmd.Flags().StringP("ai", "a", "", "AI provider to configure (claude-api, claude-cli, openai, gemini)")
	AgentInitCmd.MarkFlagRequired("ai")
}

var AgentInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI provider configuration",
	Long: `Initialize AI provider configuration for the project.

Creates .r2r/agent-config.yml with the specified AI provider settings.
The configuration file is safe to commit as it only contains environment variable references.

Available providers:
  - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)
  - claude-cli: Claude via CLI subscription (no API key needed)
  - openai: OpenAI via API (requires OPENAI_API_KEY)
  - gemini: Google Gemini via API (requires GOOGLE_API_KEY)

Example:
  r2r agent init --ai claude-cli
  r2r agent init --ai claude-api`,
	RunE: func(cmd *cobra.Command, args []string) error {
		aiProvider, _ := cmd.Flags().GetString("ai")

		// Get workspace root
		workspaceRoot, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to find repository root: %w", err)
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
		if err := createAgentDirectoryStructure(workspaceRoot); err != nil {
			return fmt.Errorf("failed to create directory structure: %w", err)
		}

		// Configure agent using --ai flag
		config, err := configureAgentProvider(aiProvider)
		if err != nil {
			return err
		}

		// Write configuration
		if err := writeAgentConfigFile(configPath, config); err != nil {
			return fmt.Errorf("failed to write configuration: %w", err)
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
		fmt.Println("   3. Run AI-powered commands (e.g., specs create, commit-ai)")
		fmt.Println("")

		return nil
	},
}

// agentProviderConfig holds configuration for an AI provider
type agentProviderConfig struct {
	providerName string // "claude-api", "claude-cli", "openai", "gemini"
	envVarName   string // "ANTHROPIC_API_KEY", etc. (empty for claude-cli)
	model        string // "claude-3-haiku-20240307", etc.
	endpoint     string // API endpoint URL (empty for claude-cli)
}

// configureAgentProvider configures the AI provider based on user input
func configureAgentProvider(aiProvider string) (*agentProviderConfig, error) {
	config := &agentProviderConfig{}

	switch strings.ToLower(aiProvider) {
	case "claude-api":
		config.providerName = "claude-api"
		config.envVarName = "ANTHROPIC_API_KEY"
		config.model = "claude-3-haiku-20240307"
		config.endpoint = "https://api.anthropic.com/v1"

	case "claude-cli":
		config.providerName = "claude-cli"
		config.envVarName = ""
		config.model = "sonnet"
		config.endpoint = ""

	case "openai":
		config.providerName = "openai"
		config.envVarName = "OPENAI_API_KEY"
		config.model = "gpt-4-turbo"
		config.endpoint = "https://api.openai.com/v1"

	case "gemini":
		config.providerName = "gemini"
		config.envVarName = "GOOGLE_API_KEY"
		config.model = "gemini-1.5-pro"
		config.endpoint = "https://generativelanguage.googleapis.com"

	default:
		return nil, fmt.Errorf("unsupported provider: %s\nSupported: claude-api, claude-cli, openai, gemini", aiProvider)
	}

	displayAgentProviderInfo(config)
	return config, nil
}

// displayAgentProviderInfo shows information about the selected provider
func displayAgentProviderInfo(config *agentProviderConfig) {
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

// createAgentDirectoryStructure creates the .r2r directory structure
func createAgentDirectoryStructure(workspaceRoot string) error {
	// Create .r2r directory
	r2rDir := filepath.Join(workspaceRoot, ".r2r")
	if err := os.MkdirAll(r2rDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r directory: %w", err)
	}

	// Create .r2r/logs directory
	logsDir := filepath.Join(r2rDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/logs directory: %w", err)
	}

	return nil
}

// writeAgentConfigFile writes the agent configuration to .r2r/agent-config.yml
func writeAgentConfigFile(configPath string, config *agentProviderConfig) error {
	var content strings.Builder

	content.WriteString("# Agent Configuration\n")
	content.WriteString("# Generated by: r2r agent init\n")
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
