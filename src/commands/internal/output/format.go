// Package output provides unified formatting utilities for command output.
// This ensures consistent "look and feel" across test, build, and other commands.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/ready-to-release/eac/src/core/platform"
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

// ResultLine formats a completion line with aligned columns.
// Format: "✅ name                          type     result   time"
//
// Parameters:
//   - icon: Status icon (IconPass, IconFail, etc.)
//   - name: Module or package name (truncated/padded to NameWidth)
//   - typeStr: Type identifier (go, godog, go-cli, mkdocs, etc.)
//   - result: Result string (e.g., "12/12" for tests, empty string for builds)
//   - duration: Time taken
func ResultLine(icon, name, typeStr, result string, duration time.Duration) string {
	// Truncate name if too long
	displayName := truncateOrPad(name, NameWidth)
	displayType := truncateOrPad(typeStr, TypeWidth)
	displayResult := truncateOrPad(result, ResultWidth)
	displayTime := formatDuration(duration)

	return fmt.Sprintf("%s %s %s %s %s",
		icon, displayName, displayType, displayResult, displayTime)
}

// ResultLineWithSuffix formats a completion line with an optional suffix (e.g., warnings).
func ResultLineWithSuffix(icon, name, typeStr, result string, duration time.Duration, suffix string) string {
	base := ResultLine(icon, name, typeStr, result, duration)
	if suffix != "" {
		return base + "  " + suffix
	}
	return base
}

// ResultLineNoTime formats a completion line without timing (timing shown in summary).
// Format: "✅ name                          type     result"
func ResultLineNoTime(icon, name, typeStr, result string) string {
	displayName := truncateOrPad(name, NameWidth)
	displayType := truncateOrPad(typeStr, TypeWidth)
	displayResult := truncateOrPad(result, ResultWidth)

	return fmt.Sprintf("%s %s %s %s", icon, displayName, displayType, displayResult)
}

// ResultLineNoTimeWithSuffix formats a completion line without timing but with suffix.
func ResultLineNoTimeWithSuffix(icon, name, typeStr, result, suffix string) string {
	base := ResultLineNoTime(icon, name, typeStr, result)
	if suffix != "" {
		return base + "  " + suffix
	}
	return base
}

// TimingLine formats a timing entry for the timing summary.
// Format: "  2.3s  module-name"
func TimingLine(duration time.Duration, name string) string {
	return fmt.Sprintf("%6.1fs  %s", duration.Seconds(), name)
}

// TimingTotal formats the total timing line.
// Format: "  27.8s  TOTAL"
func TimingTotal(duration time.Duration) string {
	return fmt.Sprintf("%6.1fs  TOTAL", duration.Seconds())
}

// SummaryLine formats a summary statistic line.
// Format: "  Label:  value"
func SummaryLine(label string, value interface{}) string {
	return fmt.Sprintf("  %-20s %v", label+":", value)
}

// SummaryCount formats a count summary with pass/fail breakdown.
// Format: "Modules: 5 total, 4 passed, 1 failed"
func SummaryCount(label string, total, passed, failed int) string {
	return fmt.Sprintf("%s: %d total, %d passed, %d failed", label, total, passed, failed)
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

// Writeln writes a formatted line with platform-specific line ending.
func Writeln(w fmt.Stringer, format string, args ...interface{}) {
	// This is a helper for io.Writer - actual implementation below
}

// FormatLine formats a line with platform-specific line ending.
func FormatLine(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...) + platform.LineEnding
}

// truncateOrPad ensures a string is exactly the specified width.
// If truncation is needed, the last character becomes "…"
func truncateOrPad(s string, width int) string {
	if len(s) > width {
		return s[:width-1] + "…"
	}
	return s + strings.Repeat(" ", width-len(s))
}

// formatDuration formats a duration compactly.
// Examples: "0.8s", "15.2s", "1m 23s"
func formatDuration(d time.Duration) string {
	seconds := d.Seconds()
	if seconds < 60 {
		return fmt.Sprintf("%5.1fs", seconds)
	}
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%dm %02ds", minutes, secs)
}

// FormatDurationShort formats duration as seconds with one decimal.
func FormatDurationShort(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
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
