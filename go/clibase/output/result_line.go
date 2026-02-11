package output

import (
	"fmt"
	"strings"
	"time"
)

// ResultLine formats a completion line with aligned columns.
// Format: "✅ name                          type     result   time"
//
// Parameters:
//   - icon: Status icon (IconPass, IconFail, etc.)
//   - name: Module or package name (truncated/padded to NameWidth)
//   - typeStr: Type identifier (go, godog, container, typescript, static, etc.)
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

// truncateOrPad ensures a string is exactly the specified width.
// If truncation is needed, the last character becomes "\u2026"
func truncateOrPad(s string, width int) string {
	if len(s) > width {
		return s[:width-1] + "\u2026"
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
