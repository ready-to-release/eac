package output

import "strings"

// ListInline formats a list as a single line with truncation.
// Example: " (core, eac, ...)" or " (core)"
// Returns empty string if items is empty.
func ListInline(items []string, maxLen int) string {
	if len(items) == 0 {
		return ""
	}
	if maxLen <= 0 {
		maxLen = 50
	}

	result := " ("
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		// Check if adding this item would exceed max length
		if len(result)+len(item)+4 > maxLen { // 4 for ", ...)"
			result += "..."
			break
		}
		result += item
	}
	result += ")"
	return result
}

// ListFormat formats a list of items for display.
// If the total length is short (<=maxInlineLen), returns inline format: "item1, item2, item3"
// Otherwise returns multi-line format with itemsPerLine items per line:
//
//	item1, item2, item3, item4,
//	item5, item6, item7, item8
//
// Parameters:
//   - items: list of items to format
//   - maxInlineLen: maximum length for inline format (default 60 if 0)
//   - itemsPerLine: items per line in multi-line format (default 4 if 0)
func ListFormat(items []string, maxInlineLen, itemsPerLine int) string {
	if len(items) == 0 {
		return ""
	}

	// Apply defaults
	if maxInlineLen <= 0 {
		maxInlineLen = 60
	}
	if itemsPerLine <= 0 {
		itemsPerLine = 4
	}

	// Try inline format first
	inline := strings.Join(items, ", ")
	if len(inline) <= maxInlineLen {
		return inline
	}

	// Multi-line format
	var lines []string
	for i := 0; i < len(items); i += itemsPerLine {
		end := i + itemsPerLine
		if end > len(items) {
			end = len(items)
		}
		chunk := items[i:end]
		line := "  " + strings.Join(chunk, ", ")
		// Add trailing comma if not last line
		if end < len(items) {
			line += ","
		}
		lines = append(lines, line)
	}

	return "\n" + strings.Join(lines, "\n")
}

// ListFormatWithPrefix formats a list with a prefix label.
// If inline: "Prefix: item1, item2, item3"
// If multi-line:
//
//	Prefix:
//	  item1, item2, item3, item4,
//	  item5, item6, item7, item8
func ListFormatWithPrefix(prefix string, items []string, maxInlineLen, itemsPerLine int) string {
	if len(items) == 0 {
		return prefix + ": (none)"
	}

	// Apply defaults
	if maxInlineLen <= 0 {
		maxInlineLen = 60
	}
	if itemsPerLine <= 0 {
		itemsPerLine = 4
	}

	// Try inline format first
	inline := strings.Join(items, ", ")
	if len(prefix)+2+len(inline) <= maxInlineLen {
		return prefix + ": " + inline
	}

	// Multi-line format
	formatted := ListFormat(items, maxInlineLen, itemsPerLine)
	return prefix + ":" + formatted
}
