package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeJinja2(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double braces",
			input: "Use {{ variable }} in templates",
			want:  "Use { { variable } } in templates",
		},
		{
			name:  "statement tags",
			input: "{% if condition %}",
			want:  "{ % if condition % }",
		},
		{
			name:  "no templates",
			input: "Normal text without templates",
			want:  "Normal text without templates",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "multiple occurrences",
			input: "{{ a }} and {{ b }}",
			want:  "{ { a } } and { { b } }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeJinja2(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEscapeTableCell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "pipe character",
			input: "format: json|yaml",
			want:  "format: json\\|yaml",
		},
		{
			name:  "newline",
			input: "Line one\nLine two",
			want:  "Line one Line two",
		},
		{
			name:  "both",
			input: "Option: a|b\nDefault: a",
			want:  "Option: a\\|b Default: a",
		},
		{
			name:  "no special chars",
			input: "Normal text",
			want:  "Normal text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeTableCell(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
