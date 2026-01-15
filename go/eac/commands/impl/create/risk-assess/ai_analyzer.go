package riskassess

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"
	coreai "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// defaultRiskAssessPrompt is the fallback prompt when template is not found
const defaultRiskAssessPrompt = `# Risk Assessment - Executive Summary

Generate an executive summary analyzing overall risk posture. Output ONLY valid JSON with 2 fields: executive_summary and confidence.

## Output Format

{
  "executive_summary": {
    "overall_risk_posture": "moderate",
    "summary_narrative": "2-3 paragraph narrative (100-1000 chars)",
    "key_findings": ["Finding 1 (20+ chars)", "Finding 2", "Finding 3"],
    "strategic_recommendations": ["Recommendation 1 (30+ chars)", "Recommendation 2"]
  },
  "confidence": 0.85
}

Assessment data:`

// AIRiskAssessmentInput holds input for unified AI analysis
type AIRiskAssessmentInput struct {
	Modules              []ModuleAnalysisInput
	ProfileName          string
	TotalControls        int
	SatisfiedControls    int
	NotSatisfiedControls int
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
	ModuleAnalyses   []ModuleAnalysisData `json:"module_analyses,omitempty"`
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
	// Build comprehensive prompt with template
	prompt, err := buildRiskAssessmentPrompt(config.WorkspaceRoot, input)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Create AI executor
	executor := ai.NewExecutor(config.WorkspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match validation.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Load AI config for retry strategy
	aiConfig, err := coreai.LoadAIConfig(config.WorkspaceRoot)
	if err != nil {
		logging.C().Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	// Build retry configuration using factory (validator auto-loaded for FormatJSON)
	retryConfig, err := coreai.BuildRetryConfig(
		coreai.TypeRiskAssess,
		coreai.FormatJSON, // JSON format with automatic enhanced validation
		executorAdapter,
		nil, // Let BuildRetryConfig auto-load JSON validator
		config.WorkspaceRoot,
		aiConfig,
		coreai.WithDebug(config.Debug),
		coreai.WithLogger(logging.C().Zap()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry
	result, err := coreai.GenerateWithRetry(ctx, retryConfig, prompt)
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

// buildRiskAssessmentPrompt loads template and appends aggregate statistics
func buildRiskAssessmentPrompt(workspaceRoot string, input *AIRiskAssessmentInput) (string, error) {
	// Load prompt template with three-tier priority:
	// 1. Team override (.r2r/eac/templates/ai/risk-assess/risk-assess.md)
	// 2. System default (templates/ai/risk-assess/risk-assess.md)
	loader := coreai.NewContractLoader(workspaceRoot, coreai.TypeRiskAssess, "")
	promptTemplate, _, err := loader.LoadPrompt("", defaultRiskAssessPrompt)
	if err != nil {
		promptTemplate = defaultRiskAssessPrompt
	}

	var sb strings.Builder
	sb.WriteString(promptTemplate)
	sb.WriteString("\n\n")

	// Add aggregate summary information only
	sb.WriteString("Profile: ")
	sb.WriteString(input.ProfileName)
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf("Modules assessed: %d\n", len(input.Modules)))
	sb.WriteString(fmt.Sprintf("Total controls: %d\n", input.TotalControls))
	sb.WriteString(fmt.Sprintf("Controls satisfied: %d\n", input.SatisfiedControls))
	sb.WriteString(fmt.Sprintf("Controls not satisfied: %d\n", input.NotSatisfiedControls))

	// Calculate satisfaction rate
	satisfactionRate := 0.0
	if input.TotalControls > 0 {
		satisfactionRate = float64(input.SatisfiedControls) / float64(input.TotalControls) * 100
	}
	sb.WriteString(fmt.Sprintf("Control satisfaction rate: %.1f%%\n\n", satisfactionRate))

	// Count modules by control status
	modulesWithControls := 0
	modulesWithoutControls := 0
	for _, mod := range input.Modules {
		if mod.ControlsSatisfied > 0 {
			modulesWithControls++
		} else {
			modulesWithoutControls++
		}
	}

	sb.WriteString(fmt.Sprintf("Modules with controls implemented: %d\n", modulesWithControls))
	sb.WriteString(fmt.Sprintf("Modules without controls: %d\n", modulesWithoutControls))

	return sb.String(), nil
}

// ApplyAIRiskAssessment applies AI-generated results to module assessment results
func ApplyAIRiskAssessment(results []*ModuleAssessmentResult, aiOutput *AIRiskAssessmentOutput, log *logging.ComponentLogger) {
	// Check if per-module analysis was generated
	if len(aiOutput.ModuleAnalyses) == 0 {
		return
	}

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

			log.Debugf("Applied AI risk score: module=%s, likelihood=%d, score=%d",
				result.Module, analysis.ComputedLikelihood, result.RiskScore.Score)
		} else {
			log.Warnf("No AI analysis found for module: %s", result.Module)
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
