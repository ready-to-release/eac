package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
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
// Note: Only matches single-brace attr_list { }, NOT Jinja2 macros {{ }}
// The negative character class [^{}] ensures we don't match nested/double braces.
var imagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)(\s*\{[^{}\n]*\})?`)

// imageWithAttrsPattern matches markdown images with attr_list: ![alt](path){attrs}
// Supports: {width=100}, {width="100"}, {width="100px"}, {: width="100" }.
var imageWithAttrsPattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)\s*\{:?\s*([^}]+)\}`)

// convertAttrListImagesToHTML converts markdown images with attr_list syntax to HTML img tags
// Input:  ![alt](image.png){width=100}
// Output: <img src="image.png" width="100" alt="alt">
//
// This ensures compatibility with:
// - GitHub Pages (doesn't support attr_list)
// - MkDocs HTML rendering
// - PDF generation via Playwright.
func (p *Preprocessor) convertAttrListImagesToHTML() error {
	p.log("    Converting attr_list images to HTML...")

	processed := 0
	converted := 0

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
		modified, count := convertAttrListImages(original, false) // Don't adjust paths - link translator handles it

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			converted += count
		}
		processed++
		return nil
	})
	if err != nil {
		return err
	}

	p.log("    Processed %d files, converted %d images to HTML", processed, converted)
	return nil
}

// convertAttrListImages converts ![alt](path){attrs} to <img> tags
// If adjustPaths is true, prepends ../ to relative paths for MkDocs compatibility
// (MkDocs converts file.md to file/index.html, changing relative path depth)
// NOTE: When using LinkTranslator for path management, set adjustPaths=false
// as the translator handles all path adjustments based on source→staging mapping
// NOTE: Raw .drawio files are NOT converted - they need markdown syntax for the drawio plugin.
func convertAttrListImages(content string, adjustPaths bool) (string, int) {
	count := 0

	result := imageWithAttrsPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := imageWithAttrsPattern.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		alt := parts[1]
		src := parts[2]
		attrs := parts[3]

		// Skip raw .drawio files - they need markdown syntax for the MkDocs drawio plugin
		// Note: .drawio.png files ARE images and should be converted
		if strings.HasSuffix(src, ".drawio") {
			return match
		}

		// Parse attributes from attr_list format
		htmlAttrs := parseAttrListToHTML(attrs)
		if htmlAttrs == "" {
			return match // No recognizable attributes, keep original
		}

		// Optionally adjust paths for MkDocs (file.md -> file/index.html changes depth)
		adjustedSrc := src
		if adjustPaths && !strings.HasPrefix(src, "/") && !strings.HasPrefix(src, "http") {
			adjustedSrc = "../" + src
		}

		count++
		return fmt.Sprintf(`<img src="%s" %s alt="%s">`, adjustedSrc, htmlAttrs, alt)
	})

	return result, count
}

// parseAttrListToHTML converts attr_list attributes to HTML attributes
// Handles: width=100, width="100", width="100px", height=50, style="...".
func parseAttrListToHTML(attrs string) string {
	var htmlParts []string

	// Pattern for key=value or key="value"
	attrPattern := regexp.MustCompile(`(\w+)\s*=\s*"?([^"\s}]+)"?`)
	matches := attrPattern.FindAllStringSubmatch(attrs, -1)

	for _, m := range matches {
		if len(m) >= 3 {
			key := strings.ToLower(m[1])
			value := m[2]

			// Only include recognized HTML img attributes
			switch key {
			case "width", "height", "style", "class", "id":
				htmlParts = append(htmlParts, fmt.Sprintf(`%s="%s"`, key, value))
			}
		}
	}

	return strings.Join(htmlParts, " ")
}

// addImageWidthConstraints adds width="100%" to images that don't have explicit dimensions
// This ensures large diagrams (especially drawio exports) scale properly in PDFs.
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

		// Skip raw .drawio files - they're handled by the drawio plugin which doesn't support attr_list
		// Note: .drawio.png files ARE images and can have width constraints
		if strings.HasSuffix(path, ".drawio") {
			return match
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

// convertDrawioToLinks converts .drawio image references to GitHub Pages links
// Since .drawio files are interactive diagrams that can't display in PDFs,
// we convert them to clickable links to the online version
