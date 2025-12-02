package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/ai/providers"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
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

// RiskData represents a parsed risk from an assessment
type RiskData struct {
	RiskID        string   `json:"risk_id"`
	ControlName   string   `json:"control_name"`
	Domain        string   `json:"domain"`
	Severity      string   `json:"severity"`
	Description   string   `json:"description"`
	AffectedFiles []string `json:"affected_files"`
	RelatedSpecs  []string `json:"related_specs"`
	Impact        string   `json:"impact"`
}

// parseAssessment parses a risk assessment file to extract risks
func parseAssessment(config *Config, assessmentPath string) ([]RiskData, error) {
	// Read assessment file
	fullPath := assessmentPath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(config.WorkspaceRoot, assessmentPath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("assessment file not found: %w", err)
	}

	assessmentContent := string(content)

	// Check if it's a valid assessment (contains "Risk Assessment Report")
	if !strings.Contains(assessmentContent, "Risk Assessment Report") {
		return nil, fmt.Errorf("failed to parse assessment: invalid format (missing 'Risk Assessment Report' header)")
	}

	// Use AI to parse the assessment
	risks, err := parseWithAI(config, assessmentPath, assessmentContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assessment with AI: %w", err)
	}

	return risks, nil
}

// parseWithAI uses AI to extract structured risk data from assessment
func parseWithAI(config *Config, assessmentPath string, assessmentContent string) ([]RiskData, error) {
	var response string

	// Check for mock response from file-based mock system (subprocess testing)
	if mock, ok := aimock.GetMockResponseWithSubcommand("risks", "create"); ok {
		response = mock
	} else if mockAIResponse != "" {
		// Check for mock response (test mode - in-process testing)
		response = mockAIResponse
	} else {
		// Load AI contract
		promptPath := config.PromptPath
		if promptPath == "" {
			promptPath = filepath.Join(config.WorkspaceRoot, "contracts", "ai", "risks-create", "0.1.0", "create.md")
		}

		// Read prompt template
		templateBytes, err := os.ReadFile(promptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt template: %w", err)
		}
		template := string(templateBytes)

		// Build prompt
		prompt := strings.ReplaceAll(template, "{{.AssessmentPath}}", assessmentPath)
		prompt = prompt + "\n\n---\n\n### Assessment Report Content\n\n" + assessmentContent

		// Debug: save prompt if requested
		if config.Debug {
			saveDebugFile(config, "parse-prompt.md", prompt)
		}

		// Create AI executor
		executor := ai.NewExecutor(config.WorkspaceRoot)
		providers.RegisterBuiltIn(executor)

		// Execute AI call
		ctx := context.Background()
		response, err = executor.Execute(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("AI parsing failed: %w", err)
		}

		// Debug: save response
		if config.Debug {
			saveDebugFile(config, "parse-response.json", response)
		}
	}

	// Clean response
	cleaned := cleanJSONResponse(response)

	// Parse JSON
	var risks []RiskData
	if err := json.Unmarshal([]byte(cleaned), &risks); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w (response: %s)", err, cleaned)
	}

	return risks, nil
}

// cleanJSONResponse cleans AI response to extract JSON
func cleanJSONResponse(response string) string {
	// Remove code fences
	cleaned := strings.ReplaceAll(response, "```json", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Find JSON array start
	startIdx := strings.Index(cleaned, "[")
	if startIdx == -1 {
		return cleaned
	}

	// Find JSON array end
	endIdx := strings.LastIndex(cleaned, "]")
	if endIdx == -1 {
		return cleaned
	}

	return cleaned[startIdx : endIdx+1]
}

// findAssessmentFiles finds all assessment files (handles both file and folder input)
func findAssessmentFiles(config *Config) ([]string, error) {
	path := config.AssessmentPath
	if !filepath.IsAbs(path) {
		path = filepath.Join(config.WorkspaceRoot, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("assessment file not found: %w", err)
	}

	// Single file
	if !info.IsDir() {
		return []string{config.AssessmentPath}, nil
	}

	// Directory - find all .md files
	var files []string
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(filePath, ".md") {
			// Make relative to workspace root
			relPath, err := filepath.Rel(config.WorkspaceRoot, filePath)
			if err != nil {
				relPath = filePath
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// saveDebugFile saves debug output
func saveDebugFile(config *Config, filename string, content string) error {
	logDir := filepath.Join(config.WorkspaceRoot, "out", "logs", "risks")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, filename)
	return os.WriteFile(logFile, []byte(content), 0644)
}
