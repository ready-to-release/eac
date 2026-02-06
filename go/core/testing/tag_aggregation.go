package testing

import (
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// AggregateTagsForTests computes the union TagSummary for a set of tests.
// Each tag is classified into the appropriate category. The result is
// deterministic: all slices are sorted and deduplicated.
func AggregateTagsForTests(tests []TestReference) workunit.TagSummary {
	if len(tests) == 0 {
		return workunit.TagSummary{}
	}

	allSet := make(map[string]bool)
	levelSet := make(map[string]bool)
	verifSet := make(map[string]bool)
	controlSet := make(map[string]bool)
	sysDeps := make(map[string]bool)
	modDeps := make(map[string]bool)
	isGxP := false
	hasManual := false

	for i := range tests {
		ref := &tests[i]
		for _, tag := range ref.Tags {
			allSet[tag] = true

			switch {
			case isLevelTag(tag):
				levelSet[tag] = true
			case isVerificationTag(tag):
				verifSet[tag] = true
			case strings.HasPrefix(tag, "@risk-control:"):
				controlSet[strings.TrimPrefix(tag, "@risk-control:")] = true
			case strings.HasPrefix(tag, "@deps:"):
				sysDeps[strings.TrimPrefix(tag, "@deps:")] = true
			case strings.HasPrefix(tag, "@depm:"):
				modDeps[strings.TrimPrefix(tag, "@depm:")] = true
			}

			if tag == "@gxp" {
				isGxP = true
			}
			if tag == "@Manual" {
				hasManual = true
			}
		}

		// Also include structured fields (already extracted by discovery)
		for _, dep := range ref.SystemDependencies {
			sysDeps[dep] = true
		}
		for _, dep := range ref.ModuleDependencies {
			modDeps[dep] = true
		}
		for _, ctrl := range ref.RiskControls {
			controlSet[ctrl] = true
		}
		if ref.IsGxP {
			isGxP = true
		}
		if ref.IsManual {
			hasManual = true
		}
	}

	return workunit.TagSummary{
		All:          sortedKeys(allSet),
		Levels:       sortedKeys(levelSet),
		Verification: sortedKeys(verifSet),
		RiskControls: sortedKeys(controlSet),
		SystemDeps:   sortedKeys(sysDeps),
		ModuleDeps:   sortedKeys(modDeps),
		IsGxP:        isGxP,
		HasManual:    hasManual,
	}
}

// sortedKeys returns the keys of a map as a sorted slice, or nil if empty.
func sortedKeys(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
