package staging

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirrorSync_NewFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "file1.txt"), []byte("content1"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(src, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "subdir", "file2.txt"), []byte("content2"), 0o644))

	stats, err := MirrorSync(src, dst)
	require.NoError(t, err)

	assert.Equal(t, 2, stats.Copied)
	assert.Equal(t, 0, stats.Skipped)
	assert.Equal(t, 0, stats.Deleted)

	content1, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	content2, err := os.ReadFile(filepath.Join(dst, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
}

func TestMirrorSync_UnchangedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "file.txt"), []byte("content"), 0o644))

	stats1, err := MirrorSync(src, dst)
	require.NoError(t, err)
	assert.Equal(t, 1, stats1.Copied)

	stats2, err := MirrorSync(src, dst)
	require.NoError(t, err)
	assert.Equal(t, 0, stats2.Copied)
	assert.Equal(t, 1, stats2.Skipped)
}

func TestMirrorSync_ChangedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "file.txt")
	dstFile := filepath.Join(dst, "file.txt")

	require.NoError(t, os.WriteFile(srcFile, []byte("original"), 0o644))
	_, err := MirrorSync(src, dst)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	require.NoError(t, os.WriteFile(srcFile, []byte("modified content"), 0o644))

	stats, err := MirrorSync(src, dst)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Copied)
	assert.Equal(t, 0, stats.Skipped)

	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "modified content", string(content))
}

func TestMirrorSync_DeletedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "keep.txt"), []byte("keep"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "delete.txt"), []byte("delete"), 0o644))

	_, err := MirrorSync(src, dst)
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(src, "delete.txt")))

	stats, err := MirrorSync(src, dst)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Deleted)

	_, err = os.Stat(filepath.Join(dst, "delete.txt"))
	assert.True(t, os.IsNotExist(err))
}

func TestMirrorSync_EmptyDirectoryCleanup(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(src, "a", "b"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "a", "b", "file.txt"), []byte("x"), 0o644))

	_, err := MirrorSync(src, dst)
	require.NoError(t, err)

	require.NoError(t, os.Remove(filepath.Join(src, "a", "b", "file.txt")))
	require.NoError(t, os.Remove(filepath.Join(src, "a", "b")))
	require.NoError(t, os.Remove(filepath.Join(src, "a")))

	stats, err := MirrorSync(src, dst)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Deleted)

	_, err = os.Stat(filepath.Join(dst, "a"))
	assert.True(t, os.IsNotExist(err))
}

func TestMirrorSync_PreservesMtime(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "file.txt")
	dstFile := filepath.Join(dst, "file.txt")

	require.NoError(t, os.WriteFile(srcFile, []byte("content"), 0o644))
	mtime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(srcFile, mtime, mtime))

	_, err := MirrorSync(src, dst)
	require.NoError(t, err)

	dstInfo, err := os.Stat(dstFile)
	require.NoError(t, err)
	assert.WithinDuration(t, mtime, dstInfo.ModTime(), time.Second)
}

func TestNeedsCopy(t *testing.T) {
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "src.txt")
	dstFile := filepath.Join(tempDir, "dst.txt")

	require.NoError(t, os.WriteFile(srcFile, []byte("content"), 0o644))
	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)

	// dst doesn't exist
	assert.True(t, NeedsCopy(srcInfo, dstFile))

	// Create identical dst
	require.NoError(t, os.WriteFile(dstFile, []byte("content"), 0o644))
	require.NoError(t, os.Chtimes(dstFile, srcInfo.ModTime(), srcInfo.ModTime()))
	assert.False(t, NeedsCopy(srcInfo, dstFile))

	// Different size
	require.NoError(t, os.WriteFile(dstFile, []byte("different"), 0o644))
	require.NoError(t, os.Chtimes(dstFile, srcInfo.ModTime(), srcInfo.ModTime()))
	assert.True(t, NeedsCopy(srcInfo, dstFile))

	// Different mtime
	require.NoError(t, os.WriteFile(dstFile, []byte("content"), 0o644))
	newMtime := srcInfo.ModTime().Add(time.Hour)
	require.NoError(t, os.Chtimes(dstFile, newMtime, newMtime))
	assert.True(t, NeedsCopy(srcInfo, dstFile))
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.Mkdir(srcDir, 0o755))

	srcFile := filepath.Join(srcDir, "test.txt")
	content := "Test file content\nLine 2\nLine 3"
	require.NoError(t, os.WriteFile(srcFile, []byte(content), 0o644))

	dstFile := filepath.Join(tempDir, "dst", "subdir", "test.txt")
	err := CopyFile(srcFile, dstFile)
	require.NoError(t, err)

	dstContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(dstContent))
}

func TestCopyFile_MissingSource(t *testing.T) {
	tempDir := t.TempDir()
	err := CopyFile(filepath.Join(tempDir, "nonexistent.txt"), filepath.Join(tempDir, "dest.txt"))
	assert.Error(t, err)
}
