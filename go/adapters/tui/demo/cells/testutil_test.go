package cells

import (
	"regexp"
	"strings"
)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// countFilledDots counts filled dots in output.
func countFilledDots(s string, ascii bool) int {
	if ascii {
		return strings.Count(s, "*")
	}
	return strings.Count(s, "●")
}

// countEmptyDots counts empty dots in output.
func countEmptyDots(s string, ascii bool) int {
	if ascii {
		return strings.Count(s, "o")
	}
	return strings.Count(s, "○")
}
