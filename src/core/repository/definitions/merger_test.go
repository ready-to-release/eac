//go:build L2
// +build L2

package definitions

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateYAMLPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{filepath.Join("definitions.yml"), ""},
		{filepath.Join("test", "definitions.yml"), "test"},
		{filepath.Join("test", "mytest", "definitions.yml"), "test.mytest"},
		{filepath.Join("a", "b", "c", "definitions.yml"), "a.b.c"},
	}

	for _, test := range tests {
		result := generateYAMLPath(test.input)
		if result != test.expected {
			t.Errorf("generateYAMLPath(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestMergeDefinitions(t *testing.T) {
	// Create yaml.Node content for testing
	rootContent := &yaml.Node{Kind: yaml.MappingNode}
	rootContent.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "root"},
		{Kind: yaml.ScalarNode, Value: "value"},
		{Kind: yaml.ScalarNode, Value: "shared"},
		{Kind: yaml.ScalarNode, Value: "root"},
	}

	nestedContent := &yaml.Node{Kind: yaml.MappingNode}
	nestedContent.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "nested"},
		{Kind: yaml.ScalarNode, Value: "value"},
	}

	deepContent := &yaml.Node{Kind: yaml.MappingNode}
	deepContent.Content = []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "deep"},
		{Kind: yaml.ScalarNode, Value: "value"},
	}

	definitions := []DefinitionFile{
		{
			Path:     filepath.Join(filepath.VolumeName("/root"), "root", "definitions.yml"),
			Content:  rootContent,
			YAMLPath: "",
		},
		{
			Path:     filepath.Join(filepath.VolumeName("/root"), "root", "test", "definitions.yml"),
			Content:  nestedContent,
			YAMLPath: "test",
		},
		{
			Path:     filepath.Join(filepath.VolumeName("/root"), "root", "test", "mytest", "definitions.yml"),
			Content:  deepContent,
			YAMLPath: "test.mytest",
		},
	}

	result, err := MergeDefinitions(definitions)
	if err != nil {
		t.Fatalf("MergeDefinitions failed: %v", err)
	}

	// Convert result to map for easier testing
	var resultMap map[string]interface{}
	err = result.Decode(&resultMap)
	if err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	if resultMap["root"] != "value" {
		t.Errorf("Expected root value, got %v", resultMap["root"])
	}

	if resultMap["shared"] != "root" {
		t.Errorf("Expected shared value, got %v", resultMap["shared"])
	}

	testNode, ok := resultMap["test"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected test node to be map[string]interface{}, got %T", resultMap["test"])
	}

	if testNode["nested"] != "value" {
		t.Errorf("Expected nested value, got %v", testNode["nested"])
	}

	mytestNode, ok := testNode["mytest"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mytest node to be map[string]interface{}, got %T", testNode["mytest"])
	}

	if mytestNode["deep"] != "value" {
		t.Errorf("Expected deep value, got %v", mytestNode["deep"])
	}
}

func TestProcessDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "definitions_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	rootDef := `root: value
shared: root`
	err = os.WriteFile(filepath.Join(tempDir, "definitions.yml"), []byte(rootDef), 0644)
	if err != nil {
		t.Fatalf("Failed to write root definition: %v", err)
	}

	testDir := filepath.Join(tempDir, "test")
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	testDef := `nested: value`
	err = os.WriteFile(filepath.Join(testDir, "definitions.yml"), []byte(testDef), 0644)
	if err != nil {
		t.Fatalf("Failed to write test definition: %v", err)
	}

	mytestDir := filepath.Join(testDir, "mytest")
	err = os.MkdirAll(mytestDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create mytest dir: %v", err)
	}

	mytestDef := `deep: value`
	err = os.WriteFile(filepath.Join(mytestDir, "definitions.yml"), []byte(mytestDef), 0644)
	if err != nil {
		t.Fatalf("Failed to write mytest definition: %v", err)
	}

	result, err := ProcessDirectory(tempDir)
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}

	// Convert result to map for easier testing
	var resultMap map[string]interface{}
	err = result.Decode(&resultMap)
	if err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	if resultMap["root"] != "value" {
		t.Errorf("Expected root value, got %v", resultMap["root"])
	}

	testNode, ok := resultMap["test"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected test node to be map[string]interface{}, got %T", resultMap["test"])
	}

	if testNode["nested"] != "value" {
		t.Errorf("Expected nested value, got %v", testNode["nested"])
	}

	mytestNode, ok := testNode["mytest"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mytest node to be map[string]interface{}, got %T", testNode["mytest"])
	}

	if mytestNode["deep"] != "value" {
		t.Errorf("Expected deep value, got %v", mytestNode["deep"])
	}
}

func TestIsTemplateSkeletonPath(t *testing.T) {
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

func TestTemplateSkeletonExclusion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "definitions_template_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a normal definition file that should be processed
	normalDef := `normal: value`
	err = os.WriteFile(filepath.Join(tempDir, "definitions.yml"), []byte(normalDef), 0644)
	if err != nil {
		t.Fatalf("Failed to write normal definition: %v", err)
	}

	// Create a template skeleton directory structure with invalid YAML (Handlebars)
	skeletonDir := filepath.Join(tempDir, "automation", "pwsh-templates", "assets", "templates", "test-template", "skeleton")
	err = os.MkdirAll(skeletonDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create skeleton dir: %v", err)
	}

	// Create a template definition file with Handlebars syntax that would fail YAML parsing
	templateDef := `name: {{PROJECT_NAME}}
tags:
  - ${{ values.ProjectType }}
  {{#each Languages}}- {{this}}
  {{/each}}`
	err = os.WriteFile(filepath.Join(skeletonDir, "definitions.yml"), []byte(templateDef), 0644)
	if err != nil {
		t.Fatalf("Failed to write template definition: %v", err)
	}

	// Process directory - should succeed without parsing the template file
	result, err := ProcessDirectory(tempDir)
	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}

	// Convert result to map for testing
	var resultMap map[string]interface{}
	err = result.Decode(&resultMap)
	if err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	// Normal definition should be processed
	if resultMap["normal"] != "value" {
		t.Errorf("Expected normal value, got %v", resultMap["normal"])
	}

	// Template content should NOT be present in the result
	if _, exists := resultMap["name"]; exists {
		t.Errorf("Template content should not be processed - found 'name' key from template")
	}

	if _, exists := resultMap["tags"]; exists {
		t.Errorf("Template content should not be processed - found 'tags' key from template")
	}
}
