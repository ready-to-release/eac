package commitmessage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commitmessageinternal "github.com/ready-to-release/eac/go/eac/commands/impl/create/commit-message/internal"
	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/ai/providers"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/paths"
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

// loadPromptWithFallback implements prompt loading from AI configs:
// 1. AI config: .r2r/eac/ai/commit-message/<name>.md
// 2. Built-in: embedded prompts/<name>.md
func loadPromptWithFallback(promptName string, workspaceRoot string) (string, error) {
	// Create contract loader for AI config
	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "")

	// No embedded prompt - load from .r2r/eac/ai/
	var embeddedPrompt string

	// Load prompt from AI config (prompts organized in folders)
	agentContent, _, err := loader.LoadPrompt("commit/"+promptName+".md", embeddedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	return agentContent, nil
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
	if mock, ok := aimock.GetMockResponse("commit"); ok {
		return &GenerationResult{Output: mock, ProviderName: "mock-file"}, nil
	}

	// Check for mock response (test mode - in-process testing)
	if mockAIResponse != "" {
		return &GenerationResult{Output: mockAIResponse, ProviderName: "mock"}, nil
	}

	// Load prompt template using three-tier system
	promptTemplate, err := loadPromptWithFallback(promptName, workspaceRoot)
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

	// Load contract and anti-corruption rules for validation only
	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")
	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		// If anti-corruption rules fail, fall back to non-validated generation
		log.Warnf("Could not load anti-corruption rules, proceeding without validation: %v", err)
		return generateWithoutValidationResult(executor, fullPrompt, model, promptName, workspaceRoot)
	}

	// Create appropriate validator based on prompt type
	var validator contracts.Validator
	switch promptName {
	case "top-level":
		// Top-level validator checks header, auditor-summary, body (no module sections)
		validator = commitmessageinternal.NewTopLevelValidator(affectedModules)
	case "module":
		// Module validator checks module section format only
		validator = commitmessageinternal.NewModuleSectionValidator("")
	default:
		// Fallback to full commit message validator for final assembly validation
		commitContract, err := commitmessageinternal.LoadContractFromConfig(workspaceRoot)
		if err != nil {
			log.Warnf("Could not load commit contract, proceeding without validation: %v", err)
			return generateWithoutValidationResult(executor, fullPrompt, model, promptName, workspaceRoot)
		}
		validator = commitmessageinternal.NewCommitMessageValidator(
			commitContract,
			antiCorruptionRules,
			affectedModules,
			workspaceRoot,
		)
	}

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := &aiExecutorAdapter{executor: executor, model: model}

	// Setup debug directory if needed
	debugOutputDir := ""
	if debugEnabled {
		debugOutputDir = filepath.Join(paths.LogsPath(workspaceRoot), "commit")
		if err := os.MkdirAll(debugOutputDir, 0755); err != nil {
			log.Warnf("Failed to create debug directory: %v", err)
		}
	}

	// Configure retry behavior
	retryConfig := buildRetryConfig(executorAdapter, validator, antiCorruptionRules, affectedModules, debugEnabled, debugOutputDir)

	// Generate with retry
	ctx := context.Background()
	result, err := contracts.GenerateWithRetry(ctx, retryConfig, fullPrompt)
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

	return &GenerationResult{
		Output:       result.Output,
		ProviderName: result.ProviderName,
	}, nil
}

// generateWithoutValidation is a fallback when contract/validation cannot be loaded
func generateWithoutValidation(executor *ai.Executor, fullPrompt string, model string, promptName string, workspaceRoot string) (string, error) {
	result, err := generateWithoutValidationResult(executor, fullPrompt, model, promptName, workspaceRoot)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}

// generateWithoutValidationResult is a fallback when contract/validation cannot be loaded.
// Returns GenerationResult with provider metadata.
func generateWithoutValidationResult(executor *ai.Executor, fullPrompt string, model string, promptName string, workspaceRoot string) (*GenerationResult, error) {
	// Prepare options
	var opts []ai.Option
	if model != "" {
		opts = append(opts, ai.WithModel(model))
	}

	// Execute with context
	ctx := context.Background()
	output, err := executor.Execute(ctx, fullPrompt, opts...)
	if err != nil {
		return nil, fmt.Errorf("AI execution failed: %w", err)
	}

	// Get provider name after execution
	providerName := ""
	if provider := executor.GetLastUsedProvider(); provider != nil {
		providerName = provider.Name()
		// Log provider in debug mode (check for debug output dir as indicator)
		if workspaceRoot != "" {
			log.Debugf("AI provider used: %s", providerName)
		}
	}

	// Trim output
	output = strings.TrimSpace(output)

	// Remove common agent initialization noise using contract-based filtering
	output = StripAgentNoise(output, promptName, workspaceRoot)

	return &GenerationResult{
		Output:       output,
		ProviderName: providerName,
	}, nil
}

// buildRetryConfig creates a RetryConfig with standard settings for commit message generation
func buildRetryConfig(
	executor contracts.AIExecutor,
	validator contracts.Validator,
	antiCorruption *contracts.AntiCorruptionRules,
	affectedModules []string,
	debug bool,
	debugOutputDir string,
) *contracts.RetryConfig {
	return &contracts.RetryConfig{
		Executor:       executor,
		Validator:      validator,
		PromptBuilder:  &contracts.DefaultRetryPromptBuilder{},
		AntiCorruption: antiCorruption,
		Fixer:          commitmessageinternal.AutoCleanup, // Apply line wrapping and formatting fixes before validation
		ContentMarker:  "", // Commit messages don't have a specific content marker
		MaxAttempts:    2,
		Debug:          debug,
		DebugOutputDir: debugOutputDir,
		ValidationContext: map[string]interface{}{
			"affectedModules": affectedModules,
		},
	}
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
