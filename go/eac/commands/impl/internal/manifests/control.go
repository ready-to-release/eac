package manifests

import (
	"sort"

	implinternal "github.com/ready-to-release/eac/go/eac/commands/impl/internal"
)

// ControlSummary represents a summary of tests tagged with a specific control ID
type ControlSummary struct {
	ControlID    string   `json:"control_id" yaml:"controlid"`
	TestCount    int      `json:"test_count" yaml:"testcount"`
	ModuleCount  int      `json:"module_count" yaml:"modulecount"`
	Modules      []string `json:"modules" yaml:"modules"`
	PassedCount  int      `json:"passed_count" yaml:"passedcount"`
	FailedCount  int      `json:"failed_count" yaml:"failedcount"`
	SkippedCount int      `json:"skipped_count" yaml:"skippedcount"`
}

// BuildControlSummary builds control summaries from all tests across manifests.
// Groups tests by control tag and aggregates counts across modules.
func BuildControlSummary(manifests []*implinternal.TestManifest) []ControlSummary {
	// Map: controlID -> summary data
	summaryMap := make(map[string]*ControlSummary)

	// Map: controlID -> set of modules
	modulesByControl := make(map[string]map[string]bool)

	for _, manifest := range manifests {
		for _, test := range manifest.Tests {
			// Extract control tags from this test
			controlTags := ExtractControlTags(test.Tags)

			// Process each control tag
			for _, controlID := range controlTags {
				// Get or create summary
				summary, exists := summaryMap[controlID]
				if !exists {
					summary = &ControlSummary{
						ControlID: controlID,
						Modules:   []string{},
					}
					summaryMap[controlID] = summary
					modulesByControl[controlID] = make(map[string]bool)
				}

				// Increment test count
				summary.TestCount++

				// Increment status counts
				switch test.Status {
				case implinternal.TestStatusPassed:
					summary.PassedCount++
				case implinternal.TestStatusFailed:
					summary.FailedCount++
				case implinternal.TestStatusSkipped:
					summary.SkippedCount++
				}

				// Track module
				modulesByControl[controlID][manifest.Moniker] = true
			}
		}
	}

	// Convert map to slice and populate module lists
	var result []ControlSummary
	for controlID, summary := range summaryMap {
		// Get unique modules for this control
		modules := make([]string, 0, len(modulesByControl[controlID]))
		for module := range modulesByControl[controlID] {
			modules = append(modules, module)
		}

		// Sort modules for consistent output
		sort.Strings(modules)

		summary.Modules = modules
		summary.ModuleCount = len(modules)

		result = append(result, *summary)
	}

	// Sort by control ID
	sort.Slice(result, func(i, j int) bool {
		return result[i].ControlID < result[j].ControlID
	})

	return result
}
