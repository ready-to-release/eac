//go:build L1 && ov
// +build L1,ov

package render

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

// ---- Test types for TOML round-trip testing ----

type tomlTestPerson struct {
	Name  string `yaml:"name"`
	Age   int    `yaml:"age"`
	Email string `yaml:"email,omitempty"`
}

type tomlTestConfig struct {
	Host    string   `yaml:"host"`
	Port    int      `yaml:"port"`
	Enabled bool     `yaml:"enabled"`
	Tags    []string `yaml:"tags"`
}

type tomlTestNested struct {
	Name    string       `yaml:"name"`
	Address tomlTestAddr `yaml:"address"`
}

type tomlTestAddr struct {
	Street string `yaml:"street"`
	City   string `yaml:"city"`
}

// ---- RenderAsTOML tests ----

func TestRenderAsTOML_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantKeys []string
	}{
		{
			name: "simple struct with string and int",
			input: tomlTestPerson{
				Name: "Alice",
				Age:  30,
			},
			wantKeys: []string{"name", "age"},
		},
		{
			name: "struct with all fields populated",
			input: tomlTestPerson{
				Name:  "Bob",
				Age:   25,
				Email: "bob@example.com",
			},
			wantKeys: []string{"name", "age", "email"},
		},
		{
			name: "struct with boolean and slice",
			input: tomlTestConfig{
				Host:    "localhost",
				Port:    8080,
				Enabled: true,
				Tags:    []string{"dev", "test"},
			},
			wantKeys: []string{"host", "port", "enabled", "tags"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderAsTOML(tt.input)
			if err != nil {
				t.Fatalf("RenderAsTOML() returned error: %v", err)
			}

			if got == "" {
				t.Fatal("RenderAsTOML() returned empty string")
			}

			// Verify the output is valid TOML by parsing it back
			var parsed map[string]interface{}
			if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("RenderAsTOML() output is not valid TOML: %v\nOutput:\n%s", err, got)
			}

			// Verify all expected keys are present
			for _, key := range tt.wantKeys {
				if _, exists := parsed[key]; !exists {
					t.Errorf("expected key %q not found in TOML output.\nParsed: %v\nRaw:\n%s", key, parsed, got)
				}
			}
		})
	}
}

func TestRenderAsTOML_RoundTripValues(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantPairs map[string]interface{}
	}{
		{
			name:  "string and int values preserved",
			input: tomlTestPerson{Name: "Alice", Age: 30},
			wantPairs: map[string]interface{}{
				"name": "Alice",
				"age":  int64(30), // TOML integers unmarshal as int64
			},
		},
		{
			name:  "boolean value preserved",
			input: tomlTestConfig{Host: "db.local", Port: 5432, Enabled: false, Tags: []string{}},
			wantPairs: map[string]interface{}{
				"host":    "db.local",
				"port":    int64(5432),
				"enabled": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderAsTOML(tt.input)
			if err != nil {
				t.Fatalf("RenderAsTOML() returned error: %v", err)
			}

			var parsed map[string]interface{}
			if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("output is not valid TOML: %v\nOutput:\n%s", err, got)
			}

			for key, wantVal := range tt.wantPairs {
				gotVal, exists := parsed[key]
				if !exists {
					t.Errorf("key %q not found in TOML output", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("key %q: got %v (%T), want %v (%T)", key, gotVal, gotVal, wantVal, wantVal)
				}
			}
		})
	}
}

func TestRenderAsTOML_SliceValues(t *testing.T) {
	input := tomlTestConfig{
		Host:    "localhost",
		Port:    8080,
		Enabled: true,
		Tags:    []string{"alpha", "beta", "gamma"},
	}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid TOML: %v\nOutput:\n%s", err, got)
	}

	tags, ok := parsed["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected tags to be a slice, got %T", parsed["tags"])
	}

	expectedTags := []string{"alpha", "beta", "gamma"}
	if len(tags) != len(expectedTags) {
		t.Fatalf("expected %d tags, got %d", len(expectedTags), len(tags))
	}

	for i, expected := range expectedTags {
		if tags[i] != expected {
			t.Errorf("tag[%d]: got %v, want %v", i, tags[i], expected)
		}
	}
}

func TestRenderAsTOML_NestedStruct(t *testing.T) {
	input := tomlTestNested{
		Name: "Alice",
		Address: tomlTestAddr{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	// Verify valid TOML
	var parsed map[string]interface{}
	if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid TOML: %v\nOutput:\n%s", err, got)
	}

	if parsed["name"] != "Alice" {
		t.Errorf("name: got %v, want Alice", parsed["name"])
	}

	// Nested struct should produce a TOML table
	addr, ok := parsed["address"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected address to be a TOML table (map), got %T", parsed["address"])
	}

	if addr["street"] != "123 Main St" {
		t.Errorf("address.street: got %v, want '123 Main St'", addr["street"])
	}
	if addr["city"] != "NYC" {
		t.Errorf("address.city: got %v, want 'NYC'", addr["city"])
	}
}

func TestRenderAsTOML_OmitEmpty(t *testing.T) {
	input := tomlTestPerson{
		Name: "Alice",
		Age:  30,
		// Email is omitempty and left empty
	}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	// YAML omitempty should cause the field to be absent from the intermediate representation
	if strings.Contains(got, "email") {
		t.Errorf("expected email to be omitted from TOML output, got:\n%s", got)
	}
}

func TestRenderAsTOML_ContainsTOMLSyntax(t *testing.T) {
	input := tomlTestPerson{Name: "Test", Age: 1}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	// TOML uses key = value syntax
	if !strings.Contains(got, "=") {
		t.Errorf("expected TOML key=value syntax, got:\n%s", got)
	}

	// Should contain the actual values
	if !strings.Contains(got, "Test") {
		t.Errorf("expected 'Test' in TOML output, got:\n%s", got)
	}
}

// ---- RenderAsTOMLOrPanic tests ----

func TestRenderAsTOMLOrPanic_Success(t *testing.T) {
	input := tomlTestPerson{Name: "Test", Age: 42}

	result := RenderAsTOMLOrPanic(input)

	if !strings.Contains(result, "Test") {
		t.Errorf("expected output to contain 'Test', got: %s", result)
	}
}

func TestRenderAsTOMLOrPanic_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected RenderAsTOMLOrPanic to panic on invalid input, but it did not")
		}
	}()

	// Channels cannot be marshaled to YAML
	ch := make(chan int)
	RenderAsTOMLOrPanic(ch)
}

// ---- TOML-specific format tests ----

func TestRenderAsTOML_EmptySliceField(t *testing.T) {
	input := tomlTestConfig{
		Host:    "localhost",
		Port:    80,
		Enabled: true,
		Tags:    []string{},
	}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	// Should still produce valid TOML
	var parsed map[string]interface{}
	if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid TOML: %v\nOutput:\n%s", err, got)
	}
}

func TestRenderAsTOML_SpecialCharacters(t *testing.T) {
	input := tomlTestPerson{
		Name:  `O'Brien "The Great"`,
		Age:   40,
		Email: "o.brien@example.com",
	}

	got, err := RenderAsTOML(input)
	if err != nil {
		t.Fatalf("RenderAsTOML() returned error: %v", err)
	}

	// Verify the TOML is still valid with special characters
	var parsed map[string]interface{}
	if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid TOML with special chars: %v\nOutput:\n%s", err, got)
	}

	// The parsed name should match the original (TOML handles escaping)
	if parsed["name"] != `O'Brien "The Great"` {
		t.Errorf("name with special chars: got %v, want %v", parsed["name"], `O'Brien "The Great"`)
	}
}
