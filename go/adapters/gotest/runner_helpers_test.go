//go:build L1
// +build L1

package gotest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractGoBuildTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "single L-level tag",
			input:    "@L2",
			expected: "L2",
		},
		{
			name:     "multiple L-level tags",
			input:    "@L0,@L1",
			expected: "L0,L1",
		},
		{
			name:     "L-level with skip filter",
			input:    "@L0,@L1 && ~@skip:wip",
			expected: "L0,L1",
		},
		{
			name:     "single deps tag",
			input:    "@deps:docker",
			expected: "deps_docker",
		},
		{
			name:     "deps tag with hyphen",
			input:    "@deps:gh-token",
			expected: "deps_gh_token",
		},
		{
			name:     "L-level and deps tag",
			input:    "@L2 && @deps:gh-token",
			expected: "L2,deps_gh_token",
		},
		{
			name:     "multiple deps tags",
			input:    "@deps:docker && @deps:gh-token",
			expected: "deps_docker,deps_gh_token",
		},
		{
			name:     "complex filter with L-level and deps",
			input:    "@L2,@L3 && @deps:go && ~@skip:wip",
			expected: "L2,L3,deps_go",
		},
		{
			name:     "deps tag with multiple hyphens",
			input:    "@deps:az-cli-tools",
			expected: "deps_az_cli_tools",
		},
		{
			name:     "negated L-level tags should be excluded",
			input:    "~@L2 && ~@L3 && ~@L4 && @L0,@L1",
			expected: "L0,L1",
		},
		{
			name:     "all negated L-level tags",
			input:    "~@L0 && ~@L1 && ~@L2",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGoBuildTags(tt.input)
			if result != tt.expected {
				t.Errorf("extractGoBuildTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHasGenerateDirectives(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected bool
	}{
		{
			name: "no go files",
			files: map[string]string{
				"readme.txt": "just text",
			},
			expected: false,
		},
		{
			name: "go file without directive",
			files: map[string]string{
				"main.go": "package main\n\nfunc main() {}\n",
			},
			expected: false,
		},
		{
			name: "go file with directive",
			files: map[string]string{
				"main.go": "package main\n\n//go:generate stringer -type=Pill\n\nfunc main() {}\n",
			},
			expected: true,
		},
		{
			name: "directive in subdirectory",
			files: map[string]string{
				"main.go":       "package main\n\nfunc main() {}\n",
				"sub/gen.go":    "package sub\n\n//go:generate mockgen -source=iface.go\n",
			},
			expected: true,
		},
		{
			name: "comment mentioning go:generate but not a directive",
			files: map[string]string{
				"main.go": "package main\n\n// see //go:generate docs for more\n\nfunc main() {}\n",
			},
			expected: false,
		},
		{
			name: "bare directive without args",
			files: map[string]string{
				"main.go": "package main\n\n//go:generate\n\nfunc main() {}\n",
			},
			expected: true,
		},
		{
			name: "non-go file containing directive text",
			files: map[string]string{
				"notes.txt": "//go:generate should be ignored in non-go files\n",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(root, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
			}
			got, err := hasGenerateDirectives(root)
			if err != nil {
				t.Fatalf("hasGenerateDirectives: %v", err)
			}
			if got != tt.expected {
				t.Errorf("hasGenerateDirectives(%q) = %v, want %v", root, got, tt.expected)
			}
		})
	}
}
