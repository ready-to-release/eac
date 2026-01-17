package books

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
		relFileDir, relErr := filepath.Rel(p.stagingDir, fileDir)
		if relErr != nil {
			relFileDir = fileDir // fallback to absolute
		}

		original := string(content)
		modified, count := fixLinksInContent(original, relFileDir, p.stagingDir, siteURL)

		if count > 0 {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
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
// Captures: [1]=preceding char or empty, [2]=link text, [3]=path.
var linkWithContextPattern = regexp.MustCompile(`(^|[^!])\[([^\]]+)\]\(([^)]+)\)`)

// fixLinksInContent processes a single file's content and fixes broken links.
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

// ============================================================================
// Link Translation Architecture
// ============================================================================
//
// This section implements comprehensive path translation for source → staging.
// When copying markdown from source to staging with different directory structures,
// relative paths must be recalculated to point to the correct locations.
//
// Flow:
// 1. During copy, track source → staging file mappings
// 2. Parse source markdown to extract all relative references
// 3. Resolve references to absolute paths from source directory
// 4. Calculate new relative paths from staging directory
// 5. Build translation map: old_path → new_path
// 6. Apply translations to all staging files
//
// This fixes:
// - Image references: ![](../../assets/diagram.png)
// - Markdown links: [Other](../other.md)
// - HTML img tags: <img src="../image.png">
// - Mermaid SVG paths: dynamically calculated during rendering

// LinkTranslation holds path translations for a single file
