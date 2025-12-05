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
	TestSuite         string

	// Executive summary
	Summary ExecutiveSummary

	// Module results
	ModuleResults []ModuleReportData

	// Footer
	AggregatedReportPath string
}

// ExecutiveSummary holds summary statistics
type ExecutiveSummary struct {
	TotalControls      int
	Satisfied          int
	NotSatisfied       int
	SatisfactionRate   float64
	NotSatisfiedRate   float64
	ModulesAssessed    int
}

// ModuleReportData holds per-module report data
type ModuleReportData struct {
	Module                    string
	Satisfied                 int
	NotSatisfied              int
	RiskScore                 *scoring.RiskScore
	RiskScoreFormatted        string
	TestEvidenceFormatted     string
	SecurityEvidenceFormatted string
	SatisfiedControls         []string
	NotSatisfiedControls      []string
	NotSatisfiedFindings      []FindingData
	ReportPath                string
}

// FindingData holds simplified finding information
type FindingData struct {
	ControlID string
	Title     string
}
