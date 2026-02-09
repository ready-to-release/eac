// Command: create spec
// Short: Generate Gherkin specifications from natural language descriptions
// Long: The create spec command uses AI to transform natural language feature descriptions into
// Long: properly formatted Gherkin specifications following Rules/Scenarios patterns. The generated specifications
// Long: include Feature, Rule, and Scenario blocks with appropriate tags and structure.
// Long: All specifications are validated against the specification contract to ensure they meet quality standards.
// Long: The command automatically saves the specification to the specs/ directory, organized by module.
// Long: Use --debug to inspect intermediate outputs and understand how the AI generates specifications.
// Long:
// Long: Expected Output:
// Long: - Gherkin .feature file in specs/ directory
// Long: - Feature, Rule, and Scenario blocks with tags
// Long: - Validated against specification contract
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs (prompts, raw AI responses, validation results) to the 'out' directory for debugging and analysis
// Flag.module: type=string, shorthand=m, usage=Target module for the specification (e.g., eac-cli, core). If not provided, the module will be inferred from the description
// Flag.output: type=string, shorthand=o, usage=Custom output path for the specification file. If not provided, the path is determined from the feature name and module
// Flag.prompt: type=string, usage=Path to a custom system prompt file. Overrides both user override prompts and built-in prompts
// Usage: create spec <description>
package spec

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/specs"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/clibase/registry"
	aimock "github.com/ready-to-release/eac/go/core/ai"
	configpkg "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/validation/formats/gherkin"
)

var log = logging.C()

func init() {
	registry.Register(CreateSpec)
}

// Intent: Create a Gherkin specification from natural language description using AI
//
// CreateSpec orchestrates the specification generation workflow.
func CreateSpec() int {
	return createSpec(defaultDeps())
}

// createSpec is the internal implementation that accepts injectable dependencies.
func createSpec(deps *Deps) int {
	// Parse configuration
	specsConfig, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Load EAC configuration for path access
	// Skip workflow validation for test environments where workflow files may not exist
	// We only need template path configuration, not full workflow information
	eacCfg, err := configpkg.Load(configpkg.LoadOptions{
		RepoRoot:               specsConfig.TemplateRoot,
		SkipWorkflowValidation: true, // Not needed for content generation
	})
	if err != nil {
		log.Errorf("Error: failed to load EAC config: %v", err)
		return 1
	}
	specsConfig.EACConfig = eacCfg

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(specsConfig.TemplateRoot, "commands", nil, specsConfig.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	// Log command start
	log.Infof("Starting specs create: description=%s, module=%s, debug=%v",
		truncateForLog(specsConfig.Description, 100), specsConfig.Module, specsConfig.Debug)

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(specsConfig)
	if err != nil {
		log.Errorf("Failed to build prompt: %v", err)
		log.Errorf("Error: %v", err)
		return 1
	}

	// Generate and clean output
	cleanedOutput, err := generateAndClean(specsConfig, fullPrompt)
	if err != nil {
		log.Errorf("AI generation failed: %v", err)
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Determine and validate output path
	finalOutputPath, err := determineAndValidateOutputPath(specsConfig, cleanedOutput)
	if err != nil {
		log.Errorf("Output path validation failed: %v", err)
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Check if file exists and --force not specified
	if !specsConfig.Force {
		if _, err := os.Stat(finalOutputPath); err == nil {
			log.Warnf("File already exists: path=%s, force=%v", finalOutputPath, specsConfig.Force)
			log.Errorf("Error: File already exists: %s", finalOutputPath)
			log.Error("Use --force to overwrite")
			return 1
		}
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(finalOutputPath, cleanedOutput, specsConfig); err != nil {
		log.Errorf("Failed to write output: %v", err)
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Log successful completion
	log.Infof("Specification created successfully: path=%s, size=%d", finalOutputPath, len(cleanedOutput))

	return 0
}

// truncateForLog truncates a string for logging purposes.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SpecsConfig holds configuration for specs create command.
type SpecsConfig struct {
	Description  string
	Debug        bool   // -d, --debug: Save intermediate outputs
	Force        bool   // --force: Overwrite existing files
	Module       string // -m, --module: Target module
	OutputPath   string // -o, --output: Custom output path
	PromptPath   string // --prompt: Custom system prompt file
	TemplateRoot string
	EACConfig    *configpkg.EACConfig // Loaded repository configuration
}

// loadAndBuildPrompt loads the contract and builds the AI prompt
//
// This function:
// - Loads specification contract from YAML files
// - Builds the full AI prompt with contract context
// - Optionally saves debug output.
func loadAndBuildPrompt(config *SpecsConfig) (string, error) {
	log.Debug("Loading specification contract")
	log.Info("📋 Loading specification contract...")

	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		log.Errorf("Failed to build contract-based prompt: %v", err)
		return "", err
	}

	log.Debugf("Prompt built successfully: promptLength=%d", len(fullPrompt))

	if config.Debug {
		// Log debug content to logger instead of writing to file
		log.Debugf("=== full-prompt START ===")
		log.Debug(fullPrompt)
		log.Debugf("=== full-prompt END ===")
		log.Errorf("🔍 DEBUG: Full prompt logged to commands.log")
	}

	return fullPrompt, nil
}

// generateAndClean generates AI output with retry and applies anti-corruption filtering
//
// This function:
// - Uses the generic retry framework for AI generation
// - Applies anti-corruption layer to remove noise
// - Validates Gherkin structure with automatic retry on errors
// - Optionally saves debug outputs.
func generateAndClean(config *SpecsConfig, prompt string) (string, error) {
	// Check for subprocess mock response (used in acceptance tests)
	if mock, ok := aimock.GetMockResponse("specs"); ok {
		log.Debug("Using subprocess mock response")
		return mock, nil
	}

	log.Debug("Starting AI generation with retry")

	// Load contract and anti-corruption rules for validator
	loader := aimock.NewContractLoader(config.TemplateRoot, aimock.TypeSpecs, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		log.Errorf("Failed to load contract: %v", err)
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Load tags config for tag validation
	// The EAC config was already loaded in CreateSpec, use it to get tags config
	var tagsConfig *configpkg.TestingTagsConfig
	if config.EACConfig != nil && config.EACConfig.TestingTags != nil {
		tagsConfig = config.EACConfig.TestingTags
	}

	// Create validator with tags config (JSON schema validated in Phase 1)
	validator := gherkin.NewValidator(contractData, tagsConfig)

	// Load AI config to get retry strategy
	var aiConfig *aimock.AIConfig
	aiConfig, err = aimock.LoadAIConfig(config.TemplateRoot)
	if err != nil {
		log.Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	// Build retry configuration using factory
	retryConfig, err := aimock.BuildRetryConfig(
		aimock.TypeSpecs,
		aimock.FormatGherkin,
		executorAdapter,
		validator,
		config.TemplateRoot,
		aiConfig,
		aimock.WithDebug(config.Debug),
		aimock.WithLogger(logging.C().Zap()),
		aimock.WithTagsConfig(tagsConfig),
	)
	if err != nil {
		log.Errorf("Failed to build retry config: %v", err)
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	log.Debugf("Configured retry behavior: maxAttempts=%d, debug=%v", retryConfig.MaxAttempts, config.Debug)

	// Generate with retry
	ctx := context.Background()
	result, err := aimock.GenerateWithRetry(ctx, retryConfig, prompt)
	if err != nil {
		log.Errorf("AI generation failed after retries: %v", err)
		log.Error("\nTroubleshooting:")
		log.Error("  1. Ensure AI provider is configured: clie agent init --ai <provider>")
		log.Error("  2. Check API key environment variable is set")
		log.Error("  3. Verify network connectivity to AI provider")
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	log.Debugf("AI generation completed: attempts=%d, outputLength=%d, validationErrors=%d",
		result.Attempts, len(result.Output), len(result.ValidationErrors))

	// Handle validation errors
	if len(result.ValidationErrors) > 0 {
		criticalErrors := domain.CountCriticalErrors(result.ValidationErrors)

		if criticalErrors > 0 {
			log.Errorf("Generated specification has critical validation errors: criticalErrors=%d, attempts=%d",
				criticalErrors, result.Attempts)
			log.Error("")
			log.Error("⚠️  Generated specification has validation errors:\n")
			log.Errorf("%s", domain.FormatValidationErrors(result.ValidationErrors))
			log.Errorf("\nThe AI attempted %d time(s) but could not generate valid output.", result.Attempts)
			log.Error("\nTroubleshooting:")
			log.Error("  1. Try rephrasing your description to be more specific")
			log.Error("  2. Use --debug to inspect the generated output")
			log.Error("  3. Review the validation errors above and manually fix the output")
			return "", fmt.Errorf("generated specification has %d critical validation error(s)", criticalErrors)
		}

		// Only warnings - log them but continue
		log.Warnf("Generated specification has warnings: warnings=%d", len(result.ValidationErrors))
		log.Errorf("\nℹ️  Generated specification has %d warning(s):", len(result.ValidationErrors))
		log.Errorf("%s", domain.FormatValidationErrors(result.ValidationErrors))
	}

	return result.Output, nil
}

// determineAndValidateOutputPath determines the output file path and validates security
//
// This function:
// - Uses user-specified path or extracts from feature name
// - Validates path is within repository (prevents path traversal)
// - Returns absolute path to output file.
func determineAndValidateOutputPath(config *SpecsConfig, content string) (string, error) {
	var finalOutputPath string

	if config.OutputPath != "" {
		// User specified output path
		finalOutputPath = config.OutputPath
		if !filepath.IsAbs(finalOutputPath) {
			finalOutputPath = filepath.Join(config.TemplateRoot, config.OutputPath)
		}
		log.Debugf("Using user-specified output path: path=%s", finalOutputPath)
	} else {
		// Extract feature name and determine path
		moduleName, featureName, err := specs.ExtractFeatureName(content)
		if err != nil {
			log.Errorf("Failed to extract feature name: %v", err)
			return "", err
		}

		log.Debugf("Extracted feature info: module=%s, feature=%s", moduleName, featureName)

		// Default to specs directory
		finalOutputPath = specs.DetermineOutputPath(config.TemplateRoot, moduleName, featureName, config.EACConfig)
	}

	// Security: Validate that output path is within repository
	if err := specs.ValidateOutputPath(finalOutputPath, config.TemplateRoot); err != nil {
		log.Errorf("Output path security validation failed: path=%s, error=%v", finalOutputPath, err)
		return "", fmt.Errorf("security error: %w", err)
	}

	log.Debugf("Output path validated: path=%s", finalOutputPath)

	return finalOutputPath, nil
}

// writeOutputAndReportSuccess writes the specification file and displays success message
//
// This function:
// - Creates necessary directories
// - Writes specification content to file
// - Displays success message with next steps.
func writeOutputAndReportSuccess(outputPath, content string, config *SpecsConfig) error {
	// Create directories
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Log the write operation
	log.Debugf("Writing specification file: path=%s, size=%d", outputPath, len(content))

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write specification: %w", err)
	}

	// Report success
	log.Info("")
	log.Info("✅ Specification created")
	log.Infof("   File: %s", outputPath)
	log.Info("")
	log.Info("ℹ️  Next steps:")
	log.Info("   1. Review the generated specification")
	log.Info("   2. Refine Rules and Scenarios as needed")
	log.Info("   3. Implement step definitions in src/<module>/tests/")
	log.Info("")

	return nil
}

// parseConfig parses command line arguments into configuration
// specFlags defines valid flags for the create spec command

func parseConfig() (*SpecsConfig, error) {
	config := &SpecsConfig{}
	var description string

	args := os.Args[3:] // Skip program name, "create", and "spec"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	// Parse flags using shared package
	config.Debug = flags.ParseDebugFlag(args)
	config.Force = flags.HasFlag(args, "--force", "")

	// Parse value flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-m", "--module":
			if i+1 < len(args) {
				config.Module = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				config.OutputPath = args[i+1]
				i++
			}
		case "--prompt":
			if i+1 < len(args) {
				config.PromptPath = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(arg, "-") {
				// First non-flag argument is the description
				if description == "" {
					description = arg
				} else {
					description += " " + arg
				}
			}
		}
	}

	// Validate description
	if description == "" {
		return nil, fmt.Errorf("description is required\n\nUsage: create spec <description> [--module <module>] [--output <path>] [--prompt <file>]\nExample: create spec Add user authentication with email and password")
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	// Truncate very long descriptions
	const maxDescLength = 1000
	if len(description) > maxDescLength {
		log.Errorf("⚠️  Warning: Description truncated to %d characters", maxDescLength)
		description = description[:maxDescLength]
	}

	config.Description = description

	// Get repository root
	templateRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w\n\nPlease run this command from within a git repository", err)
	}

	config.TemplateRoot = templateRoot

	// Validate module if specified
	if config.Module != "" {
		moduleReport, err := reports.GetModuleContracts(templateRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to load module contracts: %w", err)
		}

		mod, exists := moduleReport.Registry.Get(config.Module)
		if !exists {
			return nil, fmt.Errorf("module not found: %s\n\nAvailable modules:\n%s",
				config.Module, formatModuleList(moduleReport))
		}

		config.Module = mod.Moniker // Store validated moniker
	}

	return config, nil
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

// loadPromptWithFallback implements prompt loading:
// 1. Custom path (if specified via --prompt flag)
// 2. AI config: .eac/aimock.TypeSpecs/specification.md
// 3. Built-in: embedded prompts/specification.md.
func loadPromptWithFallback(templateRoot, customPath string) (string, error) {
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
		log.Errorf("📋 Using custom prompt: %s", customPath)
		return string(content), nil
	}

	// Load prompt with three-tier priority:
	// 1. Command flag (--prompt)
	// 2. Team override (.eac/templates/aimock.TypeSpecs/specs.md)
	// 3. System default (templates/aimock.TypeSpecs/specs.md)
	// Convention: Empty string uses type name (specs.md)
	loader := aimock.NewContractLoader(templateRoot, aimock.TypeSpecs, "")
	prompt, source, err := loader.LoadPromptWithPriority("", "")
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Log source if not default
	if source != "embedded default" {
		log.Errorf("ℹ️  Using %s prompt", source)
	}

	return prompt, nil
}

// buildContractBasedPrompt loads contract files and builds comprehensive AI prompt.
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

// loadPromptTemplates loads all template files and builds the prompt template.
func loadPromptTemplates(config *SpecsConfig) (string, error) {
	// Load custom or default prompt
	promptContent, err := loadPromptWithFallback(config.TemplateRoot, config.PromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Use generalized contract loader
	loader := aimock.NewContractLoader(config.TemplateRoot, aimock.TypeSpecs, paths.DefaultsVersion)

	// Load contract
	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Load referenced files (tags)
	tagsPath := filepath.Join(domain.EACConfigRelPath, "testing-tags.yml")
	tagsContent, err := loader.LoadReferencedFile(tagsPath)
	if err != nil {
		return "", fmt.Errorf("failed to load tags: %w", err)
	}

	// Load available OSCAL controls for module (if profile exists)
	availableControls := loadModuleControlsContext(config.TemplateRoot, config.Module, config.EACConfig)

	// Build prompt with contract using Go templates
	customData := map[string]string{
		"TagsSpec":          string(tagsContent),
		"AvailableControls": availableControls,
	}

	promptTemplate, err := aimock.BuildPromptWithTemplate(
		promptContent,
		contractData,
		customData,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt template: %w", err)
	}

	return promptTemplate, nil
}

// buildUserInputSection builds the user input section of the prompt.
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
		prompt.WriteString("Common modules include: eac-cli, core, clie, eac-mcp-server\n\n")
	}

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete Gherkin specification now.\n")
	prompt.WriteString("Return ONLY the Gherkin content starting with 'Feature:' - no markdown fences, no explanations.\n")

	return prompt.String()
}

// loadModuleControlsContext loads OSCAL profile and formats controls for AI prompt.
func loadModuleControlsContext(workspaceRoot, moduleName string, cfg *configpkg.EACConfig) string {
	// Only load if module is specified
	if moduleName == "" {
		return "(No module specified - control tags optional)"
	}

	// Get profile path for module
	profilePath := filepath.Join(cfg.Repository.RiskControlsPathAbs(workspaceRoot), moduleName+".profile.json")

	// Check if profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return "(No risk profile found for this module - control tags optional)"
	}

	// Load profile
	profile, err := oscal.LoadProfile(profilePath)
	if err != nil {
		log.Errorf("Warning: Failed to load profile for module %s: %v", moduleName, err)
		return "(Failed to load risk profile - control tags optional)"
	}

	// Load catalog to get control descriptions
	catalogPath := cfg.Repository.RiskCatalogPathAbs(workspaceRoot)
	catalog, err := oscal.LoadCatalog(catalogPath)
	if err != nil {
		log.Errorf("Warning: Failed to load catalog: %v", err)
		return "(Failed to load control catalog - control tags optional)"
	}

	// Get control IDs from profile
	controlIDs := oscal.GetControlIDsFromProfile(profile)
	if len(controlIDs) == 0 {
		return "(Profile contains no controls - control tags optional)"
	}

	// Format controls for AI prompt
	var sb strings.Builder
	sb.WriteString("This module has a risk profile with the following controls:\n\n")
	sb.WriteString("| Control ID | Title | Description |\n")
	sb.WriteString("|------------|-------|-------------|\n")

	for _, controlID := range controlIDs {
		control := oscal.GetControl(catalog, controlID)
		if control != nil {
			// Get title
			title := control.Title
			if title == "" {
				title = "N/A"
			}

			// Get statement description
			desc := oscal.GetControlStatement(control)
			if desc == "" {
				desc = "N/A"
			}

			// Truncate description for prompt
			if len(desc) > 120 {
				desc = desc[:117] + "..."
			}

			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n",
				controlID, title, desc))
		} else {
			// Control not found in catalog
			sb.WriteString(fmt.Sprintf("| `%s` | (Not found in catalog) | |\n", controlID))
		}
	}

	sb.WriteString("\n**When to tag scenarios with these controls:**\n\n")
	sb.WriteString("- If a scenario validates security, access control, or compliance behavior\n")
	sb.WriteString("- Tag with `@control:<id>` if it provides evidence for that control\n")
	sb.WriteString("- Tag with `@controls:<id1>,<id2>` if it covers multiple controls\n")
	sb.WriteString("- Omit control tags for non-security scenarios\n")

	return sb.String()
}
