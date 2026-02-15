package navigation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nopLog(string, ...any) {}

// TestEnsureRootIndex_ExistingFile verifies that existing index.md is preserved.
func TestEnsureRootIndex_ExistingFile(t *testing.T) {
	stagingDir := t.TempDir()
	indexPath := filepath.Join(stagingDir, "index.md")
	existingContent := "# Existing Index\nThis should be preserved."
	require.NoError(t, os.WriteFile(indexPath, []byte(existingContent), 0o644))

	err := EnsureRootIndex(&config.Book{}, stagingDir, nopLog)

	require.NoError(t, err)
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.Equal(t, existingContent, string(content))
}

// TestEnsureRootIndex_GenerateNew verifies that index.md is generated when missing.
func TestEnsureRootIndex_GenerateNew(t *testing.T) {
	stagingDir := t.TempDir()

	book := &config.Book{
		Name:        "Test Book",
		Title:       "Test Book Title",
		Description: "Test book description",
	}

	err := EnsureRootIndex(book, stagingDir, nopLog)

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
	stagingDir := t.TempDir()
	toc := GenerateTOC(&config.Book{}, stagingDir)
	assert.Empty(t, toc)
}

// TestGenerateTOC_WithFiles verifies TOC generation with markdown files.
func TestGenerateTOC_WithFiles(t *testing.T) {
	stagingDir := t.TempDir()

	file1 := filepath.Join(stagingDir, "intro.md")
	file2 := filepath.Join(stagingDir, "guide.md")

	require.NoError(t, os.WriteFile(file1, []byte("# Introduction\nIntro content"), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte("# User Guide\nGuide content"), 0o644))

	toc := GenerateTOC(&config.Book{}, stagingDir)

	assert.NotEmpty(t, toc)
	assert.Contains(t, toc, "Introduction")
	assert.Contains(t, toc, "User Guide")
}

// TestGetTitleFromFile_Frontmatter verifies title extraction from frontmatter.
func TestGetTitleFromFile_Frontmatter(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	content := "---\ntitle: \"Frontmatter Title\"\n---\n# Heading Title\nContent here."
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	title := GetTitleFromFile(filePath)
	assert.Equal(t, "Frontmatter Title", title)
}

// TestGetTitleFromFile_Heading verifies title extraction from H1 heading.
func TestGetTitleFromFile_Heading(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	content := "# Heading Title\nContent here."
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	title := GetTitleFromFile(filePath)
	assert.Equal(t, "Heading Title", title)
}

// TestGetTitleFromFile_Filename verifies title extraction from filename fallback.
func TestGetTitleFromFile_Filename(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "user-guide.md")
	content := "No title here."
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	title := GetTitleFromFile(filePath)
	assert.Equal(t, "User Guide", title)
}

// TestFilenameToTitle verifies filename to title conversion.
func TestFilenameToTitle(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"simple filename", "readme.md", "Readme"},
		{"dashed filename", "user-guide.md", "User Guide"},
		{"underscored filename", "api_reference.md", "Api Reference"},
		{"mixed separators", "quick-start_guide.md", "Quick Start Guide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilenameToTitle(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetOrderForFile verifies command source ordering.
func TestGetOrderForFile(t *testing.T) {
	book := &config.Book{
		Sources: []config.Source{
			{Type: "command", Target: "intro.md", Order: 100},
			{Type: "command", Target: "guide.md", Order: 200},
		},
	}

	tests := []struct {
		name     string
		relPath  string
		expected int
	}{
		{"file with order", "intro.md", 100},
		{"file with different order", "guide.md", 200},
		{"file without order", "other.md", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOrderForFile(book, tt.relPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCollectTOCFromFilesystem verifies filesystem-based TOC collection.
func TestCollectTOCFromFilesystem(t *testing.T) {
	stagingDir := t.TempDir()

	file1 := filepath.Join(stagingDir, "a-first.md")
	file2 := filepath.Join(stagingDir, "z-last.md")
	subdir := filepath.Join(stagingDir, "subdir")
	require.NoError(t, os.Mkdir(subdir, 0o755))
	file3 := filepath.Join(subdir, "index.md")

	require.NoError(t, os.WriteFile(file1, []byte("# First\nContent"), 0o644))
	require.NoError(t, os.WriteFile(file2, []byte("# Last\nContent"), 0o644))
	require.NoError(t, os.WriteFile(file3, []byte("# Subdir Index\nContent"), 0o644))

	entries := collectTOCFromFilesystem(&config.Book{}, stagingDir, stagingDir, 0)

	assert.NotEmpty(t, entries)

	var fileTitles []string
	for _, entry := range entries {
		if entry.depth == 0 && filepath.Ext(entry.path) == ".md" {
			fileTitles = append(fileTitles, entry.title)
		}
	}

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
		{"lowercase", "hello world", "Hello World"},
		{"uppercase", "HELLO WORLD", "Hello World"},
		{"mixed case", "HeLLo WoRLd", "Hello World"},
		{"single word", "test", "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToTitleCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateNavEntry_ValidFile verifies validation of valid markdown file.
func TestValidateNavEntry_ValidFile(t *testing.T) {
	actualFiles := map[string]bool{"test.md": true}
	actualDirs := map[string]bool{}

	result, referenced := ValidateNavEntry("test.md", "", actualFiles, actualDirs)

	assert.NotNil(t, result)
	assert.Equal(t, "test.md", result)
	assert.True(t, referenced["test.md"])
}

// TestValidateNavEntry_MissingFile verifies validation rejects missing files.
func TestValidateNavEntry_MissingFile(t *testing.T) {
	actualFiles := map[string]bool{}
	actualDirs := map[string]bool{}

	result, _ := ValidateNavEntry("missing.md", "", actualFiles, actualDirs)

	assert.Nil(t, result)
}

// TestValidateNavEntry_ValidDirectory verifies validation of valid directory.
func TestValidateNavEntry_ValidDirectory(t *testing.T) {
	actualFiles := map[string]bool{}
	actualDirs := map[string]bool{"subdir": true}

	result, referenced := ValidateNavEntry("subdir/", "", actualFiles, actualDirs)

	assert.NotNil(t, result)
	assert.Equal(t, "subdir/", result)
	assert.True(t, referenced["subdir/"])
}

// TestValidateNavEntry_TitledEntry verifies validation of titled entries.
func TestValidateNavEntry_TitledEntry(t *testing.T) {
	actualFiles := map[string]bool{"intro.md": true}
	actualDirs := map[string]bool{}

	entry := map[string]any{
		"Introduction": "intro.md",
	}

	result, referenced := ValidateNavEntry(entry, "", actualFiles, actualDirs)

	assert.NotNil(t, result)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "intro.md", resultMap["Introduction"])
	assert.True(t, referenced["intro.md"])
}

// TestValidateNavEntry_DirectoryWithoutSlash verifies directory ref without trailing slash.
func TestValidateNavEntry_DirectoryWithoutSlash(t *testing.T) {
	actualFiles := map[string]bool{}
	actualDirs := map[string]bool{"specs": true}

	result, referenced := ValidateNavEntry("specs", "", actualFiles, actualDirs)

	assert.NotNil(t, result)
	assert.Equal(t, "specs", result)
	assert.True(t, referenced["specs/"])
}

// TestGenerateNavForDir_EmptyDir verifies nav generation for empty directory.
func TestGenerateNavForDir_EmptyDir(t *testing.T) {
	stagingDir := t.TempDir()
	emptyDir := filepath.Join(stagingDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	err := GenerateNavForDir(&config.Book{}, stagingDir, emptyDir, nopLog)

	require.NoError(t, err)
	navPath := filepath.Join(emptyDir, ".nav.yml")
	_, err = os.Stat(navPath)
	assert.True(t, os.IsNotExist(err))
}

// TestGenerateNavForDir_WithFiles verifies nav generation with files.
func TestGenerateNavForDir_WithFiles(t *testing.T) {
	stagingDir := t.TempDir()
	dir := filepath.Join(stagingDir, "docs")
	require.NoError(t, os.Mkdir(dir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Index"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# Guide"), 0o644))

	err := GenerateNavForDir(&config.Book{}, stagingDir, dir, nopLog)

	require.NoError(t, err)
	navPath := filepath.Join(dir, ".nav.yml")
	_, err = os.Stat(navPath)
	assert.NoError(t, err)

	content, err := os.ReadFile(navPath)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, "index.md")
	assert.Contains(t, contentStr, "guide.md")
}

// TestScanMarkdownFiles verifies markdown file scanning.
func TestScanMarkdownFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("Text"), 0o644))

	files := ScanMarkdownFiles(dir)

	assert.Len(t, files, 2)
	assert.True(t, files["test.md"])
	assert.True(t, files["other.md"])
	assert.False(t, files["readme.txt"])
}

// TestScanSubdirectories verifies subdirectory scanning.
func TestScanSubdirectories(t *testing.T) {
	dir := t.TempDir()
	subdir1 := filepath.Join(dir, "docs")
	subdir2 := filepath.Join(dir, "guides")
	emptyDir := filepath.Join(dir, "empty")

	require.NoError(t, os.Mkdir(subdir1, 0o755))
	require.NoError(t, os.Mkdir(subdir2, 0o755))
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(subdir1, "test.md"), []byte("# Test"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(subdir2, "guide.md"), []byte("# Guide"), 0o644))

	dirs := ScanSubdirectories(dir)

	assert.True(t, dirs["docs"])
	assert.True(t, dirs["guides"])
	assert.False(t, dirs["empty"])
}

// TestHasMarkdownContent verifies markdown content detection.
func TestHasMarkdownContent(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	require.NoError(t, os.Mkdir(subdir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "doc.md"), []byte("# Doc"), 0o644))

	assert.True(t, HasMarkdownContent(dir))

	emptyDir := t.TempDir()
	assert.False(t, HasMarkdownContent(emptyDir))
}

// TestStripMacros verifies macro stripping from files.
func TestStripMacros(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	content := "# Title\n\n{{ page_breadcrumb() }}\n\nContent\n\n{{ diataxis_footer() }}\n"
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	idx, err := newTestFileIndex(dir)
	require.NoError(t, err)

	err = StripMacros(idx, nopLog)
	require.NoError(t, err)

	result, err := os.ReadFile(filePath)
	require.NoError(t, err)

	resultStr := string(result)
	assert.NotContains(t, resultStr, "page_breadcrumb")
	assert.NotContains(t, resultStr, "diataxis_footer")
	assert.Contains(t, resultStr, "# Title")
	assert.Contains(t, resultStr, "Content")
}

// TestInjectMacros verifies macro injection into Diataxis section files.
func TestInjectMacros(t *testing.T) {
	dir := t.TempDir()
	tutorialsDir := filepath.Join(dir, "tutorials")
	require.NoError(t, os.Mkdir(tutorialsDir, 0o755))

	filePath := filepath.Join(tutorialsDir, "intro.md")
	content := "# Getting Started\n\nSome content here.\n"
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	idx, err := newTestFileIndex(dir)
	require.NoError(t, err)

	err = InjectMacros(idx, dir, nopLog)
	require.NoError(t, err)

	result, err := os.ReadFile(filePath)
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "{{ page_breadcrumb() }}")
	assert.Contains(t, resultStr, "{{ diataxis_footer() }}")
}

// TestInjectMacros_SkipsNonDiataxis verifies files outside Diataxis sections are skipped.
func TestInjectMacros_SkipsNonDiataxis(t *testing.T) {
	dir := t.TempDir()
	otherDir := filepath.Join(dir, "other")
	require.NoError(t, os.Mkdir(otherDir, 0o755))

	filePath := filepath.Join(otherDir, "page.md")
	content := "# Other Page\n\nContent here.\n"
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0o644))

	idx, err := newTestFileIndex(dir)
	require.NoError(t, err)

	err = InjectMacros(idx, dir, nopLog)
	require.NoError(t, err)

	result, err := os.ReadFile(filePath)
	require.NoError(t, err)

	resultStr := string(result)
	assert.NotContains(t, resultStr, "page_breadcrumb")
	assert.NotContains(t, resultStr, "diataxis_footer")
}

// helper to create a file index for testing.
func newTestFileIndex(dir string) (*staging.FileIndex, error) {
	idx, err := staging.NewFileIndex(dir)
	if err != nil {
		return nil, err
	}
	if err := idx.Refresh(); err != nil {
		return nil, err
	}
	return idx, nil
}
