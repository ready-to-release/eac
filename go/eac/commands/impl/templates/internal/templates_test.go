//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates command
package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadValuesFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		expectError bool
		expected    TemplateValues
	}{
		{
			name: "valid JSON",
			jsonContent: `{
				"ProjectName": "test-project",
				"Version": "1.0.0"
			}`,
			expectError: false,
			expected: TemplateValues{
				"ProjectName": "test-project",
				"Version":     "1.0.0",
			},
		},
		{
			name:        "invalid JSON",
			jsonContent: `{ invalid }`,
			expectError: true,
		},
		{
			name:        "empty JSON",
			jsonContent: `{}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "values.json")
			err := os.WriteFile(tmpFile, []byte(tt.jsonContent), 0644)
			require.NoError(t, err)

			// Test
			values, err := LoadValuesFromJSON(tmpFile)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, values)
			}
		})
	}
}

func TestLoadValuesFromJSON_FileNotFound(t *testing.T) {
	_, err := LoadValuesFromJSON("nonexistent.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read values file")
}

func TestValidateValues(t *testing.T) {
	values := TemplateValues{
		"ProjectName": "test",
		"Version":     "1.0.0",
	}

	tests := []struct {
		name        string
		required    []string
		expectError bool
	}{
		{
			name:        "all values present",
			required:    []string{"ProjectName", "Version"},
			expectError: false,
		},
		{
			name:        "missing required value",
			required:    []string{"ProjectName", "Author"},
			expectError: true,
		},
		{
			name:        "no required values",
			required:    []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValues(values, tt.required)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRenderer_RenderString(t *testing.T) {
	renderer := NewRenderer("", "", TemplateValues{
		"ProjectName": "my-project",
		"Version":     "1.0.0",
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple replacement",
			input:    "{{ .ProjectName }}",
			expected: "my-project",
		},
		{
			name:     "multiple replacements",
			input:    "{{ .ProjectName }}-v{{ .Version }}",
			expected: "my-project-v1.0.0",
		},
		{
			name:     "no template markers",
			input:    "plain-text",
			expected: "plain-text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.renderString(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderer_RenderString_InvalidTemplate(t *testing.T) {
	renderer := NewRenderer("", "", TemplateValues{})
	_, err := renderer.renderString("{{ .UnclosedTag")
	assert.Error(t, err)
}

func TestRenderer_RenderFile(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file
	tmplFile := filepath.Join(templateDir, "README.md")
	tmplContent := "# {{ .ProjectName }}\n\nVersion: {{ .Version }}"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer
	values := TemplateValues{
		"ProjectName": "test-project",
		"Version":     "1.0.0",
	}
	renderer := NewRenderer(templateDir, outputDir, values)

	// Render file
	outputFile := filepath.Join(outputDir, "README.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(outputFile), 0755))
	err := renderer.renderFile(tmplFile, outputFile)
	require.NoError(t, err)

	// Verify output
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	expected := "# test-project\n\nVersion: 1.0.0"
	assert.Equal(t, expected, string(content))
}

func TestRenderer_RenderTemplates(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	// Create template structure
	require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "src"), 0755))

	// Create template files
	files := map[string]string{
		"README.md":       "# {{ .ProjectName }}",
		"src/main.go":     "package main\n// Version: {{ .Version }}",
		"{{ .Name }}.txt": "Name is {{ .Name }}",
	}

	for fileName, content := range files {
		filePath := filepath.Join(templateDir, fileName)
		dir := filepath.Dir(filePath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	// Create renderer
	values := TemplateValues{
		"ProjectName": "test-project",
		"Version":     "1.0.0",
		"Name":        "config",
	}
	renderer := NewRenderer(templateDir, outputDir, values)

	// Render templates
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify outputs
	// Check README.md
	content, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# test-project", string(content))

	// Check src/main.go
	content, err = os.ReadFile(filepath.Join(outputDir, "src", "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Version: 1.0.0")

	// Check config.txt (rendered file name)
	content, err = os.ReadFile(filepath.Join(outputDir, "config.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Name is config", string(content))
}

func TestRenderer_RenderTemplates_InvalidTemplate(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file with invalid syntax
	tmplFile := filepath.Join(templateDir, "bad.txt")
	tmplContent := "{{ .UnclosedTag"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer
	values := TemplateValues{
		"ProjectName": "test",
	}
	renderer := NewRenderer(templateDir, outputDir, values)

	// Render templates
	err := renderer.RenderTemplates()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

// NOTE: PlaceholderScanner and IsGitRepository tests removed
// These functions were removed from the implementation during refactoring
// Tests preserved above still validate the core template rendering functionality

func TestRenderer_PreservePlaceholdersWhenNoValues(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file with placeholders
	tmplFile := filepath.Join(templateDir, "README.md")
	tmplContent := "# {{ .ProjectName }}\n\nVersion: {{ .Version }}\n\nRequirements: {{ .changed_requirements }}"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer with nil values (no JSON file provided)
	renderer := NewRenderer(templateDir, outputDir, nil)

	// Render templates
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify output preserves placeholders
	content, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)

	// Placeholders should be preserved exactly as-is
	assert.Equal(t, tmplContent, string(content), "Placeholders should be preserved when no values provided")
	assert.Contains(t, string(content), "{{ .ProjectName }}")
	assert.Contains(t, string(content), "{{ .Version }}")
	assert.Contains(t, string(content), "{{ .changed_requirements }}")
	assert.NotContains(t, string(content), "<no value>")
}

func TestRenderer_PreservePlaceholdersWhenEmptyValues(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file with placeholders
	tmplFile := filepath.Join(templateDir, "config.yaml")
	tmplContent := "name: {{ .AppName }}\nport: {{ .Port }}"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer with empty values map
	renderer := NewRenderer(templateDir, outputDir, TemplateValues{})

	// Render templates
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify output preserves placeholders
	content, err := os.ReadFile(filepath.Join(outputDir, "config.yaml"))
	require.NoError(t, err)

	// Placeholders should be preserved exactly as-is
	assert.Equal(t, tmplContent, string(content), "Placeholders should be preserved when empty values provided")
	assert.Contains(t, string(content), "{{ .AppName }}")
	assert.Contains(t, string(content), "{{ .Port }}")
	assert.NotContains(t, string(content), "<no value>")
}

func TestRenderer_ReplaceValuesWhenProvided(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file with placeholders
	tmplFile := filepath.Join(templateDir, "README.md")
	tmplContent := "# {{ .ProjectName }}\n\nVersion: {{ .Version }}"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0644))

	// Create renderer with actual values
	values := TemplateValues{
		"ProjectName": "MyProject",
		"Version":     "2.0.0",
	}
	renderer := NewRenderer(templateDir, outputDir, values)

	// Render templates
	err := renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify output has replaced values
	content, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)

	expected := "# MyProject\n\nVersion: 2.0.0"
	assert.Equal(t, expected, string(content))
	assert.Contains(t, string(content), "MyProject")
	assert.Contains(t, string(content), "2.0.0")
	assert.NotContains(t, string(content), "{{ .ProjectName }}")
	assert.NotContains(t, string(content), "{{ .Version }}")
}

func TestRenderer_PreserveFilePermissionsWhenCopying(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template file with specific permissions
	tmplFile := filepath.Join(templateDir, "script.sh")
	tmplContent := "#!/bin/bash\n# {{ .ScriptName }}"
	require.NoError(t, os.WriteFile(tmplFile, []byte(tmplContent), 0755))

	// Get original file permissions
	origInfo, err := os.Stat(tmplFile)
	require.NoError(t, err)

	// Create renderer with nil values
	renderer := NewRenderer(templateDir, outputDir, nil)

	// Render templates
	err = renderer.RenderTemplates()
	require.NoError(t, err)

	// Verify permissions are preserved (use original permissions for comparison)
	info, err := os.Stat(filepath.Join(outputDir, "script.sh"))
	require.NoError(t, err)
	assert.Equal(t, origInfo.Mode(), info.Mode(), "File permissions should be preserved")

	// Verify content is unchanged
	content, err := os.ReadFile(filepath.Join(outputDir, "script.sh"))
	require.NoError(t, err)
	assert.Equal(t, tmplContent, string(content))
}
