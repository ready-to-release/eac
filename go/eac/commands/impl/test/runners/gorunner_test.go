//go:build L1
// +build L1

package runners

import "testing"

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
