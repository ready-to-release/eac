//go:build L0

package defaults

import (
	"testing"
)

// All paths use forward slashes (/) for cross-platform config file compatibility.

func TestTestImplPath(t *testing.T) {
	tests := []struct {
		moniker  string
		expected string
	}{
		{"cli", "src/cli/tests"},
		{"core", "src/core/tests"},
		{"commands", "src/commands/tests"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			got := TestImplPath(tt.moniker)
			if got != tt.expected {
				t.Errorf("TestImplPath(%q) = %q, want %q", tt.moniker, got, tt.expected)
			}
		})
	}
}

func TestDesignPath(t *testing.T) {
	tests := []struct {
		moniker  string
		expected string
	}{
		{"cli", "specs/cli/.design"},
		{"core", "specs/core/.design"},
		{"commands", "specs/commands/.design"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			got := DesignPath(tt.moniker)
			if got != tt.expected {
				t.Errorf("DesignPath(%q) = %q, want %q", tt.moniker, got, tt.expected)
			}
		})
	}
}

func TestSpecsPath(t *testing.T) {
	tests := []struct {
		moniker  string
		expected string
	}{
		{"cli", "specs/cli"},
		{"core", "specs/core"},
		{"commands", "specs/commands"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			got := SpecsPath(tt.moniker)
			if got != tt.expected {
				t.Errorf("SpecsPath(%q) = %q, want %q", tt.moniker, got, tt.expected)
			}
		})
	}
}

func TestSpecsPattern(t *testing.T) {
	tests := []struct {
		moniker  string
		expected string
	}{
		{"cli", "specs/cli/**"},
		{"core", "specs/core/**"},
		{"commands", "specs/commands/**"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			got := SpecsPattern(tt.moniker)
			if got != tt.expected {
				t.Errorf("SpecsPattern(%q) = %q, want %q", tt.moniker, got, tt.expected)
			}
		})
	}
}

func TestFeaturePath(t *testing.T) {
	tests := []struct {
		moniker  string
		feature  string
		expected string
	}{
		{"cli", "init", "specs/cli/init/init.feature"},
		{"commands", "build", "specs/commands/build/build.feature"},
		{"core", "validation", "specs/core/validation/validation.feature"},
	}

	for _, tt := range tests {
		name := tt.moniker + "/" + tt.feature
		t.Run(name, func(t *testing.T) {
			got := FeaturePath(tt.moniker, tt.feature)
			if got != tt.expected {
				t.Errorf("FeaturePath(%q, %q) = %q, want %q", tt.moniker, tt.feature, got, tt.expected)
			}
		})
	}
}

func TestFeatureDir(t *testing.T) {
	tests := []struct {
		moniker  string
		feature  string
		expected string
	}{
		{"cli", "init", "specs/cli/init"},
		{"commands", "build", "specs/commands/build"},
		{"core", "validation", "specs/core/validation"},
	}

	for _, tt := range tests {
		name := tt.moniker + "/" + tt.feature
		t.Run(name, func(t *testing.T) {
			got := FeatureDir(tt.moniker, tt.feature)
			if got != tt.expected {
				t.Errorf("FeatureDir(%q, %q) = %q, want %q", tt.moniker, tt.feature, got, tt.expected)
			}
		})
	}
}

// TestConstants verifies the default constant values are as expected
func TestConstants(t *testing.T) {
	if ModuleType != "no-module-type" {
		t.Errorf("ModuleType = %q, want %q", ModuleType, "no-module-type")
	}
	if Changelog != "CHANGELOG.md" {
		t.Errorf("Changelog = %q, want %q", Changelog, "CHANGELOG.md")
	}
}
