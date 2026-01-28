//go:build L0
// +build L0

package builders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeStagingHash(t *testing.T) {
	// Create temp staging directory with test files
	stagingDir := t.TempDir()

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "index.md"), []byte("# Hello"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "guide.md"), []byte("# Guide"), 0o644))

	// Compute hash
	hash1, err := ComputeStagingHash(stagingDir)
	require.NoError(t, err)
	assert.Len(t, hash1, 64) // SHA256 hex = 64 chars

	// Same content = same hash
	hash2, err := ComputeStagingHash(stagingDir)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Change content = different hash
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "index.md"), []byte("# Changed"), 0o644))
	hash3, err := ComputeStagingHash(stagingDir)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestComputeStagingHash_Deterministic(t *testing.T) {
	// Create staging directory
	stagingDir := t.TempDir()

	// Create files in arbitrary order
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "z.md"), []byte("# Z"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "a.md"), []byte("# A"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "m.md"), []byte("# M"), 0o644))

	// Hash should be same regardless of creation order (sorted internally)
	hash1, err := ComputeStagingHash(stagingDir)
	require.NoError(t, err)

	hash2, err := ComputeStagingHash(stagingDir)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2, "hash should be deterministic")
}

func TestBuildStateRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	state := &BookBuildState{
		BookName:    "test-book",
		ContentHash: "abc123def456",
		Theme:       "dark",
		OutputPath:  "/path/to/test-book-dark.pdf",
	}

	// Save state (uses tempDir as workspaceRoot, cache goes to tempDir/.cache/eac/build/state/)
	err := SaveBookBuildState(state, tempDir)
	require.NoError(t, err)

	// Load state
	loaded, err := LoadBookBuildState("test-book", "dark", tempDir)
	require.NoError(t, err)

	assert.Equal(t, state.BookName, loaded.BookName)
	assert.Equal(t, state.ContentHash, loaded.ContentHash)
	assert.Equal(t, state.Theme, loaded.Theme)
	assert.Equal(t, state.OutputPath, loaded.OutputPath)
}

func TestShouldSkipPDFGeneration(t *testing.T) {
	workspaceRoot := t.TempDir()
	stagingDir := filepath.Join(workspaceRoot, "staging")
	outputDir := filepath.Join(workspaceRoot, "out")
	require.NoError(t, os.MkdirAll(stagingDir, 0o755))
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	// Create staging content
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "index.md"), []byte("# Hello"), 0o644))

	// First build - no previous state
	canSkip, reason := ShouldSkipPDFGeneration("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	assert.False(t, canSkip)
	assert.Equal(t, "no previous build state", reason)

	// Record build complete
	err := RecordPDFBuildComplete("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	require.NoError(t, err)

	// Create fake PDF output
	pdfPath := filepath.Join(outputDir, "test-book-dark.pdf")
	require.NoError(t, os.WriteFile(pdfPath, []byte("PDF"), 0o644))

	// Second build - should skip (staging unchanged)
	canSkip, reason = ShouldSkipPDFGeneration("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	assert.True(t, canSkip)
	assert.Contains(t, reason, "staging unchanged")

	// Modify staging content - should rebuild
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "index.md"), []byte("# Changed"), 0o644))
	canSkip, reason = ShouldSkipPDFGeneration("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	assert.False(t, canSkip)
	assert.Equal(t, "staging content changed", reason)
}

func TestShouldSkipPDFGeneration_MissingPDF(t *testing.T) {
	workspaceRoot := t.TempDir()
	stagingDir := filepath.Join(workspaceRoot, "staging")
	outputDir := filepath.Join(workspaceRoot, "out")
	require.NoError(t, os.MkdirAll(stagingDir, 0o755))
	require.NoError(t, os.MkdirAll(outputDir, 0o755))

	// Create staging content
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "index.md"), []byte("# Hello"), 0o644))

	// Record build but don't create PDF
	err := RecordPDFBuildComplete("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	require.NoError(t, err)

	// Should not skip - PDF missing
	canSkip, reason := ShouldSkipPDFGeneration("test-book", "dark", stagingDir, workspaceRoot, outputDir)
	assert.False(t, canSkip)
	assert.Equal(t, "output PDF missing", reason)
}

func TestBuildStateCacheLocation(t *testing.T) {
	workspaceRoot := t.TempDir()

	// Cache should be in .cache/eac/build/state/, not out/build/
	expectedDir := filepath.Join(workspaceRoot, ".cache", "eac", "build", "state")
	actualDir := getBuildStateCacheDir(workspaceRoot)

	assert.Equal(t, expectedDir, actualDir)
}
