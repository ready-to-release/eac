package contracts

import (
	"testing"
)

func TestEACCoreSchemaEmbedded(t *testing.T) {
	path := EACCorePath("repository.schema.json")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded schema %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded schema %s is empty", path)
	}
}

func TestEACCoreDefaultsEmbedded(t *testing.T) {
	path := EACCoreDefaultPath("repository.yml")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded default %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded default %s is empty", path)
	}
}

func TestR2RCLISchemaEmbedded(t *testing.T) {
	path := R2RCLIPath("r2r-cli.schema.json")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded schema %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded schema %s is empty", path)
	}
}

func TestR2RCLIEBNFEmbedded(t *testing.T) {
	path := R2RCLIPath("command.ebnf")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded EBNF %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded EBNF %s is empty", path)
	}
}

func TestEACDocsSchemaEmbedded(t *testing.T) {
	path := EACDocsPath("manifest.schema.json")
	data, err := FS.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read embedded schema %s: %v", path, err)
	}
	if len(data) == 0 {
		t.Errorf("Embedded schema %s is empty", path)
	}
}

func TestPathHelperVersions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		input    string
		expected string
	}{
		{"EACCorePath", EACCorePath, "repository.schema.json", "core/0.1.0/repository.schema.json"},
		{"EACCoreDefaultPath", EACCoreDefaultPath, "repository.yml", "core/0.1.0/defaults/repository.yml"},
		{"EACDocsPath", EACDocsPath, "manifest.schema.json", "docs/0.1.0/manifest.schema.json"},
		{"R2RCLIPath", R2RCLIPath, "command.ebnf", "r2r-cli/0.1.0/command.ebnf"},
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
