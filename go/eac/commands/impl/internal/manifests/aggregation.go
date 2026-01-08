package manifests

import (
	"sort"
	"time"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
)

const (
	// Test status constants from manifest specification
	StatusPassed = "passed"
	StatusFailed = "failed"
)

// TotalCounts holds aggregated pass/fail counts across all tests
type TotalCounts struct {
	TotalPassed int `json:"total_passed" yaml:"total_passed"`
	TotalFailed int `json:"total_failed" yaml:"total_failed"`
}

// CompleteTestData holds all test data prepared for display or serialization
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
	SummaryBySuite []SuiteSummary     `json:"summary_by_suite" yaml:"summary_by_suite"`
}

// ModuleStats represents test statistics for a module
type ModuleStats struct {
	Module          string            `json:"module" yaml:"module"`
	Total           int               `json:"total" yaml:"total"`
	Passed          int               `json:"passed" yaml:"passed"`
	Failed          int               `json:"failed" yaml:"failed"`
	Skipped         int               `json:"skipped" yaml:"skipped"`
	DurationSeconds float64           `json:"duration_seconds" yaml:"duration_seconds"`
	SuiteCounts     map[string]int    `json:"suite_counts" yaml:"suite_counts"`
	Controls        []string          `json:"controls" yaml:"controls"`
	Tests           []TestResult      `json:"tests" yaml:"tests"`
}

// TypeSummary represents test counts by type
type TypeSummary struct {
	Type   string `json:"type" yaml:"type"`
	Count  int    `json:"count" yaml:"count"`
	Passed int    `json:"passed" yaml:"passed"`
	Failed int    `json:"failed" yaml:"failed"`
}

// SuiteSummary represents test counts by suite
type SuiteSummary struct {
	Suite  string `json:"suite" yaml:"suite"`
	Count  int    `json:"count" yaml:"count"`
	Passed int    `json:"passed" yaml:"passed"`
	Failed int    `json:"failed" yaml:"failed"`
}

// BuildModuleStats aggregates test data by module
func BuildModuleStats(manifestList []*implinternal.TestManifest, allTests []TestResult) []ModuleStats {
	stats := make([]ModuleStats, 0, len(manifestList))

	for _, m := range manifestList {
		// Filter tests for this module
		moduleTests := make([]TestResult, 0)
		for _, test := range allTests {
			if test.Module == m.Moniker {
				moduleTests = append(moduleTests, test)
			}
		}

		stat := ModuleStats{
			Module:          m.Moniker,
			Total:           m.Summary.Total,
			Passed:          m.Summary.Passed,
			Failed:          m.Summary.Failed,
			Skipped:         m.Summary.Skipped,
			DurationSeconds: m.DurationSeconds,
			SuiteCounts:     make(map[string]int),
			Controls:        []string{},
			Tests:           moduleTests,
		}

		// Count by suite - dynamically handle any suite from configuration
		for suiteName, suite := range m.Suites {
			stat.SuiteCounts[suiteName] = suite.Tests.Total
		}

		// Collect unique control tags
		controlSet := make(map[string]bool)
		for _, test := range m.Tests {
			for _, tag := range ExtractControlTags(test.Tags) {
				controlSet[tag] = true
			}
		}
		for control := range controlSet {
			stat.Controls = append(stat.Controls, control)
		}
		sort.Strings(stat.Controls)

		stats = append(stats, stat)
	}

	// Sort by module name
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Module < stats[j].Module
	})

	return stats
}

// BuildTypeSummary aggregates test counts by type
func BuildTypeSummary(tests []TestResult) []TypeSummary {
	typeCounts := make(map[string]*TypeSummary)

	for _, test := range tests {
		summary, exists := typeCounts[test.Type]
		if !exists {
			summary = &TypeSummary{Type: test.Type}
			typeCounts[test.Type] = summary
		}

		summary.Count++
		if test.Status == StatusPassed {
			summary.Passed++
		} else if test.Status == StatusFailed {
			summary.Failed++
		}
	}

	// Convert to slice and sort
	result := make([]TypeSummary, 0, len(typeCounts))
	for _, summary := range typeCounts {
		result = append(result, *summary)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})

	return result
}

// BuildSuiteSummary aggregates test counts by suite
func BuildSuiteSummary(tests []TestResult) []SuiteSummary {
	suiteCounts := make(map[string]*SuiteSummary)

	for _, test := range tests {
		summary, exists := suiteCounts[test.Suite]
		if !exists {
			summary = &SuiteSummary{Suite: test.Suite}
			suiteCounts[test.Suite] = summary
		}

		summary.Count++
		if test.Status == StatusPassed {
			summary.Passed++
		} else if test.Status == StatusFailed {
			summary.Failed++
		}
	}

	// Convert to slice and sort
	result := make([]SuiteSummary, 0, len(suiteCounts))
	for _, summary := range suiteCounts {
		result = append(result, *summary)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Suite < result[j].Suite
	})

	return result
}

// GetLatestRunTime returns the most recent test time across all manifests
func GetLatestRunTime(manifestList []*implinternal.TestManifest) time.Time {
	var latest time.Time
	for _, m := range manifestList {
		if m.TestTime.After(latest) {
			latest = m.TestTime
		}
	}
	return latest
}

// CalculateTotalCounts aggregates pass/fail counts from test results
func CalculateTotalCounts(tests []TestResult) TotalCounts {
	counts := TotalCounts{}
	for _, test := range tests {
		if test.Status == StatusPassed {
			counts.TotalPassed++
		} else if test.Status == StatusFailed {
			counts.TotalFailed++
		}
	}
	return counts
}

// BuildCompleteTestData aggregates all test data from manifests
// This is the single source of truth for test data preparation used by both get and show commands
func BuildCompleteTestData(manifestList []*implinternal.TestManifest) *CompleteTestData {
	// Extract base test results
	tests := ExtractTestResults(manifestList)

	// Calculate totals
	totals := CalculateTotalCounts(tests)

	// Build all aggregations
	data := &CompleteTestData{
		ModulesTested:  len(manifestList),
		LastRun:        GetLatestRunTime(manifestList),
		TotalTests:     len(tests),
		TotalPassed:    totals.TotalPassed,
		TotalFailed:    totals.TotalFailed,
		Tests:          tests,
		SpecCoverage:   BuildSpecCoverage(manifestList),
		ControlSummary: BuildControlSummary(manifestList),
		ModuleStats:    BuildModuleStats(manifestList, tests),
		SummaryByType:  BuildTypeSummary(tests),
		SummaryBySuite: BuildSuiteSummary(tests),
	}

	return data
}
