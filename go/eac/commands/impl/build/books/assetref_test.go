//go:build L0
// +build L0

package books

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanAssetReferences(t *testing.T) {
	// Create temp directory with test markdown files
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs")
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "explanation"), 0o755))

	// Create test markdown with various asset references
	md1 := `# Test Document

![Diagram](../assets/assisted/diagram.png)

Some text here.

![Another](../assets/cd-model/overview.png)

[Link to PDF](../assets/lfs/report.pdf)

<img src="../assets/cache/mermaid/abc123.svg" alt="Mermaid">
`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "explanation", "test.md"), []byte(md1), 0o644))

	// Create another markdown with different references
	md2 := `# Another Document

![Logo](../assets/logo/icon.png)

Text with inline image: ![](../assets/architecture/component.png)
`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "explanation", "other.md"), []byte(md2), 0o644))

	// Scan for asset references
	refs, err := ScanAssetReferences(docsDir)
	require.NoError(t, err)

	// Verify expected references found
	// The scanner should return normalized paths relative to workspace
	assert.True(t, len(refs) >= 5, "Expected at least 5 asset references, got %d", len(refs))

	// Check that key asset directories are represented
	hasAssisted := false
	hasCdModel := false
	hasLogo := false
	hasCacheMermaid := false
	hasArchitecture := false

	for ref := range refs {
		if containsPath(ref, "assets/assisted") || containsPath(ref, "assets\\assisted") {
			hasAssisted = true
		}
		if containsPath(ref, "assets/cd-model") || containsPath(ref, "assets\\cd-model") {
			hasCdModel = true
		}
		if containsPath(ref, "assets/logo") || containsPath(ref, "assets\\logo") {
			hasLogo = true
		}
		if containsPath(ref, "assets/cache/mermaid") || containsPath(ref, "assets\\cache\\mermaid") {
			hasCacheMermaid = true
		}
		if containsPath(ref, "assets/architecture") || containsPath(ref, "assets\\architecture") {
			hasArchitecture = true
		}
	}

	assert.True(t, hasAssisted, "Should find reference to assets/assisted")
	assert.True(t, hasCdModel, "Should find reference to assets/cd-model")
	assert.True(t, hasLogo, "Should find reference to assets/logo")
	assert.True(t, hasCacheMermaid, "Should find reference to assets/cache/mermaid")
	assert.True(t, hasArchitecture, "Should find reference to assets/architecture")
}

func TestScanAssetReferences_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	refs, err := ScanAssetReferences(tempDir)
	require.NoError(t, err)
	assert.Empty(t, refs)
}

func TestScanAssetReferences_NoAssetRefs(t *testing.T) {
	tempDir := t.TempDir()

	md := `# Document with no assets

Just plain text.

[External link](https://example.com)

[Another file](./other.md)
`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "test.md"), []byte(md), 0o644))

	refs, err := ScanAssetReferences(tempDir)
	require.NoError(t, err)
	assert.Empty(t, refs)
}

func TestIsAssetPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"../assets/image.png", true},
		{"assets/logo/icon.png", true},
		{"./assets/cache/mermaid/abc.svg", true},
		{"../../assets/cd-model/diagram.png", true},
		{"./other.md", false},
		{"https://example.com/image.png", false},
		{"#anchor", false},
		{"mailto:test@example.com", false},
		{"../docs/guide.md", false},
		{"assets.md", false}, // file named assets.md, not in assets folder
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsAssetPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeAssetPath(t *testing.T) {
	tests := []struct {
		name         string
		ref          string
		mdFilePath   string
		workspaceDir string
		expected     string
	}{
		{
			name:         "relative path from nested markdown",
			ref:          "../assets/logo/icon.png",
			mdFilePath:   filepath.FromSlash("/workspace/docs/explanation/test.md"),
			workspaceDir: filepath.FromSlash("/workspace/docs"),
			expected:     filepath.FromSlash("assets/logo/icon.png"),
		},
		{
			name:         "double parent path",
			ref:          "../../assets/cd-model/diagram.png",
			mdFilePath:   filepath.FromSlash("/workspace/docs/explanation/sub/test.md"),
			workspaceDir: filepath.FromSlash("/workspace/docs"),
			expected:     filepath.FromSlash("assets/cd-model/diagram.png"),
		},
		{
			name:         "same directory reference",
			ref:          "assets/image.png",
			mdFilePath:   filepath.FromSlash("/workspace/docs/test.md"),
			workspaceDir: filepath.FromSlash("/workspace/docs"),
			expected:     filepath.FromSlash("assets/image.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAssetPath(tt.ref, tt.mdFilePath, tt.workspaceDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// containsPath checks if a path contains a substring (handles OS-specific separators)
func containsPath(path, substr string) bool {
	normalizedPath := filepath.ToSlash(path)
	normalizedSubstr := filepath.ToSlash(substr)
	return len(normalizedPath) > 0 && len(normalizedSubstr) > 0 &&
		(filepath.ToSlash(path) == normalizedSubstr ||
			len(normalizedPath) > len(normalizedSubstr) &&
				(normalizedPath[len(normalizedPath)-len(normalizedSubstr)-1:] == "/"+normalizedSubstr ||
					normalizedPath[:len(normalizedSubstr)+1] == normalizedSubstr+"/" ||
					containsSubstring(normalizedPath, "/"+normalizedSubstr+"/")))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
