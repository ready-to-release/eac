package commit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/ai/providers"
	"github.com/ready-to-release/eac/src/core/contracts"
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

// loadPromptWithFallback implements three-tier prompt loading:
// 1. Local contract: .r2r/contracts/ai/commit-message/0.1.0/<name>.md
// 2. Repo contract: contracts/ai/commit-message/0.1.0/<name>.md
// 3. Built-in: embedded prompts/<name>.md
func loadPromptWithFallback(promptName string, workspaceRoot string) (string, error) {
	// Create contract loader
	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")

	// No embedded prompt - load from .r2r/contracts or contracts/ai
	var embeddedPrompt string

	// Load prompt using three-tier system
	agentContent, _, err := loader.LoadPrompt(promptName+".md", embeddedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	return agentContent, nil
}

// generateWithPrompt generates output using the three-tier prompt loading system with validation and retry
// If testExecutor is provided (non-nil), it will be used instead of creating a new executor (for testing)
func generateWithPrompt(promptName string, userPrompt string, workspaceRoot string, affectedModules []string, debugEnabled bool, testExecutor *ai.Executor) (string, error) {
	// Check for mock response (test mode)
	if mockAIResponse != "" {
		return mockAIResponse, nil
	}

	// Load prompt template using three-tier system
	promptTemplate, err := loadPromptWithFallback(promptName, workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt template: %w", err)
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
		fmt.Fprintf(os.Stderr, "⚠️  Could not load anti-corruption rules, proceeding without validation: %v\n", err)
		return generateWithoutValidation(executor, fullPrompt, model, promptName, workspaceRoot)
	}

	// Create appropriate validator based on prompt type
	var validator contracts.Validator
	switch promptName {
	case "top-level":
		// Top-level validator checks header, auditor-summary, body (no module sections)
		validator = commitmessage.NewTopLevelValidator(affectedModules)
	case "module":
		// Module validator checks module section format only
		validator = commitmessage.NewModuleSectionValidator("")
	default:
		// Fallback to full commit message validator for final assembly validation
		contractPath := filepath.Join(loader.GetContractPath(), "contract.yml")
		commitContract, err := commitmessage.LoadContract(contractPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Could not load commit contract, proceeding without validation: %v\n", err)
			return generateWithoutValidation(executor, fullPrompt, model, promptName, workspaceRoot)
		}
		validator = commitmessage.NewCommitMessageValidator(
			commitContract,
			antiCorruptionRules,
			affectedModules,
			loader.GetContractPath(),
		)
	}

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := &aiExecutorAdapter{executor: executor, model: model}

	// Setup debug directory if needed
	debugOutputDir := ""
	if debugEnabled {
		debugOutputDir = filepath.Join(workspaceRoot, "out", "logs", "commit")
		if err := os.MkdirAll(debugOutputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to create debug directory: %v\n", err)
		}
	}

	// Configure retry behavior
	retryConfig := buildRetryConfig(executorAdapter, validator, antiCorruptionRules, affectedModules, debugEnabled, debugOutputDir)

	// Generate with retry
	ctx := context.Background()
	result, err := contracts.GenerateWithRetry(ctx, retryConfig, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Report validation errors if any
	if len(result.ValidationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Generated commit message has validation issues:\n")
		for _, verr := range result.ValidationErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", verr.Message)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	return result.Output, nil
}

// generateWithoutValidation is a fallback when contract/validation cannot be loaded
func generateWithoutValidation(executor *ai.Executor, fullPrompt string, model string, promptName string, workspaceRoot string) (string, error) {
	// Prepare options
	var opts []ai.Option
	if model != "" {
		opts = append(opts, ai.WithModel(model))
	}

	// Execute with context
	ctx := context.Background()
	output, err := executor.Execute(ctx, fullPrompt, opts...)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	// Trim output
	output = strings.TrimSpace(output)

	// Remove common agent initialization noise using contract-based filtering
	output = StripAgentNoise(output, promptName, workspaceRoot)

	return output, nil
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
		Fixer:          commitmessage.AutoCleanup, // Apply line wrapping and formatting fixes before validation
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
