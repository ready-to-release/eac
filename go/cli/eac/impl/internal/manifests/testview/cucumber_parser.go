package testview

import (
	"encoding/json"
	"os"
	"strings"
)

// cucumberFeature is a minimal struct for parsing cucumber.json files.
type cucumberFeature struct {
	URI      string            `json:"uri"`
	Elements []cucumberElement `json:"elements"`
}

type cucumberElement struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	Tags  []cucumberTag  `json:"tags"`
	Steps []cucumberStep `json:"steps"`
}

type cucumberTag struct {
	Name string `json:"name"`
}

type cucumberStep struct {
	Result cucumberResult `json:"result"`
}

type cucumberResult struct {
	Status   string `json:"status"`
	Duration int64  `json:"duration"` // nanoseconds
}

// parseCucumberFile parses a cucumber.json file and returns test entries.
func parseCucumberFile(path, module string) []TestEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var features []cucumberFeature
	if err := json.Unmarshal(data, &features); err != nil {
		return nil
	}

	var entries []TestEntry
	for _, feature := range features {
		for _, element := range feature.Elements {
			if element.Type != "scenario" {
				continue
			}

			status := deriveScenarioStatus(element.Steps)
			durationNs := sumStepDurations(element.Steps)

			var tags []string
			for _, tag := range element.Tags {
				tags = append(tags, tag.Name)
			}

			// Determine suite from L-level tags
			suite := deriveSuiteFromTags(tags)

			entries = append(entries, TestEntry{
				Name:       element.Name,
				Module:     module,
				Package:    feature.URI,
				Type:       "godog",
				Suite:      suite,
				Status:     status,
				DurationMs: durationNs / 1_000_000,
				Tags:       tags,
				FilePath:   feature.URI,
			})
		}
	}

	return entries
}

// deriveScenarioStatus determines scenario status from step results.
func deriveScenarioStatus(steps []cucumberStep) string {
	for _, step := range steps {
		switch step.Result.Status {
		case "failed":
			return StatusFailed
		case "undefined":
			return StatusFailed
		case "pending":
			return StatusSkipped
		case "skipped":
			// A skipped step after a failure doesn't change status
			continue
		}
	}
	return StatusPassed
}

// sumStepDurations sums step durations in nanoseconds.
func sumStepDurations(steps []cucumberStep) int64 {
	var total int64
	for _, step := range steps {
		total += step.Result.Duration
	}
	return total
}

// deriveSuiteFromTags extracts suite name from L-level tags.
func deriveSuiteFromTags(tags []string) string {
	for _, tag := range tags {
		tag = strings.TrimPrefix(tag, "@")
		switch {
		case tag == "L0":
			return "unit"
		case tag == "L1":
			return "integration"
		case tag == "L2" || tag == "HE2E":
			return "e2e"
		}
	}
	return "unit" // default
}
