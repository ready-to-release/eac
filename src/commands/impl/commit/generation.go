package commit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commitmessage "github.com/ready-to-release/eac/src/commands/impl/commit/internal"
	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/core/ai/providers"
)

// loadPromptWithFallback implements three-tier prompt loading:
// 1. Local contract: .r2r/contracts/ai/commit-message/0.1.0/<name>.md
// 2. Repo contract: contracts/ai/commit-message/0.1.0/<name>.md
// 3. Built-in: embedded prompts/<name>.md
func loadPromptWithFallback(promptName string, workspaceRoot string) (string, error) {
	// Create contract loader
	loader := contracts.NewAIContractLoader(workspaceRoot, "commit-message", "0.1.0")

	// No embedded prompt - load from .r2r/contracts or contracts/ai
	var embeddedPrompt string

	// Load prompt with fallback chain
	prompt, source, err := loader.LoadPrompt(promptName+".md", embeddedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Log source if using override (for transparency)
	if source != "embedded default" && source != "repository contract" {
		fmt.Fprintf(os.Stderr, "ℹ️  Using %s prompt\n", source)
	}

	return prompt, nil
}

// generateWithPrompt generates output using the three-tier prompt loading system with validation and retry
func generateWithPrompt(promptName string, userPrompt string, workspaceRoot string, affectedModules []string, debugEnabled bool) (string, error) {
	// Load prompt using three-tier system
	agentContent, err := loadPromptWithFallback(promptName, workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	model := extractModelFromAgent(agentContent)

	// Create executor and register providers
	executor := ai.NewExecutor(workspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Build full prompt: agent instructions + user input
	fullPrompt := agentContent + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + userPrompt

	// Load contract and anti-corruption rules
	loader := contracts.NewSpecContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")
	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		// If anti-corruption rules fail, fall back to non-validated generation
		fmt.Fprintf(os.Stderr, "⚠️  Could not load anti-corruption rules, proceeding without validation: %v\n", err)
		return generateWithoutValidation(executor, fullPrompt, model, promptName, workspaceRoot)
	}

	// Load commit message contract data for validation
	contractPath := filepath.Join(loader.GetContractPath(), "contract.yml")
	commitContract, err := commitmessage.LoadContract(contractPath)
	if err != nil {
		// If commit contract fails, fall back to non-validated generation
		fmt.Fprintf(os.Stderr, "⚠️  Could not load commit contract, proceeding without validation: %v\n", err)
		return generateWithoutValidation(executor, fullPrompt, model, promptName, workspaceRoot)
	}

	// Create validator
	validator := commitmessage.NewCommitMessageValidator(
		commitContract,
		antiCorruptionRules,
		affectedModules,
		loader.GetContractPath(),
	)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := &aiExecutorAdapter{executor: executor, model: model}

	// Setup debug directory if needed
	debugOutputDir := ""
	if debugEnabled {
		debugOutputDir = filepath.Join(workspaceRoot, "out")
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
		ContentMarker:  "", // Commit messages don't have a specific content marker
		MaxAttempts:    2,
		Debug:          debugEnabled,
		DebugOutputDir: debugOutputDir,
		ValidationContext: map[string]interface{}{
			"affectedModules": affectedModules,
		},
	}

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
	output = stripAgentNoise(output, promptName, workspaceRoot)

	return output, nil
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
