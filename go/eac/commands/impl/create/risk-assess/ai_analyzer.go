package riskassess

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	"github.com/ready-to-release/eac/go/eac/core/ai/generation"
	"go.uber.org/zap"
)

// AIRiskAssessmentInput holds input for unified AI analysis
type AIRiskAssessmentInput struct {
	Modules               []ModuleAnalysisInput
	ProfileName           string
	TotalControls         int
	SatisfiedControls     int
	NotSatisfiedControls  int
}

// ModuleAnalysisInput holds per-module input data
type ModuleAnalysisInput struct {
	Module                string
	VulnerabilityFindings []scoring.VulnerabilityInput
	Context               scoring.ModuleContext
	ControlsSatisfied     int
	ControlsNotSatisfied  int
	Impact                int
}

// AIRiskAssessmentOutput holds complete AI analysis result
type AIRiskAssessmentOutput struct {
	ExecutiveSummary ExecutiveSummaryData `json:"executive_summary"`
	ModuleAnalyses   []ModuleAnalysisData `json:"module_analyses"`
	Confidence       float64              `json:"confidence"`
}

// ExecutiveSummaryData holds executive summary from AI
type ExecutiveSummaryData struct {
	OverallRiskPosture       string               `json:"overall_risk_posture"`
	SummaryNarrative         string               `json:"summary_narrative"`
	KeyFindings              []string             `json:"key_findings"`
	CriticalModules          []CriticalModuleInfo `json:"critical_modules"`
	Trends                   []string             `json:"trends"`
	StrategicRecommendations []string             `json:"strategic_recommendations"`
}

// ModuleAnalysisData holds per-module analysis from AI
type ModuleAnalysisData struct {
	Module              string   `json:"module"`
	ComputedLikelihood  int      `json:"computed_likelihood"`
	Reasoning           string   `json:"reasoning"`
	RiskSummary         string   `json:"risk_summary"`
	RecommendedControls []string `json:"recommended_controls"`
}

// GenerateRiskAssessment performs unified AI risk assessment
func GenerateRiskAssessment(ctx context.Context, config *AssessConfig, input *AIRiskAssessmentInput) (*AIRiskAssessmentOutput, error) {
	// Build comprehensive prompt
	prompt, err := buildRiskAssessmentPrompt(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Create AI executor
	executor := ai.NewExecutor(config.WorkspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match validation.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Call AI with generation layer validation
	result, err := generation.GenerateWithRetry(ctx, &generation.RetryConfig{
		TypeName:     generation.TypeRiskAssessment,
		OutputFormat: generation.FormatJSON,
		Executor:     executorAdapter,
		TemplateRoot: config.WorkspaceRoot,
		MaxAttempts:  3,
		Debug:        config.Debug,
		Logger:       nil, // Will use no-op logger
	}, prompt)

	if err != nil {
		return nil, fmt.Errorf("failed to generate risk assessment: %w", err)
	}

	// Parse validated JSON
	var output AIRiskAssessmentOutput
	if err := json.Unmarshal([]byte(result.Output), &output); err != nil {
		return nil, fmt.Errorf("failed to parse risk assessment JSON: %w", err)
	}

	return &output, nil
}

// buildRiskAssessmentPrompt builds comprehensive prompt with all module data
func buildRiskAssessmentPrompt(input *AIRiskAssessmentInput) (string, error) {
	// Build structured JSON input with all module data
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}

	var sb strings.Builder

	// Add summary information
	sb.WriteString("# Assessment Scope\n\n")
	sb.WriteString(fmt.Sprintf("- Profile: %s\n", input.ProfileName))
	sb.WriteString(fmt.Sprintf("- Modules Assessed: %d\n", len(input.Modules)))
	sb.WriteString(fmt.Sprintf("- Total Controls: %d\n", input.TotalControls))
	sb.WriteString(fmt.Sprintf("- Controls Satisfied: %d\n", input.SatisfiedControls))
	sb.WriteString(fmt.Sprintf("- Controls Not Satisfied: %d\n\n", input.NotSatisfiedControls))

	// Add module details
	sb.WriteString("# Module Assessment Data\n\n")
	sb.WriteString("```json\n")
	sb.Write(inputJSON)
	sb.WriteString("\n```\n")

	return sb.String(), nil
}

// ApplyAIRiskAssessment applies AI-generated results to module assessment results
func ApplyAIRiskAssessment(results []*ModuleAssessmentResult, aiOutput *AIRiskAssessmentOutput, logger *zap.SugaredLogger) {
	// Map AI module analyses to results
	analysisMap := make(map[string]*ModuleAnalysisData)
	for i := range aiOutput.ModuleAnalyses {
		analysisMap[aiOutput.ModuleAnalyses[i].Module] = &aiOutput.ModuleAnalyses[i]
	}

	// Apply AI risk scores to each module result
	for _, result := range results {
		if analysis, ok := analysisMap[result.Module]; ok {
			// Get impact (default to 3 if not specified)
			impact := 3

			result.RiskScore = scoring.ComputeRiskScore(
				result.Module,
				analysis.ComputedLikelihood,
				impact,
				analysis.Reasoning,
			)

			logger.Debugf("Applied AI risk score to %s: likelihood=%d, score=%d",
				result.Module, analysis.ComputedLikelihood, result.RiskScore.Score)
		} else {
			logger.Warnf("No AI analysis found for module %s", result.Module)
		}
	}
}

// BuildExecutiveSummary builds executive summary data from AI output
func BuildExecutiveSummary(
	aiOutput *AIRiskAssessmentOutput,
	basicStats ExecutiveSummary,
) ExecutiveSummary {
	summary := basicStats

	// Add AI-generated content
	summary.OverallRiskPosture = aiOutput.ExecutiveSummary.OverallRiskPosture
	summary.SummaryNarrative = aiOutput.ExecutiveSummary.SummaryNarrative
	summary.KeyFindings = aiOutput.ExecutiveSummary.KeyFindings
	summary.CriticalModules = aiOutput.ExecutiveSummary.CriticalModules
	summary.Trends = aiOutput.ExecutiveSummary.Trends
	summary.StrategicRecommendations = aiOutput.ExecutiveSummary.StrategicRecommendations
	summary.AIConfidence = aiOutput.Confidence
	summary.HasAISummary = true

	return summary
}
