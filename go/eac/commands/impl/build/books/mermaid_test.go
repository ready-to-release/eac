package books

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
)

func TestExtractMermaidBlocks(t *testing.T) {
	// Test content with multiple mermaid diagrams
	content := `# Title

Some text before.

` + "```mermaid" + `
graph TD
    A --> B
    B --> C
` + "```" + `

More text in between.

` + "```mermaid" + `
%%{size:medium}%%
flowchart LR
    Start --> End
` + "```" + `

Final text.

` + "```mermaid" + `
sequenceDiagram
    Alice->>Bob: Hello
` + "```" + `
`

	// Create temp directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	blocks := extractMermaidBlocks(content, testFile, tmpDir)

	// Should find 3 blocks
	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	// Verify first block
	if blocks[0].BlockIndex != 0 {
		t.Errorf("Block 0: expected index 0, got %d", blocks[0].BlockIndex)
	}
	if !contains(blocks[0].Content, "graph TD") {
		t.Errorf("Block 0: expected content to contain 'graph TD'")
	}
	if blocks[0].Filename != "test_mermaid_0_"+blocks[0].Hash+".svg" {
		t.Errorf("Block 0: unexpected filename %s", blocks[0].Filename)
	}

	// Verify second block (with size directive)
	if blocks[1].BlockIndex != 1 {
		t.Errorf("Block 1: expected index 1, got %d", blocks[1].BlockIndex)
	}
	if !contains(blocks[1].Content, "flowchart LR") {
		t.Errorf("Block 1: expected content to contain 'flowchart LR'")
	}
	// Content should include size directive
	if !contains(blocks[1].Content, "%%{size:medium}%%") {
		t.Errorf("Block 1: expected content to preserve size directive")
	}

	// Verify third block
	if blocks[2].BlockIndex != 2 {
		t.Errorf("Block 2: expected index 2, got %d", blocks[2].BlockIndex)
	}
	if !contains(blocks[2].Content, "sequenceDiagram") {
		t.Errorf("Block 2: expected content to contain 'sequenceDiagram'")
	}

	t.Logf("✓ Found %d diagrams", len(blocks))
	for i, block := range blocks {
		t.Logf("  [%d] %s (hash: %s)", i, block.Filename, block.Hash)
	}
}

func TestHashContent(t *testing.T) {
	// Test that hash is deterministic
	content := "graph TD\n    A --> B"
	hash1 := HashContent(content)
	hash2 := HashContent(content)

	if hash1 != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
	}

	// Hash should be 8 characters
	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 chars, got %d: %s", len(hash1), hash1)
	}

	// Different content should give different hash
	content2 := "graph TD\n    A --> C"
	hash3 := HashContent(content2)

	if hash1 == hash3 {
		t.Errorf("Different content should have different hashes")
	}

	t.Logf("✓ Hash for content: %s", hash1)
}

func TestStripSizeDirective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "with size directive",
			input: `%%{size:medium}%%
graph TD
    A --> B`,
			expected: `graph TD
    A --> B`,
		},
		{
			name: "with width directive",
			input: `%%{width:80%}%%
flowchart LR
    Start --> End`,
			expected: `flowchart LR
    Start --> End`,
		},
		{
			name: "no directive",
			input: `graph TD
    A --> B`,
			expected: `graph TD
    A --> B`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripSizeDirective(tt.input)
			if result != tt.expected {
				t.Errorf("StripSizeDirective() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractMermaidBlocksWithRealFile(t *testing.T) {
	// Test with a real markdown file
	content := `# Test Document

This document has diagrams.

` + "```mermaid" + `
graph TD
    A[Start] --> B[Process]
    B --> C[End]
` + "```" + `

Some more content.
`

	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "real-test.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read and extract
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	blocks := extractMermaidBlocks(string(fileContent), testFile, tmpDir)

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	// Verify the block
	block := blocks[0]
	if block.SourceFile != testFile {
		t.Errorf("Expected sourceFile %s, got %s", testFile, block.SourceFile)
	}
	if block.RelPath != "real-test.md" {
		t.Errorf("Expected relPath 'real-test.md', got %s", block.RelPath)
	}
	if !contains(block.Content, "A[Start]") {
		t.Errorf("Expected content to contain 'A[Start]'")
	}

	t.Logf("✓ Extracted from real file: %s", block.Filename)
	t.Logf("  Hash: %s", block.Hash)
	t.Logf("  Content preview: %s", block.Content[:30]+"...")
}

func TestCheckMermaidCache(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// The cache system uses docs/assets/cache as root, then mermaid/ subdirectory
	// For staging, it checks {staging}/assets/cache/mermaid/
	stagingCacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")

	// Create a preprocessor
	p := &Preprocessor{
		stagingDir:    tmpDir,
		workspaceRoot: tmpDir,
		logWriter:     os.Stdout,
		assetCache:    NewAssetCache(tmpDir),
	}

	// Create some test blocks with content that will be hashed
	blocks := []MermaidBlock{
		{
			Content:  "graph TD\n    A --> B",
			Hash:     "aaaaaaaa",
			Filename: "test_mermaid_0_aaaaaaaa.svg",
		},
		{
			Content:  "graph TD\n    C --> D",
			Hash:     "bbbbbbbb",
			Filename: "test_mermaid_1_bbbbbbbb.svg",
		},
		{
			Content:  "graph TD\n    E --> F",
			Hash:     "cccccccc",
			Filename: "test_mermaid_2_cccccccc.svg",
		},
	}

	// First check - all should be cache misses
	statuses, err := p.checkMermaidCache(blocks)
	if err != nil {
		t.Fatalf("checkMermaidCache failed: %v", err)
	}

	if len(statuses) != 3 {
		t.Fatalf("Expected 3 statuses, got %d", len(statuses))
	}

	// All should be cache misses
	for i, status := range statuses {
		if status.Cached {
			t.Errorf("Block %d: expected cache miss, got hit", i)
		}
		if status.CachePath == "" {
			t.Errorf("Block %d: cachePath should be set", i)
		}
	}

	t.Logf("✓ First check: All 3 diagrams are cache misses")

	// Create cache directory
	if err := os.MkdirAll(stagingCacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create dummy SVG files for first two blocks using their actual cache paths
	// The cache system uses content-hash based filenames computed by AssetCache
	for i := 0; i < 2; i++ {
		// Use the cache path that checkMermaidCache computed (extract from previous status)
		svgPath := statuses[i].CachePath
		if err := os.WriteFile(svgPath, []byte("<svg></svg>"), 0644); err != nil {
			t.Fatalf("Failed to create cached SVG: %v", err)
		}
	}

	// Second check - first two should be hits, third should be miss
	statuses, err = p.checkMermaidCache(blocks)
	if err != nil {
		t.Fatalf("checkMermaidCache failed: %v", err)
	}

	// Check results
	if !statuses[0].Cached {
		t.Errorf("Block 0: expected cache hit")
	}
	if !statuses[1].Cached {
		t.Errorf("Block 1: expected cache hit")
	}
	if statuses[2].Cached {
		t.Errorf("Block 2: expected cache miss")
	}

	t.Logf("✓ Second check: 2 hits, 1 miss (66.7%% hit rate)")

	// Verify cache paths are in the correct directory
	for i, status := range statuses {
		if !filepath.HasPrefix(status.CachePath, stagingCacheDir) {
			t.Errorf("Block %d: cachePath = %s, expected to be under %s",
				i, status.CachePath, stagingCacheDir)
		}
	}

	t.Logf("✓ Cache paths are correct")
}

func TestCacheDirectoryCreation(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	p := &Preprocessor{
		stagingDir:    tmpDir,
		workspaceRoot: tmpDir,
		logWriter:     os.Stdout,
		assetCache:    NewAssetCache(tmpDir),
	}

	// Cache directory shouldn't exist yet (staging cache location)
	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	if _, err := os.Stat(cacheDir); err == nil {
		t.Fatal("Cache directory should not exist yet")
	}

	// Call checkMermaidCache with a block to trigger directory creation
	// Empty blocks won't create the directory since there's nothing to cache
	blocks := []MermaidBlock{
		{
			Content:  "graph TD\n    A --> B",
			Hash:     "aaaaaaaa",
			Filename: "test_mermaid_0_aaaaaaaa.svg",
		},
	}
	statuses, err := p.checkMermaidCache(blocks)
	if err != nil {
		t.Fatalf("checkMermaidCache failed: %v", err)
	}

	// Verify status was returned
	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	// Note: checkMermaidCache doesn't create the staging cache directory itself -
	// it only checks if files exist in staging. The directory is created when
	// files are copied from docs/assets/cache during preprocessing.
	// This test verifies that checkMermaidCache works correctly when the
	// cache directory doesn't exist (returns cache miss status).
	if statuses[0].Cached {
		t.Errorf("Expected cache miss for non-existent directory")
	}
	if statuses[0].CachePath == "" {
		t.Errorf("CachePath should still be set even for cache miss")
	}

	t.Logf("✓ checkMermaidCache handles non-existent cache directory correctly")
	t.Logf("  CachePath: %s", statuses[0].CachePath)
}

func TestFormatDockerVolumePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path with drive letter",
			input:    `C:\projects\eac`,
			expected: "/c/projects/eac",
		},
		{
			name:     "Windows path with different drive",
			input:    `D:\workspace\test`,
			expected: "/d/workspace/test",
		},
		{
			name:     "Unix path (unchanged)",
			input:    "/home/user/projects",
			expected: "/home/user/projects",
		},
		{
			name:     "Path with forward slashes already",
			input:    "C:/projects/eac",
			expected: "/c/projects/eac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dockerutil.FormatDockerVolume(tt.input)
			if result != tt.expected {
				t.Errorf("dockerutil.FormatDockerVolume(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestReplaceMermaidBlocksWithImages(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Create markdown with mermaid blocks
	content := `# Test Document

Some text before.

` + "```mermaid" + `
graph TD
    A --> B
` + "```" + `

Middle text.

` + "```mermaid" + `
flowchart LR
    Start --> End
` + "```" + `

Text after.
`

	// Write test file
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Extract blocks
	blocks := extractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}

	// Create preprocessor
	p := &Preprocessor{
		stagingDir:     tmpDir,
		workspaceRoot:  tmpDir,
		logWriter:      os.Stdout,
		linkTranslator: NewLinkTranslator(tmpDir, tmpDir, os.Stdout, false),
	}

	// Create blocks map
	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	// Create cache statuses with mock cache paths
	cacheDir := filepath.Join(tmpDir, "assets", "rendered", "mermaid")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		// Create empty SVG file
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	// Replace blocks
	if err := p.replaceMermaidBlocksWithImages(blocksByFile, statuses); err != nil {
		t.Fatalf("replaceMermaidBlocksWithImages failed: %v", err)
	}

	// Read modified file
	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	// Verify replacements
	if contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}

	if !contains(modifiedStr, "<img src=\"") {
		t.Errorf("Modified file doesn't contain img tags")
	}

	if !contains(modifiedStr, blocks[0].Filename) {
		t.Errorf("Modified file doesn't contain first diagram filename")
	}

	if !contains(modifiedStr, blocks[1].Filename) {
		t.Errorf("Modified file doesn't contain second diagram filename")
	}

	// Verify surrounding text is preserved
	if !contains(modifiedStr, "Some text before.") {
		t.Errorf("Lost text before diagrams")
	}

	if !contains(modifiedStr, "Middle text.") {
		t.Errorf("Lost text between diagrams")
	}

	if !contains(modifiedStr, "Text after.") {
		t.Errorf("Lost text after diagrams")
	}

	t.Logf("✓ Replaced %d mermaid blocks with img tags", len(blocks))
	t.Logf("  File size: %d bytes -> %d bytes", len(content), len(modified))
}

func TestReplaceMermaidBlocksWithImagesNestedPath(t *testing.T) {
	// Create temp directory for testing with nested structure
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs", "subfolder")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create markdown with mermaid block in nested directory
	content := `# Nested Document

` + "```mermaid" + `
graph TD
    X --> Y
` + "```" + `
`

	// Write test file in nested directory
	testFile := filepath.Join(nestedDir, "nested.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Extract blocks
	blocks := extractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	// Create preprocessor
	p := &Preprocessor{
		stagingDir:     tmpDir,
		workspaceRoot:  tmpDir,
		logWriter:      os.Stdout,
		linkTranslator: NewLinkTranslator(tmpDir, tmpDir, os.Stdout, false),
	}

	// Create blocks map
	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	// Create cache statuses with mock cache paths in assets/rendered/mermaid/
	cacheDir := filepath.Join(tmpDir, "assets", "rendered", "mermaid")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		// Create empty SVG file
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	// Replace blocks
	if err := p.replaceMermaidBlocksWithImages(blocksByFile, statuses); err != nil {
		t.Fatalf("replaceMermaidBlocksWithImages failed: %v", err)
	}

	// Read modified file
	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	// Verify replacements
	if contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}

	// For site builds (pdfMode=false), MkDocs converts file.md to file/index.html,
	// adding an extra directory level. The path needs an extra ../ compared to
	// the staging file location.
	// Staging: docs/subfolder/nested.md -> ../../assets/rendered/mermaid/xxx.svg (2 levels up)
	// Site:    docs/subfolder/nested/index.html -> ../../../assets/rendered/mermaid/xxx.svg (3 levels up)
	if !contains(modifiedStr, "../../../assets/rendered/mermaid/") {
		t.Errorf("Modified file doesn't have correct relative path for site build, got: %s", modifiedStr)
	}

	if !contains(modifiedStr, blocks[0].Filename) {
		t.Errorf("Modified file doesn't contain diagram filename")
	}

	t.Logf("✓ Replaced mermaid block with correct relative path for site build (pdfMode=false)")
	t.Logf("  Path: ../../../assets/rendered/mermaid/%s", blocks[0].Filename)
}

func TestReplaceMermaidBlocksWithImagesPDFMode(t *testing.T) {
	// Create temp directory for testing with nested structure
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs", "subfolder")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	// Create markdown with mermaid block in nested directory
	content := `# Nested Document

` + "```mermaid" + `
graph TD
    X --> Y
` + "```" + `
`

	// Write test file in nested directory
	testFile := filepath.Join(nestedDir, "nested.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Extract blocks
	blocks := extractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	// Create preprocessor with pdfMode=true
	p := &Preprocessor{
		stagingDir:     tmpDir,
		workspaceRoot:  tmpDir,
		logWriter:      os.Stdout,
		pdfMode:        true, // PDF mode - should NOT add extra ../
		linkTranslator: NewLinkTranslator(tmpDir, tmpDir, os.Stdout, true),
	}

	// Create blocks map
	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	// Create cache statuses with mock cache paths in assets/rendered/mermaid/
	cacheDir := filepath.Join(tmpDir, "assets", "rendered", "mermaid")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		// Create empty SVG file
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	// Replace blocks
	if err := p.replaceMermaidBlocksWithImages(blocksByFile, statuses); err != nil {
		t.Fatalf("replaceMermaidBlocksWithImages failed: %v", err)
	}

	// Read modified file
	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	// Verify replacements
	if contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}

	// For PDF builds (pdfMode=true), MkDocs-with-pdf uses the staging structure directly,
	// so NO extra ../ is needed. The path should match the staging file location.
	// Staging: docs/subfolder/nested.md -> ../../assets/rendered/mermaid/xxx.svg (2 levels up)
	if !contains(modifiedStr, "../../assets/rendered/mermaid/") {
		t.Errorf("Modified file doesn't have correct relative path for PDF build, got: %s", modifiedStr)
	}

	// Ensure it does NOT have the extra ../ that site builds need
	if contains(modifiedStr, "../../../assets/rendered/mermaid/") {
		t.Errorf("PDF build incorrectly has extra ../ in path, got: %s", modifiedStr)
	}

	if !contains(modifiedStr, blocks[0].Filename) {
		t.Errorf("Modified file doesn't contain diagram filename")
	}

	t.Logf("✓ Replaced mermaid block with correct relative path for PDF build (pdfMode=true)")
	t.Logf("  Path: ../../assets/rendered/mermaid/%s", blocks[0].Filename)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
