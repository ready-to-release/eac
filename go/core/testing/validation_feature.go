package testing

import (
	"fmt"
	"os"
	"strings"
)

// ValidateFeatureLevelTags validates that Feature-level tags don't conflict with Scenario tags.
// In Gherkin, scenarios INHERIT all tags from the Feature level. This causes problems when:
//   - Feature has @L2 tag and Scenario has @L3 tag -> Scenario gets BOTH @L2 AND @L3 (invalid)
//   - Feature has @ov tag and Scenario has @iv tag -> Scenario gets BOTH @ov AND @iv (invalid)
//
// This function detects conflicts for:
//   - L-level tags (@L0-@L4) - must have exactly one
//   - Verification tags (@ov/@iv/@pv/@piv/@ppv) - validation checks for multiple
//
// This function detects these anti-patterns and returns validation errors.
func ValidateFeatureLevelTags(featureFilePath string) ([]string, error) {
	content, err := os.ReadFile(featureFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	featureLevelTags, scenarios := parseFeatureStructure(lines)
	return checkTagConflicts(featureLevelTags, scenarios, lines), nil
}

// scenarioInfo holds parsed scenario location and tags.
type scenarioInfo struct {
	lineNum int
	name    string
	tags    []string
}

// parseFeatureStructure extracts feature-level tags and scenario positions.
func parseFeatureStructure(lines []string) ([]string, []scenarioInfo) {
	var featureLevelTags []string
	var scenarios []scenarioInfo
	var inFeature bool

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Feature:") {
			inFeature = true
			continue
		}

		// Extract feature-level tags (before Feature:)
		if strings.HasPrefix(trimmed, "@") && !inFeature {
			featureLevelTags = append(featureLevelTags, extractTagsFromLine(trimmed)...)
			continue
		}

		// Record scenario positions
		if strings.HasPrefix(trimmed, "Scenario:") || strings.HasPrefix(trimmed, "Scenario Outline:") {
			scenarioTags := collectScenarioTags(lines, lineNum)
			name := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(trimmed, "Scenario Outline:"), "Scenario:"))
			scenarios = append(scenarios, scenarioInfo{
				lineNum: lineNum + 1, // 1-based
				name:    name,
				tags:    scenarioTags,
			})
		}
	}

	return featureLevelTags, scenarios
}

// collectScenarioTags walks backward from a scenario line to collect its tags.
func collectScenarioTags(lines []string, scenarioLineIdx int) []string {
	var tags []string
	checkIdx := scenarioLineIdx - 1

	for checkIdx >= 0 {
		trimmed := strings.TrimSpace(lines[checkIdx])
		if strings.HasPrefix(trimmed, "@") {
			tags = append(extractTagsFromLine(trimmed), tags...)
			checkIdx--
		} else if trimmed == "" {
			checkIdx--
		} else {
			break
		}
	}
	return tags
}

// checkTagConflicts detects L-level and verification tag conflicts between
// feature-level and scenario-level tags.
func checkTagConflicts(featureTags []string, scenarios []scenarioInfo, _ []string) []string {
	var errors []string

	featureLevelTag := findFirstTag(featureTags, isLevelTag)
	featureVerificationTag := findFirstTag(featureTags, isVerificationTag)

	for _, sc := range scenarios {
		scenarioLevelTag := findFirstTag(sc.tags, isLevelTag)
		scenarioVerificationTag := findFirstTag(sc.tags, isVerificationTag)

		if featureLevelTag != "" && scenarioLevelTag != "" && featureLevelTag != scenarioLevelTag {
			errors = append(errors, fmt.Sprintf(
				"Feature has L-tag %s and Scenario '%s' (line %d) has L-tag %s. "+
					"Scenarios inherit Feature tags, resulting in BOTH tags on the scenario. "+
					"Remove the L-tag from either the Feature or all Scenarios.",
				featureLevelTag, sc.name, sc.lineNum, scenarioLevelTag))
		}

		if featureVerificationTag != "" && scenarioVerificationTag != "" && featureVerificationTag != scenarioVerificationTag {
			errors = append(errors, fmt.Sprintf(
				"Feature has verification tag %s and Scenario '%s' (line %d) has verification tag %s. "+
					"Scenarios inherit Feature tags, resulting in BOTH tags on the scenario. "+
					"Remove the verification tag from either the Feature or all Scenarios.",
				featureVerificationTag, sc.name, sc.lineNum, scenarioVerificationTag))
		}
	}

	return errors
}

// findFirstTag returns the first tag matching the predicate, or empty string.
func findFirstTag(tags []string, predicate func(string) bool) string {
	for _, tag := range tags {
		if predicate(tag) {
			return tag
		}
	}
	return ""
}
