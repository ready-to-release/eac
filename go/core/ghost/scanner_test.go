package ghost

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFileSource implements FileSource for testing
type mockFileSource struct {
	files []string
}

func (m *mockFileSource) TrackedFiles() ([]string, error) {
	return m.files, nil
}

func newMockFileSource(files []string) *mockFileSource {
	return &mockFileSource{files: files}
}

func TestScanner_DetectsGhostFiles(t *testing.T) {
	source := newMockFileSource([]string{
		"src/main.go",
		"src/ghost-feature.go",
		"tests/test.go",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	require.Len(t, ghosts, 1)
	assert.Equal(t, "src/ghost-feature.go", ghosts[0].Path)
	assert.Equal(t, "ghost-feature.go", ghosts[0].Name)
	assert.Equal(t, GhostTypeFile, ghosts[0].Type)
	assert.Equal(t, "feature.go", ghosts[0].GhostName)
}

func TestScanner_DetectsGhostDirectories(t *testing.T) {
	// Files inside a ghost directory - directory should be detected once
	source := newMockFileSource([]string{
		"ghost-feature/main.go",
		"ghost-feature/config.yaml",
		"src/normal.go",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)

	// Should find directory once (not per-file)
	ghostDirs := filterByType(ghosts, GhostTypeDirectory)
	require.Len(t, ghostDirs, 1)
	assert.Equal(t, "ghost-feature", ghostDirs[0].Path)
	assert.Equal(t, "ghost-feature", ghostDirs[0].Name)
	assert.Equal(t, "feature", ghostDirs[0].GhostName)
}

func TestScanner_DetectsNestedGhostDirectories(t *testing.T) {
	source := newMockFileSource([]string{
		"src/ghost-monitoring/probe.go",
		"src/ghost-monitoring/config.yaml",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)

	ghostDirs := filterByType(ghosts, GhostTypeDirectory)
	require.Len(t, ghostDirs, 1)
	assert.Equal(t, "src/ghost-monitoring", ghostDirs[0].Path)
	assert.Equal(t, "monitoring", ghostDirs[0].GhostName)
}

func TestScanner_CustomAlias(t *testing.T) {
	source := newMockFileSource([]string{
		"hidden-feature.go",
		"ghost-other.go", // Should NOT match with "hidden" alias
	})

	scanner := NewScanner(source, nil, ScanOptions{Alias: "hidden"})
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	require.Len(t, ghosts, 1)
	assert.Equal(t, "hidden-feature.go", ghosts[0].Path)
	assert.Equal(t, "feature.go", ghosts[0].GhostName)
}

func TestScanner_ExactMatchAlias(t *testing.T) {
	// A file/directory named exactly "ghost" (no suffix) should match
	source := newMockFileSource([]string{
		"ghost/config.yaml",
		"other/file.go",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)

	ghostDirs := filterByType(ghosts, GhostTypeDirectory)
	require.Len(t, ghostDirs, 1)
	assert.Equal(t, "ghost", ghostDirs[0].Path)
	assert.Equal(t, "", ghostDirs[0].GhostName) // Exact match has empty ghost name
}

func TestScanner_DotSuffixPattern(t *testing.T) {
	// ghost.config should match
	source := newMockFileSource([]string{
		"ghost.config",
		"ghost.yaml",
		"ghostly.go", // Should NOT match (no dash or dot after ghost)
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	require.Len(t, ghosts, 2)

	paths := make([]string, len(ghosts))
	for i, g := range ghosts {
		paths[i] = g.Path
	}
	assert.Contains(t, paths, "ghost.config")
	assert.Contains(t, paths, "ghost.yaml")
}

func TestScanner_NoGhosts(t *testing.T) {
	source := newMockFileSource([]string{
		"src/main.go",
		"tests/test.go",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	assert.Len(t, ghosts, 0)
}

func TestScanner_EmptyRepository(t *testing.T) {
	source := newMockFileSource([]string{})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	assert.Len(t, ghosts, 0)
}

func TestScanner_MixedGhostFilesAndDirectories(t *testing.T) {
	source := newMockFileSource([]string{
		"ghost-feature.go",
		"ghost-monitoring/probe.go",
		"src/ghost-api/handler.go",
		"src/normal.go",
	})

	scanner := NewScanner(source, nil, DefaultScanOptions())
	ghosts, err := scanner.Scan()

	require.NoError(t, err)

	files := filterByType(ghosts, GhostTypeFile)
	dirs := filterByType(ghosts, GhostTypeDirectory)

	assert.Len(t, files, 1)
	assert.Len(t, dirs, 2)

	assert.Equal(t, "ghost-feature.go", files[0].Path)
}

func TestDefaultScanOptions(t *testing.T) {
	opts := DefaultScanOptions()
	assert.Equal(t, "ghost", opts.Alias)
}

func TestScanner_DefaultAliasWhenEmpty(t *testing.T) {
	source := newMockFileSource([]string{
		"ghost-feature.go",
	})

	// Empty alias should default to "ghost"
	scanner := NewScanner(source, nil, ScanOptions{Alias: ""})
	ghosts, err := scanner.Scan()

	require.NoError(t, err)
	require.Len(t, ghosts, 1)
	assert.Equal(t, "ghost-feature.go", ghosts[0].Path)
}

// Helper function to filter ghosts by type
func filterByType(ghosts []Ghost, ghostType GhostType) []Ghost {
	var result []Ghost
	for _, g := range ghosts {
		if g.Type == ghostType {
			result = append(result, g)
		}
	}
	return result
}
