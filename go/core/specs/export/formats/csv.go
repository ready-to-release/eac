// Package formats provides export formatters for manual test scenarios (CSV, JSON, Markdown).
package formats

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CSVFormatter exports scenarios as CSV.
type CSVFormatter struct{}

// NewCSVFormatter creates a new CSV formatter.
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

func (f *CSVFormatter) Write(export *ManualTestExport, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header (using snake_case to match schema)
	header := []string{
		"scenario_id",
		"feature_name",
		"scenario_name",
		"tags",
		"steps",
		"description",
		"file_path",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Write rows
	for _, scenario := range export.Scenarios {
		row := []string{
			scenario.ScenarioID,
			scenario.FeatureName,
			scenario.ScenarioName,
			strings.Join(scenario.Tags, " "),
			strings.Join(scenario.Steps, " | "),
			scenario.Description,
			scenario.FilePath,
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}

	return nil
}

func (f *CSVFormatter) FileExtension() string {
	return "csv"
}
