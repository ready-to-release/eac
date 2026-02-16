package diagrams

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/commands/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// plantumlBuildOutputPath returns the builder output directory for plantuml diagrams.
func plantumlBuildOutputPath(workspaceRoot, module, componentName string) string {
	return paths.DiagramBuildOutputPath(workspaceRoot, module, componentName, "plantuml-render", "plantuml")
}

// PlantUMLBlock is a type alias for DiagramBlock, preserving backward compatibility.
type PlantUMLBlock = DiagramBlock

// PlantUMLCacheStatus is a type alias for DiagramCacheStatus, preserving backward compatibility.
type PlantUMLCacheStatus = DiagramCacheStatus

// plantumlBlockPattern matches ```plantuml fenced code blocks in markdown.
// Captures: (1) plantuml content between the fences.
var plantumlBlockPattern = regexp.MustCompile("(?s)```plantuml\\s*\n(.*?)```")

// PlantUMLConfig is the DiagramConfig for plantuml diagram processing.
var PlantUMLConfig = DiagramConfig{
	Name:            "plantuml",
	BlockPattern:    plantumlBlockPattern,
	FilePrefix:      "plantuml",
	HashFn:          HashPlantUMLContent,
	PreHashFn:       nil, // PlantUML normalizes line endings inside HashPlantUMLContent
	BuildOutputPath: plantumlBuildOutputPath,
	IndexFilename:   "plantuml-index.json",
	BuildImgTag:     defaultImgTag("PlantUML diagram"),
	CacheSubdir:     "plantuml",
}

// ExtractPlantUMLBlocks scans a markdown file for ```plantuml fenced code blocks.
// Returns all blocks with metadata for caching and rendering.
func ExtractPlantUMLBlocks(content, absSourcePath, baseDir string) []PlantUMLBlock {
	return ExtractBlocks(PlantUMLConfig, content, absSourcePath, baseDir)
}

// HashPlantUMLContent returns first 8 chars of SHA256 hash of content.
// Line endings are normalized to LF for consistent hashing across platforms.
func HashPlantUMLContent(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h)[:8]
}

// EnsurePlantUMLWrapper wraps content with @startuml/@enduml if not already present.
func EnsurePlantUMLWrapper(content string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "@startuml") {
		return content
	}
	return "@startuml\n" + content + "\n@enduml"
}

// FindStandalonePUMLFiles finds *.puml files under docsDir.
// Returns them as PlantUMLBlock entries for unified cache processing.
func FindStandalonePUMLFiles(docsDir string) ([]PlantUMLBlock, error) {
	var blocks []PlantUMLBlock

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".puml") && !strings.HasSuffix(path, ".plantuml") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", path, readErr)
		}

		diagramContent := strings.TrimSpace(string(content))
		if diagramContent == "" {
			return nil
		}

		relPath, _ := filepath.Rel(docsDir, path)
		basename := filepath.Base(path)
		basename = strings.TrimSuffix(basename, filepath.Ext(basename))
		hash := HashPlantUMLContent(diagramContent)
		filename := fmt.Sprintf("%s_plantuml_0_%s.svg", basename, hash)

		blocks = append(blocks, PlantUMLBlock{
			Content:    diagramContent,
			Hash:       hash,
			SourceFile: path,
			RelPath:    relPath,
			BlockIndex: 0,
			Filename:   filename,
		})

		return nil
	})

	return blocks, err
}

// CheckPlantUMLCache checks which diagrams have pre-rendered SVGs from the builder output.
func CheckPlantUMLCache(workspaceRoot, module, componentName string, blocks []PlantUMLBlock, debugf func(string, ...any)) ([]PlantUMLCacheStatus, error) {
	return CheckDiagramCache(PlantUMLConfig, workspaceRoot, module, componentName, blocks, debugf)
}

// ReplacePlantUMLBlocksWithImages replaces plantuml code blocks with img references.
// extraPathPrefix is "" for PDF, "../" for site builds (from OutputMode.ExtraPathPrefix()).
func ReplacePlantUMLBlocksWithImages(
	blocksByFile map[string][]PlantUMLBlock,
	statuses []PlantUMLCacheStatus,
	extraPathPrefix string,
	logf func(string, ...any),
) error {
	return ReplaceBlocksWithImages(PlantUMLConfig, blocksByFile, statuses, extraPathPrefix, logf)
}

// ScanForPlantUMLDiagrams scans all markdown files in staging directory.
// Returns all plantuml blocks found (grouped by file) and their cache statuses.
func ScanForPlantUMLDiagrams(
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot, module, componentName string,
	logf func(string, ...any),
	debugf func(string, ...any),
) (map[string][]PlantUMLBlock, []PlantUMLCacheStatus, error) {
	return ScanForDiagrams(PlantUMLConfig, fileIndex, stagingDir, workspaceRoot, module, componentName, logf, debugf)
}
