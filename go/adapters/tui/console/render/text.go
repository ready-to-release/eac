package render

import (
	"fmt"
	"strings"
	"time"
)

// FormatElapsed formats a duration for display.
func FormatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// PadOrTruncate pads or truncates a string to exactly the specified width.
func PadOrTruncate(s string, width int) string {
	if len(s) > width {
		if width > 1 {
			return s[:width-1] + "…"
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadOrTruncateASCII pads or truncates using ASCII-only characters (no Unicode ellipsis).
func PadOrTruncateASCII(s string, width int) string {
	if len(s) > width {
		if width > 2 {
			return s[:width-2] + ".."
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PadRight pads a string to the right to reach the specified width.
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// StripMarkdownPipes removes leading/trailing pipe characters from markdown table lines.
// Converts "| col1 | col2 |" to "col1   col2" for cleaner terminal display.
func StripMarkdownPipes(line string) string {
	// Handle separator lines (dashes)
	if strings.Contains(line, "---") {
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		line = strings.ReplaceAll(line, "|", " ")
		return strings.TrimSpace(line)
	}

	// Handle data lines
	if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")
		line = strings.ReplaceAll(line, " | ", "   ")
		return strings.TrimSpace(line)
	}

	return line
}

// StripToolTypeSuffix removes :system or :container suffix from tool names.
func StripToolTypeSuffix(toolID string) string {
	if idx := strings.LastIndex(toolID, ":"); idx != -1 {
		suffix := toolID[idx+1:]
		if suffix == "system" || suffix == "container" {
			return toolID[:idx]
		}
	}
	return toolID
}

// GetModuleName extracts the module name from a full moniker (module:component).
func GetModuleName(moniker string) string {
	if idx := strings.Index(moniker, ":"); idx >= 0 {
		return moniker[:idx]
	}
	return moniker
}

// WeightDigit returns a circled Unicode digit for the given weight (1-9).
func WeightDigit(weight int) string {
	digits := []string{"❶", "❷", "❸", "❹", "❺", "❻", "❼", "❽", "❾"}
	if weight < 1 {
		return ""
	}
	if weight > 9 {
		weight = 9
	}
	return digits[weight-1]
}
