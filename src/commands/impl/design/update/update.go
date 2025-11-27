// Command: design update
// Description: Update existing workspace.dsl for a module using AI
// Short: Update existing workspace.dsl for a module using AI
// Long: Updates an existing Structurizr DSL workspace file for a module by analyzing its current
// Long: source code using AI. The AI re-analyzes the source code in src/<module>/ and updates
// Long: the architecture documentation to reflect current state, preserving the overall structure
// Long: while incorporating new components, relationships, or changes. All updated workspaces are
// Long: automatically validated against Structurizr CLI to ensure correct syntax before saving.
// Long: Use --debug to save intermediate outputs to out/logs/design/ for debugging.
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to out/logs/design/ for debugging
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite workspace.dsl even if validation fails
// Flag.output: type=string, shorthand=o, default=, usage=Custom output path for workspace.dsl (default: specs/<module>/design/workspace.dsl)
// Flag.prompt: type=string, shorthand=, default=, usage=Custom AI prompt file path
// Usage: design update <module>
// HasSideEffects: true
package update

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	design "github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/impl/design/create"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/ai"
	"github.com/ready-to-release/eac/src/ai/providers"
	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(DesignUpdate)
}

// Intent: Update an existing Structurizr DSL workspace by re-analyzing source code
//
// DesignUpdate orchestrates the architecture design update workflow
func DesignUpdate() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate module and existing workspace
	if err := validateModuleAndWorkspace(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Determine workspace path
	workspacePath := config.OutputPath
	if workspacePath == "" {
		workspacePath = filepath.Join(config.TemplateRoot, "specs", config.Module, ".design", "workspace.dsl")
	}

	// Load existing workspace
	existingWorkspace, err := loadExistingWorkspace(config, workspacePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build update prompt with existing workspace context
	fullPrompt, err := buildUpdatePrompt(config, existingWorkspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Generate updated workspace
	updatedWorkspace, err := generateUpdatedWorkspace(config, fullPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Write output and report success
	if err := writeUpdatedWorkspace(config, workspacePath, updatedWorkspace); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	return 0
}

// UpdateConfig holds configuration for design update command
type UpdateConfig struct {
	Module       string // Module name (e.g., "src-cli", "commands")
	SourcePath   string // Path to source code (e.g., "src/commands")
	OutputPath   string // Custom output path (empty = default to specs/<module>/design/workspace.dsl)
	PromptPath   string // Custom AI prompt file path (empty = default prompt)
	Debug        bool
	Force        bool
	TemplateRoot string
}

// parseConfig parses command line arguments into UpdateConfig
func parseConfig() (*UpdateConfig, error) {
	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	// Find and parse arguments
	modulePath, flags, err := parseUpdateCommandArgs(os.Args)
	if err != nil {
		return nil, err
	}

	// Create config with parsed flags
	config := &UpdateConfig{
		Debug:        flags.debug,
		Force:        flags.force,
		OutputPath:   flags.outputPath,
		PromptPath:   flags.promptPath,
		TemplateRoot: repoRoot,
	}

	// Use input as-is (moniker)
	config.Module = modulePath

	// Validate module name for security
	if err := design.ValidateModuleName(config.Module); err != nil {
		return nil, fmt.Errorf("invalid module name: %w", err)
	}

	// Load module contracts and validate moniker exists (same as build command)
	moduleReport, err := reports.GetModuleContracts(repoRoot, "0.1.0")
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
	config.SourcePath = filepath.Join(repoRoot, module.Source.Root)

	return config, nil
}

// updateFlags holds parsed command-line flags
type updateFlags struct {
	debug      bool
	force      bool
	outputPath string
	promptPath string
}

// parseUpdateCommandArgs parses args for the update subcommand
// Returns: module path, flags, error
func parseUpdateCommandArgs(args []string) (string, *updateFlags, error) {
	// Find command position
	cmdPos := -1
	for i, arg := range args {
		if arg == "update" {
			cmdPos = i
			break
		}
	}

	if cmdPos == -1 {
		return "", nil, fmt.Errorf("invalid command structure")
	}

	// Parse flags and positional arguments
	flags := &updateFlags{}
	var positionalArgs []string

	for i := cmdPos + 1; i < len(args); i++ {
		arg := args[i]

		if arg == "--debug" || arg == "-d" {
			flags.debug = true
		} else if arg == "--force" || arg == "-f" {
			flags.force = true
		} else if arg == "--output" || arg == "-o" {
			// Next argument is the output path
			if i+1 < len(args) {
				flags.outputPath = args[i+1]
				i++ // Skip next arg
			}
		} else if arg == "--prompt" {
			// Next argument is the prompt path
			if i+1 < len(args) {
				flags.promptPath = args[i+1]
				i++ // Skip next arg
			}
		} else if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Validate we have a module name
	if len(positionalArgs) == 0 {
		return "", nil, fmt.Errorf("module name required\n\nUsage: design update <module>\nExample: design update src-cli")
	}

	return positionalArgs[0], flags, nil
}

// formatModuleList returns a formatted list of available modules
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, mod.Source.Root))
	}
	return sb.String()
}

// validateModuleAndWorkspace checks if the source code and existing workspace exist
func validateModuleAndWorkspace(config *UpdateConfig) error {
	fmt.Printf("🔍 Validating module '%s'...\n", config.Module)

	// Check if source directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source code not found for module '%s'\n\nExpected at: %s\n\nUsage: design update <module>\nExample: design update src-cli",
			config.Module, config.SourcePath)
	}

	// Check if workspace exists
	workspacePath := config.OutputPath
	if workspacePath == "" {
		workspacePath = filepath.Join(config.TemplateRoot, "specs", config.Module, ".design", "workspace.dsl")
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found for module '%s'\n\nExpected at: %s\n\nCreate one first: design create %s",
			config.Module, workspacePath, config.Module)
	}

	fmt.Printf("✅ Source code found at: %s\n", config.SourcePath)
	fmt.Printf("✅ Existing workspace found at: %s\n", workspacePath)
	return nil
}

// loadExistingWorkspace loads the current workspace content
func loadExistingWorkspace(config *UpdateConfig, workspacePath string) (string, error) {
	fmt.Println("📄 Loading existing workspace...")

	content, err := os.ReadFile(workspacePath)
	if err != nil {
		return "", fmt.Errorf("failed to read workspace: %w", err)
	}

	return string(content), nil
}

// buildUpdatePrompt builds the AI prompt for updating the workspace
func buildUpdatePrompt(config *UpdateConfig, existingWorkspace string) (string, error) {
	fmt.Println("📋 Building update prompt...")

	// Load contract using generalized loader
	loader := contracts.NewContractLoader(config.TemplateRoot, "ai/design", "0.1.0")

	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	antiCorruption, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		return "", fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Load prompt (default or custom)
	promptContent, err := loadPrompt(config)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Build prompt template
	promptTemplate, err := contracts.BuildPromptWithTemplate(
		promptContent,
		contractData,
		antiCorruption,
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
		writeDebugFile(config.TemplateRoot, "update-debug-full-prompt.md", fullPrompt)
	}

	return fullPrompt, nil
}

// loadPrompt loads the AI prompt for design generation
func loadPrompt(config *UpdateConfig) (string, error) {
	// If custom prompt specified, load that
	if config.PromptPath != "" {
		content, err := os.ReadFile(config.PromptPath)
		if err != nil {
			return "", fmt.Errorf("failed to read custom prompt from %s: %w", config.PromptPath, err)
		}
		return string(content), nil
	}

	// Load default prompt from repo path
	repoPath := filepath.Join(config.TemplateRoot, "contracts", "ai", "design", "0.1.0", "design.md")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt from %s: %w", repoPath, err)
	}

	return string(content), nil
}

// generateUpdatedWorkspace generates the updated workspace using AI
func generateUpdatedWorkspace(config *UpdateConfig, prompt string) (string, error) {
	// Check for mock AI response (for testing)
	if mockResponse := create.GetMockAIResponse(); mockResponse != "" {
		fmt.Println("🤖 Using mock AI response (test mode)...")
		return mockResponse, nil
	}

	// Load contract and anti-corruption rules for validator
	loader := contracts.NewContractLoader(config.TemplateRoot, "ai/design", "0.1.0")

	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		return "", fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create composite validator (quick + full validation)
	validator, err := create.NewCompositeValidator(config.Module, config.TemplateRoot, true)
	if err != nil {
		return "", fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Generate with retry and validation
	fmt.Println("🤖 Generating updated architecture design with AI...")

	retryConfig := &contracts.RetryConfig{
		Executor:       executorAdapter,
		Validator:      validator,
		AntiCorruption: antiCorruptionRules,
		ContentMarker:  "workspace",
		MaxAttempts:    3,
		Debug:          config.Debug,
	}

	if config.Debug {
		retryConfig.DebugOutputDir = filepath.Join(config.TemplateRoot, "out", "logs", "design")
	}

	result, err := contracts.GenerateWithRetry(
		context.Background(),
		retryConfig,
		prompt,
	)

	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return result.Output, nil
}

// writeUpdatedWorkspace writes the updated workspace file
func writeUpdatedWorkspace(config *UpdateConfig, workspacePath, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(workspacePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(workspacePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Report success
	fmt.Println("\n✅ Architecture design updated")
	fmt.Printf("   File: %s\n", workspacePath)
	fmt.Printf("   Valid: ✅ Passed Structurizr validation\n\n")
	fmt.Println("ℹ️  Next steps:")
	fmt.Println("   1. Review the updated workspace")
	fmt.Println("   2. Check for any breaking changes")
	fmt.Printf("   3. View in browser: design serve %s\n", config.Module)
	fmt.Printf("   4. Validate anytime: design validate %s\n", config.Module)

	return nil
}

// writeDebugFile writes content to a debug file when debug mode is enabled.
func writeDebugFile(workspaceRoot string, filename string, content string) {
	debugDir := filepath.Join(workspaceRoot, "out", "logs", "design")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		return
	}
	debugFile := filepath.Join(debugDir, filename)
	os.WriteFile(debugFile, []byte(content), 0644)
}
