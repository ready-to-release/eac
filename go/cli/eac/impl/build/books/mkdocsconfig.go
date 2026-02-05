// mkdocsconfig.go - MkDocs configuration template loading and generation
package books

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

// TemplateType represents the type of mkdocs template to use.
type TemplateType string

const (
	TemplateSite TemplateType = "site"
	TemplatePDF  TemplateType = "pdf"
)

// GetTemplateType returns the template type for a given output format.
func GetTemplateType(outputFormat string) TemplateType {
	if outputFormat == "site" {
		return TemplateSite
	}
	return TemplatePDF
}

// GetTemplatePath returns the path to the mkdocs.yml template for the given type.
func GetTemplatePath(workspaceRoot string, templateType TemplateType) string {
	switch templateType {
	case TemplateSite:
		return paths.MkDocsSiteTemplatePath(workspaceRoot)
	case TemplatePDF:
		return paths.MkDocsPdfTemplatePath(workspaceRoot)
	default:
		return paths.MkDocsSiteTemplatePath(workspaceRoot)
	}
}

// LoadMkDocsTemplate loads the mkdocs.yml template for the given output type.
func LoadMkDocsTemplate(workspaceRoot, outputFormat string) ([]byte, error) {
	templateType := GetTemplateType(outputFormat)
	templatePath := GetTemplatePath(workspaceRoot, templateType)

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mkdocs template %s: %w", templatePath, err)
	}

	return data, nil
}

// ConfigOptions contains the options for generating a mkdocs.yml config.
type ConfigOptions struct {
	SiteName        string // Book/site name
	SiteDescription string // Site description
	BookTitle       string // Book-specific title for PDF cover page
	BookDescription string // Book-specific description for PDF cover page
	SiteURL         string // Site URL for links
	DocsDir         string // Path to docs directory (relative to container mount)
	Theme           string // Theme name: "dark" or "light" (for PDF)
	OutputFormat    string // Output format: "site", "pdf-dark", "pdf-light"
	PDFConcurrency  int    // Playwright concurrency for PDF export (CI: 4, local: 8)
	PageLimit       int    // Maximum pages to render (0 = unlimited, used in reduced mode)
}

// GenerateMkDocsConfig generates a final mkdocs.yml by loading a template
// and injecting the provided options using string replacement.
// This preserves Python-specific YAML tags like !!python/name: that would be
// lost by parsing and re-marshaling the YAML.
func GenerateMkDocsConfig(workspaceRoot string, opts ConfigOptions) ([]byte, error) {
	// Load template
	templateData, err := LoadMkDocsTemplate(workspaceRoot, opts.OutputFormat)
	if err != nil {
		return nil, err
	}

	content := string(templateData)

	// Replace values using line-based matching to preserve structure
	if opts.SiteName != "" {
		content = replaceYAMLValue(content, "site_name", opts.SiteName)
	}
	if opts.SiteDescription != "" {
		content = replaceYAMLValue(content, "site_description", opts.SiteDescription)
	}
	if opts.SiteURL != "" {
		content = replaceYAMLValue(content, "site_url", fmt.Sprintf(`"%s"`, opts.SiteURL))
	}
	if opts.DocsDir != "" {
		content = replaceYAMLValue(content, "docs_dir", opts.DocsDir)
	}

	// For PDF builds, update the theme-specific paths
	if opts.OutputFormat != "site" && opts.Theme != "" {
		content = updatePDFThemeString(content, opts.Theme, opts.DocsDir)
	}

	// For PDF builds, update concurrency if specified
	if opts.PDFConcurrency > 0 {
		content = replacePDFConcurrency(content, opts.PDFConcurrency)
	}

	// For PDF builds, update page limit if specified (reduced mode)
	if opts.PageLimit > 0 {
		content = replacePDFPageLimit(content, opts.PageLimit)
	}

	// For PDF builds, inject book_title and book_description into aggregator config
	if opts.OutputFormat != "site" {
		content = injectAggregatorBookInfo(content, opts.BookTitle, opts.BookDescription)
	}

	return []byte(content), nil
}

// replaceYAMLValue replaces a top-level YAML value using regex.
func replaceYAMLValue(content, key, value string) string {
	// Match key at start of line followed by colon and value
	pattern := fmt.Sprintf(`(?m)^(%s:\s*).*$`, key)
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(content, fmt.Sprintf("${1}%s", value))
}

// updatePDFThemeString updates theme-specific CSS paths in the config string
// docsDir is the relative path from the config file to the docs directory
// This is needed because the exporter plugin resolves stylesheet paths relative to the config file.
func updatePDFThemeString(content, theme, docsDir string) string {
	// Build the new path with theme and docsDir prefix for exporter plugin
	// The exporter plugin's stylesheets option resolves paths relative to config file,
	// so we need to prepend the docsDir to reach the assets in the staging directory
	exporterPath := fmt.Sprintf("assets/templates/pdf-%s/styles.css", theme)
	if docsDir != "" {
		// Prepend docsDir to make path relative to config file location
		exporterPath = docsDir + "/" + exporterPath
	}

	// extra_css path (MkDocs resolves this relative to docs_dir, so just update theme, no prefix)
	extraCssPath := fmt.Sprintf("assets/templates/pdf-%s/styles.css", theme)

	// Update extra_css first (2 spaces indentation - MkDocs resolves relative to docs_dir)
	// Must do this BEFORE the exporter pattern since we're matching by exact indentation
	extraCssPattern := regexp.MustCompile(`(?m)^(  - )assets/templates/pdf-(dark|light)/styles\.css\s*$`)
	content = extraCssPattern.ReplaceAllString(content, fmt.Sprintf("${1}%s", extraCssPath))

	// Update exporter plugin stylesheets (12 spaces indentation - needs docsDir prefix)
	exporterPattern := regexp.MustCompile(`(?m)^(            - )assets/templates/pdf-(dark|light)/styles\.css\s*$`)
	content = exporterPattern.ReplaceAllString(content, fmt.Sprintf("${1}%s", exporterPath))

	// Update mermaid-to-svg theme setting
	// Match "theme: dark" or "theme: light" within mermaid-to-svg config
	themePattern := regexp.MustCompile(`(?m)^(\s+theme:\s*)(dark|light)\s*$`)
	content = themePattern.ReplaceAllString(content, fmt.Sprintf("${1}%s", theme))

	return content
}

// replacePDFConcurrency updates the PDF exporter concurrency setting
// Matches "concurrency: N" within the exporter plugin's pdf formats section.
func replacePDFConcurrency(content string, concurrency int) string {
	// Match "concurrency: <number>" with proper indentation (10 spaces in pdf formats section)
	pattern := regexp.MustCompile(`(?m)^(\s+concurrency:\s*)\d+\s*$`)
	return pattern.ReplaceAllString(content, fmt.Sprintf("${1}%d", concurrency))
}

// replacePDFPageLimit updates the PDF exporter page_limit setting
// Matches "page_limit: N" within the exporter plugin's pdf formats section.
func replacePDFPageLimit(content string, pageLimit int) string {
	// Match "page_limit: <number>" with proper indentation
	pattern := regexp.MustCompile(`(?m)^(\s+page_limit:\s*)\d+\s*$`)
	return pattern.ReplaceAllString(content, fmt.Sprintf("${1}%d", pageLimit))
}

// injectAggregatorBookInfo adds book_title and book_description to the aggregator config section.
// These are used by the PDF exporter to generate cover pages, TOC, and footers.
func injectAggregatorBookInfo(content, bookTitle, bookDescription string) string {
	if bookTitle == "" && bookDescription == "" {
		return content
	}

	// Find the aggregator section and add book_title/book_description after merge_workers
	// Look for "merge_workers:" line and add after it
	lines := strings.Split(content, "\n")
	var result []string

	for i, line := range lines {
		result = append(result, line)

		// After merge_workers line, inject book_title and book_description
		if strings.Contains(line, "merge_workers:") {
			// Detect indentation (12 spaces for aggregator options)
			indent := "            "
			if bookTitle != "" {
				// Escape single quotes in title
				escaped := strings.ReplaceAll(bookTitle, "'", "''")
				result = append(result, fmt.Sprintf("%sbook_title: '%s'", indent, escaped))
			}
			if bookDescription != "" {
				// Escape single quotes in description
				escaped := strings.ReplaceAll(bookDescription, "'", "''")
				result = append(result, fmt.Sprintf("%sbook_description: '%s'", indent, escaped))
			}
		}
		_ = i // Avoid unused variable warning
	}

	return strings.Join(result, "\n")
}

// WriteMkDocsConfig writes a generated mkdocs.yml to the specified path.
func WriteMkDocsConfig(workspaceRoot, outputPath string, opts ConfigOptions) error {
	data, err := GenerateMkDocsConfig(workspaceRoot, opts)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write mkdocs config: %w", err)
	}

	return nil
}

// GetThemeFromOutput extracts the theme name from an output format
// e.g., "pdf-dark" -> "dark", "pdf-light" -> "light".
func GetThemeFromOutput(outputFormat string) string {
	if strings.HasPrefix(outputFormat, "pdf-") {
		return strings.TrimPrefix(outputFormat, "pdf-")
	}
	return "dark" // default
}
