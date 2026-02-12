//go:build L1 && ov
// +build L1,ov

package render

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---- Test types for JSON round-trip testing ----

type jsonTestPerson struct {
	Name  string `yaml:"name" json:"name"`
	Age   int    `yaml:"age" json:"age"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type jsonTestConfig struct {
	Host    string   `yaml:"host" json:"host"`
	Port    int      `yaml:"port" json:"port"`
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Tags    []string `yaml:"tags" json:"tags"`
}

type jsonTestNested struct {
	Name    string         `yaml:"name" json:"name"`
	Address jsonTestAddr   `yaml:"address" json:"address"`
	Scores  []int          `yaml:"scores" json:"scores"`
}

type jsonTestAddr struct {
	Street string `yaml:"street" json:"street"`
	City   string `yaml:"city" json:"city"`
	Zip    string `yaml:"zip" json:"zip"`
}

// ---- RenderAsJSON tests ----

func TestRenderAsJSON_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		// fields expected in the JSON output
		wantKeys []string
	}{
		{
			name: "simple struct with string and int",
			input: jsonTestPerson{
				Name: "Alice",
				Age:  30,
			},
			wantKeys: []string{"name", "age"},
		},
		{
			name: "struct with all fields populated",
			input: jsonTestPerson{
				Name:  "Bob",
				Age:   25,
				Email: "bob@example.com",
			},
			wantKeys: []string{"name", "age", "email"},
		},
		{
			name: "struct with boolean and slice",
			input: jsonTestConfig{
				Host:    "localhost",
				Port:    8080,
				Enabled: true,
				Tags:    []string{"dev", "test"},
			},
			wantKeys: []string{"host", "port", "enabled", "tags"},
		},
		{
			name: "nested struct",
			input: jsonTestNested{
				Name: "Charlie",
				Address: jsonTestAddr{
					Street: "123 Main St",
					City:   "NYC",
					Zip:    "10001",
				},
				Scores: []int{95, 87, 92},
			},
			wantKeys: []string{"name", "address", "scores"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderAsJSON(tt.input)
			if err != nil {
				t.Fatalf("RenderAsJSON() returned error: %v", err)
			}

			if got == "" {
				t.Fatal("RenderAsJSON() returned empty string")
			}

			// Verify the output is valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("RenderAsJSON() output is not valid JSON: %v\nOutput:\n%s", err, got)
			}

			// Verify all expected keys are present
			for _, key := range tt.wantKeys {
				if _, exists := parsed[key]; !exists {
					t.Errorf("expected key %q not found in JSON output.\nParsed: %v\nRaw:\n%s", key, parsed, got)
				}
			}
		})
	}
}

func TestRenderAsJSON_RoundTripValues(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantPairs map[string]interface{}
	}{
		{
			name:  "string and int values preserved",
			input: jsonTestPerson{Name: "Alice", Age: 30},
			wantPairs: map[string]interface{}{
				"name": "Alice",
				"age":  float64(30), // JSON numbers unmarshal as float64
			},
		},
		{
			name:  "boolean value preserved",
			input: jsonTestConfig{Host: "db.local", Port: 5432, Enabled: false},
			wantPairs: map[string]interface{}{
				"host":    "db.local",
				"port":    float64(5432),
				"enabled": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderAsJSON(tt.input)
			if err != nil {
				t.Fatalf("RenderAsJSON() returned error: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}

			for key, wantVal := range tt.wantPairs {
				gotVal, exists := parsed[key]
				if !exists {
					t.Errorf("key %q not found in JSON output", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("key %q: got %v (%T), want %v (%T)", key, gotVal, gotVal, wantVal, wantVal)
				}
			}
		})
	}
}

func TestRenderAsJSON_SliceValues(t *testing.T) {
	input := jsonTestConfig{
		Host:    "localhost",
		Port:    8080,
		Enabled: true,
		Tags:    []string{"alpha", "beta", "gamma"},
	}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
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

func TestRenderAsJSON_NestedStructRoundTrip(t *testing.T) {
	input := jsonTestNested{
		Name: "Alice",
		Address: jsonTestAddr{
			Street: "456 Elm Ave",
			City:   "Boston",
			Zip:    "02101",
		},
		Scores: []int{100, 90, 85},
	}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify nested address
	addr, ok := parsed["address"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected address to be a map, got %T", parsed["address"])
	}

	if addr["street"] != "456 Elm Ave" {
		t.Errorf("address.street: got %v, want %v", addr["street"], "456 Elm Ave")
	}
	if addr["city"] != "Boston" {
		t.Errorf("address.city: got %v, want %v", addr["city"], "Boston")
	}
	if addr["zip"] != "02101" {
		t.Errorf("address.zip: got %v, want %v", addr["zip"], "02101")
	}

	// Verify nested scores
	scores, ok := parsed["scores"].([]interface{})
	if !ok {
		t.Fatalf("expected scores to be a slice, got %T", parsed["scores"])
	}
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
}

func TestRenderAsJSON_Indentation(t *testing.T) {
	input := jsonTestPerson{Name: "Test", Age: 1}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	// The output should be indented (multi-line)
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Errorf("expected indented (multi-line) JSON, got single line: %s", got)
	}
}

func TestRenderAsJSON_OmitEmpty(t *testing.T) {
	input := jsonTestPerson{
		Name: "Alice",
		Age:  30,
		// Email is omitempty and left empty
	}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	// YAML omitempty should cause the field to be absent
	if strings.Contains(got, "email") {
		t.Errorf("expected email to be omitted from JSON, got:\n%s", got)
	}
}

func TestRenderAsJSON_EmptySlice(t *testing.T) {
	input := jsonTestConfig{
		Host:    "localhost",
		Port:    80,
		Enabled: true,
		Tags:    []string{},
	}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	// Should still produce valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput:\n%s", err, got)
	}
}

// ---- RenderAsJSONOrPanic tests ----

func TestRenderAsJSONOrPanic_Success(t *testing.T) {
	input := jsonTestPerson{Name: "Test", Age: 42}

	// Should not panic for valid input
	result := RenderAsJSONOrPanic(input)

	if !strings.Contains(result, "Test") {
		t.Errorf("expected output to contain 'Test', got: %s", result)
	}
}

func TestRenderAsJSONOrPanic_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected RenderAsJSONOrPanic to panic on invalid input, but it did not")
		}
	}()

	// Channels cannot be marshaled to YAML
	ch := make(chan int)
	RenderAsJSONOrPanic(ch)
}

// ---- OrderedMap tests ----

func TestOrderedMap_PreservesInsertionOrder(t *testing.T) {
	om := newOrderedMap()
	om.Set("zebra", "z")
	om.Set("alpha", "a")
	om.Set("middle", "m")

	got, err := om.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() returned error: %v", err)
	}

	result := string(got)

	// Verify the order: zebra should come before alpha, alpha before middle
	zebraIdx := strings.Index(result, "zebra")
	alphaIdx := strings.Index(result, "alpha")
	middleIdx := strings.Index(result, "middle")

	if zebraIdx < 0 || alphaIdx < 0 || middleIdx < 0 {
		t.Fatalf("missing keys in output: %s", result)
	}

	if zebraIdx >= alphaIdx {
		t.Errorf("expected zebra before alpha, got zebra at %d, alpha at %d", zebraIdx, alphaIdx)
	}
	if alphaIdx >= middleIdx {
		t.Errorf("expected alpha before middle, got alpha at %d, middle at %d", alphaIdx, middleIdx)
	}
}

func TestOrderedMap_SetOverwriteKeepsOrder(t *testing.T) {
	om := newOrderedMap()
	om.Set("first", "1")
	om.Set("second", "2")
	om.Set("first", "updated") // overwrite existing key

	got, err := om.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() returned error: %v", err)
	}

	result := string(got)

	// first should still come before second
	firstIdx := strings.Index(result, "first")
	secondIdx := strings.Index(result, "second")

	if firstIdx >= secondIdx {
		t.Errorf("expected first before second after overwrite, got first at %d, second at %d", firstIdx, secondIdx)
	}

	// Verify updated value
	if !strings.Contains(result, "updated") {
		t.Errorf("expected overwritten value 'updated', got: %s", result)
	}
}

func TestOrderedMap_EmptyMap(t *testing.T) {
	om := newOrderedMap()

	got, err := om.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() returned error: %v", err)
	}

	if string(got) != "{}" {
		t.Errorf("expected empty map to marshal as '{}', got: %s", string(got))
	}
}

func TestOrderedMap_MarshalJSON_ValidJSON(t *testing.T) {
	om := newOrderedMap()
	om.Set("name", "Alice")
	om.Set("count", 42)
	om.Set("active", true)

	got, err := om.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() returned error: %v", err)
	}

	// Verify it is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("MarshalJSON() did not produce valid JSON: %v\nOutput: %s", err, string(got))
	}

	if parsed["name"] != "Alice" {
		t.Errorf("name: got %v, want Alice", parsed["name"])
	}
}

// ---- marshalOrderedMapIndent tests ----

func TestMarshalOrderedMapIndent_EmptyMap(t *testing.T) {
	om := newOrderedMap()

	got, err := marshalOrderedMapIndent(om, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalOrderedMapIndent() returned error: %v", err)
	}

	if string(got) != "{}" {
		t.Errorf("expected '{}', got: %s", string(got))
	}
}

func TestMarshalOrderedMapIndent_NestedMaps(t *testing.T) {
	inner := newOrderedMap()
	inner.Set("key", "value")

	outer := newOrderedMap()
	outer.Set("nested", inner)

	got, err := marshalOrderedMapIndent(outer, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalOrderedMapIndent() returned error: %v", err)
	}

	result := string(got)
	if !strings.Contains(result, "nested") || !strings.Contains(result, "key") {
		t.Errorf("expected nested structure, got: %s", result)
	}

	// Verify valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, result)
	}
}

// ---- marshalSliceIndent tests ----

func TestMarshalSliceIndent_EmptySlice(t *testing.T) {
	got, err := marshalSliceIndent([]interface{}{}, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalSliceIndent() returned error: %v", err)
	}

	if string(got) != "[]" {
		t.Errorf("expected '[]', got: %s", string(got))
	}
}

func TestMarshalSliceIndent_WithValues(t *testing.T) {
	items := []interface{}{"alpha", "beta", "gamma"}

	got, err := marshalSliceIndent(items, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalSliceIndent() returned error: %v", err)
	}

	result := string(got)

	// Verify valid JSON array
	var parsed []interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("output is not valid JSON array: %v\nOutput: %s", err, result)
	}

	if len(parsed) != 3 {
		t.Errorf("expected 3 items, got %d", len(parsed))
	}
}

func TestMarshalSliceIndent_WithNestedMap(t *testing.T) {
	om := newOrderedMap()
	om.Set("id", 1)
	om.Set("label", "test")

	items := []interface{}{om}

	got, err := marshalSliceIndent(items, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalSliceIndent() returned error: %v", err)
	}

	// Verify valid JSON
	var parsed []interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, string(got))
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 item, got %d", len(parsed))
	}
}

func TestMarshalSliceIndent_NestedSlices(t *testing.T) {
	inner := []interface{}{"a", "b"}
	items := []interface{}{inner, "c"}

	got, err := marshalSliceIndent(items, "", "  ", 0)
	if err != nil {
		t.Fatalf("marshalSliceIndent() returned error: %v", err)
	}

	// Verify valid JSON
	var parsed []interface{}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, string(got))
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 items, got %d", len(parsed))
	}
}

// ---- yamlNodeToOrderedInterface tests ----

func TestRenderAsJSON_PreservesFieldOrder(t *testing.T) {
	// Use a struct with known field order; YAML serialization should preserve it
	type orderedFields struct {
		Zebra  string `yaml:"zebra"`
		Alpha  string `yaml:"alpha"`
		Middle string `yaml:"middle"`
	}

	input := orderedFields{
		Zebra:  "z-val",
		Alpha:  "a-val",
		Middle: "m-val",
	}

	got, err := RenderAsJSON(input)
	if err != nil {
		t.Fatalf("RenderAsJSON() returned error: %v", err)
	}

	// Verify order is preserved (zebra before alpha before middle)
	zebraIdx := strings.Index(got, "zebra")
	alphaIdx := strings.Index(got, "alpha")
	middleIdx := strings.Index(got, "middle")

	if zebraIdx < 0 || alphaIdx < 0 || middleIdx < 0 {
		t.Fatalf("missing keys in output:\n%s", got)
	}

	if zebraIdx >= alphaIdx {
		t.Errorf("expected zebra before alpha in JSON output")
	}
	if alphaIdx >= middleIdx {
		t.Errorf("expected alpha before middle in JSON output")
	}
}
