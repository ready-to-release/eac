package staging

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateRelativePath(t *testing.T) {
	workspaceRoot := filepath.FromSlash("/workspace")

	tests := []struct {
		name        string
		matchPath   string
		pattern     string
		expected    string
		expectError bool
	}{
		{
			name:      "simple file in directory",
			matchPath: filepath.FromSlash("/workspace/docs/guide.md"),
			pattern:   "docs/**/*.md",
			expected:  "guide.md",
		},
		{
			name:      "file in nested directory",
			matchPath: filepath.FromSlash("/workspace/docs/reference/api.md"),
			pattern:   "docs/**/*.md",
			expected:  filepath.FromSlash("reference/api.md"),
		},
		{
			name:      "file at root of pattern",
			matchPath: filepath.FromSlash("/workspace/docs/index.md"),
			pattern:   "docs/**",
			expected:  "index.md",
		},
		{
			name:      "single file pattern",
			matchPath: filepath.FromSlash("/workspace/README.md"),
			pattern:   "README.md",
			expected:  "README.md",
		},
		{
			name:      "file in subdirectory with specific extension",
			matchPath: filepath.FromSlash("/workspace/assets/images/logo.png"),
			pattern:   "assets/**/*.png",
			expected:  filepath.FromSlash("images/logo.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateRelativePath(tt.matchPath, tt.pattern, workspaceRoot)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsExcluded(t *testing.T) {
	workspaceRoot := filepath.FromSlash("/workspace")

	tests := []struct {
		name     string
		path     string
		exclude  []string
		expected bool
	}{
		{
			name:     "file not excluded",
			path:     filepath.FromSlash("/workspace/docs/guide.md"),
			exclude:  []string{"**/*.tmp", "**/*.bak"},
			expected: false,
		},
		{
			name:     "file matches extension exclusion",
			path:     filepath.FromSlash("/workspace/docs/temp.tmp"),
			exclude:  []string{"**/*.tmp"},
			expected: true,
		},
		{
			name:     "file in lfs directory",
			path:     filepath.FromSlash("/workspace/lfs/large-file.bin"),
			exclude:  []string{},
			expected: true,
		},
		{
			name:     "file matches directory exclusion",
			path:     filepath.FromSlash("/workspace/node_modules/package/file.js"),
			exclude:  []string{"node_modules/**"},
			expected: true,
		},
		{
			name:     "no exclusion patterns",
			path:     filepath.FromSlash("/workspace/docs/guide.md"),
			exclude:  []string{},
			expected: false,
		},
		{
			name:     "partial path match does not exclude",
			path:     filepath.FromSlash("/workspace/docs/temporary-guide.md"),
			exclude:  []string{"**/*.tmp"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExcluded(tt.path, workspaceRoot, tt.exclude)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAssetNeeded(t *testing.T) {
	refs := map[string]bool{
		"assets/images/logo.png": true,
	}

	// Essential dirs always needed
	assert.True(t, IsAssetNeeded("/workspace/docs/assets/css/style.css", refs))
	assert.True(t, IsAssetNeeded("/workspace/docs/assets/js/script.js", refs))
	assert.True(t, IsAssetNeeded("/workspace/docs/assets/logo/icon.png", refs))
	assert.True(t, IsAssetNeeded("/workspace/docs/assets/templates/main.html", refs))

	// manifest.json always needed
	assert.True(t, IsAssetNeeded("/workspace/docs/assets/manifest.json", refs))

	// Referenced asset
	assert.True(t, IsAssetNeeded("assets/images/logo.png", refs))

	// Unreferenced asset
	assert.False(t, IsAssetNeeded("/other/path/unreferenced.png", map[string]bool{"some/other.png": true}))
}

func TestShouldCleanOrphan(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"markdown file", "docs/guide.md", true},
		{"png image", "assets/diagram.png", true},
		{"jpg image", "assets/photo.jpg", true},
		{"jpeg image", "assets/photo.jpeg", true},
		{"gif image", "assets/animation.gif", true},
		{"webp image", "assets/modern.webp", true},
		{"svg file (excluded - generated)", "cache/mermaid.svg", false},
		{"html file (excluded)", "site/index.html", false},
		{"yaml file (excluded)", "config/nav.yml", false},
		{"json file (excluded)", "cache/state.json", false},
		{"css file (excluded)", "assets/style.css", false},
		{"no extension (excluded)", "Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldCleanOrphan(tt.path)
			assert.Equal(t, tt.expected, result, "ShouldCleanOrphan(%s)", tt.path)
		})
	}
}

func TestSimpleFileMap(t *testing.T) {
	m := NewSimpleFileMap()

	m.AddFileMapping("/staging/index.md", "/source/index.md")
	m.AddFileMapping("/staging/guide.md", "/source/guide.md")

	src, ok := m.GetSourcePath("/staging/index.md")
	assert.True(t, ok)
	assert.Equal(t, "/source/index.md", src)

	_, ok = m.GetSourcePath("/staging/nonexistent.md")
	assert.False(t, ok)
}
