package books

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

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

// LinkTranslation holds path translations for a single file.
type LinkTranslation struct {
	sourceFile   string            // Original source file path (absolute)
	stagingFile  string            // Staging file path (absolute)
	translations map[string]string // old_path → new_path
}

// LinkTranslator manages translations for all files.
type LinkTranslator struct {
	sourceRoot   string                      // Source root directory (docs/)
	stagingRoot  string                      // Staging root directory (out/staging/book/)
	fileMap      map[string]string           // staging_path → source_path (absolute)
	docsRelPath  map[string]string           // source_path (absolute) → docs-relative path
	translations map[string]*LinkTranslation // staging_path → translation
	logWriter    io.Writer                   // Logger for debug output
	pdfMode      bool                        // True if building PDF (strips external links)
}

// NewLinkTranslator creates a new link translator.
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

// AddFileMapping tracks a source → staging file mapping.
func (t *LinkTranslator) AddFileMapping(stagingPath, sourcePath string) {
	t.fileMap[stagingPath] = sourcePath

	// Calculate docs-relative path for external link generation
	// sourcePath should be under sourceRoot/docs/
	docsDir := paths.DocsSourcePath(t.sourceRoot)
	if relPath, err := filepath.Rel(docsDir, sourcePath); err == nil {
		t.docsRelPath[sourcePath] = filepath.ToSlash(relPath)
	}
}

// GetSourcePath returns the source path for a staging path, if tracked.
// Used by orphan detection to check if a file in staging was tracked during this build.
func (t *LinkTranslator) GetSourcePath(stagingPath string) (string, bool) {
	src, ok := t.fileMap[stagingPath]
	return src, ok
}

// logDebug writes a debug message if logger is available.
func (t *LinkTranslator) logDebug(format string, args ...any) {
	if t.logWriter != nil {
		fmt.Fprintf(t.logWriter, "    [LinkTranslator] "+format+"\n", args...)
	}
}

// relativeLinkPattern matches markdown images, links, and HTML img/anchor tags
// Captures relative paths (not http://, https://, /, or #).
var (
	// Link extraction patterns (used by extractRelativeLinks)
	mdImagePattern  = regexp.MustCompile(`!\[.*?\]\(([^)]+)\)`)    // ![alt](path)
	mdLinkPattern   = regexp.MustCompile(`\[.*?\]\(([^)]+)\)`)     // [text](path)
	htmlImgPattern  = regexp.MustCompile(`<img[^>]+src="([^"]+)"`) // <img src="path">
	htmlLinkPattern = regexp.MustCompile(`<a[^>]+href="([^"]+)"`)  // <a href="path">

	// Code block patterns (used by stripCodeBlocks, replaceOutsideCodeBlocks)
	fencedCodeBlockPattern      = regexp.MustCompile("(?s)```.*?```")   // fenced code blocks
	fencedCodeBlockSplitPattern = regexp.MustCompile("(?s)(```.*?```)") // fenced code blocks (with capture for split)
)

// stripCodeBlocks removes both fenced and indented code blocks from markdown content
// to prevent extracting links from code examples.
func stripCodeBlocks(content string) string {
	// Remove fenced code blocks (```...```) using pre-compiled pattern
	content = fencedCodeBlockPattern.ReplaceAllString(content, "")

	// Remove indented code blocks (lines starting with 4 spaces or tab)
	lines := strings.Split(content, "\n")
	var cleaned []string

	for _, line := range lines {
		// Skip lines that start with 4+ spaces or tab (indented code blocks)
		if line != "" && (line[0] == '\t' || strings.HasPrefix(line, "    ")) {
			continue
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// extractRelativeLinks extracts all relative link references from markdown content.
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

// isRelativePathRef checks if a path is a relative reference (not external or absolute).
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

// BuildTranslations analyzes all source files and builds translation maps.
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

// ApplyAllTranslations applies path translations to all staging files.
func (t *LinkTranslator) ApplyAllTranslations() error {
	for stagingFile, trans := range t.translations {
		if err := t.applyTranslation(stagingFile, trans); err != nil {
			return fmt.Errorf("applying translations to %s: %w", stagingFile, err)
		}
	}
	return nil
}

// applyTranslation applies translations to a single staging file.
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
		if err := os.WriteFile(stagingFile, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing staging file: %w", err)
		}
	}

	return nil
}

// replaceOutsideCodeBlocks replaces old with new, but only outside fenced code blocks.
// It replaces within link contexts (markdown links/images, HTML src/href attributes)
// to prevent corrupting plain text when path patterns might overlap.
func replaceOutsideCodeBlocks(content, old, new string) string {
	// Split content by code blocks, keeping the code blocks as separators
	// Uses pre-compiled pattern for performance
	parts := fencedCodeBlockSplitPattern.Split(content, -1)
	codeBlocks := fencedCodeBlockSplitPattern.FindAllString(content, -1)

	// Apply replacements only to non-code-block parts
	var result strings.Builder
	for i, part := range parts {
		// Replace within link contexts to avoid corrupting plain text
		result.WriteString(replaceLinkPaths(part, old, new))

		// Add back the code block (unchanged) if there is one
		if i < len(codeBlocks) {
			result.WriteString(codeBlocks[i])
		}
	}

	return result.String()
}

// replaceLinkPaths replaces old path with new path within link contexts:
// - Markdown links: [text](path) or [text](path#anchor)
// - Markdown images: ![alt](path)
// - HTML src attributes: src="path"
// - HTML href attributes: href="path"
// This prevents corrupting plain text when path patterns overlap.
// The old path must match EXACTLY (not as a prefix) to avoid corrupting URLs
// like "reference/r2r-eac" when trying to replace "reference/".
func replaceLinkPaths(content, old, new string) string {
	escapedOld := regexp.QuoteMeta(old)

	// Pattern for markdown links and images: matches EXACT path (optionally followed by #anchor)
	// The path must match exactly - not as a prefix of a longer path
	mdLinkPattern := regexp.MustCompile(`(!?)\[([^\]]*)\]\((` + escapedOld + `)(#[^)]*)?\)`)
	content = mdLinkPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := mdLinkPattern.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		bang := parts[1] // "!" for images, "" for links
		text := parts[2] // link text or alt text
		anchor := ""
		if len(parts) > 4 && parts[4] != "" {
			anchor = parts[4] // anchor like #section
		}
		return fmt.Sprintf("%s[%s](%s%s)", bang, text, new, anchor)
	})

	// Pattern for HTML src attributes: src="old_path" (exact match)
	srcPattern := regexp.MustCompile(`(src=")(` + escapedOld + `)(")`)
	content = srcPattern.ReplaceAllString(content, "${1}"+new+"${3}")

	// Pattern for HTML href attributes: href="old_path" (exact match)
	hrefPattern := regexp.MustCompile(`(href=")(` + escapedOld + `)(")`)
	content = hrefPattern.ReplaceAllString(content, "${1}"+new+"${3}")

	return content
}

// replaceMarkdownLinks finds markdown links with the specified path and either
// strips them (keeps text only) or replaces the path.
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
