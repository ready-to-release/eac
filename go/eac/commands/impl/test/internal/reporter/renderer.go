package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles test report template rendering
type Renderer struct {
	templatePath string
	outputPath   string
	data         TestSuiteReportData
}

// NewRenderer creates a new test report renderer
func NewRenderer(templatePath, outputPath string, data TestSuiteReportData) *Renderer {
	return &Renderer{
		templatePath: templatePath,
		outputPath:   outputPath,
		data:         data,
	}
}

// Render executes the template and writes the report to the output file
func (r *Renderer) Render() error {
	// Read template file
	tmplContent, err := os.ReadFile(r.templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", r.templatePath, err)
	}

	// Create template with custom functions
	tmpl, err := template.New(filepath.Base(r.templatePath)).
		Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
		}).
		Option("missingkey=zero").
		Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", r.templatePath, err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(r.outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(r.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", r.outputPath, err)
	}
	defer outFile.Close()

	// Execute template
	if err := tmpl.Execute(outFile, r.data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", r.templatePath, err)
	}

	return nil
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
