package staging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileIndex(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "page.md"), []byte("# Page"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "image.png"), []byte("PNG"), 0o644))

	idx, err := NewFileIndex(dir)
	require.NoError(t, err)

	assert.Equal(t, 3, idx.FileCount())
	assert.Equal(t, 2, idx.MarkdownCount())
}

func TestFileIndex_GetFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644))

	idx, err := NewFileIndex(dir)
	require.NoError(t, err)

	mdFiles := idx.GetMarkdownFiles()
	assert.Len(t, mdFiles, 1)

	allFiles := idx.GetAllFiles()
	assert.Len(t, allFiles, 2)
}

func TestFileIndex_Refresh(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.md"), []byte("a"), 0o644))

	idx, err := NewFileIndex(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, idx.FileCount())

	// Add a file and refresh
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.md"), []byte("b"), 0o644))
	require.NoError(t, idx.Refresh())

	assert.Equal(t, 2, idx.FileCount())
	assert.Equal(t, 2, idx.MarkdownCount())
}

func TestFileIndex_RemoveFile(t *testing.T) {
	dir := t.TempDir()

	mdFile := filepath.Join(dir, "test.md")
	txtFile := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(mdFile, []byte("md"), 0o644))
	require.NoError(t, os.WriteFile(txtFile, []byte("txt"), 0o644))

	idx, err := NewFileIndex(dir)
	require.NoError(t, err)
	assert.Equal(t, 2, idx.FileCount())
	assert.Equal(t, 1, idx.MarkdownCount())

	idx.RemoveFile(mdFile)
	assert.Equal(t, 1, idx.FileCount())
	assert.Equal(t, 0, idx.MarkdownCount())
}

func TestFileIndex_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	idx, err := NewFileIndex(dir)
	require.NoError(t, err)

	assert.Equal(t, 0, idx.FileCount())
	assert.Equal(t, 0, idx.MarkdownCount())
}
