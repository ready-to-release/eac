package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles template rendering with value substitution
type Renderer struct {
	templateDir string
	outputDir   string
	values      TemplateValues
}

// NewRenderer creates a new template renderer
func NewRenderer(templateDir, outputDir string, values TemplateValues) *Renderer {
	return &Renderer{
		templateDir: templateDir,
		outputDir:   outputDir,
		values:      values,
	}
}

// shouldRender returns true if templates should be rendered with value substitution
// Returns false if no values are provided (templates should be copied as-is)
func (r *Renderer) shouldRender() bool {
	return r.values != nil && len(r.values) > 0
}

// RenderTemplates walks the template directory and renders all templates
// If no values are provided, files are copied as-is without template processing
func (r *Renderer) RenderTemplates() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Walk template directory
	return filepath.WalkDir(r.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the template directory itself
		if path == r.templateDir {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(r.templateDir, path)
		if err != nil {
			return err
		}

		// Determine output path
		// If we should render, process path names as templates
		// Otherwise, use path as-is
		var outputPath string
		if r.shouldRender() {
			// Render path name (support {{ .ProjectName }} in file/dir names)
			renderedPath, err := r.renderString(relPath)
			if err != nil {
				return fmt.Errorf("failed to render path %s: %w", relPath, err)
			}

			// SECURITY: Validate path doesn't escape output directory
			if err := ValidatePath(r.outputDir, renderedPath); err != nil {
				return fmt.Errorf("security violation in template path: %w", err)
			}

			outputPath = filepath.Join(r.outputDir, renderedPath)
		} else {
			// No rendering - use path as-is
			// SECURITY: Validate path doesn't escape output directory
			if err := ValidatePath(r.outputDir, relPath); err != nil {
				return fmt.Errorf("security violation in template path: %w", err)
			}

			outputPath = filepath.Join(r.outputDir, relPath)
		}

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(outputPath, 0755)
		}

		// Process file: render with values or copy as-is
		if r.shouldRender() {
			return r.renderFile(path, outputPath)
		} else {
			return r.copyFile(path, outputPath)
		}
	})
}

// renderFile renders a single template file
func (r *Renderer) renderFile(inputPath, outputPath string) error {
	// Read template file
	tmplContent, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template with option to treat missing keys as zero value (empty string)
	// This allows optional variables to be left out of values.json
	tmpl, err := template.New(filepath.Base(inputPath)).Option("missingkey=zero").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", inputPath, err)
	}

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

	// Execute template
	if err := tmpl.Execute(outFile, r.values); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", inputPath, err)
	}

	// Copy file permissions
	info, err := os.Stat(inputPath)
	if err == nil {
		os.Chmod(outputPath, info.Mode())
	}

	return nil
}

// renderString renders a string template (used for file/directory names)
func (r *Renderer) renderString(input string) (string, error) {
	// Use missingkey=zero to allow optional variables in paths
	tmpl, err := template.New("string").Option("missingkey=zero").Parse(input)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, r.values); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// copyFile copies a file without template processing, preserving all placeholders
func (r *Renderer) copyFile(inputPath, outputPath string) error {
	// Read source file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write to destination without any modifications
	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	// Copy file permissions
	info, err := os.Stat(inputPath)
	if err == nil {
		os.Chmod(outputPath, info.Mode())
	}

	return nil
}
