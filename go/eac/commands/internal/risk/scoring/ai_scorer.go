// Package scoring provides AI-powered risk analysis.
package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/adapters/ai"
	"github.com/ready-to-release/eac/go/eac/adapters/ai/providers"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
)

// defaultRiskAnalysisPrompt is the fallback prompt when file is not found.
const defaultRiskAnalysisPrompt = `# Risk Analysis Assistant

Analyze security findings and compute a likelihood score (1-5) following ISO 27005.
Return JSON: {"computed_likelihood": <1-5>, "reasoning": "<text>", "risk_summary": "<text>", "confidence": <0.0-1.0>}`

// MockAIResponse holds mock response for testing.
var mockAIResponse string

// SetMockAIResponse sets a mock AI response for testing.
func SetMockAIResponse(response string) {
	mockAIResponse = response
}

// ResetMockAIResponse clears the mock AI response.
func ResetMockAIResponse() {
	mockAIResponse = ""
}

// AIScorer performs AI-powered risk analysis.
type AIScorer struct {
	workspaceRoot string
	providerType  string
	model         string
	debug         bool
}

// NewAIScorer creates a new AI scorer.
func NewAIScorer(workspaceRoot, providerType, model string, debug bool) *AIScorer {
	return &AIScorer{
		workspaceRoot: workspaceRoot,
		providerType:  providerType,
		model:         model,
		debug:         debug,
	}
}

// AnalyzeRisk performs AI-powered risk analysis on security findings.
func (s *AIScorer) AnalyzeRisk(ctx context.Context, input *AIAnalysisInput) (*AIRiskAnalysis, error) {
	// Use mock response if set (for testing)
	if mockAIResponse != "" {
		return parseAIResponse(mockAIResponse)
	}

	// Check for mock AI provider (integration testing via environment)
	if mock, ok := aimock.GetMockResponse("risk-access"); ok {
		return parseAIResponse(mock)
	}

	// Build the prompt
	prompt, err := s.buildPrompt(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Call AI provider
	response, err := s.callAI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse response
	analysis, err := parseAIResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return analysis, nil
}

// buildPrompt constructs the AI prompt for risk analysis.
func (s *AIScorer) buildPrompt(input *AIAnalysisInput) (string, error) {
	// Load prompt with three-tier priority:
	// 1. Command flag (not applicable - internal function)
	// 2. Team override (.r2r/eac/templates/ai/risk-access/risk-access.md)
	// 3. System default (templates/ai/risk-access/risk-access.md)
	// Convention: Empty string uses type name (risk-access.md)
	systemPrompt := defaultRiskAnalysisPrompt
	if s.workspaceRoot != "" {
		loader := aimock.NewContractLoader(s.workspaceRoot, "ai/risk-access", "")
		if loadedPrompt, _, err := loader.LoadPrompt("", defaultRiskAnalysisPrompt); err == nil {
			systemPrompt = loadedPrompt
		}
	}

	// Convert input to JSON for the prompt
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}

	// Combine system prompt with input data
	var builder strings.Builder
	builder.WriteString(systemPrompt)
	builder.WriteString("\n\n## Analysis Input\n\n```json\n")
	builder.Write(inputJSON)
	builder.WriteString("\n```\n")

	return builder.String(), nil
}

// callAI makes the actual AI call.
func (s *AIScorer) callAI(ctx context.Context, prompt string) (string, error) {
	// Create executor
	executor := ai.NewExecutor(s.workspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Configure options
	opts := []ai.Option{}
	if s.model != "" {
		opts = append(opts, ai.WithModel(s.model))
	}
	if s.debug {
		opts = append(opts, ai.WithDebug(true))
	}

	// Execute
	response, err := executor.Execute(ctx, prompt, opts...)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	return response, nil
}

// parseAIResponse parses the AI response into structured analysis.
func parseAIResponse(response string) (*AIRiskAnalysis, error) {
	// Try to find JSON in the response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		// No JSON found, try to extract from markdown code block
		response = extractJSONFromMarkdown(response)
		jsonStart = strings.Index(response, "{")
		jsonEnd = strings.LastIndex(response, "}")
	}

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]

		var analysis AIRiskAnalysis
		if err := json.Unmarshal([]byte(jsonStr), &analysis); err == nil {
			return &analysis, nil
		}
	}

	// Fall back to creating analysis from response text
	return &AIRiskAnalysis{
		ComputedLikelihood: 3, // Default to medium
		Reasoning:          response,
		RiskSummary:        truncateString(response, 200),
		Confidence:         0.5, // Lower confidence for non-JSON response
	}, nil
}

// extractJSONFromMarkdown extracts JSON from markdown code blocks.
func extractJSONFromMarkdown(text string) string {
	// Look for ```json ... ``` blocks
	jsonStart := strings.Index(text, "```json")
	if jsonStart != -1 {
		jsonStart += 7 // Skip "```json"
		jsonEnd := strings.Index(text[jsonStart:], "```")
		if jsonEnd != -1 {
			return strings.TrimSpace(text[jsonStart : jsonStart+jsonEnd])
		}
	}

	// Look for ``` ... ``` blocks
	codeStart := strings.Index(text, "```")
	if codeStart != -1 {
		codeStart += 3
		codeEnd := strings.Index(text[codeStart:], "```")
		if codeEnd != -1 {
			extracted := strings.TrimSpace(text[codeStart : codeStart+codeEnd])
			// Check if it looks like JSON
			if strings.HasPrefix(extracted, "{") {
				return extracted
			}
		}
	}

	return text
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ComputeRiskWithAI performs complete risk scoring with AI analysis.
func (s *AIScorer) ComputeRiskWithAI(ctx context.Context, module string, findings []VulnerabilityInput, moduleContext ModuleContext, baseImpact int) (*RiskScore, error) {
	input := &AIAnalysisInput{
		Module:   module,
		Findings: findings,
		Context:  moduleContext,
	}

	analysis, err := s.AnalyzeRisk(ctx, input)
	if err != nil {
		// Fall back to base scoring if AI fails
		baseLikelihood := CalculateBaseLikelihood(
			countBySeverity(findings, "CRITICAL"),
			countBySeverity(findings, "HIGH"),
			countBySeverity(findings, "MEDIUM"),
			countBySeverity(findings, "LOW"),
		)
		return ComputeRiskScore(module, baseLikelihood, baseImpact, "AI analysis unavailable, using base scoring"), nil
	}

	return ComputeRiskScore(module, analysis.ComputedLikelihood, baseImpact, analysis.Reasoning), nil
}

// countBySeverity counts findings with a specific severity.
func countBySeverity(findings []VulnerabilityInput, severity string) int {
	count := 0
	for _, f := range findings {
		if strings.EqualFold(f.Severity, severity) {
			count++
		}
	}
	return count
}

// DefaultAIScorer returns an AI scorer with default settings.
func DefaultAIScorer(workspaceRoot string) *AIScorer {
	return NewAIScorer(workspaceRoot, "", "", false)
}
