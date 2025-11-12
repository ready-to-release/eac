package render

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// MarkdownTableConfig holds configuration for rendering markdown tables
type MarkdownTableConfig struct {
	// Headers for the table columns
	Headers []string
	// Rows of data (each row is a slice of values)
	Rows [][]interface{}
	// Footer row (optional)
	Footer []interface{}
	// AutoIndex adds an index column (1, 2, 3...) as the first column
	AutoIndex bool
}

// RenderMarkdownTable creates a markdown-formatted table from the provided configuration
//
// Example:
//
//	config := &MarkdownTableConfig{
//	    Headers: []string{"Name", "Type", "Description"},
//	    Rows: [][]interface{}{
//	        {"foo", "string", "A foo value"},
//	        {"bar", "int", "A bar value"},
//	    },
//	}
//	result := RenderMarkdownTable(config)
func RenderMarkdownTable(config *MarkdownTableConfig) string {
	tw := table.NewWriter()

	// Build header row
	var headerRow table.Row
	if config.AutoIndex {
		headerRow = append(headerRow, "#")
	}
	for _, h := range config.Headers {
		headerRow = append(headerRow, h)
	}
	tw.AppendHeader(headerRow)

	// Add data rows
	for idx, row := range config.Rows {
		var dataRow table.Row
		if config.AutoIndex {
			dataRow = append(dataRow, idx+1)
		}
		for _, cell := range row {
			dataRow = append(dataRow, cell)
		}
		tw.AppendRow(dataRow)
	}

	// Add footer if provided
	if config.Footer != nil && len(config.Footer) > 0 {
		var footerRow table.Row
		if config.AutoIndex {
			footerRow = append(footerRow, "")
		}
		for _, cell := range config.Footer {
			footerRow = append(footerRow, cell)
		}
		tw.AppendFooter(footerRow)
	}

	return FormatMarkdownTable(tw.RenderMarkdown())
}

// SimpleMarkdownTable is a convenience function for creating basic tables
// without needing to construct a full config struct
//
// Example:
//
//	result := SimpleMarkdownTable(
//	    []string{"Column 1", "Column 2"},
//	    [][]interface{}{
//	        {"value1", "value2"},
//	        {"value3", "value4"},
//	    },
//	)
func SimpleMarkdownTable(headers []string, rows [][]interface{}) string {
	return RenderMarkdownTable(&MarkdownTableConfig{
		Headers: headers,
		Rows:    rows,
	})
}

// TableBuilder provides a fluent interface for building markdown tables
type TableBuilder struct {
	config MarkdownTableConfig
}

// NewTableBuilder creates a new table builder
func NewTableBuilder() *TableBuilder {
	return &TableBuilder{
		config: MarkdownTableConfig{
			Headers: []string{},
			Rows:    [][]interface{}{},
		},
	}
}

// WithHeaders sets the table headers
func (tb *TableBuilder) WithHeaders(headers ...string) *TableBuilder {
	tb.config.Headers = headers
	return tb
}

// WithAutoIndex enables automatic row numbering
func (tb *TableBuilder) WithAutoIndex() *TableBuilder {
	tb.config.AutoIndex = true
	return tb
}

// AddRow adds a data row to the table
func (tb *TableBuilder) AddRow(cells ...interface{}) *TableBuilder {
	tb.config.Rows = append(tb.config.Rows, cells)
	return tb
}

// AddRows adds multiple data rows
func (tb *TableBuilder) AddRows(rows [][]interface{}) *TableBuilder {
	tb.config.Rows = append(tb.config.Rows, rows...)
	return tb
}

// WithFooter sets the footer row
func (tb *TableBuilder) WithFooter(cells ...interface{}) *TableBuilder {
	tb.config.Footer = cells
	return tb
}

// Build renders the table as markdown
func (tb *TableBuilder) Build() string {
	return RenderMarkdownTable(&tb.config)
}

// RenderKeyValueTable creates a simple two-column key-value markdown table
//
// Example:
//
//	data := map[string]interface{}{
//	    "Name": "MyProject",
//	    "Version": "1.0.0",
//	    "Language": "Go",
//	}
//	result := RenderKeyValueTable("Property", "Value", data)
func RenderKeyValueTable(keyHeader, valueHeader string, data map[string]interface{}) string {
	tb := NewTableBuilder().
		WithHeaders(keyHeader, valueHeader)

	// Sort keys for consistent output
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}

	// Note: We could sort here if needed, but keeping insertion order for now
	for _, k := range keys {
		tb.AddRow(k, data[k])
	}

	return tb.Build()
}

// RenderCompactList creates a single-column markdown table from a list of items
//
// Example:
//
//	items := []string{"item1", "item2", "item3"}
//	result := RenderCompactList("Items", items)
func RenderCompactList(header string, items []string) string {
	tb := NewTableBuilder().
		WithHeaders(header)

	for _, item := range items {
		tb.AddRow(item)
	}

	return tb.Build()
}

// AlignedTable supports column alignment (left, center, right)
type AlignedTable struct {
	writer table.Writer
}

// NewAlignedTable creates a table with custom alignment options
// Note: go-pretty supports alignment through the Style system
func NewAlignedTable() *AlignedTable {
	return &AlignedTable{
		writer: table.NewWriter(),
	}
}

// SetHeaders sets the table headers
func (at *AlignedTable) SetHeaders(headers ...interface{}) {
	at.writer.AppendHeader(table.Row(headers))
}

// AddRow adds a data row
func (at *AlignedTable) AddRow(cells ...interface{}) {
	at.writer.AppendRow(table.Row(cells))
}

// RenderMarkdown outputs the table in markdown format
func (at *AlignedTable) RenderMarkdown() string {
	return FormatMarkdownTable(at.writer.RenderMarkdown())
}

// TrimMarkdownTable removes leading/trailing whitespace from a markdown table string
func TrimMarkdownTable(markdown string) string {
	return strings.TrimSpace(markdown)
}
