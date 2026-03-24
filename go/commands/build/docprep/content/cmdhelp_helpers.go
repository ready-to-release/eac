package content

import (
	"github.com/ready-to-release/eac/go/clibase/render"
)

// FormatTable generates a markdown table from headers and row data.
// Each row is a slice of cell strings.
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	tb := render.NewTableBuilder().WithMarkdown().WithHeaders(headers...)
	for _, row := range rows {
		cells := make([]interface{}, len(row))
		for i, cell := range row {
			cells[i] = cell
		}
		tb.AddRow(cells...)
	}
	return tb.BuildMarkdown() + "\n"
}
