package books

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractSourcePrefix verifies source prefix extraction from glob patterns.
func TestExtractSourcePrefix(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "double star pattern",
			pattern:  "docs/explanation/**/*.md",
			expected: "docs/explanation/",
		},
		{
			name:     "single star pattern",
			pattern:  "docs/*.md",
			expected: "docs/",
		},
		{
			name:     "question mark pattern",
			pattern:  "docs/file?.md",
			expected: "docs/",
		},
		{
			name:     "bracket pattern",
			pattern:  "docs/[abc]/file.md",
			expected: "docs/",
		},
		{
			name:     "brace pattern",
			pattern:  "docs/{a,b}/file.md",
			expected: "docs/",
		},
		{
			name:     "no glob pattern",
			pattern:  "docs/guide/intro.md",
			expected: "docs/guide/",
		},
		{
			name:     "root level glob",
			pattern:  "*.md",
			expected: "",
		},
		{
			name:     "nested path with glob",
			pattern:  "docs/reference/api/**/*.md",
			expected: "docs/reference/api/",
		},
		{
			name:     "glob at beginning",
			pattern:  "**/docs/file.md",
			expected: "",
		},
		{
			name:     "multiple directory levels",
			pattern:  "a/b/c/d/**/*.md",
			expected: "a/b/c/d/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSourcePrefix(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCountPathDepth verifies directory depth counting.
func TestCountPathDepth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "empty path",
			path:     "",
			expected: 0,
		},
		{
			name:     "dot path",
			path:     ".",
			expected: 0,
		},
		{
			name:     "single directory",
			path:     "docs",
			expected: 1,
		},
		{
			name:     "two directories",
			path:     "docs/guide",
			expected: 2,
		},
		{
			name:     "three directories",
			path:     "docs/reference/api",
			expected: 3,
		},
		{
			name:     "trailing slash",
			path:     "docs/guide/",
			expected: 2,
		},
		{
			name:     "leading slash",
			path:     "/docs/guide",
			expected: 2,
		},
		{
			name:     "both slashes",
			path:     "/docs/guide/",
			expected: 2,
		},
		{
			name:     "deeply nested",
			path:     "a/b/c/d/e/f",
			expected: 6,
		},
		{
			name:     "backslashes (Windows style)",
			path:     "docs\\guide\\api",
			expected: 3, // ToSlash converts backslashes to forward slashes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countPathDepth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAdjustRelativePaths verifies relative path adjustment in markdown content.
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
			fileDepth:   1, // File at depth 1, link goes up 1, stays in tree
			expected:    "[Link](../sibling.md)",
		},
		{
			name:        "link outside content tree, adjust",
			content:     "[Link](../../parent.md)",
			depthChange: 1,
			fileDepth:   1, // File at depth 1, link goes up 2, outside tree
			expected:    "[Link](../parent.md)",
		},
		{
			name:        "multiple links",
			content:     "[One](../../a.md) and [Two](../../../b.md)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[One](../a.md) and [Two](../../b.md)",
		},
		{
			name:        "link with attributes",
			content:     "[Link](../../file.md){: .class}",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Link](../file.md){: .class}",
		},
		{
			name:        "image with attributes",
			content:     "![Image](../../img.png){ width=\"100%\" }",
			depthChange: 1,
			fileDepth:   0,
			expected:    "![Image](../img.png){ width=\"100%\" }",
		},
		{
			name:        "no relative paths",
			content:     "[Absolute](/docs/file.md) and [External](https://example.com)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Absolute](/docs/file.md) and [External](https://example.com)",
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
			expected:    "[Link](file.md)", // Can't go below zero
		},
		{
			name:        "mixed content with text",
			content:     "Text before [Link](../../file.md) and after.",
			depthChange: 1,
			fileDepth:   0,
			expected:    "Text before [Link](../file.md) and after.",
		},
		{
			name:        "link with anchor",
			content:     "[Section](../../docs/file.md#section)",
			depthChange: 1,
			fileDepth:   0,
			expected:    "[Section](../docs/file.md#section)",
		},
		{
			name:        "file at depth 2, link goes up 2, stays in tree",
			content:     "[Link](../../file.md)",
			depthChange: 1,
			fileDepth:   2, // Link goes up 2, equals fileDepth, stays in tree
			expected:    "[Link](../../file.md)",
		},
		{
			name:        "file at depth 2, link goes up 3, outside tree",
			content:     "[Link](../../../file.md)",
			depthChange: 1,
			fileDepth:   2, // Link goes up 3, greater than fileDepth, adjust
			expected:    "[Link](../../file.md)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjustRelativePaths(tt.content, tt.depthChange, tt.fileDepth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAdjustRelativePaths_ComplexMarkdown verifies handling of complex markdown.
func TestAdjustRelativePaths_ComplexMarkdown(t *testing.T) {
	content := `# Documentation

Here's a [link to guide](../../guide/intro.md) and an image:

![Architecture](../../../assets/diagrams/arch.png){ width="100%" }

Another [reference](../../api/spec.md) with text.

Code block (note: adjustRelativePaths doesn't skip code blocks):
` + "```" + `
[Not a link](../../fake.md)
` + "```" + `

But this [real link](../../docs/file.md) should be.`

	expected := `# Documentation

Here's a [link to guide](../guide/intro.md) and an image:

![Architecture](../../assets/diagrams/arch.png){ width="100%" }

Another [reference](../api/spec.md) with text.

Code block (note: adjustRelativePaths doesn't skip code blocks):
` + "```" + `
[Not a link](../fake.md)
` + "```" + `

But this [real link](../docs/file.md) should be.`

	result := adjustRelativePaths(content, 1, 0)
	assert.Equal(t, expected, result)
}
