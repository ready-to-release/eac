package formats

import (
	"fmt"
	"io"
	"strings"
)

// MarkdownFormatter exports scenarios as simplified Markdown.
// The output is intentionally minimal since documentation auto-generation
// (via <!-- book:cmd test export-manual -->) handles full documentation.
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new Markdown formatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

func (f *MarkdownFormatter) Write(export *ManualTestExport, w io.Writer) error {
	// Write minimal header
	fmt.Fprintf(w, "# Manual Test Scenarios\n\n")
	fmt.Fprintf(w, "Module: %s | Release: %s | Exported: %s\n\n",
		export.ExportMetadata.Module,
		export.ExportMetadata.ReleaseVersion,
		export.ExportMetadata.ExportTime)

	// Write scenarios in concise format
	for i, scenario := range export.Scenarios {
		fmt.Fprintf(w, "## %d. %s\n\n", i+1, scenario.ScenarioName)
		fmt.Fprintf(w, "- **ID**: `%s`\n", scenario.ScenarioID)
		fmt.Fprintf(w, "- **Feature**: %s\n", scenario.FeatureName)
		if len(scenario.Tags) > 0 {
			fmt.Fprintf(w, "- **Tags**: %s\n", strings.Join(scenario.Tags, " "))
		}
		if scenario.Description != "" {
			fmt.Fprintf(w, "- **Description**: %s\n", scenario.Description)
		}
		fmt.Fprintf(w, "\n**Steps**:\n")
		for _, step := range scenario.Steps {
			fmt.Fprintf(w, "1. %s\n", step)
		}
		fmt.Fprintf(w, "\n")
	}

	return nil
}

func (f *MarkdownFormatter) FileExtension() string {
	return "md"
}
