package contracts

import "testing"

func TestBaseContract_Getters(t *testing.T) {
	contract := BaseContract{
		Moniker:     "test-moniker",
		Name:        "Test Name",
		Type:        "test-type",
		Description: "Test description",
		Parent:      "parent-module",
		Source: Source{
			Root:          "test/root",
			ChangelogPath: "CHANGELOG.md",
		},
		Versioning: Versioning{
			VersionScheme: "MAJOR.MINOR.PATCH",
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
		{"GetVersion", contract.GetVersion(), "MAJOR.MINOR.PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.name, tt.expected, tt.got)
			}
		})
	}
}

func TestVersioning(t *testing.T) {
	v := Versioning{
		VersionScheme: "MAJOR.MINOR.PATCH",
	}

	if v.VersionScheme != "MAJOR.MINOR.PATCH" {
		t.Errorf("Expected version scheme 'MAJOR.MINOR.PATCH', got '%s'", v.VersionScheme)
	}
}

func TestSource_ChangelogPath(t *testing.T) {
	source := Source{
		Root:          "src/test",
		ChangelogPath: "docs/CHANGELOG.md",
		Includes:      []string{"**/*.go"},
	}

	if source.ChangelogPath != "docs/CHANGELOG.md" {
		t.Errorf("Expected changelog path 'docs/CHANGELOG.md', got '%s'", source.ChangelogPath)
	}
}

func TestSource(t *testing.T) {
	source := Source{
		Includes: []string{"**/*.go", "**/*.md"},
	}

	if len(source.Includes) != 2 {
		t.Errorf("Expected 2 includes, got %d", len(source.Includes))
	}

	if source.Includes[0] != "**/*.go" {
		t.Errorf("Expected first include '**/*.go', got '%s'", source.Includes[0])
	}
}
