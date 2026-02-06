package testview

import (
	"sort"
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

// BuildCompleteTestData aggregates all test data from module views.
func BuildCompleteTestData(views []*TestModuleView) *CompleteTestData {
	tests := extractTestResults(views)

	totalPassed := 0
	totalFailed := 0
	for i := range tests {
		switch tests[i].Status {
		case StatusPassed:
			totalPassed++
		case StatusFailed:
			totalFailed++
		}
	}

	data := &CompleteTestData{
		ModulesTested:  len(views),
		LastRun:        getLatestRunTime(views),
		TotalTests:     len(tests),
		TotalPassed:    totalPassed,
		TotalFailed:    totalFailed,
		Tests:          tests,
		SpecCoverage:   buildSpecCoverage(views),
		ControlSummary: buildControlSummary(views),
		ModuleStats:    buildModuleStats(views, tests),
		SummaryByType:  buildTypeSummary(tests),
		SummaryBySuite: buildSuiteSummary(tests),
	}

	return data
}

func extractTestResults(views []*TestModuleView) []TestResult {
	var results []TestResult
	for _, view := range views {
		for i := range view.Tests {
			test := &view.Tests[i]
			result := TestResult{
				Name:        test.Name,
				Module:      view.Module,
				Package:     test.Package,
				Type:        test.Type,
				Suite:       test.Suite,
				Status:      test.Status,
				DurationMs:  test.DurationMs,
				Tags:        StripTagPrefixes(test.Tags),
				FilePath:    test.FilePath,
				ControlTags: ExtractControlTags(test.Tags),
			}

			if test.Type == "godog" && test.FilePath != "" {
				result.FeatureName = ExtractFeatureName(test.FilePath)
				result.FeaturePath = test.FilePath
			}

			results = append(results, result)
		}
	}
	return results
}

func getLatestRunTime(views []*TestModuleView) time.Time {
	var latest time.Time
	for _, v := range views {
		if v.ExecutedAt.After(latest) {
			latest = v.ExecutedAt
		}
	}
	return latest
}

func buildModuleStats(views []*TestModuleView, allTests []TestResult) []ModuleStats {
	stats := make([]ModuleStats, 0, len(views))

	for _, v := range views {
		// Filter tests for this module
		var moduleTests []TestResult
		for i := range allTests {
			if allTests[i].Module == v.Module {
				moduleTests = append(moduleTests, allTests[i])
			}
		}

		stat := ModuleStats{
			Module:          v.Module,
			Total:           v.Summary.Total,
			Passed:          v.Summary.Passed,
			Failed:          v.Summary.Failed,
			Skipped:         v.Summary.Skipped,
			DurationSeconds: v.Duration.Seconds(),
			SuiteCounts:     make(map[string]int),
			Controls:        []string{},
			Tests:           moduleTests,
		}

		// Count by suite
		for suiteName, suite := range v.Suites {
			stat.SuiteCounts[suiteName] = suite.Total
		}

		// Collect unique control tags
		controlSet := make(map[string]bool)
		for i := range v.Tests {
			for _, tag := range ExtractControlTags(v.Tests[i].Tags) {
				controlSet[tag] = true
			}
		}
		for control := range controlSet {
			stat.Controls = append(stat.Controls, control)
		}
		sort.Strings(stat.Controls)

		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Module < stats[j].Module
	})

	return stats
}

func buildTypeSummary(tests []TestResult) []TypeSummary {
	typeCounts := make(map[string]*TypeSummary)

	for i := range tests {
		summary, exists := typeCounts[tests[i].Type]
		if !exists {
			summary = &TypeSummary{Type: tests[i].Type}
			typeCounts[tests[i].Type] = summary
		}
		summary.Count++
		switch tests[i].Status {
		case StatusPassed:
			summary.Passed++
		case StatusFailed:
			summary.Failed++
		}
	}

	result := make([]TypeSummary, 0, len(typeCounts))
	for _, summary := range typeCounts {
		result = append(result, *summary)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})

	return result
}

func buildSuiteSummary(tests []TestResult) []SuiteSummaryEntry {
	suiteCounts := make(map[string]*SuiteSummaryEntry)

	for i := range tests {
		summary, exists := suiteCounts[tests[i].Suite]
		if !exists {
			summary = &SuiteSummaryEntry{Suite: tests[i].Suite}
			suiteCounts[tests[i].Suite] = summary
		}
		summary.Count++
		switch tests[i].Status {
		case StatusPassed:
			summary.Passed++
		case StatusFailed:
			summary.Failed++
		}
	}

	result := make([]SuiteSummaryEntry, 0, len(suiteCounts))
	for _, summary := range suiteCounts {
		result = append(result, *summary)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Suite < result[j].Suite
	})

	return result
}

func buildSpecCoverage(views []*TestModuleView) []SpecCoverage {
	coverageMap := make(map[string]*SpecCoverage)

	for _, view := range views {
		for i := range view.Tests {
			test := &view.Tests[i]
			if test.Type != "godog" {
				continue
			}

			featurePath := test.FilePath
			if featurePath == "" {
				featurePath = test.Package
			}
			if featurePath == "" {
				continue
			}

			featureName := ExtractFeatureName(featurePath)

			coverage, exists := coverageMap[featurePath]
			if !exists {
				coverage = &SpecCoverage{
					FeatureName: featureName,
					FeaturePath: featurePath,
					Module:      view.Module,
					Controls:    []string{},
				}
				coverageMap[featurePath] = coverage
			}

			coverage.ScenarioCount++
			switch test.Status {
			case StatusPassed:
				coverage.PassedCount++
			case StatusFailed:
				coverage.FailedCount++
			case StatusSkipped:
				coverage.SkippedCount++
			}

			controlTags := ExtractControlTags(test.Tags)
			for _, control := range controlTags {
				if !containsStr(coverage.Controls, control) {
					coverage.Controls = append(coverage.Controls, control)
				}
			}
		}
	}

	var result []SpecCoverage
	for _, coverage := range coverageMap {
		sort.Strings(coverage.Controls)
		result = append(result, *coverage)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Module != result[j].Module {
			return result[i].Module < result[j].Module
		}
		return result[i].FeatureName < result[j].FeatureName
	})

	return result
}

func buildControlSummary(views []*TestModuleView) []ControlSummary {
	summaryMap := make(map[string]*ControlSummary)
	modulesByControl := make(map[string]map[string]bool)

	for _, view := range views {
		for i := range view.Tests {
			test := &view.Tests[i]
			controlTags := ExtractControlTags(test.Tags)

			for _, controlID := range controlTags {
				summary, exists := summaryMap[controlID]
				if !exists {
					summary = &ControlSummary{
						ControlID: controlID,
						Modules:   []string{},
					}
					summaryMap[controlID] = summary
					modulesByControl[controlID] = make(map[string]bool)
				}

				summary.TestCount++
				switch test.Status {
				case StatusPassed:
					summary.PassedCount++
				case StatusFailed:
					summary.FailedCount++
				case StatusSkipped:
					summary.SkippedCount++
				}

				modulesByControl[controlID][view.Module] = true
			}
		}
	}

	var result []ControlSummary
	for controlID, summary := range summaryMap {
		mods := make([]string, 0, len(modulesByControl[controlID]))
		for module := range modulesByControl[controlID] {
			mods = append(mods, module)
		}
		sort.Strings(mods)

		summary.Modules = mods
		summary.ModuleCount = len(mods)
		result = append(result, *summary)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ControlID < result[j].ControlID
	})

	return result
}

func containsStr(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
