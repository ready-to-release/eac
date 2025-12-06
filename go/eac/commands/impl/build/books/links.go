package books

import (
	"fmt"
	"net/url"
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

// imageWithAttrsPattern matches markdown images with attr_list: ![alt](path){attrs}
// Supports: {width=100}, {width="100"}, {width="100px"}, {: width="100" }
var imageWithAttrsPattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)\s*\{:?\s*([^}]+)\}`)

// convertAttrListImagesToHTML converts markdown images with attr_list syntax to HTML img tags
// Input:  ![alt](image.png){width=100}
// Output: <img src="image.png" width="100" alt="alt">
//
// This ensures compatibility with:
// - GitHub Pages (doesn't support attr_list)
// - MkDocs HTML rendering
// - PDF generation via Playwright
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
		modified, count := convertAttrListImages(original, true) // adjustPaths=true for MkDocs

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
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
// Handles: width=100, width="100", width="100px", height=50, style="..."
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

// convertDrawioToLinks converts .drawio image references to GitHub Pages links
// Since .drawio files are interactive diagrams that can't display in PDFs,
// we convert them to clickable links to the online version
func (p *Preprocessor) convertDrawioToLinks() error {
	if p.book.SiteURL == "" {
		p.log("    Skipping drawio conversion (no site_url configured)")
		return nil
	}

	p.log("    Converting .drawio images to links...")

	siteURL := p.book.SiteURL
	if !strings.HasSuffix(siteURL, "/") {
		siteURL += "/"
	}

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

		fileDir := filepath.Dir(path)
		relFileDir, _ := filepath.Rel(p.stagingDir, fileDir)

		original := string(content)
		modified, count := convertDrawioImages(original, relFileDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
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

	p.log("    Processed %d files, converted %d .drawio images to links", processed, converted)
	return nil
}

// drawioImagePattern matches markdown images pointing to .drawio files
var drawioImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]*\.drawio)\)`)

// convertDrawioImages converts ![alt](path.drawio) to [View interactive diagram](url)
func convertDrawioImages(content, relFileDir, siteURL string) (string, int) {
	count := 0

	result := drawioImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := drawioImagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		// alt := parts[1] // Not used in link text
		imgPath := parts[2]

		// Build URL path
		var urlPath string
		if strings.HasPrefix(imgPath, "/") {
			urlPath = imgPath[1:]
		} else {
			urlPath = filepath.ToSlash(filepath.Join(relFileDir, imgPath))
		}
		urlPath = filepath.ToSlash(filepath.Clean(urlPath))

		// Handle paths that go outside staging
		if strings.HasPrefix(urlPath, "../") {
			// For explanation book, assets are at root level
			// ../assets/x -> assets/x
			urlPath = strings.TrimPrefix(urlPath, "../")
		}

		fullURL, err := url.JoinPath(siteURL, urlPath)
		if err != nil {
			return match
		}

		count++
		return fmt.Sprintf("[View interactive diagram](%s)", fullURL)
	})

	return result, count
}

// fixBrokenInternalLinks converts broken internal links to GitHub Pages URLs
// This is needed because copy-in may not include all linked files, breaking relative links.
//
// For each markdown link [text](path):
//  1. If path is external (http/https/mailto), skip
//  2. If path is an anchor-only link (#section), skip
//  3. Resolve the path relative to the current file
//  4. Check if the target exists in staging
//  5. If not, convert to absolute GitHub Pages URL
func (p *Preprocessor) fixBrokenInternalLinks() error {
	if p.book.SiteURL == "" {
		p.log("    Skipping link fix (no site_url configured)")
		return nil
	}

	p.log("    Fixing broken internal links...")

	// Ensure site URL ends with /
	siteURL := p.book.SiteURL
	if !strings.HasSuffix(siteURL, "/") {
		siteURL += "/"
	}

	processed := 0
	linksFixed := 0

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

		// Get the directory of the current file for relative path resolution
		fileDir := filepath.Dir(path)
		relFileDir, _ := filepath.Rel(p.stagingDir, fileDir)

		original := string(content)
		modified, count := fixLinksInContent(original, relFileDir, p.stagingDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			linksFixed += count
		}
		processed++
		return nil
	})

	if err != nil {
		return err
	}

	p.log("    Processed %d files, fixed %d broken links", processed, linksFixed)
	return nil
}

// linkWithContextPattern matches markdown links with optional preceding character
// This allows us to detect and skip image links (preceded by !)
// Captures: [1]=preceding char or empty, [2]=link text, [3]=path
var linkWithContextPattern = regexp.MustCompile(`(^|[^!])\[([^\]]+)\]\(([^)]+)\)`)

// fixLinksInContent processes a single file's content and fixes broken links
func fixLinksInContent(content, relFileDir, stagingDir, siteURL string) (string, int) {
	count := 0

	result := linkWithContextPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := linkWithContextPattern.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		prefix := parts[1] // Character before [ or empty
		text := parts[2]
		linkPath := parts[3]

		// Skip external links
		if strings.HasPrefix(linkPath, "http://") ||
			strings.HasPrefix(linkPath, "https://") ||
			strings.HasPrefix(linkPath, "mailto:") ||
			strings.HasPrefix(linkPath, "//") {
			return match
		}

		// Skip anchor-only links
		if strings.HasPrefix(linkPath, "#") {
			return match
		}

		// Split path and anchor
		pathPart := linkPath
		anchor := ""
		if anchorIdx := strings.Index(linkPath, "#"); anchorIdx >= 0 {
			pathPart = linkPath[:anchorIdx]
			anchor = linkPath[anchorIdx:]
		}

		// Resolve the target path
		var targetPath string
		if strings.HasPrefix(pathPart, "/") {
			// Absolute path from docs root
			targetPath = filepath.Join(stagingDir, pathPart)
		} else {
			// Relative path from current file
			targetPath = filepath.Join(stagingDir, relFileDir, pathPart)
		}

		// Clean the path
		targetPath = filepath.Clean(targetPath)

		// Check if target is a non-markdown file (like .drawio, .pdf source files)
		// These should be converted to GitHub Pages URLs since they won't work in PDFs
		ext := strings.ToLower(filepath.Ext(pathPart))
		isNonMarkdownFile := ext != "" && ext != ".md"

		// Check if target exists in staging (only keep if it's a markdown file or directory)
		if !isNonMarkdownFile {
			if _, err := os.Stat(targetPath); err == nil {
				// Target exists, keep the link as-is
				return match
			}

			// Also check with .md extension if not already present
			if !strings.HasSuffix(targetPath, ".md") {
				if _, err := os.Stat(targetPath + ".md"); err == nil {
					return match
				}
				// Check for index.md in directory
				if _, err := os.Stat(filepath.Join(targetPath, "index.md")); err == nil {
					return match
				}
			}
		}

		// Target doesn't exist - convert to GitHub Pages URL
		// Convert path to URL format
		urlPath := pathPart
		if strings.HasPrefix(urlPath, "/") {
			urlPath = urlPath[1:] // Remove leading /
		} else {
			// Relative path - make it relative to current location in site
			urlPath = filepath.ToSlash(filepath.Join(relFileDir, pathPart))
		}

		// Clean up the URL path
		urlPath = filepath.ToSlash(filepath.Clean(urlPath))
		if strings.HasPrefix(urlPath, "../") {
			// Path goes outside staging - use just the basename
			urlPath = filepath.Base(pathPart)
		}

		// Remove .md extension for GitHub Pages (MkDocs serves without extension)
		urlPath = strings.TrimSuffix(urlPath, ".md")

		// Build the full URL
		fullURL, err := url.JoinPath(siteURL, urlPath)
		if err != nil {
			return match
		}

		count++
		// Preserve the prefix character (could be newline, space, etc.)
		return fmt.Sprintf("%s[%s](%s%s)", prefix, text, fullURL, anchor)
	})

	return result, count
}
