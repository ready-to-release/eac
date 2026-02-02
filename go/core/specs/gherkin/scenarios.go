package gherkin

import (
	"fmt"
	"strings"
)

// IsManualScenario checks if a scenario has the @Manual tag.
func IsManualScenario(scenario ScenarioDetail) bool {
	return HasTag(scenario.Tags, "@Manual")
}

// HasTag checks if a tag exists in a list of tags.
func HasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// FilterScenariosByTag filters scenarios to only those with the specified tag.
func FilterScenariosByTag(scenarios []ScenarioDetail, tag string) []ScenarioDetail {
	var filtered []ScenarioDetail
	for _, s := range scenarios {
		if HasTag(s.Tags, tag) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GenerateScenarioID generates a unique ID for a scenario based on its name and file.
// Format: feature-name:scenario-name (slugified).
func GenerateScenarioID(scenario ScenarioDetail) string {
	featureSlug := slugify(scenario.FeatureName)
	scenarioSlug := slugify(scenario.Name)
	return fmt.Sprintf("%s:%s", featureSlug, scenarioSlug)
}

// DetectIDCollisions checks for duplicate scenario IDs.
// Returns a map of ID -> list of scenarios with that ID.
func DetectIDCollisions(scenarios []ScenarioDetail) map[string][]ScenarioDetail {
	idMap := make(map[string][]ScenarioDetail)

	for _, s := range scenarios {
		id := GenerateScenarioID(s)
		idMap[id] = append(idMap[id], s)
	}

	// Filter to only collisions
	collisions := make(map[string][]ScenarioDetail)
	for id, scenarios := range idMap {
		if len(scenarios) > 1 {
			collisions[id] = scenarios
		}
	}

	return collisions
}

// slugify converts a string to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	// Remove consecutive hyphens
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}
