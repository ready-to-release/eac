package workunit

import (
	"fmt"
	"sort"
	"strings"
)

// TagSummary carries classified tag data for a work unit.
// Value type — always present on test UoWs (zero value is valid empty).
type TagSummary struct {
	All          []string `json:"all"`
	Levels       []string `json:"levels"`
	Verification []string `json:"verification"`
	RiskControls []string `json:"risk_controls"`
	SystemDeps   []string `json:"system_deps"`
	ModuleDeps   []string `json:"module_deps"`
	IsGxP        bool     `json:"is_gxp"`
	HasManual    bool     `json:"has_manual"`
}

// IsEmpty returns true if no tags are present.
func (ts TagSummary) IsEmpty() bool { return len(ts.All) == 0 }

// Summary returns a compact human-readable summary of the tag data.
// Example: "L0,L1 ov gxp 2 controls"
func (ts TagSummary) Summary() string {
	if ts.IsEmpty() {
		return ""
	}

	var parts []string

	if len(ts.Levels) > 0 {
		parts = append(parts, trimPrefixAndJoin(ts.Levels, "@", ","))
	}

	if len(ts.Verification) > 0 {
		parts = append(parts, trimPrefixAndJoin(ts.Verification, "@", ","))
	}

	if ts.IsGxP {
		parts = append(parts, "gxp")
	}

	if ts.HasManual {
		parts = append(parts, "manual")
	}

	if n := len(ts.RiskControls); n == 1 {
		parts = append(parts, "1 control")
	} else if n > 0 {
		parts = append(parts, fmt.Sprintf("%d controls", n))
	}

	return strings.Join(parts, " ")
}

// Merge returns a new TagSummary that is the union of ts and other.
// All slices are deduplicated and sorted. Boolean flags are OR'd.
func (ts TagSummary) Merge(other TagSummary) TagSummary {
	return TagSummary{
		All:          mergeAndSort(ts.All, other.All),
		Levels:       mergeAndSort(ts.Levels, other.Levels),
		Verification: mergeAndSort(ts.Verification, other.Verification),
		RiskControls: mergeAndSort(ts.RiskControls, other.RiskControls),
		SystemDeps:   mergeAndSort(ts.SystemDeps, other.SystemDeps),
		ModuleDeps:   mergeAndSort(ts.ModuleDeps, other.ModuleDeps),
		IsGxP:        ts.IsGxP || other.IsGxP,
		HasManual:    ts.HasManual || other.HasManual,
	}
}

// mergeAndSort returns the sorted, deduplicated union of two slices.
func mergeAndSort(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		seen[s] = true
	}
	if len(seen) == 0 {
		return nil
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

// trimPrefixAndJoin strips a prefix from each element and joins with sep.
func trimPrefixAndJoin(tags []string, prefix, sep string) string {
	stripped := make([]string, len(tags))
	for i, t := range tags {
		stripped[i] = strings.TrimPrefix(t, prefix)
	}
	return strings.Join(stripped, sep)
}
