//go:build L1 && ov
// +build L1,ov

package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateFile_EmptyFile(t *testing.T) {
	tests := []struct {
		name         string
		allowEmpty   bool
		expectValid  bool
		expectErrors int
	}{
		{
			name:         "empty file not allowed",
			allowEmpty:   false,
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name:         "empty file allowed",
			allowEmpty:   true,
			expectValid:  true,
			expectErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := createTempMarkdown(t, "")

			opts := DefaultValidatorOptions()
			opts.AllowEmptyFiles = tt.allowEmpty

			validator := NewValidator(opts, &bytes.Buffer{})
			result := validator.ValidateFile(tmpFile)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
			}
		})
	}
}

func TestValidateFile_CodeBlockValidation(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectValid  bool
		expectErrors int
	}{
		{
			name: "valid json code block",
			content: `# Test
` + "```json\n" + `{"key": "value"}` + "\n```",
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "invalid json code block",
			content: `# Test
` + "```json\n" + `{invalid json}` + "\n```",
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "valid yaml code block",
			content: `# Test
` + "```yaml\n" + `key: value` + "\n```",
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "invalid yaml code block",
			content: `# Test
` + "```yaml\n" + `key: [invalid: yaml}` + "\n```",
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "unknown language skipped",
			content: `# Test
` + "```python\n" + `print("hello")` + "\n```",
			expectValid:  true,
			expectErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempMarkdown(t, tt.content)

			opts := DefaultValidatorOptions()
			opts.ValidateCodeBlocks = true

			validator := NewValidator(opts, &bytes.Buffer{})
			result := validator.ValidateFile(tmpFile)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v (errors: %v)", tt.expectValid, result.Valid, result.Errors)
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestValidateFile_RequiredSections(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		requiredSections []string
		expectWarnings   int
	}{
		{
			name: "all required sections present",
			content: `# Title
## Description
## Usage
## Examples`,
			requiredSections: []string{"Description", "Usage", "Examples"},
			expectWarnings:   0,
		},
		{
			name: "missing required sections",
			content: `# Title
## Description`,
			requiredSections: []string{"Description", "Usage", "Examples"},
			expectWarnings:   2,
		},
		{
			name: "case insensitive section matching",
			content: `# Title
## description
## USAGE`,
			requiredSections: []string{"Description", "Usage"},
			expectWarnings:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempMarkdown(t, tt.content)

			opts := DefaultValidatorOptions()
			opts.RequiredSections = tt.requiredSections

			validator := NewValidator(opts, &bytes.Buffer{})
			result := validator.ValidateFile(tmpFile)

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}
		})
	}
}

func TestValidateFile_HeadingHierarchy(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectWarnings int
	}{
		{
			name: "proper heading hierarchy",
			content: `# H1
## H2
### H3
## H2
### H3`,
			expectWarnings: 0,
		},
		{
			name: "heading level jump",
			content: `# H1
### H3`,
			expectWarnings: 1,
		},
		{
			name: "multiple heading jumps",
			content: `# H1
### H3
##### H5`,
			expectWarnings: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempMarkdown(t, tt.content)

			opts := DefaultValidatorOptions()
			opts.CheckHeadingHierarchy = true

			validator := NewValidator(opts, &bytes.Buffer{})
			result := validator.ValidateFile(tmpFile)

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.expectWarnings, len(result.Warnings), result.Warnings)
			}
		})
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	content := `# Test
` + "```json\n" + `{"key": "value"}` + "\n```\n" +
		"Some text\n" +
		"```yaml\n" + `key: value` + "\n```"

	tmpFile := createTempMarkdown(t, content)

	opts := DefaultValidatorOptions()
	validator := NewValidator(opts, &bytes.Buffer{})
	result := validator.ValidateFile(tmpFile)

	if len(result.CodeBlocks) != 2 {
		t.Errorf("Expected 2 code blocks, got %d", len(result.CodeBlocks))
	}

	if result.CodeBlocks[0].Language != "json" {
		t.Errorf("Expected first block language 'json', got '%s'", result.CodeBlocks[0].Language)
	}

	if result.CodeBlocks[1].Language != "yaml" {
		t.Errorf("Expected second block language 'yaml', got '%s'", result.CodeBlocks[1].Language)
	}

	if !strings.Contains(result.CodeBlocks[0].Content, `"key": "value"`) {
		t.Errorf("Expected json content, got '%s'", result.CodeBlocks[0].Content)
	}
}

func TestExtractSections(t *testing.T) {
	content := `# Main Title
## Section 1
### Subsection 1.1
## Section 2`

	tmpFile := createTempMarkdown(t, content)

	opts := DefaultValidatorOptions()
	validator := NewValidator(opts, &bytes.Buffer{})
	result := validator.ValidateFile(tmpFile)

	if len(result.Sections) != 4 {
		t.Errorf("Expected 4 sections, got %d", len(result.Sections))
	}

	expectedSections := []struct {
		heading string
		level   int
	}{
		{"Main Title", 1},
		{"Section 1", 2},
		{"Subsection 1.1", 3},
		{"Section 2", 2},
	}

	for i, expected := range expectedSections {
		if i >= len(result.Sections) {
			break
		}
		section := result.Sections[i]
		if section.Heading != expected.heading {
			t.Errorf("Section %d: expected heading '%s', got '%s'", i, expected.heading, section.Heading)
		}
		if section.Level != expected.level {
			t.Errorf("Section %d: expected level %d, got %d", i, expected.level, section.Level)
		}
	}
}

func TestValidateDirectory(t *testing.T) {
	// Create temp directory with multiple markdown files
	tmpDir := t.TempDir()

	validMd := filepath.Join(tmpDir, "valid.md")
	os.WriteFile(validMd, []byte("# Valid\n\nContent"), 0644)

	invalidMd := filepath.Join(tmpDir, "invalid.md")
	os.WriteFile(invalidMd, []byte("# Invalid\n```json\n{bad}\n```"), 0644)

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	subMd := filepath.Join(subDir, "sub.md")
	os.WriteFile(subMd, []byte("# Sub\n\nContent"), 0644)

	opts := DefaultValidatorOptions()
	opts.ValidateCodeBlocks = true
	validator := NewValidator(opts, &bytes.Buffer{})

	results, err := validator.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Count valid and invalid
	validCount := 0
	invalidCount := 0
	for _, result := range results {
		if result.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	if validCount != 2 {
		t.Errorf("Expected 2 valid files, got %d", validCount)
	}

	if invalidCount != 1 {
		t.Errorf("Expected 1 invalid file, got %d", invalidCount)
	}
}

func TestValidateDirectory_ExcludeDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create markdown in root
	os.WriteFile(filepath.Join(tmpDir, "root.md"), []byte("# Root"), 0644)

	// Create markdown in excluded directory
	nodeModules := filepath.Join(tmpDir, "node_modules")
	os.Mkdir(nodeModules, 0755)
	os.WriteFile(filepath.Join(nodeModules, "excluded.md"), []byte("# Excluded"), 0644)

	opts := DefaultValidatorOptions()
	opts.ExcludeDirs = []string{"node_modules"}
	validator := NewValidator(opts, &bytes.Buffer{})

	results, err := validator.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result (excluded node_modules), got %d", len(results))
	}

	if !strings.HasSuffix(results[0].FilePath, "root.md") {
		t.Errorf("Expected root.md, got %s", results[0].FilePath)
	}
}

// Helper functions

func createTempMarkdown(t *testing.T, content string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpFile
}
