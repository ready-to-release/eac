package content

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/render"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render.EscapeJinja2(tt.input)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render.EscapeTableCell(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTable(t *testing.T) {
	t.Run("empty headers", func(t *testing.T) {
		result := FormatTable(nil, nil)
		assert.Equal(t, "", result)
	})

	t.Run("basic table", func(t *testing.T) {
		result := FormatTable(
			[]string{"Name", "Value"},
			[][]string{
				{"foo", "bar"},
				{"baz", "qux"},
			},
		)
		assert.Contains(t, result, "Name")
		assert.Contains(t, result, "Value")
		assert.Contains(t, result, "foo")
		assert.Contains(t, result, "bar")
		assert.Contains(t, result, "baz")
		assert.Contains(t, result, "qux")
	})
}
