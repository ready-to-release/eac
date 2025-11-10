//go:build L0
// +build L0

package definitions

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// L0 Tests - Pure functions, no I/O, under 500ms

func TestGenerateYAMLPath_L0(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"root level", "definitions.yml", ""},
		{"single directory", "test/definitions.yml", "test"},
		{"nested directories", "test/mytest/definitions.yml", "test.mytest"},
		{"deep nesting", "a/b/c/definitions.yml", "a.b.c"},
		{"windows paths", "test\\mytest\\definitions.yml", "test.mytest"},
		{"mixed separators", "test/mytest\\deep/definitions.yml", "test.mytest.deep"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateYAMLPath(tt.input)
			if result != tt.expected {
				t.Errorf("generateYAMLPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsTemplateSkeletonPath_L0(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		rootDir  string
		expected bool
	}{
		{
			name:     "template skeleton path should be excluded",
			path:     "/projects/r2r-cli/automation/pwsh-templates/assets/templates/root-configurations/skeleton",
			rootDir:  "/projects/r2r-cli",
			expected: true,
		},
		{
			name:     "template skeleton subdirectory should be excluded",
			path:     "/projects/r2r-cli/automation/pwsh-templates/assets/templates/root-configurations/skeleton/subdir",
			rootDir:  "/projects/r2r-cli",
			expected: true,
		},
		{
			name:     "normal directory should not be excluded",
			path:     "/projects/r2r-cli/src/internal/definitions",
			rootDir:  "/projects/r2r-cli",
			expected: false,
		},
		{
			name:     "templates directory without skeleton should not be excluded",
			path:     "/projects/r2r-cli/automation/pwsh-templates/assets/templates/root-configurations",
			rootDir:  "/projects/r2r-cli",
			expected: false,
		},
		{
			name:     "skeleton directory not under templates should not be excluded",
			path:     "/projects/r2r-cli/some/skeleton/dir",
			rootDir:  "/projects/r2r-cli",
			expected: false,
		},
		{
			name:     "Windows path - template skeleton should be excluded",
			path:     "C:\\projects\\r2r-cli\\automation\\pwsh-templates\\assets\\templates\\r2r-cli\\skeleton",
			rootDir:  "C:\\projects\\r2r-cli",
			expected: true,
		},
		{
			name:     "relative path error handling",
			path:     "invalid:/path",
			rootDir:  "/valid/root",
			expected: false,
		},
		{
			name:     "empty path components",
			path:     "/projects//templates///skeleton",
			rootDir:  "/projects",
			expected: false, // Not a valid templates/*/skeleton pattern
		},
		{
			name:     "case sensitivity test",
			path:     "/projects/Templates/test/Skeleton",
			rootDir:  "/projects",
			expected: false, // Case sensitive matching
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTemplateSkeletonPath(tc.path, tc.rootDir)
			if result != tc.expected {
				t.Errorf("isTemplateSkeletonPath(%q, %q) = %v, expected %v", tc.path, tc.rootDir, result, tc.expected)
			}
		})
	}
}

func TestSetNestedValue_L0(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantKeys []string
	}{
		{
			name:     "single level",
			path:     "test",
			wantKeys: []string{"test"},
		},
		{
			name:     "nested path",
			path:     "test.mytest",
			wantKeys: []string{"test", "mytest"},
		},
		{
			name:     "deep nesting",
			path:     "a.b.c.d",
			wantKeys: []string{"a", "b", "c", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootMap := &yaml.Node{Kind: yaml.MappingNode}
			value := &yaml.Node{Kind: yaml.ScalarNode, Value: "testvalue"}

			setNestedValue(rootMap, tt.path, value)

			// Navigate and verify the structure was created correctly
			current := rootMap
			for i, key := range tt.wantKeys {
				found := false
				for j := 0; j < len(current.Content); j += 2 {
					if current.Content[j].Value == key {
						found = true
						if i == len(tt.wantKeys)-1 {
							// Final value should be our test value
							if current.Content[j+1].Value != "testvalue" {
								t.Errorf("Final value = %q, want %q", current.Content[j+1].Value, "testvalue")
							}
						} else {
							// Intermediate should be a mapping
							if current.Content[j+1].Kind != yaml.MappingNode {
								t.Errorf("Intermediate node should be MappingNode, got %v", current.Content[j+1].Kind)
							}
							current = current.Content[j+1]
						}
						break
					}
				}
				if !found {
					t.Errorf("Key %q not found at level %d", key, i)
					break
				}
			}
		})
	}
}

func TestAddToMapping_L0(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		valueType yaml.Kind
		valueStr  string
	}{
		{
			name:      "scalar value",
			key:       "testkey",
			valueType: yaml.ScalarNode,
			valueStr:  "testvalue",
		},
		{
			name:      "empty key",
			key:       "",
			valueType: yaml.ScalarNode,
			valueStr:  "value",
		},
		{
			name:      "special characters in key",
			key:       "key-with-dashes_and_underscores.and.dots",
			valueType: yaml.ScalarNode,
			valueStr:  "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := &yaml.Node{Kind: yaml.MappingNode}
			value := &yaml.Node{Kind: tt.valueType, Value: tt.valueStr}

			addToMapping(mapping, tt.key, value)

			// Verify the key-value pair was added
			if len(mapping.Content) != 2 {
				t.Errorf("Expected 2 content nodes (key+value), got %d", len(mapping.Content))
				return
			}

			if mapping.Content[0].Value != tt.key {
				t.Errorf("Key = %q, want %q", mapping.Content[0].Value, tt.key)
			}

			if mapping.Content[1].Value != tt.valueStr {
				t.Errorf("Value = %q, want %q", mapping.Content[1].Value, tt.valueStr)
			}
		})
	}
}

func TestAddToMapping_DocumentNode_L0(t *testing.T) {
	mapping := &yaml.Node{Kind: yaml.MappingNode}

	// Create a document node containing a scalar
	innerScalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "inner"}
	documentValue := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{innerScalar},
	}

	addToMapping(mapping, "doc", documentValue)

	// Should unwrap the document and use the inner content
	if len(mapping.Content) != 2 {
		t.Errorf("Expected 2 content nodes, got %d", len(mapping.Content))
		return
	}

	if mapping.Content[1].Value != "inner" {
		t.Errorf("Value = %q, want %q", mapping.Content[1].Value, "inner")
	}
}

func TestGetOrCreateMapping_L0(t *testing.T) {
	t.Run("create new mapping", func(t *testing.T) {
		mapping := &yaml.Node{Kind: yaml.MappingNode}

		result := getOrCreateMapping(mapping, "newkey")

		if result.Kind != yaml.MappingNode {
			t.Errorf("Result should be MappingNode, got %v", result.Kind)
		}

		if len(mapping.Content) != 2 {
			t.Errorf("Expected 2 content nodes, got %d", len(mapping.Content))
			return
		}

		if mapping.Content[0].Value != "newkey" {
			t.Errorf("Key = %q, want %q", mapping.Content[0].Value, "newkey")
		}

		if mapping.Content[1] != result {
			t.Error("Returned mapping should be the same as stored mapping")
		}
	})

	t.Run("find existing mapping", func(t *testing.T) {
		mapping := &yaml.Node{Kind: yaml.MappingNode}
		existingValue := &yaml.Node{Kind: yaml.MappingNode}

		// Pre-populate with existing key
		mapping.Content = []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "existing"},
			existingValue,
		}

		result := getOrCreateMapping(mapping, "existing")

		if result != existingValue {
			t.Error("Should return existing mapping, not create new one")
		}

		// Should not have added new content
		if len(mapping.Content) != 2 {
			t.Errorf("Expected 2 content nodes, got %d", len(mapping.Content))
		}
	})

	t.Run("multiple keys", func(t *testing.T) {
		mapping := &yaml.Node{Kind: yaml.MappingNode}

		// Add multiple keys
		first := getOrCreateMapping(mapping, "first")
		second := getOrCreateMapping(mapping, "second")

		if len(mapping.Content) != 4 {
			t.Errorf("Expected 4 content nodes (2 key-value pairs), got %d", len(mapping.Content))
		}

		if first == second {
			t.Error("Different keys should return different mappings")
		}

		// Find existing should still work
		foundFirst := getOrCreateMapping(mapping, "first")
		if foundFirst != first {
			t.Error("Should find existing first mapping")
		}
	})
}
