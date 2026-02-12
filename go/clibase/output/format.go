// Package output provides unified formatting utilities for command output.
// This ensures consistent "look and feel" across test, build, and other commands.
package output

import (
	"fmt"
	"io"
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
