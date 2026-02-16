package diagrams

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/paths"
)

func TestExtractPlantUMLBlocks(t *testing.T) {
	content := `# Title

Some text before.

` + "```plantuml" + `
@startuml
Alice -> Bob: Hello
Bob -> Alice: Hi
@enduml
` + "```" + `

More text in between.

` + "```plantuml" + `
class User {
  +name: string
  +email: string
}
` + "```" + `

Final text.
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	blocks := ExtractPlantUMLBlocks(content, testFile, tmpDir)

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].BlockIndex != 0 {
		t.Errorf("Block 0: expected index 0, got %d", blocks[0].BlockIndex)
	}
	if !strings.Contains(blocks[0].Content, "@startuml") {
		t.Errorf("Block 0: expected content to contain '@startuml'")
	}
	if !strings.Contains(blocks[0].Content, "Alice -> Bob") {
		t.Errorf("Block 0: expected content to contain 'Alice -> Bob'")
	}
	if blocks[0].Filename != "test_plantuml_0_"+blocks[0].Hash+".svg" {
		t.Errorf("Block 0: unexpected filename %s", blocks[0].Filename)
	}

	if blocks[1].BlockIndex != 1 {
		t.Errorf("Block 1: expected index 1, got %d", blocks[1].BlockIndex)
	}
	if !strings.Contains(blocks[1].Content, "class User") {
		t.Errorf("Block 1: expected content to contain 'class User'")
	}
}

func TestHashPlantUMLContent(t *testing.T) {
	content := "@startuml\nAlice -> Bob: Hello\n@enduml"
	hash1 := HashPlantUMLContent(content)
	hash2 := HashPlantUMLContent(content)

	if hash1 != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
	}

	if len(hash1) != 8 {
		t.Errorf("Hash should be 8 chars, got %d: %s", len(hash1), hash1)
	}

	content2 := "@startuml\nAlice -> Bob: Bye\n@enduml"
	hash3 := HashPlantUMLContent(content2)

	if hash1 == hash3 {
		t.Errorf("Different content should have different hashes")
	}
}

func TestHashPlantUMLContentCRLF(t *testing.T) {
	contentLF := "@startuml\nAlice -> Bob: Hello\n@enduml"
	contentCRLF := "@startuml\r\nAlice -> Bob: Hello\r\n@enduml"

	hashLF := HashPlantUMLContent(contentLF)
	hashCRLF := HashPlantUMLContent(contentCRLF)

	if hashLF != hashCRLF {
		t.Errorf("LF and CRLF content should produce same hash: %s != %s", hashLF, hashCRLF)
	}
}

func TestEnsurePlantUMLWrapper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantWrap bool
	}{
		{
			name:     "already wrapped",
			input:    "@startuml\nAlice -> Bob\n@enduml",
			wantWrap: false,
		},
		{
			name:     "not wrapped",
			input:    "Alice -> Bob\nBob -> Alice",
			wantWrap: true,
		},
		{
			name:     "wrapped with leading whitespace",
			input:    "  @startuml\nAlice -> Bob\n@enduml",
			wantWrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsurePlantUMLWrapper(tt.input)
			if !strings.Contains(result, "@startuml") {
				t.Errorf("Result should contain @startuml")
			}
			if tt.wantWrap && result == tt.input {
				t.Errorf("Expected wrapping but content unchanged")
			}
			if !tt.wantWrap && result != tt.input {
				t.Errorf("Expected no wrapping but content changed")
			}
		})
	}
}

func TestFindStandalonePUMLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "diagrams")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	puml1 := filepath.Join(tmpDir, "sequence.puml")
	puml2 := filepath.Join(subDir, "class.plantuml")
	nonPuml := filepath.Join(tmpDir, "readme.md")

	if err := os.WriteFile(puml1, []byte("@startuml\nAlice -> Bob\n@enduml"), 0o644); err != nil {
		t.Fatalf("Failed to write puml file: %v", err)
	}
	if err := os.WriteFile(puml2, []byte("@startuml\nclass User\n@enduml"), 0o644); err != nil {
		t.Fatalf("Failed to write plantuml file: %v", err)
	}
	if err := os.WriteFile(nonPuml, []byte("# Not a diagram"), 0o644); err != nil {
		t.Fatalf("Failed to write md file: %v", err)
	}

	blocks, err := FindStandalonePUMLFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindStandalonePUMLFiles failed: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 standalone files, got %d", len(blocks))
	}
}

func TestCheckPlantUMLCache(t *testing.T) {
	tmpDir := t.TempDir()
	builderOutputDir := paths.DiagramBuildOutputPath(tmpDir, "docs", "docs-plantuml", "plantuml-render", "plantuml")

	block0Content := "@startuml\nAlice -> Bob\n@enduml"
	block1Content := "@startuml\nclass User\n@enduml"

	blocks := []PlantUMLBlock{
		{
			Content:    block0Content,
			Hash:       HashPlantUMLContent(block0Content),
			BlockIndex: 0,
			Filename:   "test_plantuml_0_" + HashPlantUMLContent(block0Content) + ".svg",
		},
		{
			Content:    block1Content,
			Hash:       HashPlantUMLContent(block1Content),
			BlockIndex: 1,
			Filename:   "test_plantuml_1_" + HashPlantUMLContent(block1Content) + ".svg",
		},
	}

	// First check - no builder output, all should be cache misses
	statuses, err := CheckPlantUMLCache(tmpDir, "docs", "docs-plantuml", blocks, nil)
	if err != nil {
		t.Fatalf("CheckPlantUMLCache failed: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("Expected 2 statuses, got %d", len(statuses))
	}

	for i, status := range statuses {
		if status.Cached {
			t.Errorf("Block %d: expected cache miss, got hit", i)
		}
	}

	// Create builder output with index for first block only
	if err := os.MkdirAll(builderOutputDir, 0o755); err != nil {
		t.Fatalf("Failed to create builder output dir: %v", err)
	}

	indexJSON := fmt.Sprintf(`{
		"entries": [
			{"source_file": "test.md", "block_index": 0, "content_hash": "%s", "svg_filename": "diagram_0.svg"}
		]
	}`, HashPlantUMLContent(block0Content))

	indexPath := filepath.Join(builderOutputDir, "plantuml-index.json")
	if err := os.WriteFile(indexPath, []byte(indexJSON), 0o644); err != nil {
		t.Fatalf("Failed to write plantuml-index.json: %v", err)
	}

	svgPath := filepath.Join(builderOutputDir, "diagram_0.svg")
	if err := os.WriteFile(svgPath, []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatalf("Failed to create SVG file: %v", err)
	}

	// Second check - first should be hit, second miss
	statuses, err = CheckPlantUMLCache(tmpDir, "docs", "docs-plantuml", blocks, nil)
	if err != nil {
		t.Fatalf("CheckPlantUMLCache failed: %v", err)
	}

	if !statuses[0].Cached {
		t.Errorf("Block 0: expected cache hit")
	}
	if statuses[1].Cached {
		t.Errorf("Block 1: expected cache miss")
	}
}

func TestExtractPlantUMLBlocksEmpty(t *testing.T) {
	content := `# Document with no diagrams

Just regular markdown content.
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.md")

	blocks := ExtractPlantUMLBlocks(content, testFile, tmpDir)

	if len(blocks) != 0 {
		t.Errorf("Expected 0 blocks, got %d", len(blocks))
	}
}
