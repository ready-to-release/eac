package list

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatTable formats the linkage report as a table
func formatTable(report *LinkageReport, filter string) string {
	var output strings.Builder

	// Handle missing-links filter separately
	if filter == "missing-links" {
		output.WriteString("Specs Missing Risk Control Links\n")
		output.WriteString("═══════════════════════════════════════════════════════════\n\n")
		if len(report.MissingLinks) > 0 {
			for _, missing := range report.MissingLinks {
				output.WriteString(fmt.Sprintf("Spec: %s\n", missing.SpecPath))
				if missing.SuggestedControl != "" {
					output.WriteString(fmt.Sprintf("  Suggested control: %s\n", missing.SuggestedControl))
				}
				output.WriteString("\n")
			}
		} else {
			output.WriteString("No specs missing risk control links.\n")
		}
		return output.String()
	}

	// Standard control listing
	output.WriteString("Risk Controls Linkage Report\n")
	output.WriteString("═══════════════════════════════════════════════════════════\n\n")

	controlsShown := 0
	for _, control := range report.Controls {
		// Apply filter
		if filter == "orphaned" && control.Status != "orphaned" {
			continue
		}
		if filter == "linked" && control.Status != "linked" {
			continue
		}

		controlsShown++

		// Control header
		output.WriteString(fmt.Sprintf("Risk Control: %s\n", control.Path))

		if len(control.Tags) > 0 {
			output.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(control.Tags, ", ")))
		} else {
			output.WriteString("  Tags: (none)\n")
		}

		// References
		if len(control.ReferencedBy) > 0 {
			output.WriteString("  Referenced by:\n")
			for _, ref := range control.ReferencedBy {
				output.WriteString(fmt.Sprintf("    - %s (line %d)\n", ref.SpecPath, ref.Line))
			}
		} else {
			output.WriteString("  Referenced by: (none)\n")
		}

		// Status
		if control.Status == "linked" {
			output.WriteString("  Status: ✓ Linked\n")
		} else {
			output.WriteString("  Status: ⚠ Orphaned\n")
		}
		output.WriteString("\n")
	}

	// Summary
	output.WriteString("───────────────────────────────────────────────────────────\n")
	output.WriteString("Summary:\n")

	if filter == "all" {
		output.WriteString(fmt.Sprintf("  Total controls: %d\n", report.Summary.Total))
		output.WriteString(fmt.Sprintf("  Linked: %d\n", report.Summary.Linked))
		output.WriteString(fmt.Sprintf("  Orphaned: %d\n", report.Summary.Orphaned))
		if len(report.MissingLinks) > 0 {
			output.WriteString(fmt.Sprintf("  Specs missing links: %d\n", len(report.MissingLinks)))
		}
	} else if filter == "orphaned" {
		output.WriteString(fmt.Sprintf("  Orphaned controls shown: %d of %d total\n", controlsShown, report.Summary.Total))
	} else if filter == "linked" {
		output.WriteString(fmt.Sprintf("  Linked controls shown: %d of %d total\n", controlsShown, report.Summary.Total))
	}

	return output.String()
}

// formatJSON formats the linkage report as JSON
func formatJSON(report *LinkageReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
