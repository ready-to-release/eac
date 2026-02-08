// Package aiutil provides shared utilities for AI-powered commit message generation commands
// (create commit-message, create squash-message).
package aiutil

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/logging"
)

// LogDebugArtifact logs debug content with labeled sections to the log file.
// Used by AI generation commands to record intermediate outputs for troubleshooting.
func LogDebugArtifact(log *logging.ComponentLogger, label, content string) {
	log.Debug(fmt.Sprintf("=== %s START ===", label))
	log.Debug(content)
	log.Debug(fmt.Sprintf("=== %s END ===", label))
}

// ExtractFirstSentence extracts the first sentence from text.
// Returns the first sentence terminated by a period, question mark, or exclamation mark.
// If no sentence delimiter is found, returns the first line with a trailing period.
// The fallback parameter is used when text is empty.
func ExtractFirstSentence(text, fallback string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		if fallback != "" {
			return ensurePeriod(fallback)
		}
		return "Summary of changes."
	}

	// Find first sentence ending
	for _, delim := range []string{". ", ".\n", "? ", "?\n", "! ", "!\n"} {
		if idx := strings.Index(text, delim); idx != -1 {
			return strings.TrimSpace(text[:idx+1])
		}
	}

	// If text ends with punctuation, return first line
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "?") || strings.HasSuffix(text, "!") {
		lines := strings.Split(text, "\n")
		return strings.TrimSpace(lines[0])
	}

	// Fallback: first line with period
	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return ensurePeriod(fallback)
	}
	return ensurePeriod(firstLine)
}

func ensurePeriod(s string) string {
	if s == "" {
		return "."
	}
	if !strings.HasSuffix(s, ".") {
		return s + "."
	}
	return s
}
