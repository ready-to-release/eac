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

	// PHASE 3: Final cleanup (close code blocks, ensure trailing newline)
	result := strings.Join(lines, "\n")
	result = ensureCodeBlocksClosed(result)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
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

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track code block state
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
		}

		// FIX 1: Truncate title to 72 chars with ellipsis if needed
		if i == 0 && strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimPrefix(trimmed, "# ")
			if len("# "+title) > 72 {
				// Truncate to 69 chars to leave room for "..."
				title = title[:66]
				// Remove any trailing spaces, periods, or punctuation before adding ellipsis
				title = strings.TrimRight(title, " .")
				title = title + "..."
			} else {
				// Only remove trailing period if NOT truncated (no ellipsis)
				title = strings.TrimSuffix(title, ".")
			}
			line = "# " + title
		}

		// FIX 2: Remove trailing periods from module headers (## <module-name>)
		if strings.HasPrefix(trimmed, "## ") &&
			trimmed != "## Summary" &&
			trimmed != "## Files affected" &&
			strings.HasSuffix(trimmed, ".") {
			line = strings.TrimSuffix(line, ".")
		}

		// FIX 3: Handle subject lines (truncate if too long, remove trailing periods)
		// Format: <module>: <type>: <description>
		subjectRegex := regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)`)

		// Special case: if last line was a module header, this might be a wrapped subject line
		// Join continuation lines until we hit a blank line or code block
		if lastWasModuleHeader && subjectRegex.MatchString(trimmed) {
			subjectLine := trimmed

			// Look ahead and join continuation lines
			j := i + 1
			for j < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[j])

				// Stop at blank line, separator, code block, or next header
				if nextTrimmed == "" ||
					nextTrimmed == "---" ||
					strings.HasPrefix(nextTrimmed, "```") ||
					strings.HasPrefix(nextTrimmed, "## ") ||
					strings.HasPrefix(nextTrimmed, "| ") {
					break
				}

				subjectLine += " " + nextTrimmed
				j++
			}

			// Remove trailing period
			subjectLine = strings.TrimSuffix(subjectLine, ".")

			// Truncate if too long
			if len(subjectLine) > 72 {
				subjectLine = subjectLine[:69]
				subjectLine = strings.TrimRight(subjectLine, " .")
				subjectLine = subjectLine + "..."
			}

			cleaned = append(cleaned, subjectLine)

			// Skip the continuation lines we just processed
			for k := i + 1; k < j; k++ {
				lines[k] = "" // Mark as processed
			}

			lastWasModuleHeader = false
			continue
		}

		if subjectRegex.MatchString(trimmed) {
			// Remove trailing period
			line = strings.TrimSuffix(strings.TrimSpace(line), ".")

			// Truncate if too long
			if len(line) > 72 {
				line = line[:69]
				line = strings.TrimRight(line, " .")
				line = line + "..."
			}
		}

		// FIX 4: Track body sections for line wrapping
		// Detect start of body sections (Summary or module body text)
		if trimmed == "## Summary" || (strings.HasPrefix(trimmed, "## ") && trimmed != "## Files affected") {
			lastWasModuleHeader = true
			// Flush any buffered body text from previous section
			if len(bodyBuffer) > 0 {
				cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
				bodyBuffer = []string{}
			}
			inBodySection = true
			cleaned = append(cleaned, line)
			continue
		}

		// Detect end of body section
		if inBodySection && (trimmed == "---" || trimmed == "## Files affected" || strings.HasPrefix(trimmed, "```")) {
			// Flush buffered body text
			if len(bodyBuffer) > 0 {
				cleaned = append(cleaned, wrapBodyText(bodyBuffer)...)
				bodyBuffer = []string{}
			}
			inBodySection = false

			// Add blank line before divider if needed
			if trimmed == "---" && len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) != "" {
				cleaned = append(cleaned, "")
			}

			cleaned = append(cleaned, line)
			continue
		}

		// Buffer body text lines (skip empty lines - markdown will have blank line after header)
		if inBodySection && !inCodeBlock && trimmed != "" && !strings.HasPrefix(trimmed, "|") {
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
