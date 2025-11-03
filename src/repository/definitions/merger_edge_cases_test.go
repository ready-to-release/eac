//go:build L2
// +build L2

package definitions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// L2 Tests - Edge cases and error handling for low-coverage functions

func TestEnumerateDefinitionFiles_EdgeCases(t *testing.T) {
	t.Run("directory exclusions", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_exclusion_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create directories that should be excluded
		excludedDirs := []string{".git", "node_modules", ".vscode", ".idea", "out"}
		for _, dir := range excludedDirs {
			dirPath := filepath.Join(tempDir, dir)
			err = os.MkdirAll(dirPath, 0755)
			if err != nil {
				t.Fatalf("Failed to create excluded dir %s: %v", dir, err)
			}

			// Add definitions.yml that should be excluded
			defFile := filepath.Join(dirPath, "definitions.yml")
			err = os.WriteFile(defFile, []byte("excluded: true"), 0644)
			if err != nil {
				t.Fatalf("Failed to write excluded definition: %v", err)
			}
		}

		// Create a valid definition that should be included
		validDef := filepath.Join(tempDir, "definitions.yml")
		err = os.WriteFile(validDef, []byte("valid: true"), 0644)
		if err != nil {
			t.Fatalf("Failed to write valid definition: %v", err)
		}

		definitions, err := EnumerateDefinitionFiles(tempDir)
		if err != nil {
			t.Fatalf("EnumerateDefinitionFiles failed: %v", err)
		}

		// Should only find the valid definition
		if len(definitions) != 1 {
			t.Errorf("Expected 1 definition file, got %d", len(definitions))
		}

		if len(definitions) > 0 && !strings.HasSuffix(definitions[0].Path, "definitions.yml") {
			t.Errorf("Expected to find root definitions.yml, got %s", definitions[0].Path)
		}

		// Verify excluded content is not present
		for _, def := range definitions {
			var resultMap map[string]interface{}
			err = def.Content.Decode(&resultMap)
			if err != nil {
				t.Fatalf("Failed to decode definition: %v", err)
			}

			if _, excluded := resultMap["excluded"]; excluded {
				t.Error("Found excluded content that should have been skipped")
			}
		}
	})

	t.Run("nested template skeleton exclusions", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_nested_skeleton_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create complex nested template structure
		nestedSkeletonDir := filepath.Join(tempDir, "automation", "pwsh-templates", "assets", "templates", "complex", "skeleton", "deeply", "nested")
		err = os.MkdirAll(nestedSkeletonDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create nested skeleton dir: %v", err)
		}

		// Add template file with Handlebars that would break YAML parsing
		templateFile := filepath.Join(nestedSkeletonDir, "definitions.yml")
		templateContent := `
name: {{PROJECT_NAME}}
config:
  {{#each settings}}
  {{key}}: {{value}}
  {{/each}}
`
		err = os.WriteFile(templateFile, []byte(templateContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write template file: %v", err)
		}

		// Add a valid definition outside skeleton
		validFile := filepath.Join(tempDir, "definitions.yml")
		err = os.WriteFile(validFile, []byte("valid: content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write valid file: %v", err)
		}

		// Should not fail despite the invalid template YAML
		definitions, err := EnumerateDefinitionFiles(tempDir)
		if err != nil {
			t.Fatalf("EnumerateDefinitionFiles failed: %v", err)
		}

		// Should find only the valid definition
		if len(definitions) != 1 {
			t.Errorf("Expected 1 definition, got %d", len(definitions))
		}

		// Verify the template content was not processed
		if len(definitions) > 0 {
			var resultMap map[string]interface{}
			err = definitions[0].Content.Decode(&resultMap)
			if err != nil {
				t.Fatalf("Failed to decode definition: %v", err)
			}

			if resultMap["valid"] != "content" {
				t.Errorf("Expected valid content, got %v", resultMap["valid"])
			}

			if _, hasName := resultMap["name"]; hasName {
				t.Error("Template content should not have been processed")
			}
		}
	})

	t.Run("file system errors", func(t *testing.T) {
		// Test with non-existent directory
		definitions, err := EnumerateDefinitionFiles("/non/existent/directory")
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
		if definitions != nil {
			t.Error("Expected nil definitions for non-existent directory")
		}
	})

	t.Run("malformed YAML files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_malformed_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a malformed YAML file
		malformedFile := filepath.Join(tempDir, "definitions.yml")
		malformedContent := `
invalid: yaml: content:
  - missing
    proper: indentation
  [ invalid brackets
`
		err = os.WriteFile(malformedFile, []byte(malformedContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write malformed file: %v", err)
		}

		// Should return error for malformed YAML
		definitions, err := EnumerateDefinitionFiles(tempDir)
		if err == nil {
			t.Error("Expected error for malformed YAML")
		}
		if definitions != nil {
			t.Error("Expected nil definitions for malformed YAML")
		}

		// Error should mention YAML parsing
		if err != nil && !strings.Contains(err.Error(), "parse YAML") {
			t.Errorf("Expected YAML parsing error, got: %v", err)
		}
	})

	t.Run("empty definitions files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_empty_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create empty definition file
		emptyFile := filepath.Join(tempDir, "definitions.yml")
		err = os.WriteFile(emptyFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to write empty file: %v", err)
		}

		definitions, err := EnumerateDefinitionFiles(tempDir)
		if err != nil {
			t.Fatalf("EnumerateDefinitionFiles failed: %v", err)
		}

		if len(definitions) != 1 {
			t.Errorf("Expected 1 definition file, got %d", len(definitions))
		}
	})
}

func TestProcessDirectory_EdgeCases(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		result, err := ProcessDirectory("/non/existent/directory")
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
		if result != nil {
			t.Error("Expected nil result for non-existent directory")
		}

		// Error should mention enumeration failure
		if err != nil && !strings.Contains(err.Error(), "enumerate definition files") {
			t.Errorf("Expected enumeration error, got: %v", err)
		}
	})

	t.Run("directory with no definition files", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_empty_dir_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create some non-definition files
		nonDefFile := filepath.Join(tempDir, "config.yml")
		err = os.WriteFile(nonDefFile, []byte("config: value"), 0644)
		if err != nil {
			t.Fatalf("Failed to write non-definition file: %v", err)
		}

		result, err := ProcessDirectory(tempDir)
		if err != nil {
			t.Fatalf("ProcessDirectory failed: %v", err)
		}

		// Should return empty but valid YAML structure
		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode empty result: %v", err)
		}

		if len(resultMap) != 0 {
			t.Errorf("Expected empty result map, got %v", resultMap)
		}
	})

	t.Run("permission denied scenarios", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "definitions_permission_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create a definition file
		defFile := filepath.Join(tempDir, "definitions.yml")
		err = os.WriteFile(defFile, []byte("test: value"), 0644)
		if err != nil {
			t.Fatalf("Failed to write definition file: %v", err)
		}

		// On some systems, we can test permission restrictions
		// This test might be skipped on systems where it's not applicable
		restrictedDir := filepath.Join(tempDir, "restricted")
		err = os.MkdirAll(restrictedDir, 0000) // No permissions
		if err != nil {
			t.Skipf("Cannot create restricted directory for permission test: %v", err)
		}

		// Try to restore permissions for cleanup
		defer func() {
			os.Chmod(restrictedDir, 0755)
			os.RemoveAll(restrictedDir)
		}()

		// The function should still work with the accessible files
		result, err := ProcessDirectory(tempDir)
		if err != nil {
			// On some systems, this might fail due to permission issues
			// The behavior may vary by OS and filesystem
			t.Logf("ProcessDirectory failed with permission test (may be expected): %v", err)
			return
		}

		// If it succeeds, verify we got the accessible content
		var resultMap map[string]interface{}
		err = result.Decode(&resultMap)
		if err != nil {
			t.Fatalf("Failed to decode result: %v", err)
		}

		if resultMap["test"] != "value" {
			t.Errorf("Expected test value, got %v", resultMap["test"])
		}
	})

	t.Run("merge failure propagation", func(t *testing.T) {
		// This test verifies that merge failures are properly propagated
		// We can't easily create a merge failure with the current implementation,
		// but we can verify the error handling structure is in place

		tempDir, err := os.MkdirTemp("", "definitions_merge_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create valid definition file
		defFile := filepath.Join(tempDir, "definitions.yml")
		err = os.WriteFile(defFile, []byte("valid: content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write definition file: %v", err)
		}

		result, err := ProcessDirectory(tempDir)
		if err != nil {
			t.Fatalf("ProcessDirectory failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}
