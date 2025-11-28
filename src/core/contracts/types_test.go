//go:build L1 && ov
// +build L1,ov

package contracts

import "testing"

func TestBaseContract_Getters(t *testing.T) {
	contract := BaseContract{
		Moniker:     "test-moniker",
		Name:        "Test Name",
		Type:        "test-type",
		Description: "Test description",
		Parent:      "parent-module",
		Files: Files{
			Root:      "test/root",
			Changelog: "CHANGELOG.md",
		},
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"GetMoniker", contract.GetMoniker(), "test-moniker"},
		{"GetName", contract.GetName(), "Test Name"},
		{"GetType", contract.GetType(), "test-type"},
		{"GetDescription", contract.GetDescription(), "Test description"},
		{"GetParent", contract.GetParent(), "parent-module"},
		{"GetRoot", contract.GetRoot(), "test/root"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.name, tt.expected, tt.got)
			}
		})
	}
}

func TestFiles_Changelog(t *testing.T) {
	files := Files{
		Root:      "src/test",
		Changelog: "docs/CHANGELOG.md",
		Source:    []string{"**/*.go"},
	}

	if files.Changelog != "docs/CHANGELOG.md" {
		t.Errorf("Expected changelog 'docs/CHANGELOG.md', got '%s'", files.Changelog)
	}
}

func TestFiles_Source(t *testing.T) {
	files := Files{
		Source: []string{"**/*.go", "**/*.md"},
	}

	if len(files.Source) != 2 {
		t.Errorf("Expected 2 source patterns, got %d", len(files.Source))
	}

	if files.Source[0] != "**/*.go" {
		t.Errorf("Expected first pattern '**/*.go', got '%s'", files.Source[0])
	}
}
