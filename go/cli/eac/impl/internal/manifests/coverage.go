package manifests

import (
	"sort"

	implinternal "github.com/ready-to-release/eac/go/cli/eac/impl/internal"
)

// SpecCoverage represents test coverage for a single feature file.
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

// BuildSpecCoverage builds specification coverage from godog tests in manifests.
// Groups scenarios by feature file and aggregates pass/fail counts and control tags.
func BuildSpecCoverage(manifests []*implinternal.TestManifest) []SpecCoverage {
	// Map: featurePath -> coverage data
	coverageMap := make(map[string]*SpecCoverage)

	for _, manifest := range manifests {
		for i := range manifest.Tests {
			test := &manifest.Tests[i]
			// Only process godog tests
			if test.Type != "godog" {
				continue
			}

			// Use Package field as fallback if FilePath is empty
			// In some environments (containers), FilePath may not be populated
			featurePath := test.FilePath
			if featurePath == "" {
				featurePath = test.Package
			}

			// Skip if no path available
			if featurePath == "" {
				continue
			}

			featureName := ExtractFeatureName(featurePath)

			// Get or create coverage entry
			coverage, exists := coverageMap[featurePath]
			if !exists {
				coverage = &SpecCoverage{
					FeatureName: featureName,
					FeaturePath: featurePath,
					Module:      manifest.Moniker,
					Controls:    []string{},
				}
				coverageMap[featurePath] = coverage
			}

			// Increment scenario count
			coverage.ScenarioCount++

			// Increment status counts
			switch test.Status {
			case implinternal.TestStatusPassed:
				coverage.PassedCount++
			case implinternal.TestStatusFailed:
				coverage.FailedCount++
			case implinternal.TestStatusSkipped:
				coverage.SkippedCount++
			}

			// Collect control tags
			controlTags := ExtractControlTags(test.Tags)
			for _, control := range controlTags {
				if !contains(coverage.Controls, control) {
					coverage.Controls = append(coverage.Controls, control)
				}
			}
		}
	}

	// Convert map to slice
	var result []SpecCoverage
	for _, coverage := range coverageMap {
		// Sort controls for consistent output
		sort.Strings(coverage.Controls)
		result = append(result, *coverage)
	}

	// Sort by module then feature name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Module != result[j].Module {
			return result[i].Module < result[j].Module
		}
		return result[i].FeatureName < result[j].FeatureName
	})

	return result
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
