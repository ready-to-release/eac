package builders

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
)

func TestScanDocsForPlantUMLBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")

	// Create docs structure
	subDir := filepath.Join(docsDir, "explanation")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	// File with plantuml blocks
	md1 := filepath.Join(docsDir, "overview.md")
	content1 := `# Overview

` + "```plantuml" + `
@startuml
Alice -> Bob: Hello
@enduml
` + "```" + `
`
	if err := os.WriteFile(md1, []byte(content1), 0o644); err != nil {
		t.Fatalf("Failed to write md file: %v", err)
	}

	// File without plantuml blocks
	md2 := filepath.Join(docsDir, "plain.md")
	if err := os.WriteFile(md2, []byte("# No diagrams here"), 0o644); err != nil {
		t.Fatalf("Failed to write md file: %v", err)
	}

	// Nested file with plantuml block
	md3 := filepath.Join(subDir, "arch.md")
	content3 := `# Architecture

` + "```plantuml" + `
class User {
  +name: string
}
` + "```" + `

` + "```plantuml" + `
component [API]
component [DB]
[API] --> [DB]
` + "```" + `
`
	if err := os.WriteFile(md3, []byte(content3), 0o644); err != nil {
		t.Fatalf("Failed to write md file: %v", err)
	}

	allBlocks, blocksByFile, err := scanDocsForPlantUMLBlocks(docsDir)
	if err != nil {
		t.Fatalf("scanDocsForPlantUMLBlocks failed: %v", err)
	}

	if len(allBlocks) != 3 {
		t.Errorf("Expected 3 total blocks, got %d", len(allBlocks))
	}

	if len(blocksByFile) != 2 {
		t.Errorf("Expected 2 files with blocks, got %d", len(blocksByFile))
	}

	t.Logf("✓ Found %d blocks in %d files", len(allBlocks), len(blocksByFile))
}

func TestBuildPlantUMLIndex(t *testing.T) {
	docsDir := "/tmp/test/docs"

	blocks := []diagrams.PlantUMLBlock{
		{
			Content:    "@startuml\nAlice -> Bob\n@enduml",
			Hash:       "abc12345",
			SourceFile: filepath.Join(docsDir, "overview.md"),
			BlockIndex: 0,
			Filename:   "overview_plantuml_0_abc12345.svg",
		},
		{
			Content:    "@startuml\nclass User\n@enduml",
			Hash:       "def67890",
			SourceFile: filepath.Join(docsDir, "explanation", "arch.md"),
			BlockIndex: 0,
			Filename:   "arch_plantuml_0_def67890.svg",
		},
	}

	index := buildPlantUMLIndex(blocks, docsDir)

	if len(index.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(index.Entries))
	}

	// Verify first entry
	if index.Entries[0].SourceFile != "overview.md" {
		t.Errorf("Entry 0: expected source_file 'overview.md', got '%s'", index.Entries[0].SourceFile)
	}
	if index.Entries[0].ContentHash != "abc12345" {
		t.Errorf("Entry 0: expected content_hash 'abc12345', got '%s'", index.Entries[0].ContentHash)
	}

	// Verify second entry uses forward slashes
	if index.Entries[1].SourceFile != "explanation/arch.md" {
		t.Errorf("Entry 1: expected source_file 'explanation/arch.md', got '%s'", index.Entries[1].SourceFile)
	}

	// Verify JSON serialization
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal index: %v", err)
	}

	t.Logf("✓ Index JSON:\n%s", string(data))
}

func TestPlantUMLRenderHandlerInterface(t *testing.T) {
	h := &PlantUMLRenderHandler{}

	if h.Name() != "plantuml-render" {
		t.Errorf("Expected name 'plantuml-render', got '%s'", h.Name())
	}

	reqs := h.Requirements()
	if len(reqs) != 1 || reqs[0] != "docker" {
		t.Errorf("Expected requirements [docker], got %v", reqs)
	}

	if !h.IsContainer() {
		t.Error("Expected IsContainer() to return true")
	}

	if h.IsHostInstalled() {
		t.Error("Expected IsHostInstalled() to return false")
	}

	artifacts := h.ListArtifacts(nil, "/tmp")
	if len(artifacts) != 1 || artifacts[0] != "plantuml/" {
		t.Errorf("Expected artifacts ['plantuml/'], got %v", artifacts)
	}
}

func TestScanDocsForPlantUMLBlocksEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("Failed to create docs dir: %v", err)
	}

	allBlocks, blocksByFile, err := scanDocsForPlantUMLBlocks(docsDir)
	if err != nil {
		t.Fatalf("scanDocsForPlantUMLBlocks failed: %v", err)
	}

	if len(allBlocks) != 0 {
		t.Errorf("Expected 0 blocks for empty docs, got %d", len(allBlocks))
	}

	if len(blocksByFile) != 0 {
		t.Errorf("Expected 0 files for empty docs, got %d", len(blocksByFile))
	}
}
