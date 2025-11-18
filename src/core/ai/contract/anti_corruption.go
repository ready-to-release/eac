package contract

import (
	"strings"
)

// AntiCorruptionFilter applies anti-corruption rules to clean AI output
type AntiCorruptionFilter struct {
	rules              *AntiCorruptionRules
	contentStartMarker string // e.g., "Feature:" for Gherkin, or custom pattern
}

// NewAntiCorruptionFilter creates a new filter with the given rules
func NewAntiCorruptionFilter(rules *AntiCorruptionRules, contentStartMarker string) *AntiCorruptionFilter {
	return &AntiCorruptionFilter{
		rules:              rules,
		contentStartMarker: contentStartMarker,
	}
}

// Apply filters AI output using the anti-corruption rules
//
// The filtering algorithm:
// 1. Remove markdown code fences if entire output is wrapped
// 2. Scan for content start marker (e.g., "Feature:" or first valid line)
// 3. Remove all noise before content starts
// 4. Keep all content from start marker onwards
// 5. Trim trailing whitespace
func (f *AntiCorruptionFilter) Apply(output string) string {
	output = strings.TrimSpace(output)

	// Step 1: Remove markdown code fences if entire output is wrapped
	output = f.removeCodeFences(output)

	// Step 2-4: Filter noise before content starts
	lines := strings.Split(output, "\n")
	cleaned := f.filterLines(lines)

	// Step 5: Trim trailing whitespace
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// removeCodeFences removes wrapping markdown code fences
func (f *AntiCorruptionFilter) removeCodeFences(output string) string {
	if !strings.HasPrefix(output, "```") {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return output
	}

	// Remove opening fence (may have language identifier like ```gherkin)
	lines = lines[1:]

	// Remove closing fence
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// filterLines filters out noise lines before content starts
func (f *AntiCorruptionFilter) filterLines(lines []string) []string {
	var cleaned []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Once content starts, keep everything
		if foundContent {
			cleaned = append(cleaned, line)
			continue
		}

		// Check if content starts here
		if f.isContentStart(trimmed) {
			foundContent = true
			cleaned = append(cleaned, line)
			continue
		}

		// Skip noise before content starts
		if f.shouldSkipLine(trimmed) {
			continue
		}

		// If we reach here and line is not empty, content might have started
		if trimmed != "" {
			// Check for agent signatures (end of useful content)
			if f.isAgentSignature(trimmed) {
				break
			}

			// Line doesn't match any noise patterns - might be content
			// Let it through (conservative approach)
			foundContent = true
			cleaned = append(cleaned, line)
		}
	}

	return cleaned
}

// isContentStart checks if a line marks the start of actual content
func (f *AntiCorruptionFilter) isContentStart(trimmed string) bool {
	if f.contentStartMarker != "" {
		return strings.HasPrefix(trimmed, f.contentStartMarker)
	}
	return false
}

// shouldSkipLine determines if a line should be filtered out
func (f *AntiCorruptionFilter) shouldSkipLine(trimmed string) bool {
	if f.rules == nil {
		return false
	}

	// Empty lines before content
	if trimmed == "" {
		return true
	}

	// Lines with forbidden prefixes
	for _, prefix := range f.rules.ForbiddenPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	// Lines containing forbidden strings
	for _, contains := range f.rules.ForbiddenContains {
		if strings.Contains(trimmed, contains) {
			return true
		}
	}

	// Lines with forbidden emojis
	for _, emoji := range f.rules.ForbiddenEmojis {
		if strings.Contains(trimmed, emoji) {
			return true
		}
	}

	// Horizontal rules
	if trimmed == "---" || trimmed == "___" || trimmed == "***" {
		return true
	}

	// Bold announcements (e.g., **Title**)
	if strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**") {
		return true
	}

	// Bullet points (meta-descriptions)
	if strings.HasPrefix(trimmed, "- ") {
		return true
	}

	return false
}

// isAgentSignature checks if a line is an agent signature (end marker)
func (f *AntiCorruptionFilter) isAgentSignature(trimmed string) bool {
	if f.rules == nil {
		return false
	}

	for _, signature := range f.rules.AgentSignatures {
		if strings.HasPrefix(trimmed, signature) {
			return true
		}
	}

	return false
}

// ApplyWithFallback applies anti-corruption rules with fallback to hardcoded patterns
//
// If rules is nil, uses hardcoded fallback patterns for basic filtering.
// This ensures the filter still works even if anti-corruption.yml is missing.
func ApplyWithFallback(output string, rules *AntiCorruptionRules, contentStartMarker string) string {
	if rules != nil {
		filter := NewAntiCorruptionFilter(rules, contentStartMarker)
		return filter.Apply(output)
	}

	// Fallback: Use hardcoded basic filtering
	return applyHardcodedFiltering(output, contentStartMarker)
}

// applyHardcodedFiltering applies basic hardcoded filtering when no rules available
func applyHardcodedFiltering(output string, contentStartMarker string) string {
	output = strings.TrimSpace(output)

	// Remove code fences
	if strings.HasPrefix(output, "```") {
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			lines = lines[1:]
		}
		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
			lines = lines[:len(lines)-1]
		}
		output = strings.Join(lines, "\n")
	}

	lines := strings.Split(output, "\n")
	var cleaned []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Once we find content marker, keep everything
		if contentStartMarker != "" && strings.HasPrefix(trimmed, contentStartMarker) {
			foundContent = true
		}

		if foundContent {
			cleaned = append(cleaned, line)
			continue
		}

		// Skip common noise patterns (hardcoded)
		if trimmed == "" ||
			strings.Contains(trimmed, "**Initialized") ||
			strings.Contains(trimmed, "ready to assist") ||
			strings.Contains(trimmed, "Here is") ||
			strings.Contains(trimmed, "Here's") ||
			trimmed == "---" {
			continue
		}

		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
