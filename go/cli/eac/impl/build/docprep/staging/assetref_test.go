package staging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAssetPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"assets/images/logo.png", true},
		{"./assets/images/logo.png", true},
		{"../assets/images/logo.png", true},
		{"../../assets/css/style.css", true},
		{"http://example.com/image.png", false},
		{"https://example.com/image.png", false},
		{"#anchor", false},
		{"mailto:test@example.com", false},
		{"guide.md", false},
		{"../guide.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAssetPath(tt.path))
		})
	}
}

func TestNormalizeAssetPath(t *testing.T) {
	workspaceDir := filepath.FromSlash("/workspace/docs")

	tests := []struct {
		name     string
		ref      string
		mdFile   string
		expected string
	}{
		{
			name:     "relative path from nested dir",
			ref:      "../assets/logo/icon.png",
			mdFile:   filepath.FromSlash("/workspace/docs/explanation/test.md"),
			expected: filepath.FromSlash("assets/logo/icon.png"),
		},
		{
			name:     "direct asset reference",
			ref:      "assets/images/diagram.png",
			mdFile:   filepath.FromSlash("/workspace/docs/index.md"),
			expected: filepath.FromSlash("assets/images/diagram.png"),
		},
		{
			name:     "escaping workspace returns empty",
			ref:      "../../../../etc/passwd",
			mdFile:   filepath.FromSlash("/workspace/docs/test.md"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAssetPath(tt.ref, tt.mdFile, workspaceDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAssetReferenced(t *testing.T) {
	refs := map[string]bool{
		"assets/images/logo.png": true,
		"assets/css/style.css":   true,
	}

	// Direct match
	assert.True(t, IsAssetReferenced("assets/images/logo.png", refs))

	// Basename match
	assert.True(t, IsAssetReferenced("/full/path/to/logo.png", refs))

	// Not referenced
	assert.False(t, IsAssetReferenced("assets/images/other.png", refs))

	// Empty refs = allow all
	assert.True(t, IsAssetReferenced("anything.png", nil))
	assert.True(t, IsAssetReferenced("anything.png", map[string]bool{}))
}

func TestScanAssetReferences(t *testing.T) {
	dir := t.TempDir()

	md := `# Test
![logo](assets/images/logo.png)
[link](../guide.md)
<img src="assets/css/../images/diagram.png" />
<a href="https://example.com">external</a>
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.md"), []byte(md), 0o644))

	refs, err := ScanAssetReferences(dir)
	require.NoError(t, err)

	// Should find asset references but not external URLs or non-asset links
	assert.NotEmpty(t, refs)
}
