package books

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnsureRootIndex_ExistingFile verifies that existing index.md is preserved.
func TestEnsureRootIndex_ExistingFile(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()
	indexPath := filepath.Join(stagingDir, "index.md")
	existingContent := "# Existing Index\nThis should be preserved."
	require.NoError(t, os.WriteFile(indexPath, []byte(existingContent), 0o644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.ensureRootIndex()

	// Assert
	require.NoError(t, err)
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(content), "Existing index.md should be preserved")
}

// TestEnsureRootIndex_GenerateNew verifies that index.md is generated when missing.
func TestEnsureRootIndex_GenerateNew(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()
	p := &Preprocessor{
		stagingDir: stagingDir,
		book: &config.Book{
			Name:        "Test Book",
			Title:       "Test Book Title",
			Description: "Test book description",
		},
		logWriter: &bytes.Buffer{},
	}

	// Act
	err := p.ensureRootIndex()

	// Assert
	require.NoError(t, err)
	indexPath := filepath.Join(stagingDir, "index.md")
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "title: Test Book Title")
	assert.Contains(t, contentStr, "description: Test book description")
	assert.Contains(t, contentStr, "# Test Book Title")
}

// TestGenerateTOC_Empty verifies TOC generation with no files.
func TestGenerateTOC_Empty(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()
	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	toc := p.generateTOC()

	// Assert
	assert.Empty(t, toc, "TOC should be empty when no files exist")
}

// TestGenerateTOC_WithFiles verifies TOC generation with markdown files.
func TestGenerateTOC_WithFiles(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	// Create test markdown files
	file1 := filepath.Join(stagingDir, "intro.md")
	file2 := filepath.Join(stagingDir, "guide.md")

	require.NoError(t, os.WriteFile(file1, []byte("# Introduction\nIntro content"), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte("# User Guide\nGuide content"), 0o644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	toc := p.generateTOC()

	// Assert
	assert.NotEmpty(t, toc)
	assert.Contains(t, toc, "Introduction")
	assert.Contains(t, toc, "User Guide")
}

// TestGetTitleFromFile_Frontmatter verifies title extraction from frontmatter.
func TestGetTitleFromFile_Frontmatter(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	content := `---
title: "Frontmatter Title"
---
# Heading Title
Content here.`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	p := &Preprocessor{}

	// Act
	title := p.getTitleFromFile(filePath)

	// Assert
	assert.Equal(t, "Frontmatter Title", title, "Should extract title from frontmatter")
}

// TestGetTitleFromFile_Heading verifies title extraction from H1 heading.
func TestGetTitleFromFile_Heading(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	content := `# Heading Title
Content here.`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	p := &Preprocessor{}

	// Act
	title := p.getTitleFromFile(filePath)

	// Assert
	assert.Equal(t, "Heading Title", title, "Should extract title from H1 heading")
}

// TestGetTitleFromFile_Filename verifies title extraction from filename fallback.
func TestGetTitleFromFile_Filename(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	filePath := filepath.Join(dir, "user-guide.md")
	content := `No title here.`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	p := &Preprocessor{}

	// Act
	title := p.getTitleFromFile(filePath)

	// Assert
	assert.Equal(t, "User Guide", title, "Should generate title from filename")
}

// TestFilenameToTitle verifies filename to title conversion.
func TestFilenameToTitle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "simple filename",
			filename: "readme.md",
			expected: "Readme",
		},
		{
			name:     "dashed filename",
			filename: "user-guide.md",
			expected: "User Guide",
		},
		{
			name:     "underscored filename",
			filename: "api_reference.md",
			expected: "Api Reference",
		},
		{
			name:     "mixed separators",
			filename: "quick-start_guide.md",
			expected: "Quick Start Guide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filenameToTitle(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetOrderForFile verifies command source ordering.
func TestGetOrderForFile(t *testing.T) {
	// Arrange
	p := &Preprocessor{
		book: &config.Book{
			Sources: []config.Source{
				{
					Type:   "command",
					Target: "intro.md",
					Order:  100,
				},
				{
					Type:   "command",
					Target: "guide.md",
					Order:  200,
				},
			},
		},
		logWriter: &bytes.Buffer{},
	}

	tests := []struct {
		name     string
		relPath  string
		expected int
	}{
		{
			name:     "file with order",
			relPath:  "intro.md",
			expected: 100,
		},
		{
			name:     "file with different order",
			relPath:  "guide.md",
			expected: 200,
		},
		{
			name:     "file without order",
			relPath:  "other.md",
			expected: 500, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.getOrderForFile(tt.relPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCollectTOCFromFilesystem verifies filesystem-based TOC collection.
func TestCollectTOCFromFilesystem(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()

	// Create test structure
	file1 := filepath.Join(stagingDir, "a-first.md")
	file2 := filepath.Join(stagingDir, "z-last.md")
	subdir := filepath.Join(stagingDir, "subdir")
	require.NoError(t, os.Mkdir(subdir, 0o755))
	file3 := filepath.Join(subdir, "index.md")

	require.NoError(t, os.WriteFile(file1, []byte("# First\nContent"), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte("# Last\nContent"), 0o644))
	require.NoError(t, os.WriteFile(file3, []byte("# Subdir Index\nContent"), 0o644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	entries := p.collectTOCFromFilesystem(stagingDir, 0)

	// Assert
	assert.NotEmpty(t, entries)

	// Files should be ordered alphabetically
	var fileTitles []string
	for _, entry := range entries {
		if entry.depth == 0 && filepath.Ext(entry.path) == ".md" {
			fileTitles = append(fileTitles, entry.title)
		}
	}

	// Should have files in order
	assert.Contains(t, fileTitles, "First")
	assert.Contains(t, fileTitles, "Last")
}

// TestToTitleCase verifies title case conversion.
func TestToTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "uppercase",
			input:    "HELLO WORLD",
			expected: "Hello World",
		},
		{
			name:     "mixed case",
			input:    "HeLLo WoRLd",
			expected: "Hello World",
		},
		{
			name:     "single word",
			input:    "test",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toTitleCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateNavEntry_ValidFile verifies validation of valid markdown file.
func TestValidateNavEntry_ValidFile(t *testing.T) {
	// Arrange
	p := &Preprocessor{
		stagingDir: t.TempDir(),
		logWriter:  &bytes.Buffer{},
	}
	actualFiles := map[string]bool{"test.md": true}
	actualDirs := map[string]bool{}

	// Act
	result, referenced := p.validateNavEntry("test.md", "", actualFiles, actualDirs)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "test.md", result)
	assert.True(t, referenced["test.md"])
}

// TestValidateNavEntry_MissingFile verifies validation rejects missing files.
func TestValidateNavEntry_MissingFile(t *testing.T) {
	// Arrange
	p := &Preprocessor{
		stagingDir: t.TempDir(),
		logWriter:  &bytes.Buffer{},
	}
	actualFiles := map[string]bool{}
	actualDirs := map[string]bool{}

	// Act
	result, _ := p.validateNavEntry("missing.md", "", actualFiles, actualDirs)

	// Assert
	assert.Nil(t, result, "Missing file should return nil")
}

// TestValidateNavEntry_ValidDirectory verifies validation of valid directory.
func TestValidateNavEntry_ValidDirectory(t *testing.T) {
	// Arrange
	p := &Preprocessor{
		stagingDir: t.TempDir(),
		logWriter:  &bytes.Buffer{},
	}
	actualFiles := map[string]bool{}
	actualDirs := map[string]bool{"subdir": true}

	// Act
	result, referenced := p.validateNavEntry("subdir/", "", actualFiles, actualDirs)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "subdir/", result)
	assert.True(t, referenced["subdir/"])
}

// TestValidateNavEntry_TitledEntry verifies validation of titled entries.
func TestValidateNavEntry_TitledEntry(t *testing.T) {
	// Arrange
	p := &Preprocessor{
		stagingDir: t.TempDir(),
		logWriter:  &bytes.Buffer{},
	}
	actualFiles := map[string]bool{"intro.md": true}
	actualDirs := map[string]bool{}

	entry := map[string]any{
		"Introduction": "intro.md",
	}

	// Act
	result, referenced := p.validateNavEntry(entry, "", actualFiles, actualDirs)

	// Assert
	assert.NotNil(t, result)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "intro.md", resultMap["Introduction"])
	assert.True(t, referenced["intro.md"])
}

// TestGenerateNavForDir_EmptyDir verifies nav generation for empty directory.
func TestGenerateNavForDir_EmptyDir(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()
	emptyDir := filepath.Join(stagingDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.generateNavForDir(emptyDir)

	// Assert
	require.NoError(t, err)
	// Should not create .nav.yml for empty directory
	navPath := filepath.Join(emptyDir, ".nav.yml")
	_, err = os.Stat(navPath)
	assert.True(t, os.IsNotExist(err), ".nav.yml should not exist for empty directory")
}

// TestGenerateNavForDir_WithFiles verifies nav generation with files.
func TestGenerateNavForDir_WithFiles(t *testing.T) {
	// Arrange
	stagingDir := t.TempDir()
	dir := filepath.Join(stagingDir, "docs")
	require.NoError(t, os.Mkdir(dir, 0o755))

	// Create test files
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# Guide"), 0o644))

	p := &Preprocessor{
		stagingDir: stagingDir,
		book:       &config.Book{},
		logWriter:  &bytes.Buffer{},
	}

	// Act
	err := p.generateNavForDir(dir)

	// Assert
	require.NoError(t, err)
	navPath := filepath.Join(dir, ".nav.yml")
	_, err = os.Stat(navPath)
	assert.NoError(t, err, ".nav.yml should be created")

	// Read and verify content
	content, err := os.ReadFile(navPath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "index.md")
	assert.Contains(t, contentStr, "guide.md")
}

// TestScanMarkdownFiles verifies markdown file scanning.
func TestScanMarkdownFiles(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("Text"), 0o644))

	p := &Preprocessor{logWriter: &bytes.Buffer{}}

	// Act
	files := p.scanMarkdownFiles(dir)

	// Assert
	assert.Len(t, files, 2)
	assert.True(t, files["test.md"])
	assert.True(t, files["other.md"])
	assert.False(t, files["readme.txt"], "Non-markdown files should not be included")
}

// TestScanSubdirectories verifies subdirectory scanning.
func TestScanSubdirectories(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	subdir1 := filepath.Join(dir, "docs")
	subdir2 := filepath.Join(dir, "guides")
	emptyDir := filepath.Join(dir, "empty")

	require.NoError(t, os.Mkdir(subdir1, 0o755))
	require.NoError(t, os.Mkdir(subdir2, 0o755))
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	// Add markdown file to make subdirs non-empty
	require.NoError(t, os.WriteFile(filepath.Join(subdir1, "test.md"), []byte("# Test"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subdir2, "guide.md"), []byte("# Guide"), 0o644))

	p := &Preprocessor{logWriter: &bytes.Buffer{}}

	// Act
	dirs := p.scanSubdirectories(dir)

	// Assert
	assert.True(t, dirs["docs"])
	assert.True(t, dirs["guides"])
	assert.False(t, dirs["empty"], "Empty directories should not be included")
}
