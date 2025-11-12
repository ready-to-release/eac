package render

import (
	"strings"
	"testing"
)

type TestPerson struct {
	Name  string `yaml:"name"`
	Age   int    `yaml:"age"`
	Email string `yaml:"email,omitempty"`
}

type TestConfig struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Enabled  bool     `yaml:"enabled"`
	Tags     []string `yaml:"tags"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

func TestRenderStructAsMarkdown(t *testing.T) {
	person := TestPerson{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	result, err := RenderStructAsMarkdown(person)
	if err != nil {
		t.Fatalf("RenderStructAsMarkdown failed: %v", err)
	}

	// Verify it's a YAML code block
	if !strings.HasPrefix(result, "```yaml\n") {
		t.Errorf("Expected result to start with ```yaml\\n, got: %s", result)
	}

	if !strings.HasSuffix(result, "```") {
		t.Errorf("Expected result to end with ```, got: %s", result)
	}

	// Verify content
	if !strings.Contains(result, "name: Alice") {
		t.Errorf("Expected to find 'name: Alice', got: %s", result)
	}

	if !strings.Contains(result, "age: 30") {
		t.Errorf("Expected to find 'age: 30', got: %s", result)
	}

	if !strings.Contains(result, "email: alice@example.com") {
		t.Errorf("Expected to find 'email: alice@example.com', got: %s", result)
	}
}

func TestRenderStructWithTitle(t *testing.T) {
	person := TestPerson{
		Name: "Bob",
		Age:  25,
	}

	result, err := RenderStructWithTitle("Person Details", person)
	if err != nil {
		t.Fatalf("RenderStructWithTitle failed: %v", err)
	}

	// Verify title
	if !strings.Contains(result, "## Person Details") {
		t.Errorf("Expected to find '## Person Details', got: %s", result)
	}

	// Verify YAML block
	if !strings.Contains(result, "```yaml") {
		t.Errorf("Expected YAML code block, got: %s", result)
	}

	// Verify content
	if !strings.Contains(result, "name: Bob") {
		t.Errorf("Expected to find 'name: Bob', got: %s", result)
	}
}

func TestRenderStructSliceAsMarkdown(t *testing.T) {
	people := []TestPerson{
		{Name: "Alice", Age: 30, Email: "alice@example.com"},
		{Name: "Bob", Age: 25, Email: "bob@example.com"},
		{Name: "Charlie", Age: 35},
	}

	result, err := RenderStructSliceAsMarkdown(people)
	if err != nil {
		t.Fatalf("RenderStructSliceAsMarkdown failed: %v", err)
	}

	// Verify it's a YAML list
	if !strings.Contains(result, "- name: Alice") {
		t.Errorf("Expected YAML list format, got: %s", result)
	}

	if !strings.Contains(result, "- name: Bob") {
		t.Errorf("Expected Bob in list, got: %s", result)
	}

	if !strings.Contains(result, "- name: Charlie") {
		t.Errorf("Expected Charlie in list, got: %s", result)
	}
}

func TestRenderStructWithComplexTypes(t *testing.T) {
	config := TestConfig{
		Host:    "localhost",
		Port:    8080,
		Enabled: true,
		Tags:    []string{"dev", "test", "staging"},
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "development",
		},
	}

	result, err := RenderStructAsMarkdown(config)
	if err != nil {
		t.Fatalf("RenderStructAsMarkdown with complex types failed: %v", err)
	}

	// Verify scalar fields
	if !strings.Contains(result, "host: localhost") {
		t.Errorf("Expected 'host: localhost', got: %s", result)
	}

	if !strings.Contains(result, "port: 8080") {
		t.Errorf("Expected 'port: 8080', got: %s", result)
	}

	if !strings.Contains(result, "enabled: true") {
		t.Errorf("Expected 'enabled: true', got: %s", result)
	}

	// Verify array
	if !strings.Contains(result, "tags:") {
		t.Errorf("Expected 'tags:', got: %s", result)
	}

	// Verify map
	if !strings.Contains(result, "metadata:") {
		t.Errorf("Expected 'metadata:', got: %s", result)
	}
}

func TestRenderMultipleStructs(t *testing.T) {
	person := TestPerson{Name: "Alice", Age: 30}
	config := TestConfig{Host: "localhost", Port: 8080, Enabled: true}

	sections := map[string]interface{}{
		"Person": person,
		"Config": config,
	}

	result, err := RenderMultipleStructs(sections)
	if err != nil {
		t.Fatalf("RenderMultipleStructs failed: %v", err)
	}

	// Verify both sections exist (order may vary due to map iteration)
	if !strings.Contains(result, "## Person") && !strings.Contains(result, "## Config") {
		t.Errorf("Expected both sections, got: %s", result)
	}

	// Verify content from both structs
	if !strings.Contains(result, "name: Alice") {
		t.Errorf("Expected person data, got: %s", result)
	}

	if !strings.Contains(result, "host: localhost") {
		t.Errorf("Expected config data, got: %s", result)
	}
}

func TestRenderStructAsMarkdownOrPanic(t *testing.T) {
	person := TestPerson{Name: "Test", Age: 99}

	// Should not panic for valid struct
	result := RenderStructAsMarkdownOrPanic(person)

	if !strings.Contains(result, "name: Test") {
		t.Errorf("Expected valid output, got: %s", result)
	}
}

func TestRenderEmptyStruct(t *testing.T) {
	empty := TestPerson{}

	result, err := RenderStructAsMarkdown(empty)
	if err != nil {
		t.Fatalf("Failed to render empty struct: %v", err)
	}

	// Should still produce valid YAML
	if !strings.Contains(result, "```yaml") {
		t.Errorf("Expected YAML block for empty struct, got: %s", result)
	}
}

func TestRenderStructWithOmitEmpty(t *testing.T) {
	person := TestPerson{
		Name: "Test",
		Age:  30,
		// Email omitted
	}

	result, err := RenderStructAsMarkdown(person)
	if err != nil {
		t.Fatalf("Failed to render struct: %v", err)
	}

	// Email should not appear due to omitempty tag
	if strings.Contains(result, "email:") {
		t.Errorf("Expected email to be omitted, got: %s", result)
	}

	// But name and age should be present
	if !strings.Contains(result, "name: Test") || !strings.Contains(result, "age: 30") {
		t.Errorf("Expected name and age to be present, got: %s", result)
	}
}

func TestRenderNestedStruct(t *testing.T) {
	type Address struct {
		Street string `yaml:"street"`
		City   string `yaml:"city"`
	}

	type PersonWithAddress struct {
		Name    string  `yaml:"name"`
		Address Address `yaml:"address"`
	}

	person := PersonWithAddress{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	result, err := RenderStructAsMarkdown(person)
	if err != nil {
		t.Fatalf("Failed to render nested struct: %v", err)
	}

	// Verify nested structure
	if !strings.Contains(result, "address:") {
		t.Errorf("Expected nested address field, got: %s", result)
	}

	if !strings.Contains(result, "street: 123 Main St") {
		t.Errorf("Expected street in nested struct, got: %s", result)
	}

	if !strings.Contains(result, "city: NYC") {
		t.Errorf("Expected city in nested struct, got: %s", result)
	}
}
