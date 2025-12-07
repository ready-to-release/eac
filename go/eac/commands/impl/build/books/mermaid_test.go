package books

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/buildutil"
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
	if blocks[0].blockIndex != 0 {
		t.Errorf("Block 0: expected index 0, got %d", blocks[0].blockIndex)
	}
	if !contains(blocks[0].content, "graph TD") {
		t.Errorf("Block 0: expected content to contain 'graph TD'")
	}
	if blocks[0].filename != "test_mermaid_0_"+blocks[0].hash+".svg" {
		t.Errorf("Block 0: unexpected filename %s", blocks[0].filename)
	}

	// Verify second block (with size directive)
	if blocks[1].blockIndex != 1 {
		t.Errorf("Block 1: expected index 1, got %d", blocks[1].blockIndex)
	}
	if !contains(blocks[1].content, "flowchart LR") {
		t.Errorf("Block 1: expected content to contain 'flowchart LR'")
	}
	// Content should include size directive
	if !contains(blocks[1].content, "%%{size:medium}%%") {
		t.Errorf("Block 1: expected content to preserve size directive")
	}

	// Verify third block
	if blocks[2].blockIndex != 2 {
		t.Errorf("Block 2: expected index 2, got %d", blocks[2].blockIndex)
	}
	if !contains(blocks[2].content, "sequenceDiagram") {
		t.Errorf("Block 2: expected content to contain 'sequenceDiagram'")
	}

	t.Logf("✓ Found %d diagrams", len(blocks))
	for i, block := range blocks {
		t.Logf("  [%d] %s (hash: %s)", i, block.filename, block.hash)
	}
}

func TestHashContent(t *testing.T) {
	// Test that hash is deterministic
	content := "graph TD\n    A --> B"
	hash1 := hashContent(content)
	hash2 := hashContent(content)

	if hash1 != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
	}

	// Hash should be 8 characters
	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 chars, got %d: %s", len(hash1), hash1)
	}

	// Different content should give different hash
	content2 := "graph TD\n    A --> C"
	hash3 := hashContent(content2)

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
			result := stripSizeDirective(tt.input)
			if result != tt.expected {
				t.Errorf("stripSizeDirective() = %q, want %q", result, tt.expected)
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
	if block.sourceFile != testFile {
		t.Errorf("Expected sourceFile %s, got %s", testFile, block.sourceFile)
	}
	if block.relPath != "real-test.md" {
		t.Errorf("Expected relPath 'real-test.md', got %s", block.relPath)
	}
	if !contains(block.content, "A[Start]") {
		t.Errorf("Expected content to contain 'A[Start]'")
	}

	t.Logf("✓ Extracted from real file: %s", block.filename)
	t.Logf("  Hash: %s", block.hash)
	t.Logf("  Content preview: %s", block.content[:30]+"...")
}

func TestCheckMermaidCache(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "assets", "rendered", "mermaid")

	// Create a preprocessor
	p := &Preprocessor{
		stagingDir:    tmpDir,
		workspaceRoot: tmpDir,
		logWriter:     os.Stdout,
		assetCache:    NewAssetCache(tmpDir),
	}

	// Create some test blocks
	blocks := []mermaidBlock{
		{
			content:  "graph TD\n    A --> B",
			hash:     "aaaaaaaa",
			filename: "test_mermaid_0_aaaaaaaa.svg",
		},
		{
			content:  "graph TD\n    C --> D",
			hash:     "bbbbbbbb",
			filename: "test_mermaid_1_bbbbbbbb.svg",
		},
		{
			content:  "graph TD\n    E --> F",
			hash:     "cccccccc",
			filename: "test_mermaid_2_cccccccc.svg",
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
		if status.cached {
			t.Errorf("Block %d: expected cache miss, got hit", i)
		}
		if status.cachePath == "" {
			t.Errorf("Block %d: cachePath should be set", i)
		}
	}

	t.Logf("✓ First check: All 3 diagrams are cache misses")

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create dummy SVG files for first two blocks (simulate cache hits)
	for i := 0; i < 2; i++ {
		svgPath := filepath.Join(cacheDir, blocks[i].filename)
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
	if !statuses[0].cached {
		t.Errorf("Block 0: expected cache hit")
	}
	if !statuses[1].cached {
		t.Errorf("Block 1: expected cache hit")
	}
	if statuses[2].cached {
		t.Errorf("Block 2: expected cache miss")
	}

	t.Logf("✓ Second check: 2 hits, 1 miss (66.7%% hit rate)")

	// Verify cache paths are correct
	for i, status := range statuses {
		expectedPath := filepath.Join(cacheDir, blocks[i].filename)
		if status.cachePath != expectedPath {
			t.Errorf("Block %d: cachePath = %s, want %s",
				i, status.cachePath, expectedPath)
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

	// Cache directory shouldn't exist yet
	cacheDir := filepath.Join(tmpDir, "assets", "rendered", "mermaid")
	if _, err := os.Stat(cacheDir); err == nil {
		t.Fatal("Cache directory should not exist yet")
	}

	// Call checkMermaidCache with empty blocks
	_, err := p.checkMermaidCache([]mermaidBlock{})
	if err != nil {
		t.Fatalf("checkMermaidCache failed: %v", err)
	}

	// Cache directory should now exist
	if _, err := os.Stat(cacheDir); err != nil {
		t.Fatalf("Cache directory was not created: %v", err)
	}

	t.Logf("✓ Cache directory created: %s", cacheDir)
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
			result := buildutil.FormatDockerVolumePath(tt.input)
			if result != tt.expected {
				t.Errorf("buildutil.FormatDockerVolumePath(%q) = %q, want %q",
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
	blocksByFile := map[string][]mermaidBlock{
		testFile: blocks,
	}

	// Replace blocks
	if err := p.replaceMermaidBlocksWithImages(blocksByFile); err != nil {
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

	if !contains(modifiedStr, blocks[0].filename) {
		t.Errorf("Modified file doesn't contain first diagram filename")
	}

	if !contains(modifiedStr, blocks[1].filename) {
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
	blocksByFile := map[string][]mermaidBlock{
		testFile: blocks,
	}

	// Replace blocks
	if err := p.replaceMermaidBlocksWithImages(blocksByFile); err != nil {
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

	// Should have relative path with ../ to go up from docs/subfolder/ to assets/
	if !contains(modifiedStr, "../../assets/rendered/mermaid/") {
		t.Errorf("Modified file doesn't have correct relative path, got: %s", modifiedStr)
	}

	if !contains(modifiedStr, blocks[0].filename) {
		t.Errorf("Modified file doesn't contain diagram filename")
	}

	t.Logf("✓ Replaced mermaid block with correct relative path")
	t.Logf("  Path: ../../assets/rendered/mermaid/%s", blocks[0].filename)
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
