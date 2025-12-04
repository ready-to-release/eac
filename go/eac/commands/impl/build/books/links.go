package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cleanupLinksForPDF processes all markdown files in staging for PDF compatibility (Step 6)
//
// This step:
// 1. Adds width constraints to images for proper PDF rendering
// 2. Ensures large diagrams (like drawio exports) fit within PDF page boundaries
//
// The mkdocs-with-pdf plugin generates anchor IDs in the format:
//   - Page anchor: path/to/page/:
//   - Heading anchor: path/to/page/:heading-slug
func (p *Preprocessor) cleanupLinksForPDF() error {
	p.log("    Processing images for PDF compatibility...")

	processed := 0
	imagesFixed := 0

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
		modified, count := addImageWidthConstraints(original)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			imagesFixed += count
		}
		processed++
		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Processed %d files, added width constraints to %d images", processed, imagesFixed)
	return nil
}

// imagePattern matches markdown images: ![alt](path)
// Captures: [1]=alt text, [2]=path, [3]=optional existing attributes
var imagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)(\s*\{[^}]*\})?`)

// addImageWidthConstraints adds width="100%" to images that don't have explicit dimensions
// This ensures large diagrams (especially drawio exports) scale properly in PDFs
func addImageWidthConstraints(content string) (string, int) {
	count := 0

	result := imagePattern.ReplaceAllStringFunc(content, func(match string) string {
		// Skip if already has attributes with width/style
		if strings.Contains(match, "width=") || strings.Contains(match, "style=") {
			return match
		}

		// Extract the parts
		parts := imagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		alt := parts[1]
		path := parts[2]
		existingAttrs := ""
		if len(parts) > 3 {
			existingAttrs = parts[3]
		}

		// Skip small inline images (icons, badges, emojis)
		lowerPath := strings.ToLower(path)
		if strings.Contains(lowerPath, "icon") ||
			strings.Contains(lowerPath, "badge") ||
			strings.Contains(lowerPath, "emoji") ||
			strings.Contains(lowerPath, "favicon") {
			return match
		}

		// Add width constraint for diagrams and large images
		count++
		if existingAttrs != "" {
			// Has existing attributes, add width to them
			// { existing } -> { existing width="100%" }
			newAttrs := strings.TrimSuffix(strings.TrimSpace(existingAttrs), "}")
			return "![" + alt + "](" + path + ")" + newAttrs + ` width="100%" }`
		}

		// No existing attributes, add new ones
		return "![" + alt + "](" + path + `){ width="100%" }`
	})

	return result, count
}
