// Command: specs create
// Description: Create a new specification using AI from a natural language description
// Short: Generate Gherkin specifications from natural language descriptions
// Long: The specs create command uses AI to transform natural language feature descriptions into
// Long: properly formatted Gherkin specifications following Rules/Scenarios patterns. The generated specifications
// Long: include Feature, Rule, and Scenario blocks with appropriate tags and structure.
// Long: All specifications are validated against the specification contract to ensure they meet quality standards.
// Long: The command automatically saves the specification to the specs/ directory, organized by module.
// Long: Use --debug to inspect intermediate outputs and understand how the AI generates specifications.
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to the 'out' directory for debugging and analysis
// Flag.module: type=string, shorthand=m, usage=Target module for the specification (e.g., src-commands, src-core). If not provided, the module will be inferred from the description
// Flag.output: type=string, shorthand=o, usage=Custom output path for the specification file. If not provided, the path is determined from the feature name and module
// Flag.template: type=string, usage=Path to a custom template file for specification generation. Overrides the built-in template
// Flag.prompt: type=string, usage=Path to a custom system prompt file. Overrides both user override prompts and built-in prompts
// Usage: specs create <description>
// Flags: --debug (save intermediate outputs), --module (target module), --output (output path), --template (custom template file), --prompt (custom system prompt)
// HasSideEffects: true
package create

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/src/commands/impl/specs"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/ai"
	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/ai/providers"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(SpecsCreate)
}

// Intent: Create a Gherkin specification from natural language description using AI
//
// SpecsCreate orchestrates the specification generation workflow
func SpecsCreate() int {
	// Parse configuration
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Generate and clean output
	cleanedOutput, err := generateAndClean(config, fullPrompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Determine and validate output path
	finalOutputPath, err := determineAndValidateOutputPath(config, cleanedOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(finalOutputPath, cleanedOutput); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return 1
	}

	return 0
}

// SpecsConfig holds configuration for specs create command
type SpecsConfig struct {
	Description  string
	Debug        bool
	Module       string
	OutputPath   string
	PromptPath   string
	TemplateRoot string
}

// loadAndBuildPrompt loads the contract and builds the AI prompt
//
// This function:
// - Loads specification contract from YAML files
// - Builds the full AI prompt with contract context
// - Optionally saves debug output
func loadAndBuildPrompt(config *SpecsConfig) (string, error) {
	fmt.Println("📋 Loading specification contract...")
	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		return "", err
	}

	if config.Debug {
		debugPath := filepath.Join(config.TemplateRoot, "out", "debug-full-prompt.md")
		if err := os.WriteFile(debugPath, []byte(fullPrompt), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  DEBUG: Failed to save prompt to %s: %v\n", debugPath, err)
		} else {
			fmt.Fprintf(os.Stderr, "🔍 DEBUG: Saved full prompt to %s\n", debugPath)
		}
	}

	return fullPrompt, nil
}

// generateAndClean generates AI output with retry and applies anti-corruption filtering
//
// This function:
// - Uses the generic retry framework for AI generation
// - Applies anti-corruption layer to remove noise
// - Validates Gherkin structure with automatic retry on errors
// - Optionally saves debug outputs
func generateAndClean(config *SpecsConfig, prompt string) (string, error) {
	// Load contract and anti-corruption rules for validator
	loader := contracts.NewContractLoader(config.TemplateRoot, "ai/specifications", "0.1.0")

	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		return "", fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create validator
	validator := contracts.NewGherkinValidator(contractData, antiCorruptionRules)

	// Setup debug directory if needed
	debugOutputDir := ""
	if config.Debug {
		debugOutputDir = filepath.Join(config.TemplateRoot, "out")
		// Ensure debug directory exists
		if err := os.MkdirAll(debugOutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to create debug directory: %v\n", err)
		}
	}

	// Configure retry behavior
	retryConfig := &contracts.RetryConfig{
		Executor:       executorAdapter,
		Validator:      validator,
		PromptBuilder:  &contracts.DefaultRetryPromptBuilder{},
		AntiCorruption: antiCorruptionRules,
		ContentMarker:  "Feature:",
		MaxAttempts:    2,
		Debug:          config.Debug,
		DebugOutputDir: debugOutputDir,
	}

	// Generate with retry
	ctx := context.Background()
	result, err := contracts.GenerateWithRetry(ctx, retryConfig, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nTroubleshooting:\n")
		fmt.Fprintf(os.Stderr, "  1. Ensure AI provider is configured: r2r agent init --ai <provider>\n")
		fmt.Fprintf(os.Stderr, "  2. Check API key environment variable is set\n")
		fmt.Fprintf(os.Stderr, "  3. Verify network connectivity to AI provider\n")
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Handle validation errors
	if len(result.ValidationErrors) > 0 {
		criticalErrors := contracts.CountCriticalErrors(result.ValidationErrors)

		if criticalErrors > 0 {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "⚠️  Generated specification has validation errors:\n\n")
			fmt.Fprintf(os.Stderr, "%s\n", contracts.FormatValidationErrors(result.ValidationErrors))
			fmt.Fprintf(os.Stderr, "\nThe AI attempted %d time(s) but could not generate valid output.\n", result.Attempts)
			fmt.Fprintf(os.Stderr, "\nTroubleshooting:\n")
			fmt.Fprintf(os.Stderr, "  1. Try rephrasing your description to be more specific\n")
			fmt.Fprintf(os.Stderr, "  2. Use --debug to inspect the generated output\n")
			fmt.Fprintf(os.Stderr, "  3. Review the validation errors above and manually fix the output\n")
			return "", fmt.Errorf("generated specification has %d critical validation error(s)", criticalErrors)
		}

		// Only warnings - log them but continue
		fmt.Fprintf(os.Stderr, "\nℹ️  Generated specification has %d warning(s):\n", len(result.ValidationErrors))
		fmt.Fprintf(os.Stderr, "%s\n", contracts.FormatValidationErrors(result.ValidationErrors))
	}

	return result.Output, nil
}

// determineAndValidateOutputPath determines the output file path and validates security
//
// This function:
// - Uses user-specified path or extracts from feature name
// - Validates path is within repository (prevents path traversal)
// - Returns absolute path to output file
func determineAndValidateOutputPath(config *SpecsConfig, content string) (string, error) {
	var finalOutputPath string

	if config.OutputPath != "" {
		// User specified output path
		finalOutputPath = config.OutputPath
		if !filepath.IsAbs(finalOutputPath) {
			finalOutputPath = filepath.Join(config.TemplateRoot, config.OutputPath)
		}
	} else {
		// Extract feature name and determine path
		moduleName, featureName, err := specs.ExtractFeatureName(content)
		if err != nil {
			return "", err
		}

		// Default to specs directory
		finalOutputPath = specs.DetermineOutputPath(config.TemplateRoot, moduleName, featureName)
	}

	// Security: Validate that output path is within repository
	if err := specs.ValidateOutputPath(finalOutputPath, config.TemplateRoot); err != nil {
		return "", fmt.Errorf("security error: %w", err)
	}

	return finalOutputPath, nil
}

// writeOutputAndReportSuccess writes the specification file and displays success message
//
// This function:
// - Creates necessary directories
// - Writes specification content to file
// - Displays success message with next steps
func writeOutputAndReportSuccess(outputPath string, content string) error {
	// Create directories
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write specification: %w", err)
	}

	// Report success
	fmt.Println()
	fmt.Println("✅ Specification created")
	fmt.Printf("   File: %s\n", outputPath)
	fmt.Println()
	fmt.Println("ℹ️  Next steps:")
	fmt.Println("   1. Review the generated specification")
	fmt.Println("   2. Refine Rules and Scenarios as needed")
	fmt.Println("   3. Implement step definitions in src/<module>/tests/")
	fmt.Println()

	return nil
}

// parseConfig parses command line arguments into configuration
func parseConfig() (*SpecsConfig, error) {
	config := &SpecsConfig{}
	var description string

	args := os.Args[3:] // Skip program name, "specs", and "create"
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--debug" {
			config.Debug = true
		} else if arg == "--module" && i+1 < len(args) {
			config.Module = args[i+1]
			i++
		} else if arg == "--output" && i+1 < len(args) {
			config.OutputPath = args[i+1]
			i++
		} else if arg == "--prompt" && i+1 < len(args) {
			config.PromptPath = args[i+1]
			i++
		} else if !strings.HasPrefix(arg, "--") {
			// First non-flag argument is the description
			if description == "" {
				description = arg
			} else {
				description += " " + arg
			}
		}
	}

	// Validate description
	if description == "" {
		return nil, fmt.Errorf("description is required\n\nUsage: specs create <description> [--module <module>] [--output <path>] [--prompt <file>]\nExample: specs create Add user authentication with email and password")
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	// Truncate very long descriptions
	const maxDescLength = 1000
	if len(description) > maxDescLength {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Description truncated to %d characters\n", maxDescLength)
		description = description[:maxDescLength]
	}

	config.Description = description

	// Get repository root
	templateRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}

	config.TemplateRoot = templateRoot

	return config, nil
}

// loadPromptWithFallback implements four-tier prompt loading:
// 1. Custom path (if specified via --prompt flag)
// 2. Local contract: .r2r/contracts/ai/specifications/0.1.0/specification.md
// 3. Repo contract: contracts/ai/specifications/0.1.0/specification.md
// 4. Built-in: embedded prompts/specification.md
func loadPromptWithFallback(templateRoot string, customPath string) (string, error) {
	// Tier 1: Check for custom path (from --prompt flag)
	if customPath != "" {
		// Make path absolute if relative
		if !filepath.IsAbs(customPath) {
			customPath = filepath.Join(templateRoot, customPath)
		}

		content, err := os.ReadFile(customPath)
		if err != nil {
			return "", fmt.Errorf("custom prompt not found: %w\n\nSpecified path: %s", err, customPath)
		}
		fmt.Fprintf(os.Stderr, "📋 Using custom prompt: %s\n", customPath)
		return string(content), nil
	}

	// Load from contract with fallback: .r2r/contracts → contracts/ai
	loader := contracts.NewContractLoader(templateRoot, "ai/specifications", "0.1.0")
	prompt, source, err := loader.LoadPrompt("specification.md", "")
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Log source if using override (for transparency)
	if source != "embedded default" && source != "repository contract" {
		fmt.Fprintf(os.Stderr, "ℹ️  Using %s prompt\n", source)
	}

	return prompt, nil
}

// generateWithAI invokes the AI provider with the prompt
func generateWithAI(templateRoot string, prompt string, debug bool) (string, error) {
	// Check for test double mode
	if mockResponse := os.Getenv("R2R_TEST_AI_RESPONSE"); mockResponse != "" {
		return mockResponse, nil
	}

	// Create executor
	executor := ai.NewExecutor(templateRoot)
	providers.RegisterBuiltIn(executor)

	// Execute
	ctx := context.Background()
	var opts []ai.Option
	if debug {
		opts = append(opts, ai.WithDebug(true))
	}

	output, err := executor.Execute(ctx, prompt, opts...)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	return strings.TrimSpace(output), nil
}

// stripAgentNoise removes initialization messages and markdown code fences
func stripAgentNoise(output string) string {
	output = strings.TrimSpace(output)

	// Remove wrapping markdown code fences if entire output is wrapped
	if strings.HasPrefix(output, "```") {
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			// Remove opening fence (may have language identifier like ```gherkin)
			lines = lines[1:]
		}
		// Remove closing fence
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		output = strings.Join(lines, "\n")
	}

	// Remove common initialization messages
	lines := strings.Split(output, "\n")
	var cleaned []string
	foundFeature := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip noise before Feature: declaration
		if !foundFeature {
			// Skip initialization messages
			if strings.Contains(trimmed, "**Initialized") ||
				strings.Contains(trimmed, "ready to assist") ||
				strings.Contains(trimmed, "I'll help") ||
				strings.Contains(trimmed, "I'll create") ||
				strings.Contains(trimmed, "I'll generate") ||
				strings.Contains(trimmed, "Here's") ||
				strings.Contains(trimmed, "Here is") {
				continue
			}

			// Skip horizontal rules before content
			if trimmed == "---" || trimmed == "___" || trimmed == "***" {
				continue
			}

			// Skip empty lines before content
			if trimmed == "" {
				continue
			}

			// Found Feature: - content starts
			if strings.HasPrefix(trimmed, "Feature:") {
				foundFeature = true
			}
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// buildContractBasedPrompt loads contract files and builds comprehensive AI prompt
func buildContractBasedPrompt(config *SpecsConfig) (string, error) {
	// Load prompt templates (contract, anti-corruption, referenced files)
	promptTemplate, err := loadPromptTemplates(config)
	if err != nil {
		return "", err
	}

	// Build user input section
	userInput := buildUserInputSection(config)

	// Combine template and user input
	fullPrompt := promptTemplate + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + userInput

	return fullPrompt, nil
}

// loadPromptTemplates loads all template files and builds the prompt template
func loadPromptTemplates(config *SpecsConfig) (string, error) {
	// Load custom or default prompt
	promptContent, err := loadPromptWithFallback(config.TemplateRoot, config.PromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Use generalized contract loader
	loader := contracts.NewContractLoader(config.TemplateRoot, "ai/specifications", "0.1.0")

	// Load contract and anti-corruption rules
	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	antiCorruption, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		return "", fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Load referenced files (tags and taxonomy)
	tagsContent, err := loader.LoadReferencedFile("contracts/testing/0.1.0/tags.yml")
	if err != nil {
		return "", fmt.Errorf("failed to load tags: %w", err)
	}

	taxonomyContent, err := loader.LoadReferencedFile("contracts/testing/0.1.0/taxonomy.yml")
	if err != nil {
		return "", fmt.Errorf("failed to load taxonomy: %w", err)
	}

	// Build prompt with contract using Go templates
	customData := map[string]string{
		"TagsSpec":     string(tagsContent),
		"TaxonomySpec": string(taxonomyContent),
	}

	promptTemplate, err := contracts.BuildPromptWithTemplate(
		promptContent,
		contractData,
		antiCorruption,
		customData,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt template: %w", err)
	}

	return promptTemplate, nil
}

// buildUserInputSection builds the user input section of the prompt
func buildUserInputSection(config *SpecsConfig) string {
	var prompt strings.Builder

	// User description
	prompt.WriteString("## Description\n\n")
	prompt.WriteString(config.Description)
	prompt.WriteString("\n\n")

	// Module context (if provided)
	if config.Module != "" {
		prompt.WriteString("## Target Module\n\n")
		prompt.WriteString("The specification MUST be for the following module:\n\n")
		prompt.WriteString(config.Module)
		prompt.WriteString("\n\n")
		prompt.WriteString("Use the naming format: ")
		prompt.WriteString(config.Module)
		prompt.WriteString("_<feature-name>\n\n")
	} else {
		prompt.WriteString("## Module Inference\n\n")
		prompt.WriteString("No specific module was provided. Please infer the most appropriate module based on the description.\n")
		prompt.WriteString("Common modules include: src-commands, src-core, src-cli, src-mcp-commands\n\n")
	}

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete Gherkin specification now.\n")
	prompt.WriteString("Return ONLY the Gherkin content starting with 'Feature:' - no markdown fences, no explanations.\n")

	return prompt.String()
}

// stripAgentNoiseWithContract applies anti-corruption rules from contract using generalized framework
func stripAgentNoiseWithContract(output string, templateRoot string) string {
	// Use generalized contract loader
	loader := contracts.NewContractLoader(templateRoot, "ai/specifications", "0.1.0")

	// Try to load anti-corruption rules
	rules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		// Fall back to original stripAgentNoise if contract not available
		return stripAgentNoise(output)
	}

	// Use generalized anti-corruption filter with "Feature:" as content start marker
	return contracts.ApplyWithFallback(output, rules, "Feature:")
}

