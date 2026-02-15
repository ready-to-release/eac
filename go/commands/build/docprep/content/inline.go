package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/config"
)

// DefaultMarkerPattern is the default regex for finding markers in markdown files.
const DefaultMarkerPattern = `<!--\s*book:insert\s+([a-zA-Z0-9_-]+)\s*-->`

// InsertInlineContent replaces markers with command outputs.
func InsertInlineContent(
	book *config.Book,
	stagingDir string,
	outputs map[string]string,
	logf func(string, ...any),
	warnf func(string, ...any),
) error {
	inlineSources := book.GetInlineSources()
	if len(inlineSources) == 0 {
		logf("    No inline sources defined")
		return nil
	}

	for i := range inlineSources {
		src := &inlineSources[i]
		targetPath := filepath.Join(stagingDir, src.Target)

		content, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("inline target %s: %w", src.Target, err)
		}

		pattern := src.MarkerPattern
		if pattern == "" {
			pattern = DefaultMarkerPattern
		}

		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid marker pattern for %s: %w", src.Target, err)
		}

		modified := string(content)
		replacements := 0

		for _, insert := range src.Inserts {
			key := fmt.Sprintf("%s:%s", src.Target, insert.Marker)
			output, ok := outputs[key]
			if !ok {
				warnf("no output for marker '%s' in %s", insert.Marker, src.Target)
				continue
			}

			replacement := FormatInlineReplacement(insert.Marker, output)

			markerRe := regexp.MustCompile(
				fmt.Sprintf(`<!--\s*book:insert\s+%s\s*-->`, regexp.QuoteMeta(insert.Marker)),
			)

			if !markerRe.MatchString(modified) {
				warnf("marker '%s' not found in %s", insert.Marker, src.Target)
				continue
			}

			modified = markerRe.ReplaceAllString(modified, replacement)
			replacements++
		}

		if replacements > 0 {
			if err := os.WriteFile(targetPath, []byte(modified), 0o644); err != nil {
				return err
			}
			logf("    Modified: %s (%d markers replaced)", src.Target, replacements)
		}
	}

	return nil
}

// FormatInlineReplacement wraps command output with generation markers.
func FormatInlineReplacement(marker, output string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<!-- book:generated %s -->\n", marker))
	sb.WriteString(strings.TrimSpace(output))
	sb.WriteString("\n<!-- /book:generated -->")
	return sb.String()
}
