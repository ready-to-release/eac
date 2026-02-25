package linking

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslationMap_FileMapping(t *testing.T) {
	m := NewTranslationMap("/workspace", "/staging", false, nil)

	m.AddFileMapping("/staging/index.md", "/workspace/docs/index.md")
	m.AddFileMapping("/staging/guide/intro.md", "/workspace/docs/guide/intro.md")

	src, ok := m.GetSourcePath("/staging/index.md")
	assert.True(t, ok)
	assert.Equal(t, "/workspace/docs/index.md", src)

	src, ok = m.GetSourcePath("/staging/guide/intro.md")
	assert.True(t, ok)
	assert.Equal(t, "/workspace/docs/guide/intro.md", src)

	_, ok = m.GetSourcePath("/staging/nonexistent.md")
	assert.False(t, ok)
}

func TestTranslationMap_Empty(t *testing.T) {
	m := NewTranslationMap("/workspace", "/staging", false, nil)

	_, ok := m.GetSourcePath("/staging/file.md")
	assert.False(t, ok)
}

func TestCalculateRelativePath(t *testing.T) {
	stagingFile := filepath.Join("staging", "docs", "guide", "page.md")
	targetPath := filepath.Join("staging", "docs", "assets", "image.png")

	relPath, err := CalculateRelativePath(stagingFile, targetPath)
	require.NoError(t, err)
	assert.Equal(t, "../assets/image.png", relPath)
}

func TestCalculateRelativePath_SameDir(t *testing.T) {
	stagingFile := filepath.Join("staging", "docs", "page.md")
	targetPath := filepath.Join("staging", "docs", "other.md")

	relPath, err := CalculateRelativePath(stagingFile, targetPath)
	require.NoError(t, err)
	assert.Equal(t, "other.md", relPath)
}

func TestBuildTranslations_ReportsAllBrokenLinks(t *testing.T) {
	// Create a temp workspace with source files containing broken links.
	tmpDir := t.TempDir()
	sourceRoot := filepath.Join(tmpDir, "workspace")
	stagingRoot := filepath.Join(tmpDir, "staging")
	docsDir := filepath.Join(sourceRoot, "docs")
	sourceDir := filepath.Join(docsDir, "commands")
	stagingDir := filepath.Join(stagingRoot, "commands")

	require.NoError(t, os.MkdirAll(sourceDir, 0o755))
	require.NoError(t, os.MkdirAll(stagingDir, 0o755))

	// File with two broken links
	file1 := filepath.Join(sourceDir, "page1.md")
	require.NoError(t, os.WriteFile(file1, []byte(
		"See [foo](./missing-a.md) and [bar](./missing-b.md)\n",
	), 0o644))

	// Another file with one broken link
	file2 := filepath.Join(sourceDir, "page2.md")
	require.NoError(t, os.WriteFile(file2, []byte(
		"See [baz](./missing-c.md)\n",
	), 0o644))

	m := NewTranslationMap(sourceRoot, stagingRoot, false, nil)
	m.AddFileMapping(filepath.Join(stagingDir, "page1.md"), file1)
	m.AddFileMapping(filepath.Join(stagingDir, "page2.md"), file2)

	err := m.BuildTranslations("")
	require.Error(t, err)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "BROKEN LINKS (3 found)")
	assert.Contains(t, errMsg, "missing-a.md")
	assert.Contains(t, errMsg, "missing-b.md")
	assert.Contains(t, errMsg, "missing-c.md")

	// Verify all three are listed (one per line after the header)
	lines := strings.Split(errMsg, "\n")
	brokenCount := 0
	for _, line := range lines {
		if strings.Contains(line, "missing-") {
			brokenCount++
		}
	}
	assert.Equal(t, 3, brokenCount, "should list all 3 broken links")
}
