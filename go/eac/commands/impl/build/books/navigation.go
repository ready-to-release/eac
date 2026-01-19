package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var navLog = logging.C()

// NavFile represents a .nav.yml file structure.
type NavFile struct {
	Title string `yaml:"title,omitempty"`
	Nav   []any  `yaml:"nav"`
}

// getTitleFromFile extracts title from markdown frontmatter or first heading.
func (p *Preprocessor) getTitleFromFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return filenameToTitle(filepath.Base(path))
	}

	text := string(content)

	// Try frontmatter title
	if strings.HasPrefix(text, "---") {
		endIdx := strings.Index(text[3:], "---")
		if endIdx > 0 {
			frontmatter := text[3 : 3+endIdx]
			titleMatch := regexp.MustCompile(`(?m)^title:\s*["']?([^"'\n]+)["']?`).FindStringSubmatch(frontmatter)
			if len(titleMatch) > 1 {
				return strings.TrimSpace(titleMatch[1])
			}
		}
	}

	// Try first H1
	h1Match := regexp.MustCompile(`(?m)^#\s+(.+)$`).FindStringSubmatch(text)
	if len(h1Match) > 1 {
		return strings.TrimSpace(h1Match[1])
	}

	return filenameToTitle(filepath.Base(path))
}

// getOrderForFile returns the order for a file from command sources.
func (p *Preprocessor) getOrderForFile(relPath string) int {
	for _, src := range p.book.Sources {
		if src.Type == "command" && filepath.ToSlash(src.Target) == relPath {
			if src.Order > 0 {
				return src.Order
			}
		}
	}
	return 500 // Default order for non-command files
}

// filenameToTitle converts a filename to a readable title.
func filenameToTitle(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Replace dashes and underscores with spaces
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return toTitleCase(name)
}

// toTitleCase converts a string to title case.
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
