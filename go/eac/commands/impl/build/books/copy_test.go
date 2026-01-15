package books

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCalculateRelativePath verifies relative path calculation
func TestCalculateRelativePath(t *testing.T) {
	workspaceRoot := filepath.FromSlash("/workspace")

	tests := []struct {
		name          string
		matchPath     string
		pattern       string
		workspaceRoot string
		expected      string
		expectError   bool
	}{
		{
			name:          "simple file in directory",
			matchPath:     filepath.FromSlash("/workspace/docs/guide.md"),
			pattern:       "docs/**/*.md",
			workspaceRoot: workspaceRoot,
			expected:      "guide.md",
			expectError:   false,
		},
		{
			name:          "file in nested directory",
			matchPath:     filepath.FromSlash("/workspace/docs/reference/api.md"),
			pattern:       "docs/**/*.md",
			workspaceRoot: workspaceRoot,
			expected:      filepath.FromSlash("reference/api.md"),
			expectError:   false,
		},
		{
			name:          "file at root of pattern",
			matchPath:     filepath.FromSlash("/workspace/docs/index.md"),
			pattern:       "docs/**",
			workspaceRoot: workspaceRoot,
			expected:      "index.md",
			expectError:   false,
		},
		{
			name:          "single file pattern",
			matchPath:     filepath.FromSlash("/workspace/README.md"),
			pattern:       "README.md",
			workspaceRoot: workspaceRoot,
			expected:      "README.md",
			expectError:   false,
		},
		{
			name:          "file in subdirectory with specific extension",
			matchPath:     filepath.FromSlash("/workspace/assets/images/logo.png"),
			pattern:       "assets/**/*.png",
			workspaceRoot: workspaceRoot,
			expected:      filepath.FromSlash("images/logo.png"),
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculateRelativePath(tt.matchPath, tt.pattern, tt.workspaceRoot)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestIsExcluded verifies exclusion pattern matching
func TestIsExcluded(t *testing.T) {
	workspaceRoot := filepath.FromSlash("/workspace")

	tests := []struct {
		name            string
		path            string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "file not excluded",
			path:            filepath.FromSlash("/workspace/docs/guide.md"),
			excludePatterns: []string{"**/*.tmp", "**/*.bak"},
			expected:        false,
		},
		{
			name:            "file matches extension exclusion",
			path:            filepath.FromSlash("/workspace/docs/temp.tmp"),
			excludePatterns: []string{"**/*.tmp"},
			expected:        true,
		},
		{
			name:            "file in lfs directory",
			path:            filepath.FromSlash("/workspace/lfs/large-file.bin"),
			excludePatterns: []string{},
			expected:        true, // lfs directories are always excluded
		},
		{
			name:            "file matches directory exclusion",
			path:            filepath.FromSlash("/workspace/node_modules/package/file.js"),
			excludePatterns: []string{"node_modules/**"},
			expected:        true,
		},
		{
			name:            "file matches specific file exclusion",
			path:            filepath.FromSlash("/workspace/docs/.DS_Store"),
			excludePatterns: []string{"**/.DS_Store"},
			expected:        true,
		},
		{
			name:            "file matches multiple patterns",
			path:            filepath.FromSlash("/workspace/cache/temp.tmp"),
			excludePatterns: []string{"cache/**", "**/*.tmp"},
			expected:        true,
		},
		{
			name:            "no exclusion patterns",
			path:            filepath.FromSlash("/workspace/docs/guide.md"),
			excludePatterns: []string{},
			expected:        false,
		},
		{
			name:            "partial path match does not exclude",
			path:            filepath.FromSlash("/workspace/docs/temporary-guide.md"),
			excludePatterns: []string{"**/*.tmp"},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExcluded(tt.path, workspaceRoot, tt.excludePatterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCopyFile verifies file copying functionality
func TestCopyFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	require.NoError(t, os.Mkdir(srcDir, 0755))

	srcFile := filepath.Join(srcDir, "test.txt")
	content := "Test file content\nLine 2\nLine 3"
	require.NoError(t, os.WriteFile(srcFile, []byte(content), 0644))

	// Act
	dstFile := filepath.Join(dstDir, "subdir", "test.txt")
	err := copyFile(srcFile, dstFile)

	// Assert
	require.NoError(t, err)

	// Verify destination file exists
	_, err = os.Stat(dstFile)
	require.NoError(t, err)

	// Verify content matches
	dstContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(dstContent))
}

// TestCopyFile_NestedDirectories verifies directory creation
func TestCopyFile_NestedDirectories(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("content"), 0644))

	// Act - Copy to deeply nested path
	dstFile := filepath.Join(tempDir, "a", "b", "c", "d", "destination.txt")
	err := copyFile(srcFile, dstFile)

	// Assert
	require.NoError(t, err)

	// Verify all intermediate directories were created
	_, err = os.Stat(filepath.Join(tempDir, "a", "b", "c", "d"))
	require.NoError(t, err)

	// Verify file was copied
	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "content", string(content))
}

// TestCopyFile_PreservePermissions verifies file permissions are preserved
func TestCopyFile_PreservePermissions(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("content"), 0644))

	// Change permissions on source file
	require.NoError(t, os.Chmod(srcFile, 0600))

	// Act
	dstFile := filepath.Join(tempDir, "destination.txt")
	err := copyFile(srcFile, dstFile)

	// Assert
	require.NoError(t, err)

	// Verify permissions match
	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstFile)
	require.NoError(t, err)

	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

// TestCopyFile_MissingSource verifies error handling
func TestCopyFile_MissingSource(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "nonexistent.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	// Act
	err := copyFile(srcFile, dstFile)

	// Assert
	assert.Error(t, err)
}

// TestCopyFile_Overwrite verifies existing files are overwritten
func TestCopyFile_Overwrite(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")

	// Create source and destination with different content
	require.NoError(t, os.WriteFile(srcFile, []byte("new content"), 0644))
	require.NoError(t, os.WriteFile(dstFile, []byte("old content"), 0644))

	// Act
	err := copyFile(srcFile, dstFile)

	// Assert
	require.NoError(t, err)

	// Verify destination has new content
	content, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(content))
}
