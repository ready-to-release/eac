// Feature: commands_templates
// Unit tests for templates command
package templates

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
		"README.md":     "# {{ .ProjectName }}",
		"src/main.go":   "package main\n// Version: {{ .Version }}",
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

func TestPlaceholderScanner_Scan(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template files with various placeholders
	files := map[string]string{
		"README.md":             "# {{ .ProjectName }}\n\nVersion: {{ .Version }}",
		"config.yaml":           "name: {{ .AppName }}\nport: {{ .Port }}",
		"{{ .ConfigFile }}.txt": "Config for {{ .ProjectName }}",
		"src/main.go":           "package main\n// Author: {{ .Author }}",
	}

	for fileName, content := range files {
		filePath := filepath.Join(templateDir, fileName)
		dir := filepath.Dir(filePath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	// Scan templates
	scanner := NewPlaceholderScanner(templateDir)
	placeholders, err := scanner.Scan()
	require.NoError(t, err)

	// Verify placeholders found
	expected := []string{"AppName", "Author", "ConfigFile", "Port", "ProjectName", "Version"}
	assert.Equal(t, expected, placeholders)
	assert.Equal(t, 6, len(placeholders))
}

func TestPlaceholderScanner_Scan_NoDuplicates(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template with repeated placeholders
	content := "{{ .Name }} is {{ .Name }} and version is {{ .Version }} ({{ .Version }})"
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "test.txt"), []byte(content), 0644))

	// Scan templates
	scanner := NewPlaceholderScanner(templateDir)
	placeholders, err := scanner.Scan()
	require.NoError(t, err)

	// Should only have unique placeholders
	expected := []string{"Name", "Version"}
	assert.Equal(t, expected, placeholders)
	assert.Equal(t, 2, len(placeholders))
}

func TestPlaceholderScanner_Scan_NoPlaceholders(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	require.NoError(t, os.MkdirAll(templateDir, 0755))

	// Create template without placeholders
	content := "This is a plain text file with no placeholders"
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "plain.txt"), []byte(content), 0644))

	// Scan templates
	scanner := NewPlaceholderScanner(templateDir)
	placeholders, err := scanner.Scan()
	require.NoError(t, err)

	// Should be empty
	assert.Empty(t, placeholders)
}

func TestPlaceholderScanner_Scan_NonExistentDirectory(t *testing.T) {
	scanner := NewPlaceholderScanner("nonexistent-dir")
	_, err := scanner.Scan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template directory does not exist")
}

func TestPlaceholderScanner_Scan_FileInsteadOfDirectory(t *testing.T) {
	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("content"), 0644))

	scanner := NewPlaceholderScanner(tmpFile)
	_, err := scanner.Scan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template path is not a directory")
}

func TestIsGitRepository(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "HTTPS GitHub URL",
			url:      "https://github.com/user/repo",
			expected: true,
		},
		{
			name:     "HTTPS GitHub URL with .git",
			url:      "https://github.com/user/repo.git",
			expected: true,
		},
		{
			name:     "HTTPS GitLab URL",
			url:      "https://gitlab.com/user/repo",
			expected: true,
		},
		{
			name:     "HTTP URL",
			url:      "http://example.com/repo.git",
			expected: true,
		},
		{
			name:     "SSH URL",
			url:      "git@github.com:user/repo.git",
			expected: true,
		},
		{
			name:     "Git protocol",
			url:      "git://github.com/user/repo",
			expected: true,
		},
		{
			name:     "Local path (absolute)",
			url:      "/home/user/templates",
			expected: false,
		},
		{
			name:     "Local path (relative)",
			url:      "./templates",
			expected: false,
		},
		{
			name:     "Local path (Windows)",
			url:      "C:\\templates",
			expected: false,
		},
		{
			name:     "Empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "Short string",
			url:      "abc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitRepository(tt.url)
			assert.Equal(t, tt.expected, result, "IsGitRepository(%q) = %v, want %v", tt.url, result, tt.expected)
		})
	}
}

func TestPlaceholderScanner_ExtractPlaceholders(t *testing.T) {
	scanner := NewPlaceholderScanner("")

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "simple placeholder",
			text:     "Hello {{ .Name }}",
			expected: []string{"Name"},
		},
		{
			name:     "multiple placeholders",
			text:     "{{ .First }} and {{ .Second }}",
			expected: []string{"First", "Second"},
		},
		{
			name:     "placeholder with underscore",
			text:     "Value: {{ .My_Value }}",
			expected: []string{"My_Value"},
		},
		{
			name:     "placeholder with numbers",
			text:     "Item {{ .Item123 }}",
			expected: []string{"Item123"},
		},
		{
			name:     "no spaces around dot",
			text:     "{{.NoSpaces}}",
			expected: []string{"NoSpaces"},
		},
		{
			name:     "extra spaces",
			text:     "{{  .ExtraSpaces  }}",
			expected: []string{"ExtraSpaces"},
		},
		{
			name:     "no placeholders",
			text:     "Just plain text",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner.placeholders = make(map[string]map[string]bool)
			scanner.extractPlaceholders(tt.text, "test.txt")

			// Convert map to sorted slice
			var found []string
			for key := range scanner.placeholders {
				found = append(found, key)
			}
			if len(found) == 0 {
				found = []string{}
			}

			assert.ElementsMatch(t, tt.expected, found)
		})
	}
}
