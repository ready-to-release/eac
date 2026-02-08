package output

import (
	"testing"
)

func TestMonikerParse(t *testing.T) {
	tests := []struct {
		name     string
		input    Moniker
		expected ParsedMoniker
	}{
		{
			name:  "full 4-part moniker",
			input: "build:core:go:go",
			expected: ParsedMoniker{
				Action:    "build",
				Module:    "core",
				Component: "go",
				Tool:      "go",
				Full:      "build:core:go:go",
			},
		},
		{
			name:  "3-part moniker (module:component:tool)",
			input: "core:go:gotest",
			expected: ParsedMoniker{
				Action:    "",
				Module:    "core",
				Component: "go",
				Tool:      "gotest",
				Full:      "core:go:gotest",
			},
		},
		{
			name:  "2-part moniker (module:component)",
			input: "eac-cli:docker",
			expected: ParsedMoniker{
				Action:    "",
				Module:    "eac-cli",
				Component: "docker",
				Tool:      "",
				Full:      "eac-cli:docker",
			},
		},
		{
			name:  "1-part moniker (module only)",
			input: "core",
			expected: ParsedMoniker{
				Action:    "",
				Module:    "core",
				Component: "",
				Tool:      "",
				Full:      "core",
			},
		},
		{
			name:  "empty moniker",
			input: "",
			expected: ParsedMoniker{
				Action:    "",
				Module:    "",
				Component: "",
				Tool:      "",
				Full:      "",
			},
		},
		{
			name:  "test context",
			input: "test:adapters-tui:go:gotest",
			expected: ParsedMoniker{
				Action:    "test",
				Module:    "adapters-tui",
				Component: "go",
				Tool:      "gotest",
				Full:      "test:adapters-tui:go:gotest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Parse()
			if got.Action != tt.expected.Action {
				t.Errorf("Context: got %q, want %q", got.Action, tt.expected.Action)
			}
			if got.Module != tt.expected.Module {
				t.Errorf("Module: got %q, want %q", got.Module, tt.expected.Module)
			}
			if got.Component != tt.expected.Component {
				t.Errorf("Component: got %q, want %q", got.Component, tt.expected.Component)
			}
			if got.Tool != tt.expected.Tool {
				t.Errorf("Tool: got %q, want %q", got.Tool, tt.expected.Tool)
			}
			if got.Full != tt.expected.Full {
				t.Errorf("Full: got %q, want %q", got.Full, tt.expected.Full)
			}
		})
	}
}

func TestMonikerModuleName(t *testing.T) {
	tests := []struct {
		name     string
		input    Moniker
		expected string
	}{
		{
			name:     "full 4-part moniker",
			input:    "build:core:go:go",
			expected: "build",
		},
		{
			name:     "3-part moniker",
			input:    "core:go:gotest",
			expected: "core",
		},
		{
			name:     "2-part moniker",
			input:    "eac-cli:docker",
			expected: "eac-cli",
		},
		{
			name:     "1-part moniker (module only)",
			input:    "core",
			expected: "core",
		},
		{
			name:     "empty moniker",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ModuleName()
			if got != tt.expected {
				t.Errorf("ModuleName(): got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonikerShortname(t *testing.T) {
	tests := []struct {
		name     string
		input    Moniker
		expected string
	}{
		{
			name:     "full 4-part moniker",
			input:    "build:core:go:go",
			expected: "core:go",
		},
		{
			name:     "3-part moniker",
			input:    "core:go:gotest",
			expected: "core:go",
		},
		{
			name:     "2-part moniker",
			input:    "eac-cli:docker",
			expected: "eac-cli:docker",
		},
		{
			name:     "1-part moniker",
			input:    "core",
			expected: "core",
		},
		{
			name:     "empty moniker",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Shortname()
			if got != tt.expected {
				t.Errorf("Shortname(): got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMonikerString(t *testing.T) {
	m := Moniker("build:core:go:go")
	if m.String() != "build:core:go:go" {
		t.Errorf("String(): got %q, want %q", m.String(), "build:core:go:go")
	}
}
