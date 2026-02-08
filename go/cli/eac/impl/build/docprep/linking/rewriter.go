package linking

import (
	"fmt"
	"regexp"
	"strings"
)

// ReplaceOutsideCodeBlocks replaces old with new, but only outside fenced code blocks.
// It replaces within link contexts (markdown links/images, HTML src/href attributes)
// to prevent corrupting plain text.
func ReplaceOutsideCodeBlocks(content, old, new string) string {
	parts := FencedCodeBlockSplitPattern.Split(content, -1)
	codeBlocks := FencedCodeBlockSplitPattern.FindAllString(content, -1)

	var result strings.Builder
	for i, part := range parts {
		result.WriteString(ReplaceLinkPaths(part, old, new))

		if i < len(codeBlocks) {
			result.WriteString(codeBlocks[i])
		}
	}

	return result.String()
}

// ReplaceLinkPaths replaces old path with new path within link contexts:
// - Markdown links: [text](path) or [text](path#anchor)
// - Markdown images: ![alt](path)
// - HTML src attributes: src="path"
// - HTML href attributes: href="path"
func ReplaceLinkPaths(content, old, new string) string {
	escapedOld := regexp.QuoteMeta(old)

	mdLinkPattern := regexp.MustCompile(`(!?)\[([^\]]*)\]\((` + escapedOld + `)(#[^)]*)?\)`)
	content = mdLinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := mdLinkPattern.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		bang := parts[1]
		text := parts[2]
		anchor := ""
		if len(parts) > 4 && parts[4] != "" {
			anchor = parts[4]
		}
		return fmt.Sprintf("%s[%s](%s%s)", bang, text, new, anchor)
	})

	srcPattern := regexp.MustCompile(`(src=")(` + escapedOld + `)(")`)
	content = srcPattern.ReplaceAllString(content, "${1}"+new+"${3}")

	hrefPattern := regexp.MustCompile(`(href=")(` + escapedOld + `)(")`)
	content = hrefPattern.ReplaceAllString(content, "${1}"+new+"${3}")

	return content
}

// ReplaceMarkdownLinks finds markdown links with the specified path and either
// strips them (keeps text only) or replaces the path.
func ReplaceMarkdownLinks(content, linkPath string, stripLink bool) string {
	escapedPath := regexp.QuoteMeta(linkPath)
	pattern := regexp.MustCompile(`\[([^\]]+)\]\(` + escapedPath + `(?:#[^\)]+)?\)`)

	if stripLink {
		return pattern.ReplaceAllString(content, "$1")
	}

	return content
}

// AdjustRelativePaths adjusts all relative paths in content.
// Only adjusts links that go OUTSIDE the content tree (more ../ than fileDepth).
func AdjustRelativePaths(content string, depthChange, fileDepth int) string {
	linkRe := regexp.MustCompile(`(!?\[.+?\]\()(\.\./)+(.*?)(\)\s*(?:\{[^}]*\})?)`)

	return linkRe.ReplaceAllStringFunc(content, func(match string) string {
		submatches := linkRe.FindStringSubmatch(match)
		if len(submatches) < 5 {
			return match
		}

		prefix := submatches[1]

		afterPrefix := match[len(prefix):]
		upCount := 0
		for strings.HasPrefix(afterPrefix, "../") {
			upCount++
			afterPrefix = afterPrefix[3:]
		}

		if upCount <= fileDepth {
			return match
		}

		remaining := afterPrefix

		newUpCount := upCount - depthChange
		if newUpCount < 0 {
			newUpCount = 0
		}

		var result strings.Builder
		result.WriteString(prefix)
		for i := 0; i < newUpCount; i++ {
			result.WriteString("../")
		}
		result.WriteString(remaining)

		return result.String()
	})
}

// ExtractSourcePrefix extracts the fixed directory prefix from a glob pattern.
func ExtractSourcePrefix(pattern string) string {
	pattern = strings.ReplaceAll(pattern, "\\", "/")

	globStart := strings.IndexAny(pattern, "*?[{")
	if globStart == -1 {
		dir := pattern
		if idx := strings.LastIndex(dir, "/"); idx >= 0 {
			return dir[:idx+1]
		}
		return ""
	}

	prefix := pattern[:globStart]
	lastSlash := strings.LastIndex(prefix, "/")
	if lastSlash == -1 {
		return ""
	}

	return prefix[:lastSlash+1]
}

// CountPathDepth counts the number of directory components in a path.
func CountPathDepth(path string) int {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.Trim(path, "/")
	if path == "" || path == "." {
		return 0
	}
	return strings.Count(path, "/") + 1
}
