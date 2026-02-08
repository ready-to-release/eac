package linking

import (
	"regexp"
	"strings"
)

// Link extraction patterns.
var (
	MdImagePattern  = regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)    // ![alt](path)
	MdLinkPattern   = regexp.MustCompile(`\[.*?\]\(([^)]+)\)`)     // [text](path)
	HtmlImgPattern  = regexp.MustCompile(`<img[^>]+src="([^"]+)"`) // <img src="path">
	HtmlLinkPattern = regexp.MustCompile(`<a[^>]+href="([^"]+)"`)  // <a href="path">

	FencedCodeBlockPattern      = regexp.MustCompile("(?s)```.*?```")
	FencedCodeBlockSplitPattern = regexp.MustCompile("(?s)(```.*?```)")
)

// StripCodeBlocks removes both fenced and indented code blocks from markdown content
// to prevent extracting links from code examples.
func StripCodeBlocks(content string) string {
	content = FencedCodeBlockPattern.ReplaceAllString(content, "")

	lines := strings.Split(content, "\n")
	var cleaned []string

	for _, line := range lines {
		if line != "" && (line[0] == '\t' || strings.HasPrefix(line, "    ")) {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// ExtractRelativeLinks extracts all relative link references from markdown content.
func ExtractRelativeLinks(content string) []string {
	seen := make(map[string]bool)
	var links []string

	content = StripCodeBlocks(content)

	patterns := []*regexp.Regexp{MdImagePattern, MdLinkPattern, HtmlImgPattern, HtmlLinkPattern}
	for _, pat := range patterns {
		for _, match := range pat.FindAllStringSubmatch(content, -1) {
			if len(match) >= 2 {
				path := match[1]
				if IsRelativePathRef(path) && !seen[path] {
					links = append(links, path)
					seen[path] = true
				}
			}
		}
	}

	return links
}

// IsRelativePathRef checks if a path is a relative reference (not external or absolute).
func IsRelativePathRef(path string) bool {
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "//") ||
		strings.HasPrefix(path, "mailto:") {
		return false
	}

	if strings.HasPrefix(path, "/") {
		return false
	}

	if strings.HasPrefix(path, "#") {
		return false
	}

	pathPart := path
	if idx := strings.Index(path, "#"); idx >= 0 {
		pathPart = path[:idx]
	}

	if pathPart == "" {
		return false
	}

	return true
}
