// Command: update design
// Short: Update existing workspace.dsl for a module using AI
// Long: Updates an existing Structurizr DSL workspace file for a module by analyzing its current
// Long: source code using AI. The AI re-analyzes the source code in src/<module>/ and updates
// Long: the architecture documentation to reflect current state, preserving the overall structure
// Long: while incorporating new components, relationships, or changes. All updated workspaces are
// Long: automatically validated against Structurizr CLI to ensure correct syntax before saving.
// Long: Use --debug to save intermediate outputs to out/commands.log for debugging.
// Long:
// Long: Expected Output:
// Long:   - Updated Structurizr DSL workspace file at specs/<module>/.design/workspace.dsl
// Long:   - Preserves overall structure while incorporating code changes
// Long:   - Validated syntax (passed Structurizr CLI validation)
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to out/commands.log for debugging
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite workspace.dsl even if validation fails
// Flag.output: type=string, shorthand=o, default=, usage=Custom output path for workspace.dsl (default: specs/<module>/.design/workspace.dsl)
// Flag.prompt: type=string, shorthand=, default=, usage=Custom AI prompt file path
// Usage: update design <module>
package design

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	createDesign "github.com/ready-to-release/eac/go/cli/eac/impl/create/design"
	designHelper "github.com/ready-to-release/eac/go/cli/eac/impl/design"
	designInternal "github.com/ready-to-release/eac/go/cli/eac/impl/design/helper"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/validation/formats/structurizr"
)

var log = logging.C()

// commandFlags defines valid flags for the update design command

func init() {
	registry.Register(UpdateDesign)
}

// Intent: Update an existing Structurizr DSL workspace by re-analyzing source code
//
// UpdateDesign orchestrates the architecture design update workflow.
func UpdateDesign() int {
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

	// Initialize logger
	if err := logging.ConfigureLoggingSimple(config.TemplateRoot, "commands", nil, config.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
		// Continue without logger - output will still work
	}
	defer logging.CloseLogging()
	defer logging.CloseLogging()

	// Create output handler
	out := designHelper.NewOutput()

	// Validate module and existing workspace
	if err := validateModuleAndWorkspace(config, out); err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Determine workspace path
	workspacePath := config.OutputPath
	if workspacePath == "" {
		workspacePath = paths.WorkspaceDSLPath(config.TemplateRoot, config.Module)
	}

	// Load existing workspace
	existingWorkspace, err := loadExistingWorkspace(config, out, workspacePath)
	if err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Build update prompt with existing workspace context
	fullPrompt, err := buildUpdatePrompt(config, out, existingWorkspace)
	if err != nil {
		out.Errorf("Error: %v", err)
		return 1
	}

	// Generate updated workspace
	updatedWorkspace, err := generateUpdatedWorkspace(config, out, fullPrompt)
	if err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Write output and report success
	if err := writeUpdatedWorkspace(config, out, workspacePath, updatedWorkspace); err != nil {
		out.Errorf("\n❌ Error: %v", err)
		return 1
	}

	return 0
}

// UpdateConfig holds configuration for design update command.
type UpdateConfig struct {
	Module       string // Module name (e.g., "r2r-cli", "commands")
	SourcePath   string // Path to source code (e.g., "go/cli/eac")
	OutputPath   string // Custom output path (empty = default to specs/<module>/.design/workspace.dsl)
	PromptPath   string // Custom AI prompt file path (empty = default prompt)
	Debug        bool
	Force        bool
	TemplateRoot string
}

// parseConfig parses command line arguments into UpdateConfig.
func parseConfig() (*UpdateConfig, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find and parse arguments
	modulePath, cmdFlags, err := parseUpdateCommandArgs(os.Args)
	if err != nil {
		return nil, err
	}

	// Create config with parsed flags
	config := &UpdateConfig{
		Debug:        cmdFlags.debug,
		Force:        cmdFlags.force,
		OutputPath:   cmdFlags.outputPath,
		PromptPath:   cmdFlags.promptPath,
		TemplateRoot: repoRoot,
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
	// Get the buildable package root (go or typescript) for source analysis
	buildableRoot := module.Components.GetBuildableRoot()
	if buildableRoot == "" {
		// Fallback to first available package root
		for _, root := range module.GetComponentRoots() {
			buildableRoot = root
			break
		}
	}
	config.SourcePath = filepath.Join(repoRoot, buildableRoot)

	return config, nil
}

// updateFlags holds parsed command-line flags.
type updateFlags struct {
	debug      bool
	force      bool
	outputPath string
	promptPath string
}

// parseUpdateCommandArgs parses args for the update subcommand
// Returns: module path, flags, error.
func parseUpdateCommandArgs(args []string) (string, *updateFlags, error) {
	// Find command position (looking for "update" followed by "design")
	cmdPos := -1
	for i, arg := range args {
		if arg == "update" && i+1 < len(args) && args[i+1] == "design" {
			cmdPos = i + 1 // Position after "design"
			break
		}
	}

	if cmdPos == -1 {
		return "", nil, fmt.Errorf("invalid command structure")
	}

	// Parse flags and positional arguments (skip "design" subcommand)
	cmdArgs := args[cmdPos+1:]
	updateFlags := &updateFlags{}
	var positionalArgs []string

	// Parse debug flag using shared package
	updateFlags.debug = flags.ParseDebugFlag(cmdArgs)

	for i := 0; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]

		switch {
		case arg == "--debug" || arg == "-d":
			// Already handled by shared flags package
			continue

		case arg == "--force":
			updateFlags.force = true

		case arg == "--output" || arg == "-o":
			// Next argument is the output path
			if i+1 < len(cmdArgs) {
				updateFlags.outputPath = cmdArgs[i+1]
				i++ // Skip next arg
			}

		case arg == "--prompt":
			// Next argument is the prompt path
			if i+1 < len(cmdArgs) {
				updateFlags.promptPath = cmdArgs[i+1]
				i++ // Skip next arg
			}

		case !strings.HasPrefix(arg, "-"):
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Validate we have a module name
	if len(positionalArgs) == 0 {
		return "", nil, fmt.Errorf("module name required\n\nUsage: design update <module>\nExample: design update r2r-cli")
	}

	return positionalArgs[0], updateFlags, nil
}

// formatModuleList returns a formatted list of available modules.
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		// Get first package root for display
		var displayRoot string
		for _, root := range mod.GetComponentRoots() {
			displayRoot = root
			break
		}
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, displayRoot))
	}
	return sb.String()
}

// validateModuleAndWorkspace checks if the source code and existing workspace exist.
func validateModuleAndWorkspace(config *UpdateConfig, out *designHelper.Output) error {
	out.Progressf("🔍 Validating module '%s'...", config.Module)

	// Check if source directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source code not found for module '%s'\n\nExpected at: %s\n\nUsage: design update <module>\nExample: design update r2r-cli",
			config.Module, config.SourcePath)
	}

	// Check if workspace exists
	workspacePath := config.OutputPath
	if workspacePath == "" {
		workspacePath = paths.WorkspaceDSLPath(config.TemplateRoot, config.Module)
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found for module '%s'\n\nExpected at: %s\n\nCreate one first: design create %s",
			config.Module, workspacePath, config.Module)
	}

	out.Progressf("✅ Source code found at: %s", config.SourcePath)
	out.Progressf("✅ Existing workspace found at: %s", workspacePath)
	return nil
}

// loadExistingWorkspace loads the current workspace content.
func loadExistingWorkspace(config *UpdateConfig, out *designHelper.Output, workspacePath string) (string, error) {
	out.Progress("📄 Loading existing workspace...")

	content, err := os.ReadFile(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to read workspace: %w", err)
	}

	return string(content), nil
}

// buildUpdatePrompt builds the AI prompt for updating the workspace.
func buildUpdatePrompt(config *UpdateConfig, out *designHelper.Output, existingWorkspace string) (string, error) {
	out.Progress("📋 Building update prompt...")

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

	// Build final prompt with module context and existing workspace
	var prompt strings.Builder
	prompt.WriteString(promptTemplate)
	prompt.WriteString("\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n")

	// Module context
	prompt.WriteString("## Target Module\n\n")
	prompt.WriteString("Module: ")
	prompt.WriteString(config.Module)
	prompt.WriteString("\n\n")
	prompt.WriteString("Analyze the source code in the following directory and update the architecture documentation:\n\n")
	prompt.WriteString("Source Path: ")
	prompt.WriteString(config.SourcePath)
	prompt.WriteString("\n\n")

	// Existing workspace context
	prompt.WriteString("## Existing Workspace\n\n")
	prompt.WriteString("Here is the current workspace that needs to be updated:\n\n")
	prompt.WriteString("```\n")
	prompt.WriteString(existingWorkspace)
	prompt.WriteString("\n```\n\n")

	// Update instructions
	prompt.WriteString("## Update Instructions\n\n")
	prompt.WriteString("Examine the module's source code to understand current state:\n")
	prompt.WriteString("- New components and their responsibilities\n")
	prompt.WriteString("- Changed dependencies between components\n")
	prompt.WriteString("- New or removed external systems or services\n")
	prompt.WriteString("- Updated data flow and relationships\n\n")
	prompt.WriteString("Update the workspace to reflect these changes while:\n")
	prompt.WriteString("- Preserving the overall structure and naming conventions\n")
	prompt.WriteString("- Keeping valid existing elements that haven't changed\n")
	prompt.WriteString("- Adding new elements discovered in the code\n")
	prompt.WriteString("- Removing elements that no longer exist\n")
	prompt.WriteString("- Updating relationships to match current code\n\n")

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete updated Structurizr DSL workspace now.\n")
	prompt.WriteString("Return ONLY the DSL content starting with 'workspace' - no markdown fences, no explanations.\n")

	fullPrompt := prompt.String()

	if config.Debug {
		log.Debugf("Full AI prompt built: promptLength=%d", len(fullPrompt))
	}

	return fullPrompt, nil
}

// loadPrompt loads the AI prompt for design updates with three-tier priority:
// 1. Command flag (--prompt)
// 2. Team override (.eac/templates/coreai.TypeDesign/design.md)
// 3. System default (templates/coreai.TypeDesign/design.md)
// Convention: Empty string uses type name (design.md).
func loadPrompt(config *UpdateConfig) (string, error) {
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

// generateUpdatedWorkspace generates the updated workspace using AI.
func generateUpdatedWorkspace(config *UpdateConfig, out *designHelper.Output, prompt string) (string, error) {
	// Check for mock AI response (for testing)
	if mockResponse := createDesign.GetMockAIResponse(); mockResponse != "" {
		out.Progress("🤖 Using mock AI response (test mode)...")
		return mockResponse, nil
	}

	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create composite validator (quick + full validation)
	validator, err := structurizr.NewCompositeValidator(config.Module, config.TemplateRoot, true)
	if err != nil {
		return "", fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Load AI config to get retry strategy
	aiConfig, err := coreai.LoadAIConfig(config.TemplateRoot)
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
		coreai.WithDefaultMaxAttempts(3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry and validation
	out.Progress("🤖 Generating updated architecture design with AI...")

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

// writeUpdatedWorkspace writes the updated workspace file.
func writeUpdatedWorkspace(config *UpdateConfig, out *designHelper.Output, workspacePath, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(workspacePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(workspacePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Report success
	out.Progress("\n✅ Architecture design updated")
	out.Progressf("   File: %s", workspacePath)
	out.Progress("   Valid: ✅ Passed Structurizr validation\n")
	out.Progress("ℹ️  Next steps:")
	out.Progress("   1. Review the updated workspace")
	out.Progress("   2. Check for any breaking changes")
	out.Progressf("   3. View in browser: design serve %s", config.Module)
	out.Progressf("   4. Validate anytime: design validate %s", config.Module)

	return nil
}
