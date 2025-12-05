package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Size presets for mermaid diagrams
var mermaidSizePresets = map[string]string{
	"small":  "33%",
	"medium": "50%",
	"large":  "66%",
	"full":   "100%",
}

// mermaidBlockPattern matches mermaid code blocks with optional size directive
// Captures: (1) size directive value, (2) mermaid content
var mermaidBlockPattern = regexp.MustCompile("(?s)```mermaid\\s*\n%%\\{(?:size|width):([^}]+)\\}%%\\s*\n(.*?)```")

// mermaidBlockPlain matches plain mermaid blocks without size directive
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// processMermaidSizing wraps mermaid blocks with size directives in container divs
// This enables CSS-based sizing for both web (mermaid2) and PDF (mermaid-to-svg)
//
// Syntax in markdown:
//
//	```mermaid
//	%%{size:medium}%%
//	flowchart TD
//	    A --> B
//	```
//
// Or with explicit width:
//
//	```mermaid
//	%%{width:40%}%%
//	flowchart TD
//	    A --> B
//	```
//
// Size presets: small (33%), medium (50%), large (66%), full (100%)
func (p *Preprocessor) processMermaidSizing() error {
	p.log("    Processing mermaid diagram sizing...")

	processed := 0
	wrapped := 0

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := wrapMermaidBlocks(original)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			wrapped++
		}
		processed++
		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Processed %d files, wrapped %d mermaid blocks with sizing", processed, wrapped)
	return nil
}

// wrapMermaidBlocks finds mermaid blocks with size directives and wraps them
func wrapMermaidBlocks(content string) string {
	// Process blocks with size directives
	result := mermaidBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract size value and content
		submatches := mermaidBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		sizeValue := strings.TrimSpace(submatches[1])
		mermaidContent := submatches[2]

		// Resolve preset name to percentage
		width := sizeValue
		if preset, ok := mermaidSizePresets[strings.ToLower(sizeValue)]; ok {
			width = preset
		}

		// Ensure width ends with %
		if !strings.HasSuffix(width, "%") && !strings.HasSuffix(width, "px") {
			width = width + "%"
		}

		// Build wrapped block
		// Use both data-size attribute (for CSS) and inline style (for PDF)
		var wrapper strings.Builder
		wrapper.WriteString("<div class=\"mermaid-wrapper\" data-size=\"")
		wrapper.WriteString(strings.ToLower(sizeValue))
		wrapper.WriteString("\" style=\"max-width:")
		wrapper.WriteString(width)
		wrapper.WriteString("; margin: 0 auto;\">\n\n")
		wrapper.WriteString("```mermaid\n")
		wrapper.WriteString(mermaidContent)
		wrapper.WriteString("```\n\n</div>")

		return wrapper.String()
	})

	return result
}

// countMermaidBlocks returns the number of mermaid blocks in content (for logging)
func countMermaidBlocks(content string) int {
	return len(mermaidBlockPlain.FindAllString(content, -1))
}
