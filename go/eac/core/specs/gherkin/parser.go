// Package gherkin provides parsing utilities for Gherkin feature files.
package gherkin

import (
	"fmt"
	"os"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
)

// ParseFile parses a Gherkin feature file and returns all scenarios with full details.
func ParseFile(filePath string) ([]ScenarioDetail, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse using official gherkin-go library
	doc, err := gherkin.ParseGherkinDocument(strings.NewReader(string(content)), (&messages.Incrementing{}).NewId)
	if err != nil {
		return nil, fmt.Errorf("parsing Gherkin: %w", err)
	}

	if doc.Feature == nil {
		return nil, fmt.Errorf("no feature found in file")
	}

	return extractScenarios(doc, filePath), nil
}

// extractScenarios extracts all scenarios from a parsed Gherkin document.
func extractScenarios(doc *messages.GherkinDocument, filePath string) []ScenarioDetail {
	var scenarios []ScenarioDetail

	feature := doc.Feature
	featureName := feature.Name
	featureTags := extractTagsFromMessages(feature.Tags)

	for _, child := range feature.Children {
		// Handle Scenario
		if child.Scenario != nil {
			scenario := extractScenario(child.Scenario, featureName, featureTags, filePath)
			scenarios = append(scenarios, scenario)
		}

		// Handle Rule (contains scenarios)
		if child.Rule != nil {
			for _, ruleChild := range child.Rule.Children {
				if ruleChild.Scenario != nil {
					scenario := extractScenario(ruleChild.Scenario, featureName, featureTags, filePath)
					scenarios = append(scenarios, scenario)
				}
			}
		}
	}

	return scenarios
}

// extractScenario converts a gherkin Scenario to our ScenarioDetail type.
func extractScenario(scenario *messages.Scenario, featureName string, featureTags []string, filePath string) ScenarioDetail {
	scenarioTags := extractTagsFromMessages(scenario.Tags)

	// Combine feature and scenario tags (all tags apply to the scenario)
	allTags := CombineTags(featureTags, scenarioTags)

	steps := extractSteps(scenario.Steps)
	description := strings.TrimSpace(scenario.Description)

	return ScenarioDetail{
		Name:        scenario.Name,
		Tags:        allTags,
		Steps:       steps,
		FeatureName: featureName,
		FeatureTags: featureTags,
		FilePath:    filePath,
		Description: description,
		LineNumber:  int(scenario.Location.Line),
	}
}

// extractTagsFromMessages converts gherkin message tags to string slice.
func extractTagsFromMessages(tags []*messages.Tag) []string {
	result := make([]string, len(tags))
	for i, tag := range tags {
		result[i] = tag.Name
	}
	return result
}

// extractSteps converts gherkin steps to formatted strings.
func extractSteps(steps []*messages.Step) []string {
	result := make([]string, len(steps))
	for i, step := range steps {
		// Format: "Given I have a calculator"
		result[i] = step.Keyword + step.Text

		// Add DataTable if present
		if step.DataTable != nil {
			result[i] += "\n" + formatDataTable(step.DataTable)
		}

		// Add DocString if present
		if step.DocString != nil {
			result[i] += "\n" + formatDocString(step.DocString)
		}
	}
	return result
}

// formatDataTable formats a data table for display.
func formatDataTable(table *messages.DataTable) string {
	var sb strings.Builder
	for _, row := range table.Rows {
		sb.WriteString("| ")
		for i, cell := range row.Cells {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(cell.Value)
		}
		sb.WriteString(" |\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// formatDocString formats a doc string for display.
func formatDocString(docString *messages.DocString) string {
	delimiter := `"""`
	if docString.Delimiter != "" {
		delimiter = docString.Delimiter
	}
	return delimiter + "\n" + docString.Content + "\n" + delimiter
}

// CombineTags merges feature and scenario tags, deduplicating.
func CombineTags(featureTags, scenarioTags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range featureTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	for _, tag := range scenarioTags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}
