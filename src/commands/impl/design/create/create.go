// Command: design create
// Description: Generate workspace.dsl for a module using AI
// Short: Generate workspace.dsl for a module using AI
// Long: Generates Structurizr DSL workspace files for a module by analyzing its source code using AI.
// Long: The AI analyzes the source code in src/<module>/ and creates comprehensive architecture
// Long: documentation including system context views (external systems and actors), container views
// Long: (major components), and component views (internal structure). All generated workspaces are
// Long: automatically validated against Structurizr CLI to ensure correct syntax before saving to
// Long: specs/<module>/design/workspace.dsl (or custom path via --output). Use --debug to save AI
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to out/ for debugging
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing workspace.dsl file if it exists
// Flag.output: type=string, shorthand=o, default=, usage=Custom output path for workspace.dsl (default: specs/<module>/design/workspace.dsl)
// Usage: design create <module>
// HasSideEffects: true
package create

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	design "github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/ai/providers"
	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(DesignCreate)
}

// Intent: Create a Structurizr DSL workspace from natural language description using AI
//
// DesignCreate orchestrates the architecture design generation workflow
func DesignCreate() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Validate module exists
	if err := validateModuleExists(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Check Docker availability early (before expensive AI generation)
	if err := checkDockerAvailability(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(config.TemplateRoot, "specs", config.Module, "design", "workspace.dsl")
	}

	// Check if output exists (unless --force)
	if !config.Force && fileExists(outputPath) {
		fmt.Fprintf(os.Stderr, "Error: workspace already exists at %s\n", outputPath)
		fmt.Fprintf(os.Stderr, "Use --force to overwrite\n")
		return 1
	}

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Generate and validate with retry
	cleanedOutput, err := generateAndValidate(config, fullPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(outputPath, cleanedOutput, config.Module); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	return 0
}

// DesignConfig holds configuration for design create command
type DesignConfig struct {
	Module       string // Module name (e.g., "src-cli", "commands")
	SourcePath   string // Path to source code (e.g., "src/commands")
	OutputPath   string // Custom output path (empty = default to specs/<module>/design/workspace.dsl)
	Debug        bool
	Force        bool
	TemplateRoot string
}

// parseConfig parses command line arguments into DesignConfig
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
		Debug:        flags.debug,
		Force:        flags.force,
		OutputPath:   flags.outputPath,
		TemplateRoot: repoRoot,
	}

	// Clean module name and determine paths
	config.Module = design.CleanModuleName(modulePath)

	// Validate module name
	if err := design.ValidateModuleName(config.Module); err != nil {
		return nil, fmt.Errorf("invalid module name: %w", err)
	}

	config.SourcePath = determineSourcePath(repoRoot, config.Module, modulePath)

	return config, nil
}

// createFlags holds parsed command-line flags
type createFlags struct {
	debug      bool
	force      bool
	outputPath string
}

// parseCreateCommandArgs parses args for the create subcommand
// Returns: module path, flags, error
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

	// Parse flags and positional arguments
	flags := &createFlags{}
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
		} else if !strings.HasPrefix(arg, "-") {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	// Validate we have a module name
	if len(positionalArgs) == 0 {
		return "", nil, fmt.Errorf("module name required\n\nUsage: design create <module>\nExample: design create src-cli")
	}

	return positionalArgs[0], flags, nil
}

// determineSourcePath finds the source code directory for a module
// Tries src/<module> first, then falls back to <modulePath> if it exists
func determineSourcePath(repoRoot, module, modulePath string) string {
	// Try src/<module> first (standard location)
	sourcePath := filepath.Join(repoRoot, "src", module)
	if _, err := os.Stat(sourcePath); err == nil {
		return sourcePath
	}

	// Try the original module path (in case it's a custom location)
	altPath := filepath.Join(repoRoot, modulePath)
	if _, err := os.Stat(altPath); err == nil {
		return altPath
	}

	// Return the standard path even if it doesn't exist
	// (validation will catch this later)
	return sourcePath
}

// validateModuleExists checks if the source code exists for the specified module
func validateModuleExists(config *DesignConfig) error {
	fmt.Printf("🔍 Validating source code for module '%s'...\n", config.Module)

	// Check if source directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source code not found for module '%s'\n\nExpected at: %s\n\nUsage: design create <module>\nExample: design create src-cli (analyzes code in src/src-cli/)",
			config.Module, config.SourcePath)
	}

	fmt.Printf("✅ Source code found at: %s\n", config.SourcePath)
	return nil
}

// checkDockerAvailability checks if Docker is running before starting expensive operations
func checkDockerAvailability() error {
	fmt.Println("🐳 Checking Docker availability...")

	// Create a temporary validator to check Docker
	validator, err := NewCompositeValidator("", "", true)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Check if Docker is running
	if !validator.IsDockerRunning() {
		return fmt.Errorf("Docker is not running\n\nStructurizr validation requires Docker to be running.\nPlease start Docker and try again.")
	}

	fmt.Println("✅ Docker is available")
	return nil
}

// loadAndBuildPrompt loads the contract and builds the AI prompt
func loadAndBuildPrompt(config *DesignConfig) (string, error) {
	fmt.Println("📋 Loading design contract...")
	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		return "", err
	}

	if config.Debug {
		debugPath := filepath.Join(config.TemplateRoot, "out", "design-debug-full-prompt.md")
		if err := os.WriteFile(debugPath, []byte(fullPrompt), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  DEBUG: Failed to save prompt to %s: %v\n", debugPath, err)
		} else {
			fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved full prompt to %s\n", debugPath)
		}
	}

	return fullPrompt, nil
}

// buildContractBasedPrompt builds the AI prompt with contract context
func buildContractBasedPrompt(config *DesignConfig) (string, error) {
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

	// Load default prompt
	promptContent, err := loadPrompt(config.TemplateRoot)
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

// loadPrompt loads the default AI prompt for design generation
func loadPrompt(templateRoot string) (string, error) {
	// Load from repo path
	repoPath := filepath.Join(templateRoot, "contracts", "ai", "design", "0.1.0", "design.md")
	content, err := os.ReadFile(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt from %s: %w", repoPath, err)
	}

	return string(content), nil
}

// generateAndValidate generates AI output with retry and validates with Structurizr CLI
func generateAndValidate(config *DesignConfig, prompt string) (string, error) {
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
	// Skip expensive Docker validation if quick validation finds errors
	validator, err := NewCompositeValidator(config.Module, config.TemplateRoot, true)
	if err != nil {
		return "", fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Generate with retry and validation
	fmt.Println("🤖 Generating architecture design with AI...")

	retryConfig := &contracts.RetryConfig{
		Executor:       executorAdapter,
		Validator:      validator,
		AntiCorruption: antiCorruptionRules,
		ContentMarker:  "workspace",
		MaxAttempts:    3,
		Debug:          config.Debug,
	}

	if config.Debug {
		retryConfig.DebugOutputDir = filepath.Join(config.TemplateRoot, "out")
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

// writeOutputAndReportSuccess writes the workspace file and reports success
func writeOutputAndReportSuccess(outputPath, content, module string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Report success
	fmt.Println("\n✅ Architecture design created")
	fmt.Printf("   File: %s\n", outputPath)
	fmt.Printf("   Valid: ✅ Passed Structurizr validation\n\n")
	fmt.Println("ℹ️  Next steps:")
	fmt.Println("   1. Review the generated workspace")
	fmt.Println("   2. Refine containers and relationships as needed")
	fmt.Printf("   3. View in browser: design serve %s\n", module)
	fmt.Printf("   4. Validate anytime: design validate %s\n", module)

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

