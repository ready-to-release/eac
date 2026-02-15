package commitmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/commands/repository/create/aiutil"
	aimock "github.com/ready-to-release/eac/go/core/ai"
)

// GenerationResult holds the result of AI generation including metadata.
type GenerationResult struct {
	Output       string // The generated output
	ProviderName string // The AI provider used (e.g., "claude-cli", "openai")
}

// generateWithPrompt generates output using the three-tier prompt loading system with validation and retry
// If testExecutor is provided (non-nil), it will be used instead of creating a new executor (for testing).
func generateWithPrompt(deps *Deps, promptName, userPrompt, workspaceRoot string, affectedModules []string, debugEnabled bool, testExecutor *ai.Executor) (string, error) {
	result, err := generateWithPromptResult(deps, promptName, userPrompt, workspaceRoot, affectedModules, debugEnabled, testExecutor)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}

// generateWithPromptResult generates output and returns full metadata including provider info.
func generateWithPromptResult(deps *Deps, promptName, userPrompt, workspaceRoot string, affectedModules []string, debugEnabled bool, testExecutor *ai.Executor) (*GenerationResult, error) {
	// Check for mock response from file-based mock system (subprocess testing)
	// Use subcommand-aware mock lookup for module sections to allow different mock responses
	var mock string
	var mockFound bool
	if promptName == "module" {
		mock, mockFound = aimock.GetMockResponseWithSubcommand("commit-message", "module")
	} else {
		mock, mockFound = aimock.GetMockResponse("commit-message")
	}
	if mockFound {
		// Format mock JSON to conventional commit format
		var formattedOutput string
		var err error
		if promptName == "module" {
			formattedOutput, err = FormatModuleSection(mock)
		} else {
			formattedOutput, err = FormatCommitMessage(mock)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to format mock response: %w", err)
		}
		return &GenerationResult{Output: formattedOutput, ProviderName: "mock-file"}, nil
	}

	// Check for mock response (test mode - in-process testing)
	if deps.AIResponse != "" {
		// Format mock JSON to conventional commit format
		var formattedOutput string
		var err error
		if promptName == "module" {
			formattedOutput, err = FormatModuleSection(deps.AIResponse)
		} else {
			formattedOutput, err = FormatCommitMessage(deps.AIResponse)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to format mock response: %w", err)
		}
		return &GenerationResult{Output: formattedOutput, ProviderName: "mock"}, nil
	}

	// Load prompt template using three-tier system
	loader := aimock.NewContractLoader(workspaceRoot, aimock.TypeCommitMessage, "")
	promptTemplate, _, err := loader.LoadPrompt(promptName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	model := extractModelFromAgent(promptTemplate)

	// Build full prompt: prompt template + user input (no template rendering needed)
	fullPrompt := promptTemplate + "\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n" + userPrompt

	// Choose schema based on prompt type:
	// - "module" prompts use commit-message-module.schema.json
	// - other prompts use commit-message.schema.json
	schemaFilename := "commit-message.schema.json"
	if promptName == "module" {
		schemaFilename = "commit-message-module.schema.json"
	}

	// Execute AI generation using shared pipeline
	result, err := aiutil.ExecuteGeneration(aiutil.GenerationParams{
		WorkspaceRoot:  workspaceRoot,
		Prompt:         fullPrompt,
		TypeName:       aimock.TypeCommitMessage,
		SchemaFilename: schemaFilename,
		Model:          model,
		Debug:          debugEnabled,
		TestExecutor:   testExecutor,
	})
	if err != nil {
		return nil, err
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

// extractModelFromAgent parses agent frontmatter and extracts the model field.
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
