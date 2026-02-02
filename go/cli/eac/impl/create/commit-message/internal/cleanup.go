package commitmessage

import (
	"regexp"
	"strings"
)

// Precompiled regular expressions for performance.
var (
	conventionalHeaderRegex = regexp.MustCompile(`^(feat|fix|refactor|docs|chore|test|perf|style)\([^)]+\):\s*(.+)`)
	subjectLineRegex        = regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)`)
	subjectRegexWithPeriod  = regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)\.$`)
)

// AutoCleanup performs automatic fixes on commit message before validation
// This catches common issues that can be fixed programmatically without AI.
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

// normalizeSpacing removes duplicate blank lines and ensures proper spacing around sections.
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
// Refactored into smaller, focused functions for better maintainability.
func fixContent(lines []string) []string {
	// Create processing context
	ctx := &contentFixContext{
		lines:                       lines,
		cleaned:                     make([]string, 0, len(lines)),
		inCodeBlock:                 false,
		inBodySection:               false,
		bodyBuffer:                  []string{},
		lastWasModuleHeader:         false,
		needBlankLineAfterCodeBlock: false,
		inMultiLineSpecialField:     false,
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		ctx.currentIndex = i
		ctx.currentLine = line
		ctx.currentTrimmed = trimmed

		// Process line through different fixers
		if ctx.handleCodeBlock() {
			continue
		}
		if ctx.handleHeader() {
			continue
		}
		if ctx.handleModuleSection() {
			continue
		}
		if ctx.handleSeparator() {
			continue
		}
		if ctx.handleDashes() {
			continue
		}
		if ctx.handleSubjectLine() {
			continue
		}
		if ctx.handleSpecialField() {
			continue
		}
		if ctx.handleBodyText() {
			continue
		}

		// Default: pass through
		ctx.passThrough()
	}

	// Flush any remaining body text
	if len(ctx.bodyBuffer) > 0 {
		ctx.cleaned = append(ctx.cleaned, wrapBodyText(ctx.bodyBuffer)...)
	}

	return ctx.cleaned
}

// contentFixContext holds state for content fixing.
type contentFixContext struct {
	lines                       []string
	cleaned                     []string
	currentIndex                int
	currentLine                 string
	currentTrimmed              string
	inCodeBlock                 bool
	inBodySection               bool
	bodyBuffer                  []string
	lastWasModuleHeader         bool
	needBlankLineAfterCodeBlock bool
	inMultiLineSpecialField     bool
}

// handleCodeBlock handles code block state and spacing.
func (ctx *contentFixContext) handleCodeBlock() bool {
	trimmed := ctx.currentTrimmed
	line := ctx.currentLine

	// Add blank line after code block if needed
	if ctx.needBlankLineAfterCodeBlock && trimmed != "" && !strings.HasPrefix(trimmed, "```") {
		if len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) != "" {
			ctx.cleaned = append(ctx.cleaned, "")
		}
		ctx.needBlankLineAfterCodeBlock = false
	}

	// Track code block state
	if strings.HasPrefix(trimmed, "```") {
		if !ctx.inCodeBlock {
			// Opening fence - ensure blank line before it
			if len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) != "" {
				ctx.cleaned = append(ctx.cleaned, "")
			}
			ctx.inCodeBlock = true
			ctx.cleaned = append(ctx.cleaned, line)
			return true
		} else {
			// Closing fence
			ctx.cleaned = append(ctx.cleaned, line)
			ctx.inCodeBlock = false
			ctx.needBlankLineAfterCodeBlock = true
			return true
		}
	}

	// Handle code block end in body section
	if ctx.inBodySection && strings.HasPrefix(trimmed, "```") {
		if len(ctx.bodyBuffer) > 0 {
			ctx.cleaned = append(ctx.cleaned, wrapBodyText(ctx.bodyBuffer)...)
			ctx.bodyBuffer = []string{}
		}
		ctx.inBodySection = false
		ctx.cleaned = append(ctx.cleaned, line)
		return true
	}

	return false
}

// handleHeader handles header cleanup.
func (ctx *contentFixContext) handleHeader() bool {
	if ctx.currentIndex == 0 && conventionalHeaderRegex.MatchString(ctx.currentTrimmed) {
		// Only remove trailing period (validation will warn if too long)
		line := strings.TrimSuffix(ctx.currentTrimmed, ".")
		ctx.cleaned = append(ctx.cleaned, line)
		ctx.inBodySection = true
		return true
	}
	return false
}

// handleModuleSection handles module headers.
func (ctx *contentFixContext) handleModuleSection() bool {
	if ctx.currentIndex >= len(ctx.lines)-1 {
		return false
	}

	nextTrimmed := strings.TrimSpace(ctx.lines[ctx.currentIndex+1])
	if isDashesLine(nextTrimmed) && isModuleName(ctx.currentTrimmed) {
		// Ensure blank line before module header
		if len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) != "" {
			ctx.cleaned = append(ctx.cleaned, "")
		}
		ctx.lastWasModuleHeader = true
		ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		return true
	}
	return false
}

// handleSeparator handles horizontal rule separators.
func (ctx *contentFixContext) handleSeparator() bool {
	if ctx.currentTrimmed == "---" {
		// Flush buffered body text
		if ctx.inBodySection && len(ctx.bodyBuffer) > 0 {
			ctx.cleaned = append(ctx.cleaned, wrapBodyText(ctx.bodyBuffer)...)
			ctx.bodyBuffer = []string{}
		}
		ctx.inBodySection = false

		// Add blank line before separator
		if len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) != "" {
			ctx.cleaned = append(ctx.cleaned, "")
		}

		ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		return true
	}
	return false
}

// handleDashes handles dashes lines (module header underlines).
func (ctx *contentFixContext) handleDashes() bool {
	if isDashesLine(ctx.currentTrimmed) {
		ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		ctx.inBodySection = true
		return true
	}
	return false
}

// handleSubjectLine handles subject line wrapping.
func (ctx *contentFixContext) handleSubjectLine() bool {
	if subjectLineRegex.MatchString(ctx.currentTrimmed) {
		// Remove trailing period
		line := strings.TrimSuffix(strings.TrimSpace(ctx.currentLine), ".")

		// Wrap if too long
		if len(line) > MaxSubjectLength {
			wrapped := wrapSemanticCommitLine(line)
			ctx.cleaned = append(ctx.cleaned, wrapped...)
		} else {
			ctx.cleaned = append(ctx.cleaned, line)
		}
		ctx.lastWasModuleHeader = false
		return true
	}
	return false
}

// handleSpecialField handles special fields like Auditor-Summary, Changes, etc.
func (ctx *contentFixContext) handleSpecialField() bool {
	trimmed := ctx.currentTrimmed

	isSpecialField := strings.HasPrefix(trimmed, "Auditor-Summary:") ||
		strings.HasPrefix(trimmed, "Changes:") ||
		strings.HasPrefix(trimmed, "BREAKING CHANGE:") ||
		strings.HasPrefix(trimmed, "Refs:") ||
		strings.HasPrefix(trimmed, "Closes:") ||
		strings.HasPrefix(trimmed, "Co-authored-by:")

	if isSpecialField {
		// Flush buffered body text
		if len(ctx.bodyBuffer) > 0 {
			ctx.cleaned = append(ctx.cleaned, wrapBodyText(ctx.bodyBuffer)...)
			ctx.bodyBuffer = []string{}
		}
		ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		ctx.inMultiLineSpecialField = true
		return true
	}

	// Handle continuation lines
	if ctx.inMultiLineSpecialField {
		if trimmed == "" {
			ctx.inMultiLineSpecialField = false
			ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		} else {
			ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
		}
		return true
	}

	return false
}

// handleBodyText handles body text buffering and wrapping.
func (ctx *contentFixContext) handleBodyText() bool {
	trimmed := ctx.currentTrimmed

	if ctx.inBodySection && !ctx.inCodeBlock && trimmed != "" && !strings.HasPrefix(trimmed, "|") {
		// Ensure blank line after module header
		if ctx.lastWasModuleHeader {
			if len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) != "" {
				ctx.cleaned = append(ctx.cleaned, "")
			}
		}
		ctx.bodyBuffer = append(ctx.bodyBuffer, trimmed)
		ctx.lastWasModuleHeader = false
		return true
	}

	return false
}

// passThrough passes line through with duplicate blank line removal.
func (ctx *contentFixContext) passThrough() {
	trimmed := ctx.currentTrimmed

	// Skip duplicate blank lines
	if trimmed == "" && len(ctx.cleaned) > 0 && strings.TrimSpace(ctx.cleaned[len(ctx.cleaned)-1]) == "" {
		return
	}

	// Don't reset lastWasModuleHeader for blank lines
	if trimmed != "" {
		ctx.lastWasModuleHeader = false
	}
	ctx.cleaned = append(ctx.cleaned, ctx.currentLine)
}

// wrapSemanticCommitLine wraps a semantic commit line at MaxSubjectLength characters
// Preserves the format: <module>: <type>: <description>.
func wrapSemanticCommitLine(line string) []string {
	if len(line) <= MaxSubjectLength {
		return []string{line}
	}

	// Split at MaxSubjectLength chars and wrap the rest with proper indentation
	var wrapped []string
	currentLine := ""
	words := strings.Fields(line)

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= MaxSubjectLength {
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

// wrapBodyText joins buffered lines and reflows at MaxLineLength characters.
func wrapBodyText(lines []string) []string {
	// Join all lines into one paragraph
	paragraph := strings.Join(lines, " ")

	// Split into sentences/phrases and wrap at MaxLineLength chars
	var wrapped []string
	var currentLine string

	words := strings.Fields(paragraph)
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= MaxLineLength {
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

// ensureCodeBlocksClosed adds missing closing fences.
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

// GetCleanupStats returns what was cleaned.
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
		for i, line := range origLines {
			trimmed := strings.TrimSpace(line)
			if subjectRegexWithPeriod.MatchString(trimmed) {
				if i < len(cleanLines) && !strings.HasSuffix(strings.TrimSpace(cleanLines[i]), ".") {
					stats = append(stats, "Removed trailing period from subject line")
					break
				}
			}
		}
	}

	return stats
}
