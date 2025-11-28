package assessment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/ai"
	"github.com/ready-to-release/eac/src/ai/providers"
)

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

// ============================================================================

// generateReport generates a risk assessment report using AI
func generateReport(config *Config, input *AnalysisInput) (string, error) {
	// Check for mock response (test mode)
	if mockAIResponse != "" {
		return mockAIResponse, nil
	}

	// Load AI contract
	contractPath := filepath.Join(config.WorkspaceRoot, "contracts", "ai", "risks-assessment", "0.1.0")

	// Check if contract exists
	if _, err := os.Stat(contractPath); os.IsNotExist(err) {
		return "", fmt.Errorf("AI contract not found at %s", contractPath)
	}

	// Build AI prompt
	prompt, err := buildPrompt(config, input)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	// Debug: save prompt if requested
	if config.Debug {
		if err := saveDebugFile(config, "prompt.md", prompt); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save debug prompt: %v\n", err)
		}
	}

	// Create AI executor
	executor := ai.NewExecutor(config.WorkspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Execute AI call
	ctx := context.Background()
	response, err := executor.Execute(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Clean response
	report := cleanResponse(response)

	// Debug: save raw response if requested
	if config.Debug {
		if err := saveDebugFile(config, "raw-response.md", response); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save debug response: %v\n", err)
		}
		if err := saveDebugFile(config, "cleaned-report.md", report); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save debug report: %v\n", err)
		}
	}

	return report, nil
}

// buildPrompt builds the AI prompt from template and data
func buildPrompt(config *Config, input *AnalysisInput) (string, error) {
	// Use custom prompt if provided
	promptPath := config.PromptPath
	if promptPath == "" {
		promptPath = filepath.Join(config.WorkspaceRoot, "contracts", "ai", "risks-assessment", "0.1.0", "assessment.md")
	}

	// Read prompt template
	templateBytes, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template: %w", err)
	}
	template := string(templateBytes)

	// Prepare replacement data
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	changedFiles := formatChangedFiles(input.Files, input.WorkspaceRoot)
	specifications := formatSpecifications(input.Specifications)

	// Simple template replacement
	prompt := template
	prompt = strings.ReplaceAll(prompt, "{{.ChangedFiles}}", changedFiles)
	prompt = strings.ReplaceAll(prompt, "{{.Specifications}}", specifications)
	prompt = strings.ReplaceAll(prompt, "{{.Scope}}", input.Scope)
	prompt = strings.ReplaceAll(prompt, "{{.Timestamp}}", timestamp)
	prompt = strings.ReplaceAll(prompt, "{{.FileCount}}", fmt.Sprintf("%d", len(input.Files)))
	prompt = strings.ReplaceAll(prompt, "{{.SpecCount}}", fmt.Sprintf("%d", len(input.Specifications)))
	prompt = strings.ReplaceAll(prompt, "{{.Command}}", fmt.Sprintf("risks assessment --scope %s", input.Scope))

	// Load contract data
	contractData, err := loadContractData(config.WorkspaceRoot)
	if err != nil {
		contractData = "(Contract data not available)"
	}
	prompt = strings.ReplaceAll(prompt, "{{.ContractData}}", contractData)

	return prompt, nil
}

// loadContractData loads the contract YAML for context
func loadContractData(workspaceRoot string) (string, error) {
	contractPath := filepath.Join(workspaceRoot, "contracts", "ai", "risks-assessment", "0.1.0", "contract.yml")
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// cleanResponse cleans AI response according to anti-corruption rules
func cleanResponse(response string) string {
	// Remove common AI preambles
	cleaned := response

	// Remove code fences
	cleaned = strings.ReplaceAll(cleaned, "```markdown", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")

	// Remove AI preambles
	lines := strings.Split(cleaned, "\n")
	var outputLines []string
	started := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip preamble lines
		if !started {
			if strings.HasPrefix(trimmed, "# Risk Assessment Report") {
				started = true
			} else if strings.Contains(trimmed, "I'll help") ||
			          strings.Contains(trimmed, "I'll create") ||
			          strings.Contains(trimmed, "I'll analyze") {
				continue
			}
		}

		if started {
			outputLines = append(outputLines, line)
		}
	}

	return strings.Join(outputLines, "\n")
}

// saveDebugFile saves debug output to the logs directory
func saveDebugFile(config *Config, filename string, content string) error {
	logDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "risks")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("%s-%s", timestamp, filename))

	return os.WriteFile(logFile, []byte(content), 0644)
}
