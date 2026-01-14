// Package template provides shared template rendering functionality for all report types.
package template

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles template rendering for reports
type Renderer struct {
	templatePath   string
	customFuncs    template.FuncMap
	missingKeyMode string
}

// NewRenderer creates a new template renderer
func NewRenderer(templatePath string) *Renderer {
	return &Renderer{
		templatePath:   templatePath,
		customFuncs:    defaultFuncMap(),
		missingKeyMode: "zero", // Default: missing keys render as zero values
	}
}

// WithFuncs adds custom template functions
func (r *Renderer) WithFuncs(funcs template.FuncMap) *Renderer {
	for name, fn := range funcs {
		r.customFuncs[name] = fn
	}
	return r
}

// WithMissingKeyMode sets the missingkey option (zero, invalid, error)
func (r *Renderer) WithMissingKeyMode(mode string) *Renderer {
	r.missingKeyMode = mode
	return r
}

// RenderToString renders the template to a string
func (r *Renderer) RenderToString(data interface{}) (string, error) {
	var buf strings.Builder
	if err := r.renderToWriter(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderToFile renders the template to a file
func (r *Renderer) RenderToFile(outputPath string, data interface{}) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer outFile.Close()

	return r.renderToWriter(outFile, data)
}

// renderToWriter is the core rendering function that writes to any io.Writer
func (r *Renderer) renderToWriter(w io.Writer, data interface{}) error {
	// Verify template file exists
	if _, err := os.Stat(r.templatePath); err != nil {
		return fmt.Errorf("template not found: %s", r.templatePath)
	}

	// Read template file
	tmplContent, err := os.ReadFile(r.templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Create template with custom functions
	tmpl, err := template.New(filepath.Base(r.templatePath)).
		Funcs(r.customFuncs).
		Option(fmt.Sprintf("missingkey=%s", r.missingKeyMode)).
		Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// defaultFuncMap returns the default template functions available to all templates
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"join":  strings.Join,
		"split": strings.Split,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,

		// Comparison functions
		"gt":  func(a, b int) bool { return a > b },
		"gte": func(a, b int) bool { return a >= b },
		"lt":  func(a, b int) bool { return a < b },
		"lte": func(a, b int) bool { return a <= b },
		"eq":  func(a, b int) bool { return a == b },

		// Math functions (integer)
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Math functions (float64)
		"mulf": func(a, b float64) float64 { return a * b },
		"divf": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Percentage calculation
		"percent": func(value, total int) float64 {
			if total == 0 {
				return 0.0
			}
			return float64(value) / float64(total) * 100
		},

		// Format percentage with precision
		"percentf": func(value, total int, precision int) string {
			if total == 0 {
				return "0.0"
			}
			pct := float64(value) / float64(total) * 100
			return fmt.Sprintf("%.*f", precision, pct)
		},

		// Risk badge - format risk level with icon
		"riskBadge": func(band string) string {
			switch strings.ToLower(band) {
			case "critical":
				return "🔴 Critical"
			case "high":
				return "🟠 High"
			case "medium":
				return "🟡 Medium"
			case "low":
				return "🟢 Low"
			default:
				return band
			}
		},

		// Truncate string to max length with ellipsis
		"truncate": func(s string, maxLen int) string {
			if len(s) <= maxLen {
				return s
			}
			if maxLen <= 3 {
				return s[:maxLen]
			}
			return s[:maxLen-3] + "..."
		},

		// First N items from a slice
		"first": func(n int, items []string) []string {
			if n >= len(items) {
				return items
			}
			return items[:n]
		},

		// Last N items from a slice
		"last": func(n int, items []string) []string {
			if n >= len(items) {
				return items
			}
			return items[len(items)-n:]
		},
	}
}

// NormalizeSpecPath normalizes a feature URI to a clean spec path
// Removes relative path prefixes and ensures it starts with "specs/"
func NormalizeSpecPath(uri string) string {
	// Convert to forward slashes
	normalized := filepath.ToSlash(uri)

	// Remove leading "../" prefixes
	for strings.HasPrefix(normalized, "../") {
		normalized = strings.TrimPrefix(normalized, "../")
	}

	// Ensure it starts with "specs/"
	if !strings.HasPrefix(normalized, "specs/") {
		normalized = "specs/" + normalized
	}

	return normalized
}
