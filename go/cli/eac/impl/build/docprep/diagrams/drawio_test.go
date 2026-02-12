package diagrams

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindDrawioImages_DiscoversPngFiles(t *testing.T) {
	docsDir := t.TempDir()

	// Create .drawio.png files at various depths
	files := []string{
		filepath.Join(docsDir, "diagram.drawio.png"),
		filepath.Join(docsDir, "sub", "nested.drawio.png"),
		filepath.Join(docsDir, "sub", "deep", "deep.drawio.png"),
	}
	for _, f := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(f), 0o755))
		require.NoError(t, os.WriteFile(f, []byte("fake-png-content"), 0o644))
	}

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	assert.Len(t, images, 3)

	relPaths := make(map[string]bool)
	for _, img := range images {
		relPaths[filepath.ToSlash(img.RelPath)] = true
	}

	assert.True(t, relPaths["diagram.drawio.png"])
	assert.True(t, relPaths["sub/nested.drawio.png"])
	assert.True(t, relPaths["sub/deep/deep.drawio.png"])
}

func TestFindDrawioImages_IgnoresNonDrawioPng(t *testing.T) {
	docsDir := t.TempDir()

	// Create files that should NOT be matched
	nonDrawioFiles := []string{
		filepath.Join(docsDir, "screenshot.png"),
		filepath.Join(docsDir, "notes.drawio"),
		filepath.Join(docsDir, "readme.md"),
		filepath.Join(docsDir, "data.drawio.svg"),
	}
	for _, f := range nonDrawioFiles {
		require.NoError(t, os.WriteFile(f, []byte("content"), 0o644))
	}

	// Create one valid .drawio.png for contrast
	require.NoError(t, os.WriteFile(
		filepath.Join(docsDir, "valid.drawio.png"), []byte("png-data"), 0o644))

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	assert.Len(t, images, 1)
	assert.Equal(t, "valid.drawio.png", filepath.ToSlash(images[0].RelPath))
}

func TestFindDrawioImages_EmptyDirectory(t *testing.T) {
	docsDir := t.TempDir()

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	assert.Empty(t, images)
}

func TestFindDrawioImages_NonExistentDirectory(t *testing.T) {
	_, err := FindDrawioImages(filepath.Join(t.TempDir(), "nonexistent"))
	assert.Error(t, err)
}

func TestFindDrawioImages_PopulatesHash(t *testing.T) {
	docsDir := t.TempDir()

	content := []byte("deterministic-test-content")
	f := filepath.Join(docsDir, "hashed.drawio.png")
	require.NoError(t, os.WriteFile(f, content, 0o644))

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	require.Len(t, images, 1)

	assert.NotEmpty(t, images[0].Hash)
	assert.Len(t, images[0].Hash, 64, "SHA256 hex should be 64 characters")
}

func TestFindDrawioImages_HashIsDeterministic(t *testing.T) {
	docsDir := t.TempDir()

	content := []byte("same-content-every-time")
	f := filepath.Join(docsDir, "stable.drawio.png")
	require.NoError(t, os.WriteFile(f, content, 0o644))

	images1, err := FindDrawioImages(docsDir)
	require.NoError(t, err)

	images2, err := FindDrawioImages(docsDir)
	require.NoError(t, err)

	assert.Equal(t, images1[0].Hash, images2[0].Hash)
}

func TestFindDrawioImages_DifferentContentProducesDifferentHash(t *testing.T) {
	docsDir := t.TempDir()

	f1 := filepath.Join(docsDir, "a.drawio.png")
	f2 := filepath.Join(docsDir, "b.drawio.png")
	require.NoError(t, os.WriteFile(f1, []byte("content-a"), 0o644))
	require.NoError(t, os.WriteFile(f2, []byte("content-b"), 0o644))

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	require.Len(t, images, 2)

	assert.NotEqual(t, images[0].Hash, images[1].Hash)
}

func TestFindDrawioImages_SourceFileIsAbsolutePath(t *testing.T) {
	docsDir := t.TempDir()

	f := filepath.Join(docsDir, "abs.drawio.png")
	require.NoError(t, os.WriteFile(f, []byte("data"), 0o644))

	images, err := FindDrawioImages(docsDir)
	require.NoError(t, err)
	require.Len(t, images, 1)

	assert.True(t, filepath.IsAbs(images[0].SourceFile),
		"SourceFile should be absolute, got %s", images[0].SourceFile)
}

func TestHashFileContent_ReturnsSHA256Hex(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.bin")
	require.NoError(t, os.WriteFile(f, []byte("hello"), 0o644))

	hash, err := HashFileContent(f)
	require.NoError(t, err)

	// SHA256("hello") is a well-known value
	assert.Equal(t,
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		hash)
}

func TestHashFileContent_Deterministic(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "det.bin")
	require.NoError(t, os.WriteFile(f, []byte("deterministic"), 0o644))

	h1, err := HashFileContent(f)
	require.NoError(t, err)
	h2, err := HashFileContent(f)
	require.NoError(t, err)

	assert.Equal(t, h1, h2)
}

func TestHashFileContent_NonExistentFile(t *testing.T) {
	_, err := HashFileContent(filepath.Join(t.TempDir(), "missing.bin"))
	assert.Error(t, err)
}

func TestHashFileContent_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "empty.bin")
	require.NoError(t, os.WriteFile(f, []byte{}, 0o644))

	hash, err := HashFileContent(f)
	require.NoError(t, err)

	// SHA256 of empty input is the well-known "e3b0c44..." hash
	assert.Equal(t,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		hash)
}

func TestMaxImageWidthPDF(t *testing.T) {
	assert.Equal(t, 1200, MaxImageWidthPDF)
}
