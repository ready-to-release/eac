package testing

import (
	"time"
)

// CompleteTestData holds all aggregated test data for display/serialization.
type CompleteTestData struct {
	ModulesTested  int                `json:"modules_tested" yaml:"modules_tested"`
	LastRun        time.Time          `json:"last_run" yaml:"last_run"`
	TotalTests     int                `json:"total_tests" yaml:"total_tests"`
	TotalPassed    int                `json:"total_passed" yaml:"total_passed"`
	TotalFailed    int                `json:"total_failed" yaml:"total_failed"`
	Tests          []TestResult       `json:"tests" yaml:"tests"`
	SpecCoverage   []SpecCoverage     `json:"spec_coverage" yaml:"spec_coverage"`
	ControlSummary []ControlSummary   `json:"control_summary" yaml:"control_summary"`
	ModuleStats    []ModuleStats      `json:"module_stats" yaml:"module_stats"`
	SummaryByType  []TypeSummary      `json:"summary_by_type" yaml:"summary_by_type"`
	SummaryBySuite []SuiteSummaryEntry `json:"summary_by_suite" yaml:"summary_by_suite"`
}

// TestResult extends TestEntry with control tags for consumer use.
type TestResult struct {
	Name        string   `json:"name" yaml:"name"`
	Module      string   `json:"module" yaml:"module"`
	Package     string   `json:"package" yaml:"package"`
	Type        string   `json:"type" yaml:"type"`
	Suite       string   `json:"suite" yaml:"suite"`
	Status      string   `json:"status" yaml:"status"`
	DurationMs  int64    `json:"duration_ms,omitempty" yaml:"durationms,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	FilePath    string   `json:"file_path,omitempty" yaml:"filepath,omitempty"`
	ControlTags []string `json:"control_tags,omitempty" yaml:"controltags,omitempty"`
	FeatureName string   `json:"feature_name,omitempty" yaml:"featurename,omitempty"`
	FeaturePath string   `json:"feature_path,omitempty" yaml:"featurepath,omitempty"`
}

// ModuleStats aggregates test statistics per module.
type ModuleStats struct {
	Module          string         `json:"module" yaml:"module"`
	Total           int            `json:"total" yaml:"total"`
	Passed          int            `json:"passed" yaml:"passed"`
	Failed          int            `json:"failed" yaml:"failed"`
	Skipped         int            `json:"skipped" yaml:"skipped"`
	DurationSeconds float64        `json:"duration_seconds" yaml:"duration_seconds"`
	SuiteCounts     map[string]int `json:"suite_counts" yaml:"suite_counts"`
	Controls        []string       `json:"controls" yaml:"controls"`
	Tests           []TestResult   `json:"tests" yaml:"tests"`
}

// SpecCoverage represents test coverage for a feature file.
type SpecCoverage struct {
	FeatureName   string   `json:"feature_name" yaml:"featurename"`
	FeaturePath   string   `json:"feature_path" yaml:"featurepath"`
	Module        string   `json:"module" yaml:"module"`
	ScenarioCount int      `json:"scenario_count" yaml:"scenariocount"`
	PassedCount   int      `json:"passed_count" yaml:"passedcount"`
	FailedCount   int      `json:"failed_count" yaml:"failedcount"`
	SkippedCount  int      `json:"skipped_count" yaml:"skippedcount"`
	Controls      []string `json:"controls,omitempty" yaml:"controls,omitempty"`
}

// ControlSummary aggregates test evidence for a security control.
type ControlSummary struct {
	ControlID    string   `json:"control_id" yaml:"controlid"`
	TestCount    int      `json:"test_count" yaml:"testcount"`
	ModuleCount  int      `json:"module_count" yaml:"modulecount"`
	Modules      []string `json:"modules" yaml:"modules"`
	PassedCount  int      `json:"passed_count" yaml:"passedcount"`
	FailedCount  int      `json:"failed_count" yaml:"failedcount"`
	SkippedCount int      `json:"skipped_count" yaml:"skippedcount"`
}

// TypeSummary aggregates test counts by type.
type TypeSummary struct {
	Type   string `json:"type" yaml:"type"`
	Count  int    `json:"count" yaml:"count"`
	Passed int    `json:"passed" yaml:"passed"`
	Failed int    `json:"failed" yaml:"failed"`
}

// SuiteSummaryEntry aggregates test counts by suite.
type SuiteSummaryEntry struct {
	Suite  string `json:"suite" yaml:"suite"`
	Count  int    `json:"count" yaml:"count"`
	Passed int    `json:"passed" yaml:"passed"`
	Failed int    `json:"failed" yaml:"failed"`
}
