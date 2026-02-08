package diagrams

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/linking"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
	"github.com/ready-to-release/eac/go/core/paths"
)

// plantumlBlockPattern matches ```plantuml fenced code blocks in markdown.
// Captures: (1) plantuml content between the fences.
var plantumlBlockPattern = regexp.MustCompile("(?s)```plantuml\\s*\n(.*?)```")

// PlantUMLBlock represents a single plantuml diagram found in markdown.
type PlantUMLBlock struct {
	Content    string // The PlantUML source code
	Hash       string // First 8 chars of SHA256
	SourceFile string // Absolute path to .md file
	RelPath    string // Relative to staging/docs dir
	BlockIndex int    // 0-based index in file
	Filename   string // Generated: {basename}_plantuml_{idx}_{hash}.svg
	StartPos   int    // Byte position for replacement
	EndPos     int    // Byte position for replacement
}

// PlantUMLCacheStatus represents the cache state for a plantuml block.
type PlantUMLCacheStatus struct {
	Block     PlantUMLBlock
	Cached    bool
	CachePath string
}

// plantumlIndexEntry maps a single plantuml block to its rendered SVG in the builder output.
type plantumlIndexEntry struct {
	SourceFile  string `json:"source_file"`
	BlockIndex  int    `json:"block_index"`
	ContentHash string `json:"content_hash"`
	SVGFilename string `json:"svg_filename"`
}

// plantumlIndex is the manifest written by the plantuml builder.
type plantumlIndex struct {
	Entries []plantumlIndexEntry `json:"entries"`
}

// ExtractPlantUMLBlocks scans a markdown file for ```plantuml fenced code blocks.
// Returns all blocks with metadata for caching and rendering.
func ExtractPlantUMLBlocks(content, absSourcePath, baseDir string) []PlantUMLBlock {
	blocks := []PlantUMLBlock{}

	relPath, relErr := filepath.Rel(baseDir, absSourcePath)
	if relErr != nil || relPath == "" {
		relPath = filepath.Base(absSourcePath)
	}

	basename := filepath.Base(absSourcePath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	matches := plantumlBlockPattern.FindAllStringSubmatchIndex(content, -1)

	for idx, match := range matches {
		if len(match) < 4 {
			continue
		}

		diagramContent := strings.TrimSpace(content[match[2]:match[3]])
		if diagramContent == "" {
			continue
		}

		hash := HashPlantUMLContent(diagramContent)
		filename := fmt.Sprintf("%s_plantuml_%d_%s.svg", basename, idx, hash)

		blocks = append(blocks, PlantUMLBlock{
			Content:    diagramContent,
			Hash:       hash,
			SourceFile: absSourcePath,
			RelPath:    relPath,
			BlockIndex: idx,
			Filename:   filename,
			StartPos:   match[0],
			EndPos:     match[1],
		})
	}

	return blocks
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
func CheckPlantUMLCache(workspaceRoot string, blocks []PlantUMLBlock, debugf func(string, ...any)) ([]PlantUMLCacheStatus, error) {
	if debugf == nil {
		debugf = func(string, ...any) {}
	}

	builderOutputDir := paths.PlantUMLBuildOutputPath(workspaceRoot)
	indexPath := filepath.Join(builderOutputDir, "plantuml-index.json")

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		debugf("plantuml: no builder index at %s: %v", indexPath, err)
		statuses := make([]PlantUMLCacheStatus, 0, len(blocks))
		for _, block := range blocks {
			statuses = append(statuses, PlantUMLCacheStatus{
				Block:  block,
				Cached: false,
			})
		}
		return statuses, nil
	}

	var idx plantumlIndex
	if err := json.Unmarshal(indexData, &idx); err != nil {
		return nil, fmt.Errorf("parsing plantuml-index.json: %w", err)
	}

	hashToSVG := make(map[string]string, len(idx.Entries))
	for _, entry := range idx.Entries {
		hashToSVG[entry.ContentHash] = entry.SVGFilename
	}

	debugf("plantuml: loaded builder index with %d entries", len(idx.Entries))

	statuses := make([]PlantUMLCacheStatus, 0, len(blocks))
	for _, block := range blocks {
		hash := HashPlantUMLContent(block.Content)

		svgFilename, found := hashToSVG[hash]
		if !found {
			debugf("plantuml: builder index MISS block=%d hash=%s", block.BlockIndex, hash)
			statuses = append(statuses, PlantUMLCacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		svgPath := filepath.Join(builderOutputDir, svgFilename)
		if _, err := os.Stat(svgPath); err != nil {
			debugf("plantuml: builder SVG missing block=%d file=%s", block.BlockIndex, svgPath)
			statuses = append(statuses, PlantUMLCacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		debugf("plantuml: builder HIT block=%d hash=%s file=%s", block.BlockIndex, hash, svgFilename)
		statuses = append(statuses, PlantUMLCacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: svgPath,
		})
	}

	return statuses, nil
}

// ReplacePlantUMLBlocksWithImages replaces plantuml code blocks with img references.
// extraPathPrefix is "" for PDF, "../" for site builds (from OutputMode.ExtraPathPrefix()).
func ReplacePlantUMLBlocksWithImages(
	blocksByFile map[string][]PlantUMLBlock,
	statuses []PlantUMLCacheStatus,
	extraPathPrefix string,
	logf func(string, ...any),
) error {
	cachePathByBlock := make(map[string]string)
	for i := range statuses {
		status := &statuses[i]
		cachePathByBlock[status.Block.Filename] = status.CachePath
	}

	for filePath, blocks := range blocksByFile {
		if len(blocks) == 0 {
			continue
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", filePath, err)
		}

		modified := string(content)

		// Replace blocks in reverse order to preserve positions
		for i := len(blocks) - 1; i >= 0; i-- {
			block := blocks[i]

			svgAbsPath := cachePathByBlock[block.Filename]
			if svgAbsPath == "" {
				return fmt.Errorf("no cache path found for block %s", block.Filename)
			}

			relPath, err := linking.CalculateRelativePath(filePath, svgAbsPath)
			if err != nil {
				return fmt.Errorf("calculating relative path for %s: %w", block.Filename, err)
			}

			relPath = extraPathPrefix + relPath

			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"PlantUML diagram\" style=\"display:block; width:100%%; margin:0 auto;\">",
				relPath,
			)

			modified = modified[:block.StartPos] + imgTag + modified[block.EndPos:]
		}

		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}

		logf("      Replaced %d plantuml block(s) in %s", len(blocks), blocks[0].RelPath)
	}

	return nil
}

// ScanForPlantUMLDiagrams scans all markdown files in staging directory.
// Returns all plantuml blocks found (grouped by file) and their cache statuses.
func ScanForPlantUMLDiagrams(
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	logf func(string, ...any),
	debugf func(string, ...any),
) (map[string][]PlantUMLBlock, []PlantUMLCacheStatus, error) {
	blocksByFile := make(map[string][]PlantUMLBlock)
	allBlocks := []PlantUMLBlock{}

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := ExtractPlantUMLBlocks(string(content), path, stagingDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
	}

	if len(allBlocks) == 0 {
		logf("    No plantuml diagrams found")
		return blocksByFile, nil, nil
	}

	statuses, err := CheckPlantUMLCache(workspaceRoot, allBlocks, debugf)
	if err != nil {
		return nil, nil, fmt.Errorf("checking builder output: %w", err)
	}

	// Copy builder SVGs to staging cache directory
	cacheDir := filepath.Join(stagingDir, "assets", "cache", "plantuml")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating staging cache directory: %w", err)
	}

	hits := 0
	misses := 0
	for i := range statuses {
		status := &statuses[i]
		if status.Cached {
			stagingPath := filepath.Join(cacheDir, status.Block.Filename)
			if err := staging.CopyFile(status.CachePath, stagingPath); err != nil {
				return nil, nil, fmt.Errorf("copying SVG to staging: %w", err)
			}
			status.CachePath = stagingPath
			hits++
		} else {
			misses++
		}
	}

	logf("    Found %d diagrams in %d files", len(allBlocks), len(blocksByFile))
	logf("    Builder output: %d available, %d missing", hits, misses)

	if misses > 0 {
		return nil, nil, fmt.Errorf("%d plantuml diagram(s) not found in builder output (run plantuml build first)", misses)
	}

	return blocksByFile, statuses, nil
}
