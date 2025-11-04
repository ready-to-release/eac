package render

import (
	"strings"
)

// FormatMarkdownTable takes raw markdown table output and reformats it with proper spacing
func FormatMarkdownTable(rawMarkdown string) string {
	lines := strings.Split(rawMarkdown, "\n")
	if len(lines) < 2 {
		return rawMarkdown
	}

	// Parse all rows to find column widths
	var rows [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse cells by splitting on |
		cells := strings.Split(line, "|")

		// Remove empty first and last elements from split
		if len(cells) > 0 && cells[0] == "" {
			cells = cells[1:]
		}
		if len(cells) > 0 && cells[len(cells)-1] == "" {
			cells = cells[:len(cells)-1]
		}

		// Trim spaces from each cell
		var trimmedCells []string
		for _, cell := range cells {
			trimmedCells = append(trimmedCells, strings.TrimSpace(cell))
		}

		rows = append(rows, trimmedCells)
	}

	if len(rows) == 0 {
		return rawMarkdown
	}

	// Calculate maximum width for each column
	numCols := len(rows[0])
	colWidths := make([]int, numCols)

	for _, row := range rows {
		for i, cell := range row {
			if i < numCols {
				// For separator rows, use the actual dash count
				if isSeparatorRow(row) {
					// Remove alignment markers to get base dash count
					cleanCell := strings.ReplaceAll(cell, ":", "")
					cleanCell = strings.TrimSpace(cleanCell)
					if len(cleanCell) > colWidths[i] {
						colWidths[i] = len(cleanCell)
					}
				} else {
					if len(cell) > colWidths[i] {
						colWidths[i] = len(cell)
					}
				}
			}
		}
	}

	// Rebuild the table with proper spacing
	var result strings.Builder
	for rowIdx, row := range rows {
		result.WriteString("|")

		for colIdx, cell := range row {
			if colIdx < numCols {
				if isSeparatorRow(row) {
					// Handle separator row with alignment
					formatted := formatSeparatorCell(cell, colWidths[colIdx])
					result.WriteString(" ")
					result.WriteString(formatted)
					result.WriteString(" |")
				} else {
					// Regular data row
					padded := padCell(cell, colWidths[colIdx])
					result.WriteString(" ")
					result.WriteString(padded)
					result.WriteString(" |")
				}
			}
		}

		// Add newline except for last line
		if rowIdx < len(rows)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// isSeparatorRow checks if a row is a separator row (contains dashes)
func isSeparatorRow(row []string) bool {
	if len(row) == 0 {
		return false
	}

	for _, cell := range row {
		cleaned := strings.ReplaceAll(cell, ":", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")
		cleaned = strings.TrimSpace(cleaned)

		// If there's any non-dash, non-colon content, it's not a separator
		if cleaned != "" {
			return false
		}
	}

	return true
}

// formatSeparatorCell formats a separator cell with proper dashes and alignment markers
func formatSeparatorCell(cell string, width int) string {
	cell = strings.TrimSpace(cell)

	// Check for alignment markers
	leftAlign := strings.HasPrefix(cell, ":")
	rightAlign := strings.HasSuffix(cell, ":")

	// Remove existing alignment markers and dashes
	cleaned := strings.ReplaceAll(cell, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	// Build the separator with the correct width
	// For right-aligned, the colon takes one character, so reduce dashes by 1
	// For left-aligned or center, same logic applies
	dashCount := width
	if leftAlign {
		dashCount-- // Colon on left takes a character
	}
	if rightAlign {
		dashCount-- // Colon on right takes a character
	}

	// Ensure at least 1 dash
	if dashCount < 1 {
		dashCount = 1
	}

	dashes := strings.Repeat("-", dashCount)

	// Add back alignment markers
	if leftAlign && rightAlign {
		return ":" + dashes + ":"
	} else if rightAlign {
		return dashes + ":"
	} else if leftAlign {
		return ":" + dashes
	}

	return dashes
}

// padCell pads a cell to the specified width
func padCell(cell string, width int) string {
	cell = strings.TrimSpace(cell)
	if len(cell) >= width {
		return cell
	}

	padding := width - len(cell)
	return cell + strings.Repeat(" ", padding)
}
