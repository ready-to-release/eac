package staging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanOrphanedStagedFiles_RemovesUntrackedSourceFiles(t *testing.T) {
	stagingDir := t.TempDir()

	// Create files in staging
	trackedFile := filepath.Join(stagingDir, "tracked.md")
	orphanFile := filepath.Join(stagingDir, "orphan.md")
	require.NoError(t, os.WriteFile(trackedFile, []byte("tracked"), 0o644))
	require.NoError(t, os.WriteFile(orphanFile, []byte("orphan"), 0o644))

	// Only track one of them
	mapper := NewSimpleFileMap()
	mapper.AddFileMapping(trackedFile, "/source/tracked.md")

	noop := func(string, ...any) {}
	err := CleanOrphanedStagedFiles(stagingDir, mapper, noop)
	require.NoError(t, err)

	// Tracked file should still exist
	_, err = os.Stat(trackedFile)
	assert.NoError(t, err, "tracked file should not be removed")

	// Orphan file should be removed
	_, err = os.Stat(orphanFile)
	assert.True(t, os.IsNotExist(err), "orphan .md file should be removed")
}

func TestCleanOrphanedStagedFiles_PreservesNonSourceFiles(t *testing.T) {
	stagingDir := t.TempDir()

	// Create a non-source file type (not in orphanCleanExts)
	htmlFile := filepath.Join(stagingDir, "index.html")
	svgFile := filepath.Join(stagingDir, "diagram.svg")
	jsonFile := filepath.Join(stagingDir, "state.json")
	require.NoError(t, os.WriteFile(htmlFile, []byte("<html>"), 0o644))
	require.NoError(t, os.WriteFile(svgFile, []byte("<svg>"), 0o644))
	require.NoError(t, os.WriteFile(jsonFile, []byte("{}"), 0o644))

	mapper := NewSimpleFileMap() // nothing tracked

	noop := func(string, ...any) {}
	err := CleanOrphanedStagedFiles(stagingDir, mapper, noop)
	require.NoError(t, err)

	// Non-source files should be preserved even if untracked
	for _, f := range []string{htmlFile, svgFile, jsonFile} {
		_, err := os.Stat(f)
		assert.NoError(t, err, "non-source file %s should be preserved", f)
	}
}

func TestCleanOrphanedStagedFiles_RemovesOrphanImages(t *testing.T) {
	stagingDir := t.TempDir()

	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	var imageFiles []string
	for _, ext := range imageExts {
		f := filepath.Join(stagingDir, "orphan"+ext)
		require.NoError(t, os.WriteFile(f, []byte("data"), 0o644))
		imageFiles = append(imageFiles, f)
	}

	mapper := NewSimpleFileMap() // nothing tracked

	noop := func(string, ...any) {}
	err := CleanOrphanedStagedFiles(stagingDir, mapper, noop)
	require.NoError(t, err)

	for _, f := range imageFiles {
		_, err := os.Stat(f)
		assert.True(t, os.IsNotExist(err), "orphan image %s should be removed", f)
	}
}

func TestCleanOrphanedStagedFiles_HandlesNestedDirectories(t *testing.T) {
	stagingDir := t.TempDir()

	nestedDir := filepath.Join(stagingDir, "sub", "deep")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))

	orphanFile := filepath.Join(nestedDir, "orphan.md")
	require.NoError(t, os.WriteFile(orphanFile, []byte("content"), 0o644))

	mapper := NewSimpleFileMap()

	noop := func(string, ...any) {}
	err := CleanOrphanedStagedFiles(stagingDir, mapper, noop)
	require.NoError(t, err)

	_, err = os.Stat(orphanFile)
	assert.True(t, os.IsNotExist(err), "nested orphan should be removed")
}

func TestCleanOrphanedStagedFiles_EmptyDirectory(t *testing.T) {
	stagingDir := t.TempDir()

	mapper := NewSimpleFileMap()

	noop := func(string, ...any) {}
	err := CleanOrphanedStagedFiles(stagingDir, mapper, noop)
	require.NoError(t, err)
}

func TestCleanOrphanedStagedFiles_LogsRemovalCount(t *testing.T) {
	stagingDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "a.md"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "b.png"), []byte("b"), 0o644))

	mapper := NewSimpleFileMap()

	var logMessages []string
	logf := func(format string, args ...any) {
		logMessages = append(logMessages, format)
	}

	err := CleanOrphanedStagedFiles(stagingDir, mapper, logf)
	require.NoError(t, err)

	// Should have logged about removal
	assert.NotEmpty(t, logMessages, "should log when files are removed")
}

func TestCleanOrphanedStagedFiles_NoLogWhenNothingRemoved(t *testing.T) {
	stagingDir := t.TempDir()

	trackedFile := filepath.Join(stagingDir, "tracked.md")
	require.NoError(t, os.WriteFile(trackedFile, []byte("data"), 0o644))

	mapper := NewSimpleFileMap()
	mapper.AddFileMapping(trackedFile, "/source/tracked.md")

	var logMessages []string
	logf := func(format string, args ...any) {
		logMessages = append(logMessages, format)
	}

	err := CleanOrphanedStagedFiles(stagingDir, mapper, logf)
	require.NoError(t, err)

	// Should not log removal messages when nothing was removed
	for _, msg := range logMessages {
		assert.NotContains(t, msg, "orphaned", "should not log removal when nothing removed")
	}
}

func TestShouldCleanOrphan_SourceExtensions(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"markdown", "docs/guide.md", true},
		{"png", "assets/diagram.png", true},
		{"jpg", "assets/photo.jpg", true},
		{"jpeg", "assets/photo.jpeg", true},
		{"gif", "assets/animation.gif", true},
		{"webp", "assets/modern.webp", true},
		{"uppercase MD", "docs/GUIDE.MD", true},
		{"uppercase PNG", "assets/LOGO.PNG", true},
		{"svg excluded", "cache/mermaid.svg", false},
		{"html excluded", "site/index.html", false},
		{"yaml excluded", "config/nav.yml", false},
		{"json excluded", "cache/state.json", false},
		{"css excluded", "assets/style.css", false},
		{"no extension", "Makefile", false},
		{"go file", "main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ShouldCleanOrphan(tt.path),
				"ShouldCleanOrphan(%s)", tt.path)
		})
	}
}
