package books

import (
	"testing"
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
			content:  `See [CLI docs](../../reference/r2r-eac/cli.md) for details.`,
			old:      "../../reference/r2r-eac/cli.md",
			new:      "https://example.com/reference/r2r-eac/cli",
			expected: `See [CLI docs](https://example.com/reference/r2r-eac/cli) for details.`,
		},
		{
			name:     "markdown link with anchor",
			content:  `See [CLI docs](../../reference/r2r-eac/cli.md#section) for details.`,
			old:      "../../reference/r2r-eac/cli.md",
			new:      "https://example.com/reference/r2r-eac/cli",
			expected: `See [CLI docs](https://example.com/reference/r2r-eac/cli#section) for details.`,
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
			content:  `[Link1](reference/r2r-eac/cli.md) and [Link2](reference/other.md)`,
			old:      "reference/",
			new:      "https://example.com/reference",
			expected: `[Link1](reference/r2r-eac/cli.md) and [Link2](reference/other.md)`,
		},
		{
			name:     "exact path match only",
			content:  `[Link](reference/r2r-eac/cli.md)`,
			old:      "reference/r2r-eac/cli.md",
			new:      "https://example.com/reference/r2r-eac/cli",
			expected: `[Link](https://example.com/reference/r2r-eac/cli)`,
		},
		{
			name:     "does not corrupt plain text",
			content:  `The reference/r2r-eac directory contains CLI docs. See [Link](reference/r2r-eac/cli.md).`,
			old:      "reference/r2r-eac/cli.md",
			new:      "https://example.com/reference/r2r-eac/cli",
			expected: `The reference/r2r-eac directory contains CLI docs. See [Link](https://example.com/reference/r2r-eac/cli).`,
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
			result := replaceLinkPaths(tt.content, tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("replaceLinkPaths()\nexpected: %s\ngot:      %s", tt.expected, result)
			}
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
			result := replaceOutsideCodeBlocks(tt.content, tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("replaceOutsideCodeBlocks()\nexpected: %s\ngot:      %s", tt.expected, result)
			}
		})
	}
}

func TestAddImageWidthConstraints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		count    int
	}{
		{
			name:     "simple image without attributes",
			input:    `![CD Model](../assets/cd-model/diagram.png)`,
			expected: `![CD Model](../assets/cd-model/diagram.png){ width="100%" }`,
			count:    1,
		},
		{
			name:     "image with existing width - skip",
			input:    `![Image](path.png){ width="50%" }`,
			expected: `![Image](path.png){ width="50%" }`,
			count:    0,
		},
		{
			name:     "image with style attribute - skip",
			input:    `![Image](path.png){ style="max-width: 300px" }`,
			expected: `![Image](path.png){ style="max-width: 300px" }`,
			count:    0,
		},
		{
			name:     "icon image - skip",
			input:    `![Icon](../icons/check-icon.png)`,
			expected: `![Icon](../icons/check-icon.png)`,
			count:    0,
		},
		{
			name:     "favicon - skip",
			input:    `![](favicon.ico)`,
			expected: `![](favicon.ico)`,
			count:    0,
		},
		{
			name:     "badge image - skip",
			input:    `![Build Badge](badge.svg)`,
			expected: `![Build Badge](badge.svg)`,
			count:    0,
		},
		{
			name:     "multiple images",
			input:    "Some text\n![Img1](path1.png)\nMore text\n![Img2](path2.png)",
			expected: "Some text\n![Img1](path1.png){ width=\"100%\" }\nMore text\n![Img2](path2.png){ width=\"100%\" }",
			count:    2,
		},
		{
			name:     "drawio diagram",
			input:    `![Architecture](../assets/branching/overview.drawio.png)`,
			expected: `![Architecture](../assets/branching/overview.drawio.png){ width="100%" }`,
			count:    1,
		},
		{
			name:     "image with other attributes but no width",
			input:    `![Image](path.png){ loading="lazy" }`,
			expected: `![Image](path.png){ loading="lazy"  width="100%" }`,
			count:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := addImageWidthConstraints(tt.input)
			if result != tt.expected {
				t.Errorf("addImageWidthConstraints() result = %q, want %q", result, tt.expected)
			}
			if count != tt.count {
				t.Errorf("addImageWidthConstraints() count = %d, want %d", count, tt.count)
			}
		})
	}
}

func TestConvertAttrListImages(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		count    int
	}{
		{
			name:     "width only",
			input:    `![icon](assets/icon.png){width=48}`,
			expected: `<img src="assets/icon.png" width="48" alt="icon">`,
			count:    1,
		},
		{
			name:     "width with quotes",
			input:    `![icon](assets/icon.png){width="48"}`,
			expected: `<img src="assets/icon.png" width="48" alt="icon">`,
			count:    1,
		},
		{
			name:     "width with px",
			input:    `![icon](assets/icon.png){width="48px"}`,
			expected: `<img src="assets/icon.png" width="48px" alt="icon">`,
			count:    1,
		},
		{
			name:     "width and height",
			input:    `![diagram](diagram.png){width=600 height=400}`,
			expected: `<img src="diagram.png" width="600" height="400" alt="diagram">`,
			count:    1,
		},
		{
			name:     "with colon prefix",
			input:    `![icon](icon.png){: width="48" }`,
			expected: `<img src="icon.png" width="48" alt="icon">`,
			count:    1,
		},
		{
			name:     "no attrs - unchanged",
			input:    `![image](image.png)`,
			expected: `![image](image.png)`,
			count:    0,
		},
		{
			name:     "multiple images",
			input:    `![a](a.png){width=100} text ![b](b.png){width=200}`,
			expected: `<img src="a.png" width="100" alt="a"> text <img src="b.png" width="200" alt="b">`,
			count:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := convertAttrListImages(tt.input, false) // adjustPaths=false for unit test
			if result != tt.expected {
				t.Errorf("convertAttrListImages()\nexpected: %s\ngot:      %s", tt.expected, result)
			}
			if count != tt.count {
				t.Errorf("convertAttrListImages() count = %d, want %d", tt.count, count)
			}
		})
	}
}
