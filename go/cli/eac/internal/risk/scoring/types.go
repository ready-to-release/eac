// Package scoring provides risk scoring utilities following ISO 27005 methodology.
//
// Risk Score = Likelihood × Impact
//
// Likelihood and Impact are both rated 1-5:
//
//	1 = Very Low / Rare
//	2 = Low / Unlikely
//	3 = Medium / Possible
//	4 = High / Likely
//	5 = Very High / Almost Certain
//
// Risk Score ranges:
//
//	20-25: Critical (Red)
//	12-19: High (Orange)
//	6-11:  Medium (Yellow)
//	1-5:   Low (Green)
package scoring

import "fmt"

// RiskBand represents risk level categories.
type RiskBand string

const (
	RiskCritical RiskBand = "Critical" // 20-25
	RiskHigh     RiskBand = "High"     // 12-19
	RiskMedium   RiskBand = "Medium"   // 6-11
	RiskLow      RiskBand = "Low"      // 1-5
)

// RiskScore holds computed risk information.
type RiskScore struct {
	Module     string   `json:"module"`
	Likelihood int      `json:"likelihood"` // 1-5
	Impact     int      `json:"impact"`     // 1-5
	Score      int      `json:"score"`      // Likelihood × Impact
	Band       RiskBand `json:"band"`       // Critical/High/Medium/Low
	Reasoning  string   `json:"reasoning,omitempty"`
	Confidence float64  `json:"confidence,omitempty"` // 0.0-1.0
}

// AIRiskAnalysis represents the AI-generated risk analysis.
type AIRiskAnalysis struct {
	ComputedLikelihood  int      `json:"computed_likelihood"`
	Reasoning           string   `json:"reasoning"`
	RiskSummary         string   `json:"risk_summary"`
	RecommendedControls []string `json:"recommended_controls,omitempty"`
	Confidence          float64  `json:"confidence"` // 0.0-1.0
}

// VulnerabilityInput represents vulnerability data sent to AI for analysis.
type VulnerabilityInput struct {
	Scanner        string  `json:"scanner"`
	Severity       string  `json:"severity"`
	Vulnerability  string  `json:"vulnerability"`
	Description    string  `json:"description,omitempty"`
	CVSSScore      float64 `json:"cvss_score,omitempty"`
	Exploitability string  `json:"exploitability,omitempty"`
	Package        string  `json:"package,omitempty"`
}

// ModuleContext provides context about the module for AI analysis.
type ModuleContext struct {
	ComponentType string `json:"component_type"`
	Criticality      string   `json:"criticality"` // high, medium, low
	ExistingControls []string `json:"existing_controls,omitempty"`
}

// AIAnalysisInput is the complete input structure sent to AI.
type AIAnalysisInput struct {
	Module   string               `json:"module"`
	Findings []VulnerabilityInput `json:"findings"`
	Context  ModuleContext        `json:"context"`
}

// ControlFinding holds finding information for a specific control.
type ControlFinding struct {
	ControlID    string   `json:"control_id"`
	State        string   `json:"state"` // satisfied, not-satisfied
	TestsPassed  int      `json:"tests_passed"`
	TestsTotal   int      `json:"tests_total"`
	TestsFailed  int      `json:"tests_failed"`
	VulnCount    int      `json:"vuln_count"`
	VulnSeverity string   `json:"vuln_severity,omitempty"` // Highest severity
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

// CalculateBaseLikelihood computes likelihood from vulnerability severity counts.
// Uses a weighted severity approach:
//   - No findings: 1 (Rare)
//   - LOW severity: +1
//   - MEDIUM severity: +2
//   - HIGH severity: +3
//   - CRITICAL severity: +4
func CalculateBaseLikelihood(critical, high, medium, low int) int {
	likelihood := 1 // Base: Rare

	// Apply severity weights (cap at 5)
	if critical > 0 {
		likelihood += 4
	} else if high > 0 {
		likelihood += 3
	} else if medium > 0 {
		likelihood += 2
	} else if low > 0 {
		likelihood += 1
	}

	// Cap at 5 (Almost Certain)
	if likelihood > 5 {
		likelihood = 5
	}

	return likelihood
}

// ApplyControlReductions reduces likelihood based on passing controls.
// Each control with reduction_likelihood value reduces the score.
func ApplyControlReductions(baseLikelihood int, reductions []int) int {
	adjusted := baseLikelihood

	for _, reduction := range reductions {
		adjusted -= reduction
	}

	// Minimum likelihood is 1
	if adjusted < 1 {
		adjusted = 1
	}

	return adjusted
}

// CalculateRiskScore computes final risk score.
func CalculateRiskScore(likelihood, impact int) int {
	return likelihood * impact
}

// DetermineRiskBand maps risk score to band.
func DetermineRiskBand(score int) RiskBand {
	switch {
	case score >= 20:
		return RiskCritical
	case score >= 12:
		return RiskHigh
	case score >= 6:
		return RiskMedium
	default:
		return RiskLow
	}
}

// ComputeRiskScore calculates a complete risk score from inputs.
func ComputeRiskScore(module string, likelihood, impact int, reasoning string) *RiskScore {
	score := CalculateRiskScore(likelihood, impact)
	return &RiskScore{
		Module:     module,
		Likelihood: likelihood,
		Impact:     impact,
		Score:      score,
		Band:       DetermineRiskBand(score),
		Reasoning:  reasoning,
	}
}

// GetDefaultImpact returns default impact rating based on module type.
// Uses configurable mappings from risk-config.yml.
// Can be overridden by module metadata.
func GetDefaultImpact(moniker string) int {
	return GetRiskScoringConfig().GetImpact(moniker)
}

// FormatRiskBandColor returns ANSI color code for risk band.
func FormatRiskBandColor(band RiskBand) string {
	switch band {
	case RiskCritical:
		return "\033[31m" // Red
	case RiskHigh:
		return "\033[33m" // Orange/Yellow
	case RiskMedium:
		return "\033[93m" // Bright Yellow
	case RiskLow:
		return "\033[32m" // Green
	default:
		return "\033[0m" // Reset
	}
}

// ResetColor returns ANSI reset code.
func ResetColor() string {
	return "\033[0m"
}

// FormatRiskScore formats a risk score for display.
func FormatRiskScore(rs *RiskScore) string {
	color := FormatRiskBandColor(rs.Band)
	reset := ResetColor()
	return fmt.Sprintf("%s%s (%d)%s", color, rs.Band, rs.Score, reset)
}

// FormatRiskScorePlain formats risk score without ANSI color codes (for Markdown/text output).
func FormatRiskScorePlain(rs *RiskScore) string {
	return fmt.Sprintf("%s (%d)", rs.Band, rs.Score)
}
