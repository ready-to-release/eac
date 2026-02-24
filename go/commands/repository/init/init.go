package init

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/environments"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type initCommand struct{}

var _ core.SimpleCommandPort = (*initCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&initCommand{},
	}
}

func (c *initCommand) Name() string { return "init" }

func (c *initCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "init",
		Short:         "Initialize EAC project configuration",
		Long:          "Initialize EAC project configuration.\n\nCreates the .eac directory structure and generates configuration files\nwith calculated defaults. AI provider configuration is optional.\n\nAlways creates:\n  - .eac/repository.yml (module definitions with calculated defaults)\n  - .eac/books.yml (empty documentation books template)\n  - .eac/environments.yml (empty test environments template)\n\nWhen --ai-provider is specified, also creates:\n  - .eac/ai-provider.yml (team config) or\n  - .eac/ai-provider.personal.yml (personal config with tokens)\n\nAvailable AI providers:\n  - claude-api: Claude via Anthropic API (requires ANTHROPIC_API_KEY)\n  - openai: OpenAI via API (requires OPENAI_API_KEY)\n  - gemini: Google Gemini via API (requires GOOGLE_API_KEY)\n\nRe-running init will intelligently update existing configuration:\n  - Preserves user customizations (module names, versioning, dependencies)\n  - Updates AI-generated content (descriptions, component types)\n  - Detects new/removed modules automatically\n\nExamples:\n  init                                                      # Initialize project with config files\n  init                                                      # Re-run to update existing config\n  init --scan                                               # Scan repository and auto-generate config\n  init --scan --ai-provider claude-api                      # Scan with AI-enhanced config generation\n  init --ai-provider claude-api                             # Initialize and configure AI provider\n  init --ai-provider claude-api --ai-token sk-ant-xxx       # Configure with actual token\n  init --copy-templates                                     # Also copy system template files",
		Flags: []core.FlagSpec{
			{Name: "scan", Shorthand: "s", Type: "bool", DefaultValue: "false", Usage: "Scan repository to auto-detect modules and generate configuration"},
			{Name: "ai-provider", Shorthand: "a", Type: "string", Usage: "AI provider to configure (optional)", Completion: []string{"claude-api", "openai", "gemini"}},
			{Name: "ai-token", Type: "string", Usage: "AI provider API token (creates personal config if provided)"},
			{Name: "git-token", Type: "string", Usage: "Git provider API token for repository operations (supports GitHub, GitLab, etc.) (optional)"},
			{Name: "copy-templates", Type: "bool", DefaultValue: "false", Usage: "Copy system default configuration files to repository for customization"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Enable debug mode to save intermediate outputs to the 'out' directory for troubleshooting and analysis"},
		},
	}
}

func (c *initCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return Init()
}

var log = logging.C()

// initFlags holds parsed command-line flags for the init command.
type initFlags struct {
	scan          bool
	aiProvider    string
	aiToken       string
	gitToken      string
	copyTemplates bool
	debug         bool
}

// parseInitFlags parses command-line arguments into an initFlags struct.
func parseInitFlags(args []string) *initFlags {
	f := &initFlags{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--scan", "-s":
			f.scan = true
		case "--ai-provider", "-a":
			if i+1 < len(args) {
				f.aiProvider = args[i+1]
				i++ // Skip the value
			}
		case "--ai-token":
			if i+1 < len(args) {
				f.aiToken = args[i+1]
				i++ // Skip the value
			}
		case "--git-token":
			if i+1 < len(args) {
				f.gitToken = args[i+1]
				i++ // Skip the value
			}
		case "--copy-templates":
			f.copyTemplates = true
		case "--debug", "-d":
			f.debug = true
		}
	}

	return f
}

// showInitSuccess displays the success message and next steps after first-time initialization.
// configPath is the AI provider config path (empty if no AI provider was configured).
// config is the AI agent config (nil if no AI provider was configured).
func showInitSuccess(eacDir string, f *initFlags, configPath string, config *agentConfig) {
	log.Info("")
	log.Info("✅ EAC project initialized")
	log.Info("")
	log.Info("📁 Configuration files created:")
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "repository.yml")))
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "books.yml")))
	log.Info(fmt.Sprintf("   %s", filepath.Join(eacDir, "environments.yml")))
	if configPath != "" {
		log.Info(fmt.Sprintf("   %s", configPath))
	}
	log.Info("")

	if config == nil {
		// No AI provider configured
		log.Info("📋 Next steps:")
		log.Info("   1. Review your configuration: cat .eac/repository.yml")
		log.Info("   2. Verify modules: clie eac show modules")
		log.Info("   3. Commit to version control: git add .eac/")
		log.Info("")
		log.Info("ℹ️  To configure AI provider (optional):")
		log.Info("     clie eac init --ai-provider claude-api")
	} else if f.aiToken != "" {
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
}

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

	f := parseInitFlags(os.Args[2:])

	// Get workspace root via repository API
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(workspaceRoot, "commands", nil, f.debug); err != nil {
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
		if f.aiProvider != "" && f.aiProvider != existingConfig.AIProvider {
			// Override AI provider
			log.Info(fmt.Sprintf("🔄 Switching AI provider: %s → %s", existingConfig.AIProvider, f.aiProvider))
		} else if f.aiProvider != "" && existingConfig.AIProvider != "" {
			// Same provider explicitly requested, just log
			log.Info(fmt.Sprintf("🔄 Reusing existing AI provider: %s", f.aiProvider))
		} else if f.aiProvider == "" && existingConfig.AIProvider != "" {
			// No --ai-provider flag, but existing config has one - keep it
			log.Info(fmt.Sprintf("🔄 Reusing existing AI provider: %s", existingConfig.AIProvider))
		}

		// Re-scan and merge with existing config.
		// Use effective AI provider: flag wins if set, otherwise fall back to existing config.
		effectiveAI := f.aiProvider
		if effectiveAI == "" && existingConfig != nil {
			effectiveAI = existingConfig.AIProvider
		}
		return reinitialize(deps, workspaceRoot, eacDir, f.scan, effectiveAI)
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
	if f.scan {
		// Scan repository and generate config
		if err := generateWithScan(deps, workspaceRoot, eacDir, f.aiProvider); err != nil {
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
	if f.copyTemplates {
		log.Info("")
		log.Info("📄 Copying system template files...")
		if err := copySystemTemplates(workspaceRoot); err != nil {
			log.Error(fmt.Sprintf("Error copying templates: %v", err))
			return 1
		}
		log.Info("✅ System templates copied")
	}

	// If no AI provider specified, show success and exit
	if f.aiProvider == "" {
		showInitSuccess(eacDir, f, "", nil)
		return 0
	}

	// AI provider specified - configure it
	log.Info("")
	log.Info("🤖 Configuring AI Provider")
	log.Info("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Configure agent using --ai-provider flag
	config, err := configureAgent(f.aiProvider)
	if err != nil {
		log.Errorf("Error during configuration: %v", err)
		return 1
	}

	// Create tokens struct
	tokens := &tokenConfig{
		aiToken:  f.aiToken,
		gitToken: f.gitToken,
	}

	// Write configuration (team or personal based on token presence)
	configPath, err := writeConfig(workspaceRoot, config, tokens)
	if err != nil {
		log.Error(fmt.Sprintf("Error writing configuration: %v", err))
		return 1
	}

	showInitSuccess(eacDir, f, configPath, config)
	return 0
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
