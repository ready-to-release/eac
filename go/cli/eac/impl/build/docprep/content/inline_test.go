package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatInlineReplacement(t *testing.T) {
	result := FormatInlineReplacement("my-marker", "Hello World")

	assert.Contains(t, result, "<!-- book:generated my-marker -->")
	assert.Contains(t, result, "Hello World")
	assert.Contains(t, result, "<!-- /book:generated -->")
}

func TestFormatInlineReplacement_TrimsWhitespace(t *testing.T) {
	result := FormatInlineReplacement("test", "  output with spaces  ")

	assert.Contains(t, result, "output with spaces")
	assert.NotContains(t, result, "  output")
}

func TestFormatInlineReplacement_Structure(t *testing.T) {
	result := FormatInlineReplacement("marker", "content")

	expected := "<!-- book:generated marker -->\ncontent\n<!-- /book:generated -->"
	assert.Equal(t, expected, result)
}
