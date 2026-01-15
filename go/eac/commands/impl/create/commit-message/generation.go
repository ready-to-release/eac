package commitmessage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

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

// GenerationResult holds the result of AI generation including metadata
type GenerationResult struct {
	Output       string // The generated output
	ProviderName string // The AI provider used (e.g., "claude-cli", "openai")
}

// generateWithPrompt generates output using the three-tier prompt loading system with validation and retry
// If testExecutor is provided (non-nil), it will be used instead of creating a new executor (for testing)
func generateWithPrompt(promptName string, userPrompt string, workspaceRoot string, affectedModules []string, debugEnabled bool, testExecutor *ai.Executor) (string, error) {
	result, err := generateWithPromptResult(promptName, userPrompt, workspaceRoot, affectedModules, debugEnabled, testExecutor)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}

// generateWithPromptResult generates output and returns full metadata including provider info
func generateWithPromptResult(promptName string, userPrompt string, workspaceRoot string, affectedModules []string, debugEnabled bool, testExecutor *ai.Executor) (*GenerationResult, error) {
	// Check for mock response from file-based mock system (subprocess testing)
	if mock, ok := aimock.GetMockResponse("commit-message"); ok {
		return &GenerationResult{Output: mock, ProviderName: "mock-file"}, nil
	}

	// Check for mock response (test mode - in-process testing)
	if mockAIResponse != "" {
		return &GenerationResult{Output: mockAIResponse, ProviderName: "mock"}, nil
	}

	// Load prompt template using three-tier system
	loader := aimock.NewContractLoader(workspaceRoot, aimock.TypeCommitMessage, "")
	promptTemplate, _, err := loader.LoadPrompt(promptName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	model := extractModelFromAgent(promptTemplate)

	// Use provided executor or create new one with real providers
	var executor *ai.Executor
	if testExecutor != nil {
		executor = testExecutor
	} else {
		executor = ai.NewExecutor(workspaceRoot)
		providers.RegisterBuiltIn(executor)
	}

	// Build full prompt: prompt template + user input (no template rendering needed)
	fullPrompt := promptTemplate + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + userPrompt

	// Load AI config to get retry strategy
	aiConfig, err := aimock.LoadAIConfig(workspaceRoot)
	if err != nil {
		log.Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil // Will use defaults
	}

	// Create JSON schema validator for Phase 1 JSON validation
	// Choose schema based on prompt type:
	// - "module" prompts use commit-message-module.schema.json
	// - other prompts use commit-message.schema.json
	schemaFilename := "commit-message.schema.json"
	if promptName == "module" {
		schemaFilename = "commit-message-module.schema.json"
	}
	schemaPath := filepath.Join(paths.ContractsVersionPath(workspaceRoot, paths.EACCoreModule, paths.DefaultsVersion), schemaFilename)
	validator, err := contracts.NewJSONSchemaValidator(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON schema validator: %w", err)
	}

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapterWithModel(executor, model)

	// Configure retry behavior
	retryConfig, err := buildRetryConfig(executorAdapter, validator, workspaceRoot, aiConfig, debugEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry (Phase 1: JSON generation)
	ctx := context.Background()
	result, err := aimock.GenerateWithRetry(ctx, retryConfig, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Report validation errors if any
	if len(result.ValidationErrors) > 0 {
		log.Warn("\nGenerated commit message has validation issues:")
		for _, verr := range result.ValidationErrors {
			log.Warnf("  - %s", verr.Message)
		}
		log.Warn("")
	}

	// Format JSON → formatted text (no AI)
	// Use appropriate formatter based on prompt type
	var formattedOutput string
	if promptName == "module" {
		formattedOutput, err = FormatModuleSection(result.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to format module section: %w", err)
		}
	} else {
		formattedOutput, err = FormatCommitMessage(result.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to format commit message: %w", err)
		}
	}

	return &GenerationResult{
		Output:       formattedOutput,
		ProviderName: result.ProviderName,
	}, nil
}

// buildRetryConfig creates a RetryConfig with standard settings for commit message generation
func buildRetryConfig(
	executor contracts.AIExecutor,
	validator contracts.Validator,
	workspaceRoot string,
	aiConfig *aimock.AIConfig,
	debug bool,
) (*aimock.RetryConfig, error) {
	// Default retry strategy and max attempts
	maxAttempts := 2
	var strategy aimock.RetryStrategy

	// Load retry strategy from AI config if available
	if aiConfig != nil {
		if typeConfig, ok := aiConfig.Types[aimock.TypeCommitMessage]; ok {
			// Get max attempts from config
			if typeConfig.RetryStrategy != nil && typeConfig.RetryStrategy.MaxAttempts > 0 {
				maxAttempts = typeConfig.RetryStrategy.MaxAttempts
			}

			// Create retry strategy from config
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
				log.Infof("Using retry strategy: %s (max attempts: %d)", strategy.Name(), maxAttempts)
			}
		}
	}

	// Default to StandardStrategy if not configured
	if strategy == nil {
		strategy = &aimock.StandardStrategy{}
		log.Debugf("Using default retry strategy: standard (max attempts: %d)", maxAttempts)
	}

	return &aimock.RetryConfig{
		TypeName:     aimock.TypeCommitMessage,
		OutputFormat: aimock.FormatJSON, // Generate structured JSON (commands format to text)
		Executor:     executor,
		Validator:    validator, // Validate JSON output
		TemplateRoot: workspaceRoot,
		MaxAttempts:  maxAttempts,
		Strategy:     strategy,
		Debug:        debug, // Enable debug logging if requested
		Logger:       logging.C().Zap(), // ✅ Pass logger for retry observability
	}, nil
}

// extractModelFromAgent parses agent frontmatter and extracts the model field
func extractModelFromAgent(agentContent string) string {
	lines := strings.Split(agentContent, "\n")
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(trimmed, "model:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
