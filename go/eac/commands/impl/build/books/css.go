package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cssLinkPattern matches <link> tags with stylesheet references
// Captures: [1]=full href attribute value
var cssLinkPattern = regexp.MustCompile(`<link[^>]*href="([^"]*)"[^>]*rel="stylesheet"[^>]*>|<link[^>]*rel="stylesheet"[^>]*href="([^"]*)"[^>]*>`)

// EmbedCSSInHTML takes HTML content with external CSS <link> tags and embeds the CSS inline
// This fixes a WeasyPrint bug where large documents with file:// CSS URLs fail to render images
//
// Parameters:
//   - html: The HTML content with <link> stylesheet tags
//   - baseDir: Base directory for resolving relative CSS paths
//
// Returns:
//   - Modified HTML with CSS embedded inline
//   - Number of CSS files embedded
//   - Error if any
func EmbedCSSInHTML(html string, baseDir string) (string, int, error) {
	embedded := 0
	var allCSS strings.Builder

	// Find all CSS link tags
	matches := cssLinkPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		// Get the href (could be in group 1 or 2 depending on attribute order)
		href := match[1]
		if href == "" {
			href = match[2]
		}
		if href == "" {
			continue
		}

		// Resolve the CSS file path
		cssPath := resolveCSSPath(href, baseDir)
		if cssPath == "" {
			continue
		}

		// Read the CSS file
		cssContent, err := os.ReadFile(cssPath)
		if err != nil {
			// Skip files that can't be read
			continue
		}

		allCSS.WriteString(fmt.Sprintf("/* Embedded from: %s */\n", filepath.Base(cssPath)))
		allCSS.Write(cssContent)
		allCSS.WriteString("\n\n")
		embedded++
	}

	if embedded == 0 {
		return html, 0, nil
	}

	// Remove all CSS link tags
	html = cssLinkPattern.ReplaceAllString(html, "")

	// Insert embedded CSS before </head>
	embeddedStyle := fmt.Sprintf("<style>\n%s</style>\n</head>", allCSS.String())
	html = strings.Replace(html, "</head>", embeddedStyle, 1)

	return html, embedded, nil
}

// resolveCSSPath converts a CSS href to an absolute file path
func resolveCSSPath(href string, baseDir string) string {
	// Handle file:// URLs
	if strings.HasPrefix(href, "file://") {
		// Remove file:// prefix
		path := strings.TrimPrefix(href, "file://")
		// On Windows, file:// URLs might have an extra leading slash
		path = strings.TrimPrefix(path, "/")
		// Check if it exists
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}

	// Handle relative paths
	if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
		absPath := filepath.Join(baseDir, href)
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	return ""
}

// ProcessDebugHTMLForPDF processes the mkdocs-with-pdf debug HTML output
// by embedding CSS inline to fix WeasyPrint's image rendering issues
//
// Parameters:
//   - debugHTMLPath: Path to the debug HTML file from mkdocs-with-pdf
//   - outputPath: Path to write the processed HTML
//
// Returns the number of CSS files embedded and any error
func ProcessDebugHTMLForPDF(debugHTMLPath, outputPath string) (int, error) {
	// Read the debug HTML
	content, err := os.ReadFile(debugHTMLPath)
	if err != nil {
		return 0, fmt.Errorf("reading debug HTML: %w", err)
	}

	// Get the base directory for resolving CSS paths
	baseDir := filepath.Dir(debugHTMLPath)

	// Embed CSS inline
	processedHTML, embedded, err := EmbedCSSInHTML(string(content), baseDir)
	if err != nil {
		return 0, fmt.Errorf("embedding CSS: %w", err)
	}

	// Write the processed HTML
	if err := os.WriteFile(outputPath, []byte(processedHTML), 0644); err != nil {
		return 0, fmt.Errorf("writing processed HTML: %w", err)
	}

	return embedded, nil
}
