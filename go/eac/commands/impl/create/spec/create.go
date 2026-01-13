// Command: create spec
// Description: Create a new specification using AI from a natural language description
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
// Flag.module: type=string, shorthand=m, usage=Target module for the specification (e.g., eac-commands, eac-core). If not provided, the module will be inferred from the description
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

	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/impl/specs"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/validation/formats/gherkin"
)

var log = logging.C()

func init() {
	registry.Register(CreateSpec)
}

// ============================================================================
// Mock Support for Testing
// ============================================================================

// mockAIResponse holds the mock response for testing. When set, AI calls return this.
var mockAIResponse string

// SetMockAIResponse sets a mock AI response for testing.
func SetMockAIResponse(response string) {
	mockAIResponse = response
}

// ResetMockAIResponse clears the mock AI response.
func ResetMockAIResponse() {
	mockAIResponse = ""
}

// gitRepo holds the git repository instance for git operations.
// In production, this is initialized lazily. For tests, it can be injected via SetGitRepo.
var gitRepo git.GitRepository

// getGitRepo returns the git repository, creating one if needed
func getGitRepo(workspaceRoot string) (git.GitRepository, error) {
	if gitRepo != nil {
		return gitRepo, nil
	}
	return git.Open(workspaceRoot)
}

// SetGitRepo allows tests to inject a mock repository.
func SetGitRepo(repo git.GitRepository) {
	gitRepo = repo
}

// ResetGitRepo clears the mock git repository.
func ResetGitRepo() {
	gitRepo = nil
}

// ============================================================================

// Intent: Create a Gherkin specification from natural language description using AI
//
// CreateSpec orchestrates the specification generation workflow
func CreateSpec() int {
	// Parse configuration
	specsConfig, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Load EAC configuration for path access
	// Skip workflow validation for test environments where workflow files may not exist
	// We only need template path configuration, not full workflow information
	eacCfg, err := config.Load(config.LoadOptions{
		RepoRoot:               specsConfig.TemplateRoot,
		SkipWorkflowValidation: true, // Not needed for content generation
	})
	if err != nil {
		log.Errorf("Error: failed to load EAC config: %v", err)
		return 1
	}
	specsConfig.EACConfig = eacCfg

	// Initialize logger (logs to out/logs/specs/)
	var logger *logging.Logger
	if specsConfig.Debug {
		logger, err = logging.NewWithDebug("specs", specsConfig.TemplateRoot)
	} else {
		logger, err = logging.NewDefault("specs", specsConfig.TemplateRoot)
	}
	if err != nil {
		log.Errorf("Warning: Failed to initialize logger: %v", err)
		// Continue without logger - not fatal
	} else {
		specsConfig.Logger = logger
		defer logger.Sync()
	}

	// Log command start
	if specsConfig.Logger != nil {
		specsConfig.Logger.Info("Starting specs create",
			zap.String("description", truncateForLog(specsConfig.Description, 100)),
			zap.String("module", specsConfig.Module),
			zap.Bool("debug", specsConfig.Debug))
	}

	// Load contract and build prompt
	fullPrompt, err := loadAndBuildPrompt(specsConfig)
	if err != nil {
		if specsConfig.Logger != nil {
			specsConfig.Logger.Error("Failed to build prompt", zap.Error(err))
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Generate and clean output
	cleanedOutput, err := generateAndClean(specsConfig, fullPrompt)
	if err != nil {
		if specsConfig.Logger != nil {
			specsConfig.Logger.Error("AI generation failed", zap.Error(err))
		}
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Determine and validate output path
	finalOutputPath, err := determineAndValidateOutputPath(specsConfig, cleanedOutput)
	if err != nil {
		if specsConfig.Logger != nil {
			specsConfig.Logger.Error("Output path validation failed", zap.Error(err))
		}
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Check if file exists and --force not specified
	if !specsConfig.Force {
		if _, err := os.Stat(finalOutputPath); err == nil {
			if specsConfig.Logger != nil {
				specsConfig.Logger.Warn("File already exists",
					zap.String("path", finalOutputPath),
					zap.Bool("force", specsConfig.Force))
			}
			log.Errorf("Error: File already exists: %s", finalOutputPath)
			log.Error("Use --force to overwrite")
			return 1
		}
	}

	// Write output and report success
	if err := writeOutputAndReportSuccess(finalOutputPath, cleanedOutput, specsConfig); err != nil {
		if specsConfig.Logger != nil {
			specsConfig.Logger.Error("Failed to write output", zap.Error(err))
		}
		log.Errorf("\n❌ Error: %v", err)
		return 1
	}

	// Log successful completion
	if specsConfig.Logger != nil {
		specsConfig.Logger.Info("Specification created successfully",
			zap.String("path", finalOutputPath),
			zap.Int("size", len(cleanedOutput)))
	}

	return 0
}

// truncateForLog truncates a string for logging purposes
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SpecsConfig holds configuration for specs create command
type SpecsConfig struct {
	Description  string
	Debug        bool   // -d, --debug: Save intermediate outputs
	Force        bool   // --force: Overwrite existing files
	Module       string // -m, --module: Target module
	OutputPath   string // -o, --output: Custom output path
	PromptPath   string // --prompt: Custom system prompt file
	TemplateRoot string
	Logger       *logging.Logger
	EACConfig    *config.EACConfig // Loaded repository configuration
}

// loadAndBuildPrompt loads the contract and builds the AI prompt
//
// This function:
// - Loads specification contract from YAML files
// - Builds the full AI prompt with contract context
// - Optionally saves debug output
func loadAndBuildPrompt(config *SpecsConfig) (string, error) {
	if config.Logger != nil {
		config.Logger.Debug("Loading specification contract")
	}
	log.Info("📋 Loading specification contract...")

	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to build contract-based prompt", zap.Error(err))
		}
		return "", err
	}

	if config.Logger != nil {
		config.Logger.Debug("Prompt built successfully", zap.Int("promptLength", len(fullPrompt)))
	}

	if config.Debug && config.Logger != nil {
		// Log debug content to logger instead of writing to file
		config.Logger.LogDebugContent("full-prompt", fullPrompt)
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
// - Optionally saves debug outputs
func generateAndClean(config *SpecsConfig, prompt string) (string, error) {
	// Check for subprocess mock response (used in acceptance tests)
	if mock, ok := aimock.GetMockResponse("specs"); ok {
		if config.Logger != nil {
			config.Logger.Debug("Using subprocess mock response")
		}
		return mock, nil
	}

	if config.Logger != nil {
		config.Logger.Debug("Starting AI generation with retry")
	}

	// Load contract and anti-corruption rules for validator
	loader := aimock.NewContractLoader(config.TemplateRoot, aimock.TypeSpecs, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to load contract", zap.Error(err))
		}
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create validator (JSON schema validated in Phase 1)
	validator := gherkin.NewValidator(contractData)

	// Load AI config to get retry strategy
	var aiConfig *aimock.AIConfig
	aiConfig, err = aimock.LoadAIConfig(config.TemplateRoot)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Warn("Could not load AI config, using default retry strategy", zap.Error(err))
		}
		aiConfig = nil
	}

	// Build retry configuration with strategy
	retryConfig, err := buildRetryConfig(executorAdapter, validator, config, aiConfig)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to build retry config", zap.Error(err))
		}
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("Configured retry behavior",
			zap.Int("maxAttempts", retryConfig.MaxAttempts),
			zap.Bool("debug", config.Debug))
	}

	// Generate with retry
	ctx := context.Background()
	result, err := aimock.GenerateWithRetry(ctx, retryConfig, prompt)
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("AI generation failed after retries", zap.Error(err))
		}
		log.Error("\nTroubleshooting:")
		log.Error("  1. Ensure AI provider is configured: r2r agent init --ai <provider>")
		log.Error("  2. Check API key environment variable is set")
		log.Error("  3. Verify network connectivity to AI provider")
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("AI generation completed",
			zap.Int("attempts", result.Attempts),
			zap.Int("outputLength", len(result.Output)),
			zap.Int("validationErrors", len(result.ValidationErrors)))
	}

	// Handle validation errors
	if len(result.ValidationErrors) > 0 {
		criticalErrors := contracts.CountCriticalErrors(result.ValidationErrors)

		if criticalErrors > 0 {
			if config.Logger != nil {
				config.Logger.Error("Generated specification has critical validation errors",
					zap.Int("criticalErrors", criticalErrors),
					zap.Int("attempts", result.Attempts))
			}
			log.Error("")
			log.Error("⚠️  Generated specification has validation errors:\n")
			log.Errorf("%s", contracts.FormatValidationErrors(result.ValidationErrors))
			log.Errorf("\nThe AI attempted %d time(s) but could not generate valid output.", result.Attempts)
			log.Error("\nTroubleshooting:")
			log.Error("  1. Try rephrasing your description to be more specific")
			log.Error("  2. Use --debug to inspect the generated output")
			log.Error("  3. Review the validation errors above and manually fix the output")
			return "", fmt.Errorf("generated specification has %d critical validation error(s)", criticalErrors)
		}

		// Only warnings - log them but continue
		if config.Logger != nil {
			config.Logger.Warn("Generated specification has warnings",
				zap.Int("warnings", len(result.ValidationErrors)))
		}
		log.Errorf("\nℹ️  Generated specification has %d warning(s):", len(result.ValidationErrors))
		log.Errorf("%s", contracts.FormatValidationErrors(result.ValidationErrors))
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
		if config.Logger != nil {
			config.Logger.Debug("Using user-specified output path", zap.String("path", finalOutputPath))
		}
	} else {
		// Extract feature name and determine path
		moduleName, featureName, err := specs.ExtractFeatureName(content)
		if err != nil {
			if config.Logger != nil {
				config.Logger.Error("Failed to extract feature name", zap.Error(err))
			}
			return "", err
		}

		if config.Logger != nil {
			config.Logger.Debug("Extracted feature info",
				zap.String("module", moduleName),
				zap.String("feature", featureName))
		}

		// Default to specs directory
		finalOutputPath = specs.DetermineOutputPath(config.TemplateRoot, moduleName, featureName, config.EACConfig)
	}

	// Security: Validate that output path is within repository
	if err := specs.ValidateOutputPath(finalOutputPath, config.TemplateRoot); err != nil {
		if config.Logger != nil {
			config.Logger.Error("Output path security validation failed",
				zap.String("path", finalOutputPath),
				zap.Error(err))
		}
		return "", fmt.Errorf("security error: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("Output path validated", zap.String("path", finalOutputPath))
	}

	return finalOutputPath, nil
}

// writeOutputAndReportSuccess writes the specification file and displays success message
//
// This function:
// - Creates necessary directories
// - Writes specification content to file
// - Displays success message with next steps
func writeOutputAndReportSuccess(outputPath string, content string, config *SpecsConfig) error {
	// Create directories
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Log the write operation
	if config.Logger != nil {
		config.Logger.Debug("Writing specification file",
			zap.String("path", outputPath),
			zap.Int("size", len(content)))
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
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

// formatModuleList returns a formatted list of available modules
func formatModuleList(moduleReport *reports.ModuleContractReport) string {
	var sb strings.Builder
	for _, mod := range moduleReport.Registry.All() {
		sb.WriteString(fmt.Sprintf("  - %s (source: %s)\n", mod.Moniker, mod.Files.Root))
	}
	return sb.String()
}

// loadPromptWithFallback implements prompt loading:
// 1. Custom path (if specified via --prompt flag)
// 2. AI config: .r2r/eac/aimock.TypeSpecs/specification.md
// 3. Built-in: embedded prompts/specification.md
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
		log.Errorf("📋 Using custom prompt: %s", customPath)
		return string(content), nil
	}

	// Load prompt with three-tier priority:
	// 1. Command flag (--prompt)
	// 2. Team override (.r2r/eac/templates/aimock.TypeSpecs/specs.md)
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
	loader := aimock.NewContractLoader(config.TemplateRoot, aimock.TypeSpecs, paths.DefaultsVersion)

	// Load contract
	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Load referenced files (tags)
	tagsPath := filepath.Join(contracts.EACConfigRelPath, "testing-tags.yml")
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
		prompt.WriteString("Common modules include: eac-commands, eac-core, r2r-cli, eac-mcp-commands\n\n")
	}

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete Gherkin specification now.\n")
	prompt.WriteString("Return ONLY the Gherkin content starting with 'Feature:' - no markdown fences, no explanations.\n")

	return prompt.String()
}

// loadModuleControlsContext loads OSCAL profile and formats controls for AI prompt
func loadModuleControlsContext(workspaceRoot string, moduleName string, cfg *config.EACConfig) string {
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

// buildRetryConfig creates RetryConfig with strategy from AI config
func buildRetryConfig(
	executor contracts.AIExecutor,
	validator contracts.Validator,
	config *SpecsConfig,
	aiConfig *aimock.AIConfig,
) (*aimock.RetryConfig, error) {
	maxAttempts := 2
	var strategy aimock.RetryStrategy

	// Load retry strategy from config if available
	if aiConfig != nil {
		if typeConfig, ok := aiConfig.Types[aimock.TypeSpecs]; ok {
			if typeConfig.RetryStrategy != nil && typeConfig.RetryStrategy.MaxAttempts > 0 {
				maxAttempts = typeConfig.RetryStrategy.MaxAttempts
			}

			if typeConfig.RetryStrategy != nil {
				var focusCategories []string
				if typeConfig.RetryStrategy.FocusCategories != nil {
					focusCategories = typeConfig.RetryStrategy.FocusCategories
				}

				var err error
				strategy, err = aimock.GetRetryStrategy(typeConfig.RetryStrategy.Type, focusCategories)
				if err != nil {
					return nil, fmt.Errorf("failed to create retry strategy: %w", err)
				}
				if config.Logger != nil {
					config.Logger.Info("Using retry strategy",
						zap.String("strategy", strategy.Name()),
						zap.Int("maxAttempts", maxAttempts))
				}
			}
		}
	}

	// Default to StandardStrategy if not configured
	if strategy == nil {
		strategy = &aimock.StandardStrategy{}
	}

	return &aimock.RetryConfig{
		TypeName:        aimock.TypeSpecs,
		OutputFormat:    aimock.FormatGherkin, // Generate Gherkin directly
		Validator:       validator,            // Validate Gherkin output
		Executor:        executor,
		TemplateRoot:    config.TemplateRoot,
		MaxAttempts:     maxAttempts,
		Debug:           config.Debug,
		Strategy:        strategy,
	}, nil
}
