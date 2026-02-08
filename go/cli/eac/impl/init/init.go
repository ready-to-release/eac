// Command: init
// Short: Initialize EAC project configuration
// Long: Initialize EAC project configuration.
// Long:
// Long: Creates the .eac directory structure and generates configuration files
// Long: with calculated defaults. AI provider configuration is optional.
// Long:
// Long: Always creates:
// Long:   - .eac/repository.yml (module definitions with calculated defaults)
// Long:   - .eac/books.yml (empty documentation books template)
// Long:   - .eac/environments.yml (empty test environments template)
// Long:
// Long: When --ai-provider is specified, also creates:
// Long:   - .eac/ai-provider.yml (team config) or
// Long:   - .eac/ai-provider.personal.yml (personal config with tokens)
// Long:
// Long: Available AI providers:
// Long:   - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)
// Long:   - openai: OpenAI via API (requires OPENAI_API_KEY)
// Long:   - gemini: Google Gemini via API (requires GOOGLE_API_KEY)
// Long:
// Long: Re-running init will intelligently update existing configuration:
// Long:   - Preserves user customizations (module names, versioning, dependencies)
// Long:   - Updates AI-generated content (descriptions, component types)
// Long:   - Detects new/removed modules automatically
// Long:
// Long: Examples:
// Long:   init                                                      # Initialize project with config files
// Long:   init                                                      # Re-run to update existing config
// Long:   init --scan                                               # Scan repository and auto-generate config
// Long:   init --scan --ai-provider claude-api                      # Scan with AI-enhanced config generation
// Long:   init --ai-provider claude-api                             # Initialize and configure AI provider
// Long:   init --ai-provider claude-api --ai-token sk-ant-xxx       # Configure with actual token
// Long:   init --copy-templates                                     # Also copy system template files
// Flag.scan: type=bool, shorthand=s, default=false, usage=Scan repository to auto-detect modules and generate configuration, required=false
// Flag.ai-provider: type=string, shorthand=a, usage=AI provider to configure (optional), required=false, completion=claude-api,openai,gemini
// Flag.ai-token: type=string, usage=AI provider API token (creates personal config if provided), required=false
// Flag.git-token: type=string, usage=Git provider API token for repository operations (supports GitHub, GitLab, etc.) (optional), required=false
// Flag.copy-templates: type=bool, default=false, usage=Copy system default configuration files to repository for customization, required=false
// Flag.debug: type=bool, shorthand=d, default=false, usage=Enable debug mode to save intermediate outputs to the 'out' directory for troubleshooting and analysis, required=false
package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(Init)
}

var log = logging.C()

// Init initializes EAC project configuration.
func Init() int {
	return initImpl(defaultDeps())
}

// initImpl contains the implementation of Init with injectable dependencies.
func initImpl(deps *Deps) int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags to get all options
	scan := false
	aiProvider := ""
	aiToken := ""
	gitToken := ""
	copyTemplates := false
	debug := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--scan", "-s":
			scan = true
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

	// Check for existing configuration (smart re-initialization)
	eacDir := paths.EACConfigPath(workspaceRoot)
	existingConfig := detectExistingConfig(workspaceRoot)

	if existingConfig != nil {
		// Re-initialization mode: update existing config
		log.Info("📁 Re-initializing EAC project...")
		log.Info(fmt.Sprintf("   Repository root: %s", workspaceRoot))
		log.Info("")
		log.Info("ℹ️  Existing configuration detected:")
		if existingConfig.AIProvider != "" {
			log.Info(fmt.Sprintf("   - AI provider: %s", existingConfig.AIProvider))
		}
		if existingConfig.ModuleCount > 0 {
			log.Info(fmt.Sprintf("   - Modules: %d", existingConfig.ModuleCount))
		}
		log.Info("")

		// Determine which AI provider to use
		// Only configure AI if explicitly requested via --ai-provider flag
		if aiProvider != "" && aiProvider != existingConfig.AIProvider {
			// Override AI provider
			log.Info(fmt.Sprintf("🔄 Switching AI provider: %s → %s", existingConfig.AIProvider, aiProvider))
		} else if aiProvider != "" && existingConfig.AIProvider != "" {
			// Same provider explicitly requested, just log
			log.Info(fmt.Sprintf("🔄 Reusing existing AI provider: %s", aiProvider))
		} else if aiProvider == "" && existingConfig.AIProvider != "" {
			// No --ai-provider flag, but existing config has one - keep it
			log.Info(fmt.Sprintf("🔄 Reusing existing AI provider: %s", existingConfig.AIProvider))
		}

		// Re-scan and merge with existing config
		// Only pass aiProvider if user explicitly specified it (not from existing config)
		return reinitialize(deps, workspaceRoot, eacDir, scan, aiProvider)
	}

	// First-time initialization mode
	log.Info("📁 Initializing EAC project...")
	log.Info(fmt.Sprintf("   Repository root: %s", workspaceRoot))
	log.Info("")

	// Create .eac directory structure
	if err := createDirectoryStructure(workspaceRoot); err != nil {
		log.Error(fmt.Sprintf("Error creating directory structure: %v", err))
		return 1
	}

	// Generate config files
	if scan {
		// Scan repository and generate config
		if err := generateWithScan(deps, workspaceRoot, eacDir, aiProvider); err != nil {
			log.Error(fmt.Sprintf("Error generating configuration from scan: %v", err))
			return 1
		}
	} else {
		// Generate config files with calculated defaults
		if err := generateRepositoryYML(workspaceRoot, eacDir); err != nil {
			log.Error(fmt.Sprintf("Error generating repository.yml: %v", err))
			return 1
		}

		if err := generateBooksYML(eacDir); err != nil {
			log.Error(fmt.Sprintf("Error generating books.yml: %v", err))
			return 1
		}

		if err := generateEnvironmentsYML(eacDir); err != nil {
			log.Error(fmt.Sprintf("Error generating environments.yml: %v", err))
			return 1
		}
	}

	// Copy system templates if requested
	if copyTemplates {
		log.Info("")
		log.Info("📄 Copying system template files...")
		if err := copySystemTemplates(workspaceRoot); err != nil {
			log.Error(fmt.Sprintf("Error copying templates: %v", err))
			return 1
		}
		log.Info("✅ System templates copied")
	}

	// If no AI provider specified, show success and exit
	if aiProvider == "" {
		log.Info("")
		log.Info("✅ EAC project initialized")
		log.Info("")
		log.Info("📁 Configuration files created:")
		log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "repository.yml")))
		log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "books.yml")))
		log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "environments.yml")))
		log.Info("")
		log.Info("📋 Next steps:")
		log.Info("   1. Review your configuration: cat .eac/repository.yml")
		log.Info("   2. Verify modules: clie eac show modules")
		log.Info("   3. Commit to version control: git add .eac/")
		log.Info("")
		log.Info("ℹ️  To configure AI provider (optional):")
		log.Info("     clie eac init --ai-provider claude-api")
		log.Info("")
		return 0
	}

	// AI provider specified - configure it
	log.Info("")
	log.Info("🤖 Configuring AI Provider")
	log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

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
	log.Info("✅ EAC project initialized")
	log.Info("")
	log.Info("📁 Configuration files created:")
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "repository.yml")))
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "books.yml")))
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "environments.yml")))
	log.Info(fmt.Sprintf("   %s", configPath))
	log.Info("")

	// Provide appropriate next steps based on config type
	if aiToken != "" {
		// Personal config with tokens
		log.Info("📋 Next steps:")
		log.Info("   1. Review your configuration: cat .eac/repository.yml")
		log.Info("   2. Verify modules: clie eac show modules")
		log.Info("   3. Do NOT commit ai-provider.personal.yml (contains tokens)")
		log.Info("   4. Commit other config files: git add .eac/repository.yml .eac/books.yml .eac/environments.yml")
	} else {
		// Team config with placeholders
		log.Info("📋 Next steps:")
		log.Info("   1. Review your configuration: cat .eac/repository.yml")
		log.Info("   2. Verify modules: clie eac show modules")
		if config.envVarName != "" {
			log.Info(fmt.Sprintf("   3. Set environment variable: %s", config.envVarName))
		}
		log.Info("   4. Commit to version control: git add .eac/")
	}
	log.Info("")

	return 0
}

// agentConfig holds configuration for an AI provider.
type agentConfig struct {
	providerName string // "claude-api", "claude-cli", "openai", "gemini"
	envVarName   string // "ANTHROPIC_API_KEY", etc. (empty for claude-cli)
	model        string // "claude-3-haiku-20240307", etc.
	endpoint     string // API endpoint URL (empty for claude-cli)
}

// tokenConfig holds actual token values (for personal config).
type tokenConfig struct {
	aiToken  string // Actual AI API token
	gitToken string // Actual Git API token
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// configureAgent configures the AI provider based on user input.
func configureAgent(aiProvider string) (*agentConfig, error) {
	config := &agentConfig{}

	// Configure provider based on --ai flag
	if err := configureProvider(config, aiProvider); err != nil {
		return nil, err
	}

	displayProviderInfo(config)
	return config, nil
}

// configureProvider sets up the config based on the provider key.
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

// displayProviderInfo shows information about the selected provider.
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

// createDirectoryStructure creates the .eac directory structure.
func createDirectoryStructure(workspaceRoot string) error {
	// Create .eac directory
	eacDir := paths.EACConfigPath(workspaceRoot)
	log.Info(fmt.Sprintf("   Creating directory: %s", eacDir))

	if err := os.MkdirAll(eacDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .eac directory: %w", err)
	}

	// Verify directory was created
	if _, err := os.Stat(eacDir); os.IsNotExist(err) {
		return fmt.Errorf("directory creation reported success but directory does not exist: %s", eacDir)
	}

	return nil
}

// writeConfig writes the EAC configuration (team or personal based on tokens).
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
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// buildConfigContent builds the YAML config content.
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
		content.WriteString("# - This creates: .eac/ai-provider.personal.yml (gitignored)\n")
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

// copySystemTemplates copies system default configuration files to user repository.
func copySystemTemplates(workspaceRoot string) error {
	// Get system root (Docker container or local dev)
	systemRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	if systemRoot == "" {
		// Local dev mode - system defaults are in workspace root
		systemRoot = workspaceRoot
	}

	// List of Category B files (system defaults) to copy
	templateFiles := []string{
		"ai-config.yml",
		"logging.yml",
		"testing-tags.yml",
	}

	eacDir := paths.EACConfigPath(workspaceRoot)
	systemEacDir := paths.EACConfigPath(systemRoot)

	copiedCount := 0
	skippedCount := 0

	for _, filename := range templateFiles {
		srcPath := fmt.Sprintf("%s/%s", systemEacDir, filename)
		dstPath := fmt.Sprintf("%s/%s", eacDir, filename)

		// Check if destination exists (skip if already there - smart re-init)
		if fileExists(dstPath) {
			log.Info(fmt.Sprintf("   ⏭️  Skipping %s (already exists)", filename))
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
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dstPath, err)
		}

		log.Info(fmt.Sprintf("   ✓ Copied %s", filename))
		copiedCount++
	}

	log.Info("")
	log.Info(fmt.Sprintf("   📊 Summary: %d copied, %d skipped", copiedCount, skippedCount))

	return nil
}

// generateRepositoryYML generates repository.yml with calculated defaults.
func generateRepositoryYML(workspaceRoot, eacDir string) error {
	// Load calculated config (merges all defaults)
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        workspaceRoot,
		ValidateSchemas: false, // Don't validate during init - files may not exist yet
	})
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Serialize to YAML
	yamlBytes, err := yaml.Marshal(cfg.Repository)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Add header
	header := `# EAC Repository Configuration
# Generated by: clie eac init
#
# This file defines your repository settings and module domain.
# Edit this file to customize modules, dependencies, and build settings.
#
# Documentation: https://eac.readthedocs.io/configuration/repository

`

	path := filepath.Join(eacDir, "repository.yml")
	if err := os.WriteFile(path, []byte(header+string(yamlBytes)), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}

	log.Info("   ✓ Generated repository.yml")
	return nil
}

// generateBooksYML generates books.yml with empty template.
func generateBooksYML(eacDir string) error {
	content := `# Documentation Books Configuration
# Generated by: clie eac init
#
# Define documentation books for your project.
# Each book represents a documentation site built with MkDocs.
#
# Example:
#   books:
#     - name: docs
#       title: Project Documentation
#       description: Main documentation site
#       output: site
#       sources:
#         - path: docs
#
# Documentation: https://eac.readthedocs.io/configuration/books

books: []
`

	path := filepath.Join(eacDir, "books.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write books.yml: %w", err)
	}

	log.Info("   ✓ Generated books.yml")
	return nil
}

// generateEnvironmentsYML generates environments.yml with empty template.
func generateEnvironmentsYML(eacDir string) error {
	content := `# Test Environments Configuration
# Generated by: clie eac init
#
# Define test environments for your project.
# Environments specify where and how tests run.
#
# Example:
#   environments:
#     - moniker: local
#       name: Local Development
#       description: Local development environment
#       level: L0
#
#     - moniker: staging
#       name: Staging
#       description: Pre-production staging environment
#       level: L2
#
# Test Levels:
#   L0 - Unit tests (no external dependencies)
#   L1 - Integration tests (local dependencies)
#   L2 - System tests (staging environment)
#   L3 - Acceptance tests (production-like)
#   L4 - Production validation
#
# Documentation: https://eac.readthedocs.io/configuration/environments

environments: []
`

	path := filepath.Join(eacDir, "environments.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write environments.yml: %w", err)
	}

	log.Info("   ✓ Generated environments.yml")
	return nil
}

// generateWithScan scans the repository and generates configuration files.
func generateWithScan(deps *Deps, workspaceRoot, eacDir, aiProvider string) error {
	log.Info("")
	log.Info("🔍 Scanning repository structure...")

	// Run scanner
	scanResult, err := ScanRepository(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to scan repository: %w", err)
	}

	// Report scan results
	if len(scanResult.Modules) == 0 {
		log.Warn("   ⚠️  No modules detected")
		log.Info("   Falling back to manual configuration")

		// Generate empty templates
		if err := generateBooksYML(eacDir); err != nil {
			return err
		}
		if err := generateEnvironmentsYML(eacDir); err != nil {
			return err
		}
		return nil
	}

	log.Info(fmt.Sprintf("   ✓ Detected: %d module(s)", len(scanResult.Modules)))
	for _, mod := range scanResult.Modules {
		log.Info(fmt.Sprintf("      - %s (%s) at %s", mod.Name, mod.Language, mod.Root))
	}

	// Generate repository.yml using strategy pattern
	log.Info("")
	var repoYAML string

	if aiProvider != "" {
		log.Info(fmt.Sprintf("🤖 Generating configuration with AI (%s)...", aiProvider))
		// Use AI generator with rule-based fallback
		aiGen := newAIGenerator(aiProvider, deps)
		ruleGen := NewRuleBasedGenerator()

		// Try AI generation first
		repoYAML, err = aiGen.Generate(workspaceRoot, scanResult)
		if err != nil {
			log.Warn(fmt.Sprintf("   ⚠️  AI generation failed: %v", err))
			log.Info("   Falling back to rule-based generation")
			repoYAML, err = ruleGen.Generate(workspaceRoot, scanResult)
			if err != nil {
				return fmt.Errorf("failed to generate configuration: %w", err)
			}
		} else {
			log.Info("   ✓ AI-enhanced configuration generated")
		}
	} else {
		log.Info("🔧 Generating configuration (rule-based)...")
		generator := NewRuleBasedGenerator()
		repoYAML, err = generator.Generate(workspaceRoot, scanResult)
		if err != nil {
			return fmt.Errorf("failed to generate configuration: %w", err)
		}
		log.Info("   ✓ Configuration generated")
		log.Info("")
		log.Info("   💡 Tip: Use --ai-provider claude-api for enhanced descriptions")
	}

	// Write repository.yml
	repoPath := filepath.Join(eacDir, "repository.yml")
	if err := os.WriteFile(repoPath, []byte(repoYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write repository.yml: %w", err)
	}
	log.Info(fmt.Sprintf("   ✓ Generated %s", repoPath))

	// Generate other config files
	if err := generateBooksYML(eacDir); err != nil {
		return err
	}
	if err := generateEnvironmentsYML(eacDir); err != nil {
		return err
	}

	return nil
}

// generateRepositoryYAMLFromScan is deprecated - use NewRuleBasedGenerator().Generate() instead.
// Kept for backward compatibility.
func generateRepositoryYAMLFromScan(scanResult *ScanResult) string {
	generator := NewRuleBasedGenerator()
	result, _ := generator.Generate("", scanResult)
	return result
}
