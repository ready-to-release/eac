package linking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceLinkPaths(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		old      string
		new      string
		expected string
	}{
		{
			name:     "markdown link replacement",
			content:  `See [CLI docs](../../reference/clie-eac/cli.md) for details.`,
			old:      "../../reference/clie-eac/cli.md",
			new:      "https://example.com/reference/clie-eac/cli",
			expected: `See [CLI docs](https://example.com/reference/clie-eac/cli) for details.`,
		},
		{
			name:     "markdown link with anchor",
			content:  `See [CLI docs](../../reference/clie-eac/cli.md#section) for details.`,
			old:      "../../reference/clie-eac/cli.md",
			new:      "https://example.com/reference/clie-eac/cli",
			expected: `See [CLI docs](https://example.com/reference/clie-eac/cli#section) for details.`,
		},
		{
			name:     "markdown image replacement",
			content:  `![Diagram](../assets/diagram.png)`,
			old:      "../assets/diagram.png",
			new:      "assets/diagram.png",
			expected: `![Diagram](assets/diagram.png)`,
		},
		{
			name:     "HTML src replacement",
			content:  `<img src="../assets/icon.png" alt="icon">`,
			old:      "../assets/icon.png",
			new:      "assets/icon.png",
			expected: `<img src="assets/icon.png" alt="icon">`,
		},
		{
			name:     "HTML href replacement",
			content:  `<a href="../reference/page.md">Link</a>`,
			old:      "../reference/page.md",
			new:      "reference/page.md",
			expected: `<a href="reference/page.md">Link</a>`,
		},
		{
			name:     "no corruption of overlapping paths",
			content:  `[Link1](reference/clie-eac/cli.md) and [Link2](reference/other.md)`,
			old:      "reference/",
			new:      "https://example.com/reference",
			expected: `[Link1](reference/clie-eac/cli.md) and [Link2](reference/other.md)`,
		},
		{
			name:     "exact path match only",
			content:  `[Link](reference/clie-eac/cli.md)`,
			old:      "reference/clie-eac/cli.md",
			new:      "https://example.com/reference/clie-eac/cli",
			expected: `[Link](https://example.com/reference/clie-eac/cli)`,
		},
		{
			name:     "does not corrupt plain text",
			content:  `The reference/clie-eac directory contains CLI docs. See [Link](reference/clie-eac/cli.md).`,
			old:      "reference/clie-eac/cli.md",
			new:      "https://example.com/reference/clie-eac/cli",
			expected: `The reference/clie-eac directory contains CLI docs. See [Link](https://example.com/reference/clie-eac/cli).`,
		},
		{
			name:     "multiple links same path",
			content:  `[Link1](path.md) and [Link2](path.md)`,
			old:      "path.md",
			new:      "https://example.com/path",
			expected: `[Link1](https://example.com/path) and [Link2](https://example.com/path)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceLinkPaths(tt.content, tt.old, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReplaceOutsideCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		old      string
		new      string
		expected string
	}{
		{
			name:     "replaces link outside code block",
			content:  "See [docs](path.md) for info.",
			old:      "path.md",
			new:      "https://example.com/path",
			expected: "See [docs](https://example.com/path) for info.",
		},
		{
			name:     "skips replacement inside code block",
			content:  "Text\n```\n[link](path.md)\n```\nMore text",
			old:      "path.md",
			new:      "https://example.com/path",
			expected: "Text\n```\n[link](path.md)\n```\nMore text",
		},
		{
			name:     "replaces outside but not inside code block",
			content:  "[link](path.md)\n```\n[code](path.md)\n```\n[another](path.md)",
			old:      "path.md",
			new:      "new.md",
			expected: "[link](new.md)\n```\n[code](path.md)\n```\n[another](new.md)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceOutsideCodeBlocks(tt.content, tt.old, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReplaceMarkdownLinks(t *testing.T) {
	content := `[Link](path.md) and [Another](path.md#section)`

	result := ReplaceMarkdownLinks(content, "path.md", true)
	assert.Equal(t, "Link and Another", result)
}

func TestExtractSourcePrefix(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{"double star pattern", "docs/explanation/**/*.md", "docs/explanation/"},
		{"single star pattern", "docs/*.md", "docs/"},
		{"question mark pattern", "docs/file?.md", "docs/"},
		{"bracket pattern", "docs/[abc]/file.md", "docs/"},
		{"brace pattern", "docs/{a,b}/file.md", "docs/"},
		{"no glob pattern", "docs/guide/intro.md", "docs/guide/"},
		{"root level glob", "*.md", ""},
		{"nested path", "docs/reference/api/**/*.md", "docs/reference/api/"},
		{"glob at beginning", "**/docs/file.md", ""},
		{"multiple levels", "a/b/c/d/**/*.md", "a/b/c/d/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSourcePrefix(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountPathDepth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{"empty path", "", 0},
		{"dot path", ".", 0},
		{"single directory", "docs", 1},
		{"two directories", "docs/guide", 2},
		{"three directories", "docs/reference/api", 3},
		{"trailing slash", "docs/guide/", 2},
		{"leading slash", "/docs/guide", 2},
		{"both slashes", "/docs/guide/", 2},
		{"deeply nested", "a/b/c/d/e/f", 6},
		{"backslashes", "docs\\guide\\api", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountPathDepth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAdjustRelativePaths(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		depthChange int
		fileDepth   int
		expected    string
	}{
		{
			name:        "link with two levels up, depth change 1",
			content:     "[Link](../../other/file.md)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Link](../other/file.md)",
		},
		{
			name:        "link with three levels up, depth change 2",
			content:     "[Link](../../../assets/image.png)",
			depthChange: 2,
			fileDepth:   0,
			expected:    "[Link](../assets/image.png)",
		},
		{
			name:        "image with relative path",
			content:     "![Diagram](../../diagrams/flow.png)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "![Diagram](../diagrams/flow.png)",
		},
		{
			name:        "link within content tree, no adjustment",
			content:     "[Link](../sibling.md)",
			depthChange: 1,
			fileDepth:   1,
			expected:    "[Link](../sibling.md)",
		},
		{
			name:        "link outside content tree, adjust",
			content:     "[Link](../../parent.md)",
			depthChange: 1,
			fileDepth:   1,
			expected:    "[Link](../parent.md)",
		},
		{
			name:        "depth change reduces to zero",
			content:     "[Link](../file.md)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Link](file.md)",
		},
		{
			name:        "depth change greater than up count",
			content:     "[Link](../file.md)",
			depthChange: 5,
			fileDepth:   0,
			expected:    "[Link](file.md)",
		},
		{
			name:        "no relative paths",
			content:     "[Absolute](/docs/file.md) and [External](https://example.com)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Absolute](/docs/file.md) and [External](https://example.com)",
		},
		{
			name:        "link with attributes",
			content:     "[Link](../../file.md){: .class}",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Link](../file.md){: .class}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustRelativePaths(tt.content, tt.depthChange, tt.fileDepth)
			assert.Equal(t, tt.expected, result)
		})
	}
}
