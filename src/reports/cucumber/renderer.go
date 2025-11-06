// Package cucumber provides rendering of Cucumber test reports to markdown
package cucumber

import (
	"fmt"
	"sort"
	"strings"
)

// GitContext provides Git repository context for generating GitHub links
// This is a lightweight interface to avoid circular dependencies
type GitContext interface {
	BuildGitHubFileURL(filePath string) string
}

// RenderTestSummary renders test results in the format shown in implementation-report.md
// This renders a single feature with its rules and scenarios organized by acceptance criteria
func RenderTestSummary(feature *Feature, gitCtx GitContext) string {
	var buf strings.Builder

	// Feature header
	featureID := feature.GetFeatureID()
	buf.WriteString(fmt.Sprintf("**Feature ID**: `%s`\n", featureID))

	// Extract user story from description
	userStory := extractUserStory(feature.Description)
	if userStory != "" {
		buf.WriteString(fmt.Sprintf("**User Story**: %s\n", userStory))
	}

	// Specification link - internal link to Appendix A
	featureName := extractFeatureName(featureID)
	buf.WriteString(fmt.Sprintf("**Specification**: [specification.feature](#%s)\n\n", featureName))

	// Group scenarios by acceptance criteria
	scenariosByAC := groupScenariosByAC(feature.Elements)

	// Sort AC keys for consistent output (AC1, AC2, AC3, etc.)
	acKeys := make([]string, 0, len(scenariosByAC))
	for ac := range scenariosByAC {
		acKeys = append(acKeys, ac)
	}
	sort.Strings(acKeys)

	// Render each rule (acceptance criterion) with its scenarios
	for _, ac := range acKeys {
		scenarios := scenariosByAC[ac]
		if len(scenarios) == 0 {
			continue
		}

		// Rule header - we need to infer the rule text from scenario context
		// For now, use AC number as the rule identifier
		buf.WriteString(fmt.Sprintf("**Rule %s** (%s)\n\n", strings.TrimPrefix(ac, "AC"), ac))

		// Scenario table
		buf.WriteString("| Scenario | Tags | Result |\n")
		buf.WriteString("|----------|------|--------|\n")

		for _, scenario := range scenarios {
			name := scenario.Name
			tags := scenario.GetTagString()
			status := scenario.GetStatus()
			statusIcon := getStatusIcon(status)

			buf.WriteString(fmt.Sprintf("| %s | %s | %s |\n", name, tags, statusIcon))
		}

		buf.WriteString("\n")
	}

	return buf.String()
}

// RenderAllFeatures renders all features in test summary format
func RenderAllFeatures(report CucumberReport, gitCtx GitContext) string {
	var buf strings.Builder

	for i, feature := range report {
		buf.WriteString(RenderTestSummary(&feature, gitCtx))

		// Add separator between features (but not after the last one)
		if i < len(report)-1 {
			buf.WriteString("---\n\n")
		}
	}

	return buf.String()
}

// RenderByVerificationType renders scenarios filtered by verification type (IV/OV/PV)
// and grouped by feature
func RenderByVerificationType(report CucumberReport, verificationType string) string {
	var buf strings.Builder

	// Group by feature ID
	featureMap := make(map[string]*featureScenarios)

	for _, feature := range report {
		featureID := feature.GetFeatureID()

		// Filter scenarios for this verification type
		var filteredScenarios []Scenario
		for _, scenario := range feature.Elements {
			if scenario.GetVerificationType() == verificationType {
				filteredScenarios = append(filteredScenarios, scenario)
			}
		}

		if len(filteredScenarios) > 0 {
			featureMap[featureID] = &featureScenarios{
				Feature:   &feature,
				Scenarios: filteredScenarios,
			}
		}
	}

	// Sort feature IDs for consistent output
	featureIDs := make([]string, 0, len(featureMap))
	for id := range featureMap {
		featureIDs = append(featureIDs, id)
	}
	sort.Strings(featureIDs)

	// Render each feature with filtered scenarios
	for i, featureID := range featureIDs {
		fs := featureMap[featureID]

		// Feature header
		buf.WriteString(fmt.Sprintf("**Feature ID**: `%s`\n", featureID))

		userStory := extractUserStory(fs.Feature.Description)
		if userStory != "" {
			buf.WriteString(fmt.Sprintf("**User Story**: %s\n", userStory))
		}

		buf.WriteString(fmt.Sprintf("**Specification**: [specification.feature](%s)\n\n", fs.Feature.URI))

		// Group filtered scenarios by AC
		scenariosByAC := groupScenariosByAC(fs.Scenarios)

		acKeys := make([]string, 0, len(scenariosByAC))
		for ac := range scenariosByAC {
			acKeys = append(acKeys, ac)
		}
		sort.Strings(acKeys)

		for _, ac := range acKeys {
			scenarios := scenariosByAC[ac]
			buf.WriteString(fmt.Sprintf("**Rule %s** (%s)\n\n", strings.TrimPrefix(ac, "AC"), ac))

			buf.WriteString("| Scenario | Tags | Result |\n")
			buf.WriteString("|----------|------|--------|\n")

			for _, scenario := range scenarios {
				buf.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
					scenario.Name,
					scenario.GetTagString(),
					getStatusIcon(scenario.GetStatus())))
			}

			buf.WriteString("\n")
		}

		// Add separator between features
		if i < len(featureIDs)-1 {
			buf.WriteString("---\n\n")
		}
	}

	return buf.String()
}

// Helper types and functions

type featureScenarios struct {
	Feature   *Feature
	Scenarios []Scenario
}

// groupScenariosByAC groups scenarios by acceptance criteria tag
func groupScenariosByAC(scenarios []Scenario) map[string][]Scenario {
	result := make(map[string][]Scenario)

	for _, scenario := range scenarios {
		ac := scenario.GetAcceptanceCriteria()
		if ac == "" {
			ac = "NO_AC"
		}
		result[ac] = append(result[ac], scenario)
	}

	return result
}

// extractUserStory extracts the user story from feature description
// Expected format: "As a [role]\n  I want [capability]\n  So that [benefit]"
func extractUserStory(description string) string {
	// Remove leading/trailing whitespace
	description = strings.TrimSpace(description)

	if description == "" {
		return ""
	}

	// Replace newlines with commas for inline display
	lines := strings.Split(description, "\n")
	var parts []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts = append(parts, line)
		}
	}

	return strings.Join(parts, ", ")
}

// getStatusIcon returns the appropriate emoji for a test status
func getStatusIcon(status string) string {
	switch status {
	case "passed":
		return "🟢 Passed"
	case "failed":
		return "🔴 Failed"
	case "skipped":
		return "🟡 Skipped"
	case "pending":
		return "🟡 Pending"
	case "undefined":
		return "⚪ Undefined"
	default:
		return "⚫ " + status
	}
}

// normalizeFeaturePath normalizes a feature file path from cucumber.json
// Handles relative paths (e.g., "../../../specs/..." -> "specs/...")
func normalizeFeaturePath(filePath string) string {
	// Remove relative path prefixes (../../../ -> "")
	// Cucumber outputs paths relative to the test directory
	for strings.HasPrefix(filePath, "../") {
		filePath = strings.TrimPrefix(filePath, "../")
	}

	// Normalize path separators to forward slashes
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	return filePath
}
