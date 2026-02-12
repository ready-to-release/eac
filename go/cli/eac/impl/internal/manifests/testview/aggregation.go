package testview

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
