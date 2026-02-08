package core

import (
	"testing"
)

func TestSchemaEmbedded(t *testing.T) {
	path := SchemaPath("repository.schema.json")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded schema %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded schema %s is empty", path)
	}
}

func TestDefaultsEmbedded(t *testing.T) {
	path := DefaultPath("repository.yml")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded default %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded default %s is empty", path)
	}
}

func TestPathHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{"SchemaPath", SchemaPath, "repository.schema.json", "schemas/repository.schema.json"},
		{"DefaultPath", DefaultPath, "repository.yml", "schemas/defaults/repository.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.expected {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.input, result, tt.expected)
			}
		})
	}
}
