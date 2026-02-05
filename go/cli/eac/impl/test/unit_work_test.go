package test

import "testing"

func TestSanitizeTestname(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name unchanged",
			input:    "impl-build",
			expected: "impl-build",
		},
		{
			name:     "slashes replaced with dashes",
			input:    "impl/build/internal",
			expected: "impl-build-internal",
		},
		{
			name:     "backslashes replaced with dashes",
			input:    "impl\\build\\internal",
			expected: "impl-build-internal",
		},
		{
			name:     "colons replaced with dashes",
			input:    "impl:build:test",
			expected: "impl-build-test",
		},
		{
			name:     "spaces replaced with dashes",
			input:    "impl build test",
			expected: "impl-build-test",
		},
		{
			name:     "unsafe chars replaced with underscores",
			input:    "test<file>name",
			expected: "test_file_name",
		},
		{
			name:     "pipe replaced with underscore",
			input:    "test|name",
			expected: "test_name",
		},
		{
			name:     "question mark replaced with underscore",
			input:    "test?name",
			expected: "test_name",
		},
		{
			name:     "asterisk replaced with underscore",
			input:    "test*name",
			expected: "test_name",
		},
		{
			name:     "quotes replaced with underscore",
			input:    "test\"name",
			expected: "test_name",
		},
		{
			name:     "multiple consecutive dashes collapsed",
			input:    "impl--build---test",
			expected: "impl-build-test",
		},
		{
			name:     "leading dashes trimmed",
			input:    "--impl-build",
			expected: "impl-build",
		},
		{
			name:     "trailing dashes trimmed",
			input:    "impl-build--",
			expected: "impl-build",
		},
		{
			name:     "complex path sanitized",
			input:    "go/cli/eac/impl/build",
			expected: "go-cli-eac-impl-build",
		},
		{
			name:     "mixed separators sanitized",
			input:    "go/cli\\eac:impl build",
			expected: "go-cli-eac-impl-build",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "windows path sanitized",
			input:    "C:\\projects\\test",
			expected: "C-projects-test",
		},
		{
			name:     "path with all unsafe chars",
			input:    "<test>:name|file?\"*",
			expected: "_test_-name_file___",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTestname(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeTestname(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractSpecName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple spec name",
			input:    "build-module:go/eac/specs:feature.feature",
			expected: "build-module",
		},
		{
			name:     "spec name with colon in path parts",
			input:    "test-spec:path:to:feature",
			expected: "test-spec",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no colon returns whole string sanitized",
			input:    "simple-spec",
			expected: "simple-spec",
		},
		{
			name:     "spec name needing sanitization",
			input:    "test/spec:path:feature",
			expected: "test-spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSpecName(tt.input)
			if result != tt.expected {
				t.Errorf("extractSpecName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
