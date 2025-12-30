package books

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/paths"
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
// Note: Only matches single-brace attr_list { }, NOT Jinja2 macros {{ }}
// The negative character class [^{}] ensures we don't match nested/double braces
var imagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)(\s*\{[^{}\n]*\})?`)

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
		modified, count := convertAttrListImages(original, false) // Don't adjust paths - link translator handles it

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
// NOTE: When using LinkTranslator for path management, set adjustPaths=false
// as the translator handles all path adjustments based on source→staging mapping
// NOTE: Raw .drawio files are NOT converted - they need markdown syntax for the drawio plugin
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
		// For explanation book, assets are at root level
		// ../assets/x -> assets/x
		urlPath = strings.TrimPrefix(urlPath, "../")

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
type LinkTranslation struct {
	sourceFile   string            // Original source file path (absolute)
	stagingFile  string            // Staging file path (absolute)
	translations map[string]string // old_path → new_path
}

// LinkTranslator manages translations for all files
type LinkTranslator struct {
	sourceRoot    string                         // Source root directory (docs/)
	stagingRoot   string                         // Staging root directory (out/staging/book/)
	fileMap       map[string]string              // staging_path → source_path (absolute)
	docsRelPath   map[string]string              // source_path (absolute) → docs-relative path
	translations  map[string]*LinkTranslation    // staging_path → translation
	logWriter     io.Writer                      // Logger for debug output
	pdfMode       bool                           // True if building PDF (strips external links)
}

// NewLinkTranslator creates a new link translator
func NewLinkTranslator(sourceRoot, stagingRoot string, logWriter io.Writer, pdfMode bool) *LinkTranslator {
	return &LinkTranslator{
		sourceRoot:   filepath.Clean(sourceRoot),
		stagingRoot:  filepath.Clean(stagingRoot),
		fileMap:      make(map[string]string),
		docsRelPath:  make(map[string]string),
		translations: make(map[string]*LinkTranslation),
		logWriter:    logWriter,
		pdfMode:      pdfMode,
	}
}

// AddFileMapping tracks a source → staging file mapping
func (t *LinkTranslator) AddFileMapping(stagingPath, sourcePath string) {
	t.fileMap[stagingPath] = sourcePath

	// Calculate docs-relative path for external link generation
	// sourcePath should be under sourceRoot/docs/
	docsDir := paths.DocsSourcePath(t.sourceRoot)
	if relPath, err := filepath.Rel(docsDir, sourcePath); err == nil {
		t.docsRelPath[sourcePath] = filepath.ToSlash(relPath)
	}
}

// logDebug writes a debug message if logger is available
func (t *LinkTranslator) logDebug(format string, args ...any) {
	if t.logWriter != nil {
		fmt.Fprintf(t.logWriter, "    [LinkTranslator] "+format+"\n", args...)
	}
}

// relativeLinkPattern matches markdown images, links, and HTML img/anchor tags
// Captures relative paths (not http://, https://, /, or #)
var (
	mdImagePattern  = regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)       // ![alt](path)
	mdLinkPattern   = regexp.MustCompile(`\[.*?\]\(([^)]+)\)`)        // [text](path)
	htmlImgPattern  = regexp.MustCompile(`<img[^>]+src="([^"]+)"`)    // <img src="path">
	htmlLinkPattern = regexp.MustCompile(`<a[^>]+href="([^"]+)"`)     // <a href="path">
)

// stripCodeBlocks removes both fenced and indented code blocks from markdown content
// to prevent extracting links from code examples
func stripCodeBlocks(content string) string {
	// Remove fenced code blocks (```...```)
	// Use (?s) flag to make . match newlines
	fencedCodeBlockPattern := regexp.MustCompile("(?s)```.*?```")
	content = fencedCodeBlockPattern.ReplaceAllString(content, "")

	// Remove indented code blocks (lines starting with 4 spaces or tab)
	lines := strings.Split(content, "\n")
	var cleaned []string

	for _, line := range lines {
		// Skip lines that start with 4+ spaces or tab (indented code blocks)
		if len(line) > 0 && (line[0] == '\t' || strings.HasPrefix(line, "    ")) {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}


// extractRelativeLinks extracts all relative link references from markdown content
func extractRelativeLinks(content string) []string {
	seen := make(map[string]bool)
	var links []string

	// Strip code blocks (both fenced and indented) to avoid extracting links from examples
	content = stripCodeBlocks(content)

	// Extract from markdown images
	for _, match := range mdImagePattern.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			path := match[1]
			if isRelativePathRef(path) && !seen[path] {
				links = append(links, path)
				seen[path] = true
			}
		}
	}

	// Extract from markdown links (but skip if already captured as image)
	for _, match := range mdLinkPattern.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			path := match[1]
			if isRelativePathRef(path) && !seen[path] {
				links = append(links, path)
				seen[path] = true
			}
		}
	}

	// Extract from HTML img tags
	for _, match := range htmlImgPattern.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			path := match[1]
			if isRelativePathRef(path) && !seen[path] {
				links = append(links, path)
				seen[path] = true
			}
		}
	}

	// Extract from HTML anchor tags
	for _, match := range htmlLinkPattern.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			path := match[1]
			if isRelativePathRef(path) && !seen[path] {
				links = append(links, path)
				seen[path] = true
			}
		}
	}

	return links
}

// isRelativePathRef checks if a path is a relative reference (not external or absolute)
func isRelativePathRef(path string) bool {
	// Skip external URLs
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "//") ||
		strings.HasPrefix(path, "mailto:") {
		return false
	}

	// Skip absolute paths (start with /)
	if strings.HasPrefix(path, "/") {
		return false
	}

	// Skip anchor-only links (start with #)
	if strings.HasPrefix(path, "#") {
		return false
	}

	// Strip anchor from path for checking
	pathPart := path
	if idx := strings.Index(path, "#"); idx >= 0 {
		pathPart = path[:idx]
	}

	// Must have a path part (not just anchor)
	if pathPart == "" {
		return false
	}

	return true
}

// BuildTranslations analyzes all source files and builds translation maps
func (t *LinkTranslator) BuildTranslations(siteURL string) error {
	// Build reverse map: source file → staging file
	sourceToStaging := make(map[string]string)
	for stagingPath, sourcePath := range t.fileMap {
		sourceToStaging[sourcePath] = stagingPath
	}

	for stagingFile, sourceFile := range t.fileMap {
		// Only process markdown files for link translation
		if !strings.HasSuffix(strings.ToLower(sourceFile), ".md") {
			continue
		}

		// Read source file content
		content, err := os.ReadFile(sourceFile)
		if err != nil {
			return fmt.Errorf("reading source file %s: %w", sourceFile, err)
		}

		// Extract all relative links from source
		links := extractRelativeLinks(string(content))

		// Create translation entry
		trans := &LinkTranslation{
			sourceFile:   sourceFile,
			stagingFile:  stagingFile,
			translations: make(map[string]string),
		}

		// Calculate translations for each link
		sourceDir := filepath.Dir(sourceFile)
		stagingDir := filepath.Dir(stagingFile)

		for _, oldLink := range links {
			// Skip template variables (e.g. {{.Values.Docs.URL}})
			if strings.Contains(oldLink, "{{") && strings.Contains(oldLink, "}}") {
				t.logDebug("Skipping template variable link: %s", oldLink)
				continue
			}

			// Strip anchor from path
			pathPart := oldLink
			anchor := ""
			if idx := strings.Index(oldLink, "#"); idx >= 0 {
				pathPart = oldLink[:idx]
				anchor = oldLink[idx:]
			}

			// Resolve to absolute path from source directory
			absSourcePath := filepath.Join(sourceDir, pathPart)
			absSourcePath = filepath.Clean(absSourcePath)

			var newLink string

			// Check if target file exists in our source → staging map
			targetStagingFile, existsInStaging := sourceToStaging[absSourcePath]

			if !existsInStaging {
				// File not in staging - check if it exists in source docs/
				// If it exists in docs/, handle as external reference
				// If it doesn't exist anywhere, FAIL FAST (broken link)

				docsDir := paths.DocsSourcePath(t.sourceRoot)

				// Check if target exists: either as file, directory, or directory with index.md
				targetExists := false
				if info, err := os.Stat(absSourcePath); err == nil {
					targetExists = true
					// If it's a directory, also check for index.md
					if info.IsDir() {
						indexPath := filepath.Join(absSourcePath, "index.md")
						if _, err := os.Stat(indexPath); err != nil {
							// Directory exists but no index.md - still valid for web links
							targetExists = true
						}
					}
				} else {
					// Path doesn't exist as-is, try with .md extension
					if _, err := os.Stat(absSourcePath + ".md"); err == nil {
						targetExists = true
					} else {
						// Try as directory with index.md
						indexPath := filepath.Join(absSourcePath, "index.md")
						if _, err := os.Stat(indexPath); err == nil {
							targetExists = true
						}
					}
				}

				if targetExists && strings.HasPrefix(absSourcePath, docsDir) {
					// File exists in docs/ but wasn't copied
					if t.pdfMode {
						// PDF mode: strip the link (keep text, remove href)
						// Use special marker that will be replaced during apply
						newLink = "##EXTERNAL_STRIPPED##" + oldLink
						t.logDebug("External link stripped (PDF mode): %s", oldLink)
					} else {
						// Web mode: convert to external GitHub Pages URL
						if siteURL == "" {
							return fmt.Errorf("link '%s' in %s points to file not in staging, but no site_url configured for external links",
								oldLink, sourceFile)
						}

						// Calculate docs-relative path
						relPath, err := filepath.Rel(docsDir, absSourcePath)
						if err != nil {
							return fmt.Errorf("calculating docs-relative path for %s: %w", absSourcePath, err)
						}

						// Remove .md extension and build URL
						relPath = filepath.ToSlash(relPath)
						relPath = strings.TrimSuffix(relPath, ".md")

						baseURL := siteURL
						if !strings.HasSuffix(baseURL, "/") {
							baseURL += "/"
						}
						newLink = baseURL + relPath + anchor

						t.logDebug("External link: %s -> %s", oldLink, newLink)
					}
				} else {
					// File doesn't exist in source - FAIL FAST
					return fmt.Errorf("BROKEN LINK in %s: '%s' points to '%s' which doesn't exist in source",
						sourceFile, oldLink, absSourcePath)
				}
			} else {
				// File exists in staging - calculate new relative path in staging structure
				relPath, err := filepath.Rel(stagingDir, targetStagingFile)
				if err != nil {
					return fmt.Errorf("calculating relative path from %s to %s: %w",
						stagingDir, targetStagingFile, err)
				}
				newLink = filepath.ToSlash(relPath) + anchor

				// Log the translation for debugging
				if newLink != oldLink {
					t.logDebug("Translate: %s -> %s", oldLink, newLink)
				}
			}

			// Only add to map if path actually changed
			if newLink != oldLink {
				trans.translations[oldLink] = newLink
			}
		}

		// Only store translation if there are actual changes
		if len(trans.translations) > 0 {
			t.translations[stagingFile] = trans
		}
	}

	return nil
}

// ApplyAllTranslations applies path translations to all staging files
func (t *LinkTranslator) ApplyAllTranslations() error {
	for stagingFile, trans := range t.translations {
		if err := t.applyTranslation(stagingFile, trans); err != nil {
			return fmt.Errorf("applying translations to %s: %w", stagingFile, err)
		}
	}
	return nil
}

// applyTranslation applies translations to a single staging file
func (t *LinkTranslator) applyTranslation(stagingFile string, trans *LinkTranslation) error {
	// Read staging file
	content, err := os.ReadFile(stagingFile)
	if err != nil {
		return fmt.Errorf("reading staging file: %w", err)
	}

	modified := string(content)

	// Apply all translations
	// Sort by length (longest first) to avoid partial replacements
	type pathPair struct {
		old string
		new string
	}
	var pairs []pathPair
	for old, new := range trans.translations {
		pairs = append(pairs, pathPair{old: old, new: new})
	}

	// Sort by old path length (descending) to replace longer paths first
	// This prevents "../../assets/x.png" from being partially replaced when
	// "../assets/x.png" is also in the map
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if len(pairs[j].old) > len(pairs[i].old) {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Apply replacements only outside code blocks
	for _, pair := range pairs {
		// Check if this is a stripped external link (PDF mode)
		if strings.HasPrefix(pair.new, "##EXTERNAL_STRIPPED##") {
			// Find and replace all markdown links with this path
			// Pattern: [text](path) or [text](path#anchor)
			// We need to extract the text and keep it, but remove the link
			modified = replaceMarkdownLinks(modified, pair.old, true)
		} else {
			// Normal path replacement - but skip code blocks
			modified = replaceOutsideCodeBlocks(modified, pair.old, pair.new)
		}
	}

	// Write back if changed
	if modified != string(content) {
		if err := os.WriteFile(stagingFile, []byte(modified), 0644); err != nil {
			return fmt.Errorf("writing staging file: %w", err)
		}
	}

	return nil
}

// replaceOutsideCodeBlocks replaces old with new, but only outside fenced code blocks
func replaceOutsideCodeBlocks(content, old, new string) string {
	// Pattern to match fenced code blocks: ```...```
	// Uses (?s) to make . match newlines, .*? for non-greedy matching
	codeBlockPattern := regexp.MustCompile("(?s)(```.*?```)")

	// Split content by code blocks, keeping the code blocks as separators
	parts := codeBlockPattern.Split(content, -1)
	codeBlocks := codeBlockPattern.FindAllString(content, -1)

	// Apply replacements only to non-code-block parts
	var result strings.Builder
	for i, part := range parts {
		// Replace in this non-code part
		result.WriteString(strings.ReplaceAll(part, old, new))

		// Add back the code block (unchanged) if there is one
		if i < len(codeBlocks) {
			result.WriteString(codeBlocks[i])
		}
	}

	return result.String()
}

// replaceMarkdownLinks finds markdown links with the specified path and either
// strips them (keeps text only) or replaces the path
func replaceMarkdownLinks(content, linkPath string, stripLink bool) string {
	// Escape special regex characters in the path
	escapedPath := regexp.QuoteMeta(linkPath)

	// Match markdown links: [text](path) or [text](path#anchor)
	// Capture the link text in group 1
	pattern := regexp.MustCompile(`\[([^\]]+)\]\(` + escapedPath + `(?:#[^\)]+)?\)`)

	if stripLink {
		// Replace with just the text (no link)
		return pattern.ReplaceAllString(content, "$1")
	}

	return content
}

// CalculateRelativePath calculates the relative path from a staging markdown file
// to a target absolute path. This is used for dynamically generated content
// (like mermaid SVG files) that aren't in the source → staging translation map.
func (t *LinkTranslator) CalculateRelativePath(stagingFile, targetAbsPath string) (string, error) {
	// Ensure paths are absolute
	absStaging, err := filepath.Abs(stagingFile)
	if err != nil {
		return "", fmt.Errorf("making staging path absolute: %w", err)
	}

	absTarget, err := filepath.Abs(targetAbsPath)
	if err != nil {
		return "", fmt.Errorf("making target path absolute: %w", err)
	}

	// Calculate relative path from directory of staging file to target
	stagingDir := filepath.Dir(absStaging)
	relPath, err := filepath.Rel(stagingDir, absTarget)
	if err != nil {
		return "", fmt.Errorf("calculating relative path: %w", err)
	}

	result := filepath.ToSlash(relPath)

	t.logDebug("CalculateRelativePath: %s -> %s = %s", filepath.Base(stagingFile), filepath.Base(targetAbsPath), result)

	return result, nil
}
