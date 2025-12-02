// Package cucumber provides parsing and processing of Cucumber JSON test reports
package cucumber

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseFile reads and parses a Cucumber JSON file
func ParseFile(filePath string) (CucumberReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var report CucumberReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return report, nil
}

// GroupByFeature organizes scenarios by feature
// Returns a map of Feature ID -> Feature with its scenarios
func (report CucumberReport) GroupByFeature() map[string]*Feature {
	result := make(map[string]*Feature)

	for i := range report {
		feature := &report[i]
		featureID := feature.GetFeatureID()
		result[featureID] = feature
	}

	return result
}

// FilterByVerificationType filters scenarios by verification type (IV/OV/PV)
func (report CucumberReport) FilterByVerificationType(verificationType string) []ScenarioWithFeature {
	var results []ScenarioWithFeature

	for _, feature := range report {
		for _, scenario := range feature.Elements {
			if scenario.GetVerificationType() == verificationType {
				results = append(results, ScenarioWithFeature{
					Feature:  &feature,
					Scenario: &scenario,
				})
			}
		}
	}

	return results
}

// ScenarioWithFeature pairs a scenario with its parent feature
type ScenarioWithFeature struct {
	Feature  *Feature
	Scenario *Scenario
}

// GroupScenariosByRule groups scenarios by their acceptance criteria (Rule)
// Returns a map of AC tag (e.g., "AC1") -> scenarios
func GroupScenariosByRule(scenarios []ScenarioWithFeature) map[string][]ScenarioWithFeature {
	result := make(map[string][]ScenarioWithFeature)

	for _, sw := range scenarios {
		ac := sw.Scenario.GetAcceptanceCriteria()
		if ac == "" {
			ac = "NO_AC" // Scenarios without AC tag
		}
		result[ac] = append(result[ac], sw)
	}

	return result
}
