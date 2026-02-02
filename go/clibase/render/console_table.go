package render

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"
)

// ConsoleTableConfig holds configuration for console-aware table rendering.
type ConsoleTableConfig struct {
	// Headers for the table columns
	Headers []string
	// Rows of data (each row is a slice of values)
	Rows [][]interface{}
	// MaxWidth overrides auto-detected terminal width (0 = auto-detect)
	MaxWidth int
	// ColumnMaxWidths sets max width per column by index (0 = no limit)
	ColumnMaxWidths map[int]int
	// WrapLongText wraps text instead of truncating (default: true)
	WrapLongText bool
}

// DefaultTerminalWidth is used when terminal width cannot be detected.
const DefaultTerminalWidth = 120

// MinColumnWidth is the minimum width for any column.
const MinColumnWidth = 8

// GetTerminalWidth returns the current terminal width or a default.
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// RenderConsoleTable creates a console-friendly table with proper width handling.
func RenderConsoleTable(config *ConsoleTableConfig) string {
	tw := table.NewWriter()

	// Set up style for console output (not markdown)
	tw.SetStyle(table.StyleLight)

	// Detect terminal width
	maxWidth := config.MaxWidth
	if maxWidth <= 0 {
		maxWidth = GetTerminalWidth()
	}

	// Build header row
	var headerRow table.Row
	for _, h := range config.Headers {
		headerRow = append(headerRow, h)
	}
	tw.AppendHeader(headerRow)

	// Add data rows
	for _, row := range config.Rows {
		if row == nil {
			// nil row is a separator
			tw.AppendSeparator()
			continue
		}
		var dataRow table.Row
		for _, cell := range row {
			dataRow = append(dataRow, cell)
		}
		tw.AppendRow(dataRow)
	}

	// Calculate and set column widths
	numCols := len(config.Headers)
	colConfigs := make([]table.ColumnConfig, numCols)

	// Calculate available width (accounting for borders and padding)
	// Each column has: | <space> content <space> |
	// So overhead is: numCols * 3 (for "| ") + 1 (final |)
	borderOverhead := numCols*3 + 1
	availableWidth := maxWidth - borderOverhead

	// Distribute width among columns
	colWidths := distributeColumnWidths(config, availableWidth, numCols)

	for i := 0; i < numCols; i++ {
		colConfigs[i] = table.ColumnConfig{
			Number:   i + 1, // 1-indexed
			WidthMax: colWidths[i],
		}

		// Enable text wrapping
		if config.WrapLongText {
			colConfigs[i].WidthMaxEnforcer = text.WrapSoft
		} else {
			colConfigs[i].WidthMaxEnforcer = text.Trim
		}
	}

	tw.SetColumnConfigs(colConfigs)

	return tw.Render()
}

// distributeColumnWidths calculates optimal width for each column.
func distributeColumnWidths(config *ConsoleTableConfig, availableWidth, numCols int) []int {
	if numCols == 0 {
		return nil
	}

	// Calculate natural width for each column (max content width)
	naturalWidths := make([]int, numCols)

	// Check headers
	for i, h := range config.Headers {
		if len(h) > naturalWidths[i] {
			naturalWidths[i] = len(h)
		}
	}

	// Check all rows
	for _, row := range config.Rows {
		if row == nil {
			continue // skip separator rows
		}
		for i, cell := range row {
			if i < numCols {
				cellStr := cellToString(cell)
				if len(cellStr) > naturalWidths[i] {
					naturalWidths[i] = len(cellStr)
				}
			}
		}
	}

	// Apply explicit max widths from config
	for i, maxW := range config.ColumnMaxWidths {
		if i < numCols && maxW > 0 && naturalWidths[i] > maxW {
			naturalWidths[i] = maxW
		}
	}

	// Calculate total natural width
	totalNatural := 0
	for _, w := range naturalWidths {
		totalNatural += w
	}

	// If fits, use natural widths with minimum
	if totalNatural <= availableWidth {
		for i := range naturalWidths {
			if naturalWidths[i] < MinColumnWidth {
				naturalWidths[i] = MinColumnWidth
			}
		}
		return naturalWidths
	}

	// Need to shrink - distribute available width proportionally
	// but protect small columns and prioritize early columns
	result := make([]int, numCols)

	// First pass: give each column at least MinColumnWidth
	remaining := availableWidth
	for i := range result {
		result[i] = MinColumnWidth
		remaining -= MinColumnWidth
	}

	if remaining <= 0 {
		return result
	}

	// Second pass: distribute remaining space proportionally to natural widths
	// but cap large columns (like dependency lists)
	maxColWidth := availableWidth / 2 // No single column gets more than half

	for i := range result {
		natural := naturalWidths[i]
		if natural <= MinColumnWidth {
			continue
		}

		// Calculate fair share of remaining space
		extraNeeded := natural - MinColumnWidth
		if extraNeeded > maxColWidth-MinColumnWidth {
			extraNeeded = maxColWidth - MinColumnWidth
		}

		// Proportional allocation
		proportion := float64(natural) / float64(totalNatural)
		extra := int(float64(remaining) * proportion)

		if extra > extraNeeded {
			extra = extraNeeded
		}

		result[i] += extra
	}

	return result
}

// cellToString converts a cell value to string for width calculation.
func cellToString(cell interface{}) string {
	switch v := cell.(type) {
	case string:
		return v
	case int:
		return string(rune('0' + v%10)) // rough estimate
	default:
		return ""
	}
}

// ConsoleTableBuilder provides a fluent interface for building console tables.
type ConsoleTableBuilder struct {
	config ConsoleTableConfig
}

// NewConsoleTableBuilder creates a new console table builder.
func NewConsoleTableBuilder() *ConsoleTableBuilder {
	return &ConsoleTableBuilder{
		config: ConsoleTableConfig{
			Headers:      []string{},
			Rows:         [][]interface{}{},
			WrapLongText: true,
		},
	}
}

// WithHeaders sets the table headers.
func (tb *ConsoleTableBuilder) WithHeaders(headers ...string) *ConsoleTableBuilder {
	tb.config.Headers = headers
	return tb
}

// WithMaxWidth sets a fixed max width (0 = auto-detect terminal).
func (tb *ConsoleTableBuilder) WithMaxWidth(width int) *ConsoleTableBuilder {
	tb.config.MaxWidth = width
	return tb
}

// WithColumnMaxWidth sets max width for a specific column (0-indexed).
func (tb *ConsoleTableBuilder) WithColumnMaxWidth(col, maxWidth int) *ConsoleTableBuilder {
	if tb.config.ColumnMaxWidths == nil {
		tb.config.ColumnMaxWidths = make(map[int]int)
	}
	tb.config.ColumnMaxWidths[col] = maxWidth
	return tb
}

// WithTruncate disables text wrapping and truncates instead.
func (tb *ConsoleTableBuilder) WithTruncate() *ConsoleTableBuilder {
	tb.config.WrapLongText = false
	return tb
}

// AddRow adds a data row to the table.
func (tb *ConsoleTableBuilder) AddRow(cells ...interface{}) *ConsoleTableBuilder {
	tb.config.Rows = append(tb.config.Rows, cells)
	return tb
}

// AddSeparator adds a horizontal separator row to the table.
func (tb *ConsoleTableBuilder) AddSeparator() *ConsoleTableBuilder {
	tb.config.Rows = append(tb.config.Rows, nil) // nil row signals separator
	return tb
}

// Build renders the table for console output.
func (tb *ConsoleTableBuilder) Build() string {
	return RenderConsoleTable(&tb.config)
}

// FormatListForColumn formats a list of items for display in a table column.
// It shows the first few items and a count if there are more.
func FormatListForColumn(items []string, maxItems int) string {
	if len(items) == 0 {
		return "-"
	}

	if len(items) <= maxItems {
		return strings.Join(items, ", ")
	}

	shown := items[:maxItems]
	remaining := len(items) - maxItems
	return strings.Join(shown, ", ") + " +" + itoa(remaining) + " more"
}

// itoa is a simple int to string conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
