package testing

import "time"

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

func containsStr(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
