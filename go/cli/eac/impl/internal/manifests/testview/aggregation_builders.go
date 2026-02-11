package testview

import "sort"

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
