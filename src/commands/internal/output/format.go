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
	NameWidth   = 32 // Module/package name
	TypeWidth   = 8  // Type (go, godog, go-cli, mkdocs, etc.)
	ResultWidth = 7  // Result (12/12, -, etc.)
	TimeWidth   = 7  // Duration (0.8s, 15.2s, etc.)
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
//   - result: Result string (e.g., "12/12", "-" for builds)
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
