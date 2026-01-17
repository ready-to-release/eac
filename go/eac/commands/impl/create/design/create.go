// Command: create design
// Short: Generate workspace.dsl for a module using AI
// Long: Generates Structurizr DSL workspace files for a module by analyzing its source code using AI.
// Long: The AI analyzes the source code in go/eac/<module>/ and creates comprehensive architecture
// Long: documentation including system context views (external systems and actors), container views
// Long: (major components), and component views (internal structure). All generated workspaces are
// Long: automatically validated against Structurizr CLI to ensure correct syntax before saving to
// Long: specs/<module>/.design/workspace.dsl (or custom path via --output). Use --debug to save AI
// Long: prompts and outputs.
// Long:
// Long: Expected Output:
// Long: - Structurizr DSL workspace file at output path
// Long: - System context, container, and component views
// Long: - Validation results if Docker available
// Long: - Debug logs in out/commands.log if --debug enabled
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to out/commands.log for debugging
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing workspace.dsl file if it exists
// Flag.output: type=string, shorthand=o, default=, usage=Custom output path for workspace.dsl (default: specs/<module>/.design/workspace.dsl)
// Flag.prompt: type=string, shorthand=, default=, usage=Custom AI prompt file path
// Flag.skip-validation: type=bool, shorthand=, default=false, usage=Skip Docker validation (useful when Docker is unavailable)
// Usage: create design <module>
package design

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	design "github.com/ready-to-release/eac/go/eac/commands/impl/design"
	designInternal "github.com/ready-to-release/eac/go/eac/commands/impl/design/helper"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	coreai "github.com/ready-to-release/eac/go/eac/core/ai"
	eacConfig "github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/validation/formats/structurizr"
)

func init() {
	registry.Register(CreateDesign)
}

// Intent: Create a Structurizr DSL workspace from natural language description using AI
//
// CreateDesign orchestrates the architecture design generation workflow.
var log = logging.C()

// commandFlags defines valid flags for the create design command

func CreateDesign() int {
	// Validate flags
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(config.TemplateRoot, "commands", nil, config.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Create output handler
	out := design.NewOutput()

	// Validate module exists
	if err := validateModuleExists(config, out); err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Check Docker availability early (before expensive AI generation) unless skipped
	if !config.SkipValidation {
		if err := checkDockerAvailability(config, out); err != nil {
			out.Errorf("Error: %v", err)
			return 1
		}
	} else {
		out.Progress("⚠️  Skipping Docker validation")
	}

	// Load EAC config for path resolution (uses helper with clean fallback for tests)
	cfg := eacConfig.LoadOrNil(config.TemplateRoot)

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		if cfg == nil {
			// Fallback to default path if config loading fails (e.g., in tests)
			outputPath = filepath.Join(config.TemplateRoot, "specs", config.Module, ".design", "workspace.dsl")
		} else {
			outputPath = filepath.Join(cfg.Repository.SpecsPathAbs(config.TemplateRoot, config.Module), ".design", "workspace.dsl")
		}
	}

	// Check if output exists (unless --force)
	if !config.Force && fileExists(outputPath) {
		out.Errorf("Error: workspace already exists at %s", outputPath)
		out.Error("Use --force to overwrite")
		return 1
	}

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(config, out)
	if err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Generate and validate with retry
	cleanedOutput, err := generateAndValidate(config, fullPrompt, out)
	if err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(config, outputPath, cleanedOutput, out); err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	return 0
}

// DesignConfig holds configuration for design create command.
type DesignConfig struct {
	Module         string // Module name (e.g., "r2r-cli", "commands")
	SourcePath     string // Path to source code (e.g., "go/eac/commands")
	OutputPath     string // Custom output path (empty = default to specs/<module>/.design/workspace.dsl)
	PromptPath     string // Custom AI prompt file path (empty = default prompt)
	Debug          bool
	Force          bool
	SkipValidation bool
	TemplateRoot   string
}

// parseConfig parses command line arguments into DesignConfig.
func parseConfig() (*DesignConfig, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find and parse arguments
	modulePath, flags, err := parseCreateCommandArgs(os.Args)
	if err != nil {
		return nil, err
	}

	// Create config with parsed flags
	config := &DesignConfig{
		Debug:          flags.debug,
		Force:          flags.force,
		OutputPath:     flags.outputPath,
		PromptPath:     flags.promptPath,
		SkipValidation: flags.skipValidation,
		TemplateRoot:   repoRoot,
	}

	// Use input as-is (moniker)
	config.Module = modulePath

	// Validate module name for security
	if err := designInternal.ValidateModuleName(config.Module); err != nil {
		return nil, fmt.Errorf("invalid module name: %w", err)
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load module contracts: %w", err)
	}

	module, exists := moduleReport.Registry.Get(config.Module)
	if !exists {
		return nil, fmt.Errorf("module not found: %s\n\nAvailable modules:\n%s",
			config.Module, formatModuleList(moduleReport))
	}

	// Store validated moniker and source path from contract
	config.Module = module.Moniker
	config.SourcePath = filepath.Join(repoRoot, module.Files.Root)

	return config, nil
}

// createFlags holds parsed command-line flags.
type createFlags struct {
	debug          bool
	force          bool
	outputPath     string
	promptPath     string
	skipValidation bool
}

// parseCreateCommandArgs parses args for the create subcommand
// Returns: module path, flags, error.
func parseCreateCommandArgs(args []string) (string, *createFlags, error) {
	// Find command position
	cmdPos := -1
	for i, arg := range args {
		if arg == "create" {
			cmdPos = i
			break
		}
	}

	if cmdPos == -1 {
		return "", nil, fmt.Errorf("invalid command structure")
	}

	// Skip "create" and "design" tokens - start parsing from the module name
	// Args structure: [... "create" "design" <module> <flags>...]
	startPos := cmdPos + 1
	if startPos < len(args) && args[startPos] == "design" {
		startPos++ // Skip "design" subcommand token
	}

	// Parse flags and positional arguments
	cmdFlags := &createFlags{}
	var positionalArgs []string

	// Extract args from startPos onward for flag parsing
	argsToParse := args[startPos:]

	// Use shared flags package for debug flag
	cmdFlags.debug = flags.ParseDebugFlag(argsToParse)
	cmdFlags.force = flags.HasFlag(argsToParse, "--force", "")
	cmdFlags.skipValidation = flags.HasFlag(argsToParse, "--skip-validation", "")

	// Parse value flags and positional args
	for i := 0; i < len(argsToParse); i++ {
		arg := argsToParse[i]

		if arg == "--output" || arg == "-o" {
			// Next argument is the output path
			if i+1 < len(argsToParse) {
				cmdFlags.outputPath = argsToParse[i+1]
				i++ // Skip next arg
			}
		} else if arg == "--prompt" {
			// Next argument is the prompt path
			if i+1 < len(argsToParse) {
				cmdFlags.promptPath = argsToParse[i+1]
				i++ // Skip next arg
			}
		} else if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Validate we have a module name
	if len(positionalArgs) == 0 {
		return "", nil, fmt.Errorf("module name required\n\nUsage: create design <module>\nExample: create design r2r-cli")
	}

	return positionalArgs[0], cmdFlags, nil
}

// formatModuleList returns a formatted list of available modules.
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, mod.Files.Root))
	}
	return sb.String()
}

// validateModuleExists checks if the source code exists for the specified module.
func validateModuleExists(config *DesignConfig, out *design.Output) error {
	out.Progressf("🔍 Validating source code for module '%s'...", config.Module)

	// Check if source directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source code not found for module '%s'\n\nExpected at: %s\n\nUsage: design create <module>\nExample: design create r2r-cli (analyzes code in go/r2r/cli/)",
			config.Module, config.SourcePath)
	}

	out.Progressf("✅ Source code found at: %s", config.SourcePath)
	return nil
}

// checkDockerAvailability checks if Docker is running before starting expensive operations.
func checkDockerAvailability(config *DesignConfig, out *design.Output) error {
	out.Progress("🐳 Checking Docker availability...")

	// Create a temporary validator to check Docker
	validator, err := structurizr.NewCompositeValidator("", "", true)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Check if Docker is running
	if !validator.IsDockerRunning() {
		return fmt.Errorf("Docker is not running\n\nStructurizr validation requires Docker to be running.\nPlease start Docker and try again.")
	}

	out.Progress("✅ Docker is available")
	return nil
}

// loadAndBuildPrompt loads the contract and builds the AI prompt.
func loadAndBuildPrompt(config *DesignConfig, out *design.Output) (string, error) {
	out.Progress("📋 Loading design contract...")

	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		return "", err
	}

	if config.Debug {
		log.Debugf("Full AI prompt built: promptLength=%d", len(fullPrompt))
	}

	return fullPrompt, nil
}

// buildContractBasedPrompt builds the AI prompt with contract context.
func buildContractBasedPrompt(config *DesignConfig) (string, error) {
	// Load contract using generalized loader
	loader := coreai.NewContractLoader(config.TemplateRoot, coreai.TypeDesign, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Load prompt (default or custom)
	promptContent, err := loadPrompt(config)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Build prompt template
	promptTemplate, err := coreai.BuildPromptWithTemplate(
		promptContent,
		contractData,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt template: %w", err)
	}

	// Build final prompt with module context
	var prompt strings.Builder
	prompt.WriteString(promptTemplate)
	prompt.WriteString("\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n")

	// Module context
	prompt.WriteString("## Target Module\n\n")
	prompt.WriteString("Module: ")
	prompt.WriteString(config.Module)
	prompt.WriteString("\n\n")
	prompt.WriteString("Analyze the source code in the following directory and generate architecture documentation:\n\n")
	prompt.WriteString("Source Path: ")
	prompt.WriteString(config.SourcePath)
	prompt.WriteString("\n\n")
	prompt.WriteString("Use the naming format: ")
	prompt.WriteString(config.Module)
	prompt.WriteString(" Architecture\n\n")

	// Analysis instructions
	prompt.WriteString("## Analysis Instructions\n\n")
	prompt.WriteString("Examine the module's source code to understand:\n")
	prompt.WriteString("- Main components and their responsibilities\n")
	prompt.WriteString("- Dependencies between components\n")
	prompt.WriteString("- External systems or services the module interacts with\n")
	prompt.WriteString("- Data flow and relationships\n\n")

	// Generation requirements
	prompt.WriteString("## Generation Requirements\n\n")
	prompt.WriteString("Generate a COMPLETE architecture including:\n")
	prompt.WriteString("- System Context view: Show the main system with external actors and systems\n")
	prompt.WriteString("- Container view: Break down the main system into major containers\n")
	prompt.WriteString("- Component views: For each significant container, show internal components\n")
	prompt.WriteString("- Relationships: Connect all elements with descriptive relationships\n\n")

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete Structurizr DSL workspace now.\n")
	prompt.WriteString("Return ONLY the DSL content starting with 'workspace' - no markdown fences, no explanations.\n")

	return prompt.String(), nil
}

// loadPrompt loads the AI prompt for design generation with three-tier priority:
// 1. Command flag (--prompt)
// 2. Team override (.r2r/eac/templates/coreai.TypeDesign/design.md)
// 3. System default (templates/coreai.TypeDesign/design.md)
// Convention: Empty string uses type name (design.md).
func loadPrompt(config *DesignConfig) (string, error) {
	// Load prompt with three-tier priority system
	loader := coreai.NewContractLoader(config.TemplateRoot, coreai.TypeDesign, "")
	prompt, source, err := loader.LoadPromptWithPriority("", config.PromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Log source if not default
	if source != "embedded fallback" && config.Debug {
		log.Errorf("ℹ️  Using %s prompt", source)
	}

	return prompt, nil
}

// generateAndValidate generates AI output with retry and validates with Structurizr CLI.
func generateAndValidate(config *DesignConfig, prompt string, out *design.Output) (string, error) {
	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create composite validator (quick + full validation)
	// Skip expensive Docker validation if quick validation finds errors or if --skip-validation
	validator, err := structurizr.NewCompositeValidator(config.Module, config.TemplateRoot, !config.SkipValidation)
	if err != nil {
		return "", fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Load AI config to get retry strategy
	var aiConfig *coreai.AIConfig
	aiConfig, err = coreai.LoadAIConfig(config.TemplateRoot)
	if err != nil {
		log.Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	// Build retry configuration using factory
	retryConfig, err := coreai.BuildRetryConfig(
		coreai.TypeDesign,
		coreai.FormatStructurizr,
		executorAdapter,
		validator,
		config.TemplateRoot,
		aiConfig,
		coreai.WithDebug(config.Debug),
		coreai.WithLogger(logging.C().Zap()),
		coreai.WithDefaultMaxAttempts(3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry and validation
	out.Progress("🤖 Generating architecture design with AI...")

	result, err := coreai.GenerateWithRetry(
		context.Background(),
		retryConfig,
		prompt,
	)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return result.Output, nil
}

// writeOutputAndReportSuccess writes the workspace file and reports success.
func writeOutputAndReportSuccess(config *DesignConfig, outputPath, content string, out *design.Output) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Report success
	out.Progress("\n✅ Architecture design created")
	out.Progressf("   File: %s", outputPath)
	if !config.SkipValidation {
		out.Progress("   Valid: ✅ Passed Structurizr validation\n")
	} else {
		out.Progress("   Valid: ⚠️  Validation skipped\n")
	}
	out.Progress("ℹ️  Next steps:")
	out.Progress("   1. Review the generated workspace")
	out.Progress("   2. Refine containers and relationships as needed")
	out.Progressf("   3. View in browser: design serve %s", config.Module)
	out.Progressf("   4. Validate anytime: design validate %s", config.Module)

	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
