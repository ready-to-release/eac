// Package output provides unified formatting utilities for command output.
// This ensures consistent "look and feel" across test, build, and other commands.
package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/core/platform"
)

// Column widths for aligned output
const (
	NameWidth   = 46 // Module/package name
	TypeWidth   = 15 // Type (go, godog, scripts-package - truncated if longer)
	ResultWidth = 6  // Result (12/12, etc.)
	TimeWidth   = 6  // Duration (0.8s, 15.2s, etc.) - only used in timing summary
)

// Status icons
const (
	IconPass    = "✅"
	IconFail    = "❌"
	IconWarn    = "⚠️"
	IconSkip    = "⏭️"
	IconRunning = "🔄"
	IconPending = "⏳"
)

// PhaseHeader formats a phase header line.
// Format: "=== Phase N: Name ==="
func PhaseHeader(phase int, name string) string {
	return fmt.Sprintf("=== Phase %d: %s ===", phase, name)
}

// SectionHeader formats a section header without phase number.
// Format: "=== Name ==="
func SectionHeader(name string) string {
	return fmt.Sprintf("=== %s ===", name)
}

// DependencyLine formats a dependency check result.
// Format: "  ✅ go (1.21.0)" or "  ❌ docker - not available"
func DependencyLine(available bool, name, version string) string {
	if available {
		if version != "" {
			return fmt.Sprintf("  %s %s (%s)", IconPass, name, version)
		}
		return fmt.Sprintf("  %s %s", IconPass, name)
	}
	return fmt.Sprintf("  %s %s - not available", IconFail, name)
}

// OutputDir formats the output directory announcement.
// Format: "Output: out/build/"
func OutputDir(path string) string {
	return fmt.Sprintf("Output: %s", path)
}

// Writeln writes a formatted line with platform-specific line ending to the writer.
func Writeln(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+platform.LineEnding, args...)
}

// FormatLine formats a line with platform-specific line ending.
func FormatLine(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...) + platform.LineEnding
}

// FormatDurationShort formats duration as seconds with one decimal.
func FormatDurationShort(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

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
