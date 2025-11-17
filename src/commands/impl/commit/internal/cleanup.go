package commitmessage

import (
	"regexp"
	"strings"
)

// AutoCleanup performs automatic fixes on commit message before validation
// This catches common issues that can be fixed programmatically without AI
func AutoCleanup(commitMessage string) string {
	// PHASE 1: Normalize all spacing and blank lines first
	// This creates a stable foundation for content fixes
	lines := normalizeSpacing(commitMessage)

	// PHASE 2: Fix content (titles, subject lines, body wrapping)
	lines = fixContent(lines)

	// PHASE 3: Final cleanup (close code blocks, ensure trailing blank line)
	result := strings.Join(lines, "\n")
	result = ensureCodeBlocksClosed(result)

	// Remove trailing separators and blank lines
	result = strings.TrimRight(result, "\n\t ")

	// Remove trailing --- separator if present
	for strings.HasSuffix(strings.TrimRight(result, "\n\t "), "---") {
		result = strings.TrimSuffix(strings.TrimRight(result, "\n\t "), "---")
		result = strings.TrimRight(result, "\n\t ")
	}

	// Ensure file ends with exactly one blank line
	if result != "" {
		result += "\n\n"
	}

	return result
}

// normalizeSpacing removes duplicate blank lines and ensures proper spacing around sections
func normalizeSpacing(commitMessage string) []string {
	lines := strings.Split(commitMessage, "\n")
	normalized := make([]string, 0, len(lines))

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip duplicate blank lines
		if trimmed == "" {
			// Keep blank line if previous line wasn't blank
			if i == 0 || len(normalized) == 0 || strings.TrimSpace(normalized[len(normalized)-1]) != "" {
				normalized = append(normalized, "")
			}
			continue
		}

		normalized = append(normalized, line)
	}

	return normalized
}

// fixContent handles title truncation, subject line joining/truncation, and body wrapping
func fixContent(lines []string) []string {
	cleaned := make([]string, 0, len(lines))

	inCodeBlock := false
	inBodySection := false
	bodyBuffer := []string{}
	lastWasModuleHeader := false
	needBlankLineAfterCodeBlock := false
	inMultiLineSpecialField := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// If we need a blank line after code block and this is non-empty content, add it
		if needBlankLineAfterCodeBlock && trimmed != "" && !strings.HasPrefix(trimmed, "```") {
			if len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
				cleaned = append(cleaned, "")
			}
			needBlankLineAfterCodeBlock = false
		}

		// Track code block state and ensure blank lines before/after
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Opening fence - ensure blank line before it
				if len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
					cleaned = append(cleaned, "")
				}
				inCodeBlock = true
				cleaned = append(cleaned, line)
				continue
			} else {
				// Closing fence - add it, then mark that we need blank line after
				cleaned = append(cleaned, line)
				inCodeBlock = false
				needBlankLineAfterCodeBlock = true
				continue
			}
		}

		// FIX 1: Truncate header to 72 chars with ellipsis if needed
		// New format: <type>(<scope>): <summary>
		conventionalHeaderRegex := regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)\([^)]+\):\s*(.+)`)
		if i == 0 && conventionalHeaderRegex.MatchString(trimmed) {
			if len(trimmed) > 72 {
				// Truncate to 69 chars to leave room for "..."
				truncated := trimmed[:69]
				truncated = strings.TrimRight(truncated, " .")
				line = truncated + "..."
			} else {
				// Only remove trailing period
				line = strings.TrimSuffix(trimmed, ".")
			}
			cleaned = append(cleaned, line)
			// After header, we're in top-level body section
			inBodySection = true
			continue
		}

		// FIX 2: Handle module headers
		// New format: plain text module name (followed by dashes on next line)
		if i < len(lines)-1 {
			nextTrimmed := strings.TrimSpace(lines[i+1])
			if isDashesLine(nextTrimmed) && isModuleName(trimmed) {
				// This is a module header in new format
				// Ensure blank line before if needed
				if len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
					cleaned = append(cleaned, "")
				}
				lastWasModuleHeader = true
				cleaned = append(cleaned, line)
				// Add the dashes line on next iteration
				continue
			}
		}

		// FIX 3: Detect separators BEFORE dashes (since "---" is both a separator and dashes)
		if trimmed == "---" {
			// Flush buffered body text
			if inBodySection && len(bodyBuffer) > 0 {
				cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
				bodyBuffer = []string{}
			}
			inBodySection = false

			// Add blank line before separator if needed
			if len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
				cleaned = append(cleaned, "")
			}

			cleaned = append(cleaned, line)
			continue
		}

		// Pass through dashes line (part of module header)
		if isDashesLine(trimmed) {
			cleaned = append(cleaned, line)
			inBodySection = true
			continue
		}

		// FIX 4: Handle subject lines (WRAP if too long, remove trailing periods)
		// Format: <module>: <type>: <description>
		subjectRegex := regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)`)

		if subjectRegex.MatchString(trimmed) {
			// Remove trailing period
			line = strings.TrimSuffix(strings.TrimSpace(line), ".")

			// WRAP if too long (don't truncate semantic commits)
			if len(line) > 72 {
				wrapped := wrapSemanticCommitLine(line)
				cleaned = append(cleaned, wrapped...)
			} else {
				cleaned = append(cleaned, line)
			}
			lastWasModuleHeader = false
			continue
		}

		// FIX 5: Detect code block end
		if inBodySection && strings.HasPrefix(trimmed, "```") {
			// Flush buffered body text
			if len(bodyBuffer) > 0 {
				cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
				bodyBuffer = []string{}
			}
			inBodySection = false
			cleaned = append(cleaned, line)
			continue
		}

		// Check if this is a special field line (Auditor-Summary, Changes, etc.)
		isSpecialField := strings.HasPrefix(trimmed, "Auditor-Summary:") ||
		   strings.HasPrefix(trimmed, "Changes:") ||
		   strings.HasPrefix(trimmed, "BREAKING CHANGE:") ||
		   strings.HasPrefix(trimmed, "Refs:") ||
		   strings.HasPrefix(trimmed, "Closes:") ||
		   strings.HasPrefix(trimmed, "Co-authored-by:")

		if isSpecialField {
			// Flush any buffered body text first
			if len(bodyBuffer) > 0 {
				cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
				bodyBuffer = []string{}
			}
			cleaned = append(cleaned, line)
			inMultiLineSpecialField = true
			continue
		}

		// Handle continuation lines of multi-line special fields
		if inMultiLineSpecialField {
			if trimmed == "" {
				// Blank line ends the special field
				inMultiLineSpecialField = false
				cleaned = append(cleaned, line)
			} else {
				// Continuation line - pass through without buffering
				cleaned = append(cleaned, line)
			}
			continue
		}

		// Buffer body text lines (ensuring blank line after header)
		if inBodySection && !inCodeBlock && trimmed != "" && !strings.HasPrefix(trimmed, "|") {
			// If this is the first body text after a header, ensure blank line separator
			if lastWasModuleHeader {
				// Add blank line after header if not already present
				if len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
					cleaned = append(cleaned, "")
				}
			}
			bodyBuffer = append(bodyBuffer, trimmed)
			lastWasModuleHeader = false
			continue
		}

		// Skip duplicate blank lines
		if trimmed == "" && len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) == "" {
			continue
		}

		// Don't reset lastWasModuleHeader for blank lines (subject might be after blank line)
		if trimmed != "" {
			lastWasModuleHeader = false
		}
		cleaned = append(cleaned, line)
	}

	// Flush any remaining body text
	if len(bodyBuffer) > 0 {
		cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
	}

	return cleaned
}

// wrapSemanticCommitLine wraps a semantic commit line at 72 characters
// Preserves the format: <module>: <type>: <description>
func wrapSemanticCommitLine(line string) []string {
	if len(line) <= 72 {
		return []string{line}
	}

	// Split at 72 chars and wrap the rest with proper indentation
	var wrapped []string
	currentLine := ""
	words := strings.Fields(line)

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= 72 {
			currentLine = testLine
		} else {
			// Flush current line
			if currentLine != "" {
				wrapped = append(wrapped, currentLine)
			}
			currentLine = word
		}
	}

	// Add remaining line
	if currentLine != "" {
		wrapped = append(wrapped, currentLine)
	}

	return wrapped
}

// wrapBodyText joins buffered lines and reflows at 72 characters
func wrapBodyText(lines []string) []string {
	// Join all lines into one paragraph
	paragraph := strings.Join(lines, " ")

	// Split into sentences/phrases and wrap at 72 chars
	var wrapped []string
	var currentLine string

	words := strings.Fields(paragraph)
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= 72 {
			currentLine = testLine
		} else {
			// Current word would exceed limit, flush current line
			if currentLine != "" {
				wrapped = append(wrapped, currentLine)
			}
			currentLine = word
		}
	}

	// Add remaining line
	if currentLine != "" {
		wrapped = append(wrapped, currentLine)
	}

	return wrapped
}

// ensureCodeBlocksClosed adds missing closing fences
func ensureCodeBlocksClosed(content string) string {
	lines := strings.Split(content, "\n")
	openFences := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			openFences++
		}
	}

	// If odd number of fences, we have unclosed blocks
	if openFences%2 != 0 {
		// Find last code block opening and add closing fence before agent approval or end
		result := strings.TrimSpace(content)

		// Add closing fence before agent approval if it exists
		if strings.Contains(result, "Agent:") {
			agentIndex := strings.LastIndex(result, "Agent:")
			before := result[:agentIndex]
			after := result[agentIndex:]
			return strings.TrimSpace(before) + "\n```\n\n" + after
		}

		// Otherwise add at end
		return result + "\n```\n"
	}

	return content
}

// GetCleanupStats returns what was cleaned
func GetCleanupStats(original, cleaned string) []string {
	stats := []string{}

	// Check what changed
	if original != cleaned {
		origLines := strings.Split(original, "\n")
		cleanLines := strings.Split(cleaned, "\n")

		// Title period removed?
		if len(origLines) > 0 && len(cleanLines) > 0 {
			if strings.HasSuffix(origLines[0], ".") && !strings.HasSuffix(cleanLines[0], ".") {
				stats = append(stats, "Removed trailing period from title")
			}
		}

		// Agent approval added?
		if !strings.Contains(original, "Agent:") && strings.Contains(cleaned, "Agent:") {
			stats = append(stats, "Added missing agent approval line")
		}

		// Code blocks closed?
		origFences := strings.Count(original, "```")
		cleanFences := strings.Count(cleaned, "```")
		if cleanFences > origFences {
			stats = append(stats, "Closed unclosed code block")
		}

		// Subject line periods?
		subjectRegex := regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)\.$`)
		for i, line := range origLines {
			trimmed := strings.TrimSpace(line)
			if subjectRegex.MatchString(trimmed) {
				if i < len(cleanLines) && !strings.HasSuffix(strings.TrimSpace(cleanLines[i]), ".") {
					stats = append(stats, "Removed trailing period from subject line")
					break
				}
			}
		}
	}

	return stats
}
