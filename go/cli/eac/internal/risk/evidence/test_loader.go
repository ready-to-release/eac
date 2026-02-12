// Package evidence provides test evidence loading for risk assessment.
package evidence

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// GetControlTestEvidenceFromManifest builds comprehensive control evidence from test entries.
// This function extracts @control: tags from test entries and builds evidence with test status.
//
// Suite Filtering:
//   - testSuite = "" → Include ALL tests (no filtering)
//   - testSuite = "unit" → Include only unit tests
//   - testSuite = "unit+integration" → Include tests from both unit and integration suites
func GetControlTestEvidenceFromManifest(tests []TestEntryData, testSuite string) map[string]*ControlTestEvidence {
	evidenceMap := make(map[string]*ControlTestEvidence)

	// Parse suite filter (supports composite suites like "unit+integration")
	allowedSuites := parseSuiteFilter(testSuite)

	// Build evidence from manifest test entries
	for _, test := range tests {
		// Filter by suite if specified
		if len(allowedSuites) > 0 && !slices.Contains(allowedSuites, test.Suite) {
			continue
		}

		// Extract control IDs from tags
		for _, tag := range test.Tags {
			controlIDs := extractControlIDsFromTag(tag)

			for _, controlID := range controlIDs {
				// Initialize control evidence if not exists
				if evidenceMap[controlID] == nil {
					evidenceMap[controlID] = &ControlTestEvidence{
						ControlID: controlID,
						Tests:     []TestEvidence{},
					}
				}

				// Create test evidence entry
				testEv := TestEvidence{
					FeaturePath:  test.FilePath,
					FeatureName:  test.Package,
					ScenarioName: test.Name,
					Status:       test.Status,
					Location:     fmt.Sprintf("%s:%s", test.Package, test.Name),
				}

				// Add to control evidence
				evidenceMap[controlID].Tests = append(evidenceMap[controlID].Tests, testEv)
			}
		}
	}

	// Calculate aggregate counts for each control
	for _, evidence := range evidenceMap {
		evidence.TotalTests = len(evidence.Tests)
		evidence.HasEvidence = evidence.TotalTests > 0

		for _, test := range evidence.Tests {
			switch test.Status {
			case "passed":
				evidence.PassedTests++
			case "failed":
				evidence.FailedTests++
			case "skipped":
				evidence.SkippedTests++
			}
		}
	}

	return evidenceMap
}

// TestEntryData represents a single test entry with metadata.
// See TestManifestData in types_testing.go for import-cycle workaround explanation.
type TestEntryData struct {
	Name     string
	Package  string
	Type     string // gotest, godog, mocha, tscucumber
	Suite    string
	Status   string
	Tags     []string
	FilePath string
}

// extractControlIDsFromTag extracts control IDs from a single tag string.
// Supports both @control:ac-2 and @controls:ac-2,au-3 formats.
// Returns a slice of control IDs found in the tag.
func extractControlIDsFromTag(tag string) []string {
	var controlIDs []string

	// Pattern for single control: @control:ac-2 or @control:ac-2(1)
	controlTagPattern := regexp.MustCompile(`@control:([a-z]{2,4}-\d+(?:\(\d+\))?)`)

	// Pattern for multiple controls: @controls:ac-2,au-3 or @controls:ac-2(1),au-3(2)
	controlsTagPattern := regexp.MustCompile(`@controls:((?:[a-z]{2,4}-\d+(?:\(\d+\))?,)*[a-z]{2,4}-\d+(?:\(\d+\))?)`)

	// Check @control:<id>
	if matches := controlTagPattern.FindStringSubmatch(tag); len(matches) > 1 {
		controlIDs = append(controlIDs, matches[1])
	}

	// Check @controls:<id1>,<id2>
	if matches := controlsTagPattern.FindStringSubmatch(tag); len(matches) > 1 {
		ids := strings.Split(matches[1], ",")
		for _, id := range ids {
			controlIDs = append(controlIDs, strings.TrimSpace(id))
		}
	}

	return controlIDs
}

// parseSuiteFilter parses a suite filter string into a list of allowed suites.
// Returns nil for empty string (no filtering), or a slice of suite names.
// Supports composite suites: "unit+integration" → ["unit", "integration"].
func parseSuiteFilter(suiteFilter string) []string {
	if suiteFilter == "" {
		return nil // No filter = include all tests
	}
	// Split by "+" to support composite suites
	return strings.Split(suiteFilter, "+")
}

