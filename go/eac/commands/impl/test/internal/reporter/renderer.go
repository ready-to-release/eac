package reporter

import (
	sharedTemplate "github.com/ready-to-release/eac/go/eac/commands/internal/template"
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
	// Use shared template renderer
	renderer := sharedTemplate.NewRenderer(r.templatePath)
	return renderer.RenderToFile(r.outputPath, r.data)
}

// NormalizeSpecPath normalizes a feature URI to a clean spec path
// Removes relative path prefixes and ensures it starts with "specs/"
func NormalizeSpecPath(uri string) string {
	return sharedTemplate.NormalizeSpecPath(uri)
}
