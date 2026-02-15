package linking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRelativeLinks(t *testing.T) {
	content := `# Test Page

![Diagram](../assets/diagram.png)
[Other page](../guide/intro.md)
<img src="images/icon.png" alt="icon">
<a href="reference/api.md">API Docs</a>

[External](https://example.com)
[Anchor](#section)
![Absolute](/assets/logo.png)
`
	links := ExtractRelativeLinks(content)

	assert.Contains(t, links, "../assets/diagram.png")
	assert.Contains(t, links, "../guide/intro.md")
	assert.Contains(t, links, "images/icon.png")
	assert.Contains(t, links, "reference/api.md")

	// Should NOT contain external/absolute/anchor links
	for _, l := range links {
		assert.False(t, IsRelativePathRef("https://example.com"))
		assert.NotEqual(t, "#section", l)
		assert.NotEqual(t, "/assets/logo.png", l)
	}
}

func TestExtractRelativeLinks_SkipsCodeBlocks(t *testing.T) {
	content := "Before\n```\n[code link](../code/file.md)\n```\nAfter [real](real.md)"

	links := ExtractRelativeLinks(content)

	assert.Contains(t, links, "real.md")
	assert.NotContains(t, links, "../code/file.md")
}

func TestIsRelativePathRef(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"../guide.md", true},
		{"./local.md", true},
		{"sub/page.md", true},
		{"page.md#anchor", true},
		{"http://example.com", false},
		{"https://example.com", false},
		{"//cdn.example.com/file.js", false},
		{"mailto:test@example.com", false},
		{"/absolute/path.md", false},
		{"#anchor", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRelativePathRef(tt.path))
		})
	}
}

func TestStripCodeBlocks(t *testing.T) {
	content := "Text before\n```\ncode line 1\ncode line 2\n```\nText after\n    indented code line"

	result := StripCodeBlocks(content)

	assert.Contains(t, result, "Text before")
	assert.Contains(t, result, "Text after")
	assert.NotContains(t, result, "code line 1")
	assert.NotContains(t, result, "indented code line")
}
