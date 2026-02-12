package initsummary

import (
	"fmt"
	"strings"
)

// formatModuleListInline formats a list of modules for inline display.
func formatModuleListInline(modules []string, maxLen int) string {
	if len(modules) == 0 {
		return ""
	}
	list := truncateList(modules, maxLen)
	return " (" + list + ")"
}

// truncateList joins items with ", " and truncates with "..." if too long.
func truncateList(items []string, maxLen int) string {
	if len(items) == 0 {
		return ""
	}

	result := strings.Join(items, ", ")
	if len(result) <= maxLen {
		return result
	}

	// Truncate and add "..."
	for i := len(items) - 1; i >= 1; i-- {
		result = strings.Join(items[:i], ", ") + ", ..."
		if len(result) <= maxLen {
			return result
		}
	}
	return items[0][:min(len(items[0]), maxLen-3)] + "..."
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// countTotalMissingArtifacts counts total number of missing artifacts across all modules.
func countTotalMissingArtifacts(details map[string][]string) int {
	total := 0
	for _, artifacts := range details {
		total += len(artifacts)
	}
	return total
}

// countAvailableDeps counts the number of available dependencies from results.
func countAvailableDeps(results []DepsResult) int {
	count := 0
	for _, r := range results {
		if r.Available {
			count++
		}
	}
	return count
}

// depmSummary returns a shared depm status string for both formatters.
// Returns ("", false) if nothing should be shown.
func depmSummary(s DepmStatus) (string, bool) {
	if s.Skipped {
		return "skipped (--skip-depm)", true
	}
	if s.Verified && len(s.Missing) > 0 {
		return fmt.Sprintf("%d/%d resolved (%d missing)",
			len(s.Resolved), s.Total, len(s.Missing)), true
	}
	if s.Verified && s.Total > 0 {
		return fmt.Sprintf("%d/%d resolved", len(s.Resolved), s.Total), true
	}
	return "", false
}

// depsSummary returns a shared deps status string for both formatters.
// Returns ("", false) if nothing should be shown.
func depsSummary(s DepsStatus) (string, bool) {
	if s.Skipped {
		return "skipped (--skip-deps)", true
	}
	if !s.Verified || len(s.Required) == 0 {
		return "", false
	}
	available := countAvailableDeps(s.Available)
	required := len(s.Required)
	missing := len(s.Missing)

	if available+missing != required {
		panic(fmt.Sprintf("DepsStatus invariant violated: available(%d) + missing(%d) != required(%d)",
			available, missing, required))
	}

	if missing > 0 {
		return fmt.Sprintf("%d/%d available (%s missing)",
			available, required, strings.Join(s.Missing, ", ")), true
	}
	return fmt.Sprintf("%d/%d available (%s)",
		available, required, strings.Join(s.Required, ", ")), true
}
