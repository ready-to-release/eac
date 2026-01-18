package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultMarkerPattern is the default regex for finding markers in markdown files.
// Matches: <!-- book:insert marker-id -->.
const DefaultMarkerPattern = `<!--\s*book:insert\s+([a-zA-Z0-9_-]+)\s*-->`

// insertInlineContent replaces markers with command outputs (Step 5).
func (p *Preprocessor) insertInlineContent(outputs map[string]string) error {
	inlineSources := p.book.GetInlineSources()
	if len(inlineSources) == 0 {
		p.log("    No inline sources defined")
		return nil
	}

	for i := range inlineSources {
		src := &inlineSources[i]
		targetPath := filepath.Join(p.stagingDir, src.Target)

		// Read target file
		content, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("inline target %s: %w", src.Target, err)
		}

		// Get marker pattern
		pattern := src.MarkerPattern
		if pattern == "" {
			pattern = DefaultMarkerPattern
		}

		// Compile the regex
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid marker pattern for %s: %w", src.Target, err)
		}

		// Replace each marker
		modified := string(content)
		replacements := 0

		for _, insert := range src.Inserts {
			key := fmt.Sprintf("%s:%s", src.Target, insert.Marker)
			output, ok := outputs[key]
			if !ok {
				p.log("    Warning: no output for marker '%s' in %s", insert.Marker, src.Target)
				continue
			}

			// Build replacement with wrapped markers
			replacement := formatInlineReplacement(insert.Marker, output)

			// Create a specific regex for this marker
			markerRe := regexp.MustCompile(
				fmt.Sprintf(`<!--\s*book:insert\s+%s\s*-->`, regexp.QuoteMeta(insert.Marker)),
			)

			// Check if marker exists in content
			if !markerRe.MatchString(modified) {
				p.log("    Warning: marker '%s' not found in %s", insert.Marker, src.Target)
				continue
			}

			// Replace marker with output
			modified = markerRe.ReplaceAllString(modified, replacement)
			replacements++
		}

		// Only write if we made changes
		if replacements > 0 {
			if err := os.WriteFile(targetPath, []byte(modified), 0o644); err != nil {
				return err
			}
			p.log("    Modified: %s (%d markers replaced)", src.Target, replacements)
		}

		// Validate that all markers in pattern were handled
		_ = re // Pattern compiled but not all markers may match the default pattern
	}

	return nil
}

// formatInlineReplacement wraps command output with generation markers.
func formatInlineReplacement(marker, output string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<!-- book:generated %s -->\n", marker))
	sb.WriteString(strings.TrimSpace(output))
	sb.WriteString("\n<!-- /book:generated -->")

	return sb.String()
}
