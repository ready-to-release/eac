package linking

import (
	"path/filepath"
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
