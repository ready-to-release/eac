package diagrams

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/paths"
)

func TestExtractMermaidBlocks(t *testing.T) {
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

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	blocks := ExtractMermaidBlocks(content, testFile, tmpDir)

	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	if blocks[0].BlockIndex != 0 {
		t.Errorf("Block 0: expected index 0, got %d", blocks[0].BlockIndex)
	}
	if !strings.Contains(blocks[0].Content, "graph TD") {
		t.Errorf("Block 0: expected content to contain 'graph TD'")
	}
	if blocks[0].Filename != "test_mermaid_0_"+blocks[0].Hash+".svg" {
		t.Errorf("Block 0: unexpected filename %s", blocks[0].Filename)
	}

	if blocks[1].BlockIndex != 1 {
		t.Errorf("Block 1: expected index 1, got %d", blocks[1].BlockIndex)
	}
	if !strings.Contains(blocks[1].Content, "flowchart LR") {
		t.Errorf("Block 1: expected content to contain 'flowchart LR'")
	}
	if !strings.Contains(blocks[1].Content, "%%{size:medium}%%") {
		t.Errorf("Block 1: expected content to preserve size directive")
	}

	if blocks[2].BlockIndex != 2 {
		t.Errorf("Block 2: expected index 2, got %d", blocks[2].BlockIndex)
	}
	if !strings.Contains(blocks[2].Content, "sequenceDiagram") {
		t.Errorf("Block 2: expected content to contain 'sequenceDiagram'")
	}
}

func TestExtractMermaidBlocksWithRealFile(t *testing.T) {
	content := `# Test Document

This document has diagrams.

` + "```mermaid" + `
graph TD
    A[Start] --> B[Process]
    B --> C[End]
` + "```" + `

Some more content.
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "real-test.md")
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	blocks := ExtractMermaidBlocks(string(fileContent), testFile, tmpDir)

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	block := blocks[0]
	if block.SourceFile != testFile {
		t.Errorf("Expected sourceFile %s, got %s", testFile, block.SourceFile)
	}
	if block.RelPath != "real-test.md" {
		t.Errorf("Expected relPath 'real-test.md', got %s", block.RelPath)
	}
	if !strings.Contains(block.Content, "A[Start]") {
		t.Errorf("Expected content to contain 'A[Start]'")
	}
}

func TestExtractMermaidBlocksEmpty(t *testing.T) {
	content := "# No diagrams here\n\nJust text.\n"
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.md")

	blocks := ExtractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 0 {
		t.Errorf("Expected 0 blocks, got %d", len(blocks))
	}
}

func TestHashContent(t *testing.T) {
	content := "graph TD\n    A --> B"
	hash1 := HashContent(content)
	hash2 := HashContent(content)

	if hash1 != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
	}

	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 chars, got %d: %s", len(hash1), hash1)
	}

	content2 := "graph TD\n    A --> C"
	hash3 := HashContent(content2)

	if hash1 == hash3 {
		t.Errorf("Different content should have different hashes")
	}
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

func TestWrapMermaidBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		excludes string
	}{
		{
			name: "with size preset",
			input: "```mermaid\n%%{size:medium}%%\ngraph TD\n    A --> B\n```",
			contains: `data-size="medium"`,
		},
		{
			name: "with width value",
			input: "```mermaid\n%%{width:40%}%%\ngraph TD\n    A --> B\n```",
			contains: `data-size="40%"`,
		},
		{
			name:     "no directive unchanged",
			input:    "```mermaid\ngraph TD\n    A --> B\n```",
			excludes: "mermaid-wrapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapMermaidBlocks(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("Expected result to NOT contain %q, got: %s", tt.excludes, result)
			}
		})
	}
}

func TestCheckMermaidCache(t *testing.T) {
	tmpDir := t.TempDir()
	builderOutputDir := paths.MermaidBuildOutputPath(tmpDir)

	block0Content := "graph TD\n    A --> B"
	block1Content := "graph TD\n    C --> D"
	block2Content := "graph TD\n    E --> F"

	blocks := []MermaidBlock{
		{
			Content:    block0Content,
			Hash:       HashContent(block0Content),
			BlockIndex: 0,
			Filename:   "test_mermaid_0_" + HashContent(block0Content) + ".svg",
		},
		{
			Content:    block1Content,
			Hash:       HashContent(block1Content),
			BlockIndex: 1,
			Filename:   "test_mermaid_1_" + HashContent(block1Content) + ".svg",
		},
		{
			Content:    block2Content,
			Hash:       HashContent(block2Content),
			BlockIndex: 2,
			Filename:   "test_mermaid_2_" + HashContent(block2Content) + ".svg",
		},
	}

	noop := func(string, ...any) {}

	// First check - no builder output, all should be cache misses
	statuses, err := CheckMermaidCache(tmpDir, blocks, noop)
	if err != nil {
		t.Fatalf("CheckMermaidCache failed: %v", err)
	}

	if len(statuses) != 3 {
		t.Fatalf("Expected 3 statuses, got %d", len(statuses))
	}

	for i, status := range statuses {
		if status.Cached {
			t.Errorf("Block %d: expected cache miss, got hit", i)
		}
	}

	// Create builder output with index for first two blocks
	if err := os.MkdirAll(builderOutputDir, 0o755); err != nil {
		t.Fatalf("Failed to create builder output dir: %v", err)
	}

	indexJSON := fmt.Sprintf(`{
		"entries": [
			{"source_file": "test.md", "block_index": 0, "content_hash": "%s", "svg_filename": "diagram_0.svg"},
			{"source_file": "test.md", "block_index": 1, "content_hash": "%s", "svg_filename": "diagram_1.svg"}
		]
	}`, HashContent(block0Content), HashContent(block1Content))

	indexPath := filepath.Join(builderOutputDir, "mermaid-index.json")
	if err := os.WriteFile(indexPath, []byte(indexJSON), 0o644); err != nil {
		t.Fatalf("Failed to write mermaid-index.json: %v", err)
	}

	for i := 0; i < 2; i++ {
		svgPath := filepath.Join(builderOutputDir, fmt.Sprintf("diagram_%d.svg", i))
		if err := os.WriteFile(svgPath, []byte("<svg></svg>"), 0o644); err != nil {
			t.Fatalf("Failed to create SVG file: %v", err)
		}
	}

	// Second check - first two should be hits, third should be miss
	statuses, err = CheckMermaidCache(tmpDir, blocks, noop)
	if err != nil {
		t.Fatalf("CheckMermaidCache failed: %v", err)
	}

	if !statuses[0].Cached {
		t.Errorf("Block 0: expected cache hit")
	}
	if !statuses[1].Cached {
		t.Errorf("Block 1: expected cache hit")
	}
	if statuses[2].Cached {
		t.Errorf("Block 2: expected cache miss")
	}

	for i := 0; i < 2; i++ {
		expectedPath := filepath.Join(builderOutputDir, fmt.Sprintf("diagram_%d.svg", i))
		if statuses[i].CachePath != expectedPath {
			t.Errorf("Block %d: cachePath = %s, expected %s", i, statuses[i].CachePath, expectedPath)
		}
	}
}

func TestCheckMermaidCacheNoBuilderOutput(t *testing.T) {
	tmpDir := t.TempDir()
	noop := func(string, ...any) {}

	blocks := []MermaidBlock{
		{
			Content:    "graph TD\n    A --> B",
			Hash:       HashContent("graph TD\n    A --> B"),
			BlockIndex: 0,
			Filename:   "test_mermaid_0.svg",
		},
	}
	statuses, err := CheckMermaidCache(tmpDir, blocks, noop)
	if err != nil {
		t.Fatalf("CheckMermaidCache failed: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	if statuses[0].Cached {
		t.Errorf("Expected cache miss when builder output doesn't exist")
	}
}

func TestCheckMermaidCacheCorruptIndex(t *testing.T) {
	tmpDir := t.TempDir()
	builderOutputDir := paths.MermaidBuildOutputPath(tmpDir)
	noop := func(string, ...any) {}

	if err := os.MkdirAll(builderOutputDir, 0o755); err != nil {
		t.Fatalf("Failed to create builder output dir: %v", err)
	}

	indexPath := filepath.Join(builderOutputDir, "mermaid-index.json")
	if err := os.WriteFile(indexPath, []byte("not json"), 0o644); err != nil {
		t.Fatalf("Failed to write corrupt index: %v", err)
	}

	blocks := []MermaidBlock{
		{Content: "graph TD\n    A --> B", Hash: "abcdefgh", BlockIndex: 0, Filename: "test.svg"},
	}

	_, err := CheckMermaidCache(tmpDir, blocks, noop)
	if err == nil {
		t.Errorf("Expected error for corrupt index")
	}
}

func TestReplaceMermaidBlocksWithImages(t *testing.T) {
	tmpDir := t.TempDir()

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

	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	blocks := ExtractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}

	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0o644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	noop := func(string, ...any) {}
	if err := ReplaceMermaidBlocksWithImages(blocksByFile, statuses, "", noop); err != nil {
		t.Fatalf("ReplaceMermaidBlocksWithImages failed: %v", err)
	}

	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	if strings.Contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}
	if !strings.Contains(modifiedStr, "<img src=\"") {
		t.Errorf("Modified file doesn't contain img tags")
	}
	if !strings.Contains(modifiedStr, blocks[0].Filename) {
		t.Errorf("Modified file doesn't contain first diagram filename")
	}
	if !strings.Contains(modifiedStr, blocks[1].Filename) {
		t.Errorf("Modified file doesn't contain second diagram filename")
	}
	if !strings.Contains(modifiedStr, "Some text before.") {
		t.Errorf("Lost text before diagrams")
	}
	if !strings.Contains(modifiedStr, "Middle text.") {
		t.Errorf("Lost text between diagrams")
	}
	if !strings.Contains(modifiedStr, "Text after.") {
		t.Errorf("Lost text after diagrams")
	}
}

func TestReplaceMermaidBlocksNestedWithExtraPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs", "subfolder")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	content := `# Nested Document

` + "```mermaid" + `
graph TD
    X --> Y
` + "```" + `
`

	testFile := filepath.Join(nestedDir, "nested.md")
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	blocks := ExtractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0o644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	// Site mode: ExtraPathPrefix = "../"
	noop := func(string, ...any) {}
	if err := ReplaceMermaidBlocksWithImages(blocksByFile, statuses, "../", noop); err != nil {
		t.Fatalf("ReplaceMermaidBlocksWithImages failed: %v", err)
	}

	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	if strings.Contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}

	// Site build adds extra "../" so nested/index.html -> ../../../assets/cache/mermaid/xxx.svg
	if !strings.Contains(modifiedStr, "../../../assets/cache/mermaid/") {
		t.Errorf("Modified file doesn't have correct relative path for site build, got: %s", modifiedStr)
	}

	if !strings.Contains(modifiedStr, blocks[0].Filename) {
		t.Errorf("Modified file doesn't contain diagram filename")
	}
}

func TestReplaceMermaidBlocksPDFMode(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "docs", "subfolder")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	content := `# Nested Document

` + "```mermaid" + `
graph TD
    X --> Y
` + "```" + `
`

	testFile := filepath.Join(nestedDir, "nested.md")
	if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	blocks := ExtractMermaidBlocks(content, testFile, tmpDir)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	blocksByFile := map[string][]MermaidBlock{
		testFile: blocks,
	}

	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	var statuses []CacheStatus
	for _, block := range blocks {
		cachePath := filepath.Join(cacheDir, block.Filename)
		if err := os.WriteFile(cachePath, []byte("<svg></svg>"), 0o644); err != nil {
			t.Fatalf("Failed to write mock SVG: %v", err)
		}
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: cachePath,
		})
	}

	// PDF mode: ExtraPathPrefix = ""
	noop := func(string, ...any) {}
	if err := ReplaceMermaidBlocksWithImages(blocksByFile, statuses, "", noop); err != nil {
		t.Fatalf("ReplaceMermaidBlocksWithImages failed: %v", err)
	}

	modified, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modifiedStr := string(modified)

	if strings.Contains(modifiedStr, "```mermaid") {
		t.Errorf("Modified file still contains mermaid blocks")
	}

	// PDF mode: no extra "../" so nested.md -> ../../assets/cache/mermaid/xxx.svg
	if !strings.Contains(modifiedStr, "../../assets/cache/mermaid/") {
		t.Errorf("Modified file doesn't have correct relative path for PDF build, got: %s", modifiedStr)
	}

	// Ensure it does NOT have the extra ../ that site builds need
	if strings.Contains(modifiedStr, "../../../assets/cache/mermaid/") {
		t.Errorf("PDF build incorrectly has extra ../ in path, got: %s", modifiedStr)
	}
}

func TestMermaidSizePresets(t *testing.T) {
	expected := map[string]string{
		"small":  "33%",
		"medium": "50%",
		"large":  "66%",
		"full":   "100%",
	}

	for name, expectedWidth := range expected {
		if got := MermaidSizePresets[name]; got != expectedWidth {
			t.Errorf("MermaidSizePresets[%s] = %s, want %s", name, got, expectedWidth)
		}
	}
}

// TestCheckMermaidCacheMissingSVGFile verifies cache miss when index references a non-existent SVG.
func TestCheckMermaidCacheMissingSVGFile(t *testing.T) {
	tmpDir := t.TempDir()
	builderOutputDir := paths.MermaidBuildOutputPath(tmpDir)
	noop := func(string, ...any) {}

	if err := os.MkdirAll(builderOutputDir, 0o755); err != nil {
		t.Fatalf("Failed to create builder output dir: %v", err)
	}

	blockContent := "graph TD\n    A --> B"
	hash := HashContent(blockContent)

	indexData := map[string]any{
		"entries": []map[string]any{
			{
				"source_file":  "test.md",
				"block_index":  0,
				"content_hash": hash,
				"svg_filename": "missing.svg",
			},
		},
	}
	indexJSON, _ := json.Marshal(indexData)

	indexPath := filepath.Join(builderOutputDir, "mermaid-index.json")
	if err := os.WriteFile(indexPath, indexJSON, 0o644); err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}

	blocks := []MermaidBlock{
		{Content: blockContent, Hash: hash, BlockIndex: 0, Filename: "test_mermaid_0.svg"},
	}

	statuses, err := CheckMermaidCache(tmpDir, blocks, noop)
	if err != nil {
		t.Fatalf("CheckMermaidCache failed: %v", err)
	}

	if statuses[0].Cached {
		t.Errorf("Expected cache miss when SVG file doesn't exist on disk")
	}
}
