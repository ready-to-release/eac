// File: go/eac/commands/impl/create/risk-assess/template_data.go
// Purpose: Data structures for risk assessment template rendering

package riskassess

import "github.com/ready-to-release/eac/go/eac/commands/internal/risk/scoring"

// RiskAssessmentReportData holds all data for rendering the risk assessment template
type RiskAssessmentReportData struct {
	// Header metadata
	GeneratedAt       string
	ScopeDescription  string // "Full Assessment (all 19 modules)" or "Subset Assessment (3 of 19 modules)"
	ProfileName       string
	TestSuite         string // Comma-separated list of test suites (singular for template compatibility)

	// Executive summary
	Summary ExecutiveSummary

	// Module results
	ModuleResults []ModuleReportData

	// Footer
	AggregatedReportPath string
}

// ExecutiveSummary holds summary statistics and AI-generated content
type ExecutiveSummary struct {
	// Basic stats (existing)
	TotalControls    int
	Satisfied        int
	NotSatisfied     int
	SatisfactionRate float64
	NotSatisfiedRate float64
	ModulesAssessed  int

	// AI-generated content (NEW)
	OverallRiskPosture       string                `json:"overall_risk_posture"`       // critical/high/moderate/low
	SummaryNarrative         string                `json:"summary_narrative"`          // 2-3 paragraph executive summary
	KeyFindings              []string              `json:"key_findings"`               // Bullet points
	CriticalModules          []CriticalModuleInfo  `json:"critical_modules"`           // Modules needing attention
	Trends                   []string              `json:"trends"`                     // Patterns across modules
	StrategicRecommendations []string              `json:"strategic_recommendations"`  // High-level actions
	AIConfidence             float64               `json:"ai_confidence"`              // 0.0-1.0
	HasAISummary             bool                  `json:"has_ai_summary"`             // Whether AI summary was generated
}

// CriticalModuleInfo holds critical module information from AI analysis
type CriticalModuleInfo struct {
	Module string `json:"module"`
	Reason string `json:"reason"`
}

// ModuleReportData holds per-module report data
type ModuleReportData struct {
	Module                    string
	Satisfied                 int
	NotSatisfied              int
	RiskScore                 *scoring.RiskScore
	RiskScoreFormatted        string
	TestEvidenceFormatted     string
	TestTypeBreakdown         string   // Test type breakdown (e.g., "45 godog, 12 gotest")
	SuiteSummary              string   // Suite-level summaries
	SecurityEvidenceFormatted string
	SatisfiedControls         []string
	NotSatisfiedControls      []string
	NotSatisfiedFindings      []FindingData
	ReportPath                string
	Warnings                  []string // Evidence collection warnings
}

// FindingData holds simplified finding information
type FindingData struct {
	ControlID string
	Title     string
}
