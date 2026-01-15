package books

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestConvertDrawioImages verifies Drawio image to link conversion
func TestConvertDrawioImages(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		relFileDir  string
		siteURL     string
		expected    string
		expectCount int
	}{
		{
			name:        "simple drawio image",
			content:     `![Diagram](diagram.drawio)`,
			relFileDir:  "docs",
			siteURL:     "https://example.com/",
			expected:    `[View interactive diagram](https://example.com/docs/diagram.drawio)`,
			expectCount: 1,
		},
		{
			name:        "drawio with alt text",
			content:     `![Architecture Diagram](../assets/arch.drawio)`,
			relFileDir:  "docs/guide",
			siteURL:     "https://example.com/",
			expected:    `[View interactive diagram](https://example.com/docs/assets/arch.drawio)`,
			expectCount: 1,
		},
		{
			name:        "absolute path drawio",
			content:     `![Diagram](/assets/diagram.drawio)`,
			relFileDir:  "docs",
			siteURL:     "https://example.com/",
			expected:    `[View interactive diagram](https://example.com/assets/diagram.drawio)`,
			expectCount: 1,
		},
		{
			name:        "no drawio images",
			content:     `![Regular Image](image.png)`,
			relFileDir:  "docs",
			siteURL:     "https://example.com/",
			expected:    `![Regular Image](image.png)`,
			expectCount: 0,
		},
		{
			name:        "multiple drawio images",
			content:     `![Diagram1](d1.drawio) text ![Diagram2](d2.drawio)`,
			relFileDir:  "docs",
			siteURL:     "https://example.com/",
			expected:    `[View interactive diagram](https://example.com/docs/d1.drawio) text [View interactive diagram](https://example.com/docs/d2.drawio)`,
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := convertDrawioImages(tt.content, tt.relFileDir, tt.siteURL)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectCount, count)
		})
	}
}

// TestFixLinksInContent verifies broken link fixing logic
func TestFixLinksInContent(t *testing.T) {
	// Create a temporary staging directory
	stagingDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "existing.md"), []byte("# Exists"), 0644))

	tests := []struct {
		name        string
		content     string
		relFileDir  string
		stagingDir  string
		siteURL     string
		expected    string
		expectCount int
	}{
		{
			name:        "valid link to existing file",
			content:     `[Link](existing.md)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `[Link](existing.md)`,
			expectCount: 0, // No fixes needed
		},
		{
			name:        "broken link to missing file",
			content:     `[Link](missing.md)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `[Link](https://example.com/missing)`,
			expectCount: 1,
		},
		{
			name:        "external link preserved",
			content:     `[Link](https://external.com/page)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `[Link](https://external.com/page)`,
			expectCount: 0,
		},
		{
			name:        "anchor link preserved",
			content:     `[Link](#section)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `[Link](#section)`,
			expectCount: 0,
		},
		{
			name:        "link with anchor to missing file",
			content:     `[Link](missing.md#section)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `[Link](https://example.com/missing#section)`,
			expectCount: 1,
		},
		{
			name:        "image link not processed",
			content:     `![Image](missing.png)`,
			relFileDir:  ".",
			stagingDir:  stagingDir,
			siteURL:     "https://example.com/",
			expected:    `![Image](missing.png)`,
			expectCount: 0, // Images are not processed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, count := fixLinksInContent(tt.content, tt.relFileDir, tt.stagingDir, tt.siteURL)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectCount, count)
		})
	}
}

// TestParseAttrListToHTML verifies attr_list parsing
func TestParseAttrListToHTML(t *testing.T) {
	tests := []struct {
		name     string
		attrs    string
		expected string
	}{
		{
			name:     "width attribute",
			attrs:    `width="100"`,
			expected: `width="100"`,
		},
		{
			name:     "width and height",
			attrs:    `width="100" height="50"`,
			expected: `width="100" height="50"`,
		},
		{
			name:     "style attribute (only first token captured)",
			attrs:    `style="border: 1px solid"`,
			expected: `style="border:"`, // Regex pattern only captures up to first space
		},
		{
			name:     "unquoted values",
			attrs:    `width=100`,
			expected: `width="100"`,
		},
		{
			name:     "mixed quoted and unquoted",
			attrs:    `width="100" height=50`,
			expected: `width="100" height="50"`,
		},
		{
			name:     "class and id",
			attrs:    `class="img-fluid" id="main-image"`,
			expected: `class="img-fluid" id="main-image"`,
		},
		{
			name:     "empty attrs",
			attrs:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAttrListToHTML(tt.attrs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCleanupLinksForPDF_Integration verifies PDF image cleanup on files
func TestCleanupLinksForPDF_Integration(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	// Create test markdown file with various image scenarios
	testFile := filepath.Join(stagingDir, "test.md")
	content := `# Test Document

![Diagram](diagram.png)
![Icon](icon.png)
![Badge](badge.svg)
![Architecture](architecture.drawio.png){ width="100%" }
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.cleanupLinksForPDF()

	// Assert
	require.NoError(t, err)

	result, err := os.ReadFile(testFile)
	require.NoError(t, err)
	resultStr := string(result)

	// Diagram should get width constraint
	assert.Contains(t, resultStr, `![Diagram](diagram.png){ width="100%" }`)

	// Icon should NOT get width constraint (contains "icon")
	assert.Contains(t, resultStr, `![Icon](icon.png)`)
	assert.NotContains(t, resultStr, `![Icon](icon.png){ width="100%" }`)

	// Badge should NOT get width constraint (contains "badge")
	assert.Contains(t, resultStr, `![Badge](badge.svg)`)

	// Image with existing width should be unchanged
	assert.Contains(t, resultStr, `![Architecture](architecture.drawio.png){ width="100%" }`)
}

// TestConvertAttrListImagesToHTML_Integration verifies attr_list conversion on files
func TestConvertAttrListImagesToHTML_Integration(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	// Create test markdown file with attr_list images
	testFile := filepath.Join(stagingDir, "test.md")
	content := `# Test Document

![Icon](icon.png){width=48}
![Diagram](diagram.png){width="600" height="400"}
![Regular](image.png)
![Drawio](diagram.drawio){width=100}
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.convertAttrListImagesToHTML()

	// Assert
	require.NoError(t, err)

	result, err := os.ReadFile(testFile)
	require.NoError(t, err)
	resultStr := string(result)

	// Icon with width should be converted to HTML
	assert.Contains(t, resultStr, `<img src="icon.png" width="48" alt="Icon">`)

	// Diagram with width and height should be converted
	assert.Contains(t, resultStr, `<img src="diagram.png" width="600" height="400" alt="Diagram">`)

	// Regular markdown image should be unchanged
	assert.Contains(t, resultStr, `![Regular](image.png)`)

	// Drawio file should NOT be converted (needs markdown syntax for plugin)
	assert.Contains(t, resultStr, `![Drawio](diagram.drawio){width=100}`)
}

// TestCleanupLinksForPDF_EmptyDirectory verifies handling of empty directories
func TestCleanupLinksForPDF_EmptyDirectory(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.cleanupLinksForPDF()

	// Assert
	require.NoError(t, err) // Should not error on empty directory
}

// TestConvertAttrListImagesToHTML_MultipleFiles verifies processing multiple files
func TestConvertAttrListImagesToHTML_MultipleFiles(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	// Create multiple test files
	file1 := filepath.Join(stagingDir, "doc1.md")
	file2 := filepath.Join(stagingDir, "doc2.md")

	content1 := `![Image1](img1.png){width=100}`
	content2 := `![Image2](img2.png){width=200}`

	require.NoError(t, os.WriteFile(file1, []byte(content1), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(content2), 0644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.convertAttrListImagesToHTML()

	// Assert
	require.NoError(t, err)

	// Verify both files were processed
	result1, err := os.ReadFile(file1)
	require.NoError(t, err)
	assert.Contains(t, string(result1), `<img src="img1.png" width="100" alt="Image1">`)

	result2, err := os.ReadFile(file2)
	require.NoError(t, err)
	assert.Contains(t, string(result2), `<img src="img2.png" width="200" alt="Image2">`)
}
