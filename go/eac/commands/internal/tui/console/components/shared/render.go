package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Truncate truncates a string to maxWidth, adding ellipsis if needed.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	// Find position where we can fit "..."
	width := 0
	cutoff := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if width+rw > maxWidth-3 {
			cutoff = i
			break
		}
		width += rw
		cutoff = i + 1
	}
	return s[:cutoff] + "..."
}

// PadRight pads a string to the given width with spaces.
func PadRight(s string, width int) string {
	current := runewidth.StringWidth(s)
	if current >= width {
		return s
	}
	return s + strings.Repeat(" ", width-current)
}

// PadToHeight ensures output has exactly height lines.
func PadToHeight(content string, height int) string {
	lines := strings.Split(content, "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

// HorizontalJoin joins two blocks of text side by side.
func HorizontalJoin(left, right string, height int) string {
	leftLines := strings.Split(PadToHeight(left, height), "\n")
	rightLines := strings.Split(PadToHeight(right, height), "\n")

	// Find max width of left side
	maxLeftWidth := 0
	for _, line := range leftLines {
		w := runewidth.StringWidth(line)
		if w > maxLeftWidth {
			maxLeftWidth = w
		}
	}

	var result []string
	for i := 0; i < height; i++ {
		l := PadRight(leftLines[i], maxLeftWidth)
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		result = append(result, l+" "+r)
	}

	return strings.Join(result, "\n")
}

// BoxWrap wraps content in a simple box border.
func BoxWrap(content string, width int) string {
	lines := strings.Split(content, "\n")

	// Calculate content width
	contentWidth := width - 4 // Account for "│ " and " │"
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Top border
	top := "┌" + strings.Repeat("─", width-2) + "┐"
	bottom := "└" + strings.Repeat("─", width-2) + "┘"

	var result []string
	result = append(result, BorderStyle.Render(top))

	for _, line := range lines {
		// Truncate or pad line to fit
		truncated := Truncate(line, contentWidth)
		padded := PadRight(truncated, contentWidth)
		result = append(result, BorderStyle.Render("│ ")+padded+BorderStyle.Render(" │"))
	}

	result = append(result, BorderStyle.Render(bottom))
	return strings.Join(result, "\n")
}

// LabelValue renders a label: value pair.
func LabelValue(label, value string) string {
	return DimStyle.Render(label+": ") + value
}

// FormatDuration formats a duration in a human-readable way.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

// TreeBranch returns the appropriate tree branch character.
func TreeBranch(isLast bool) string {
	if isLast {
		return "└─"
	}
	return "├─"
}

// LayerHeader renders a layer header for the execution tree.
func LayerHeader(layerNum int) string {
	return lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render(
		fmt.Sprintf("Layer %d", layerNum),
	)
}

// RepeatChar repeats a character n times.
func RepeatChar(char string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(char, n)
}
