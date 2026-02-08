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
)

// DiagramBlock represents a single diagram found in markdown (mermaid, plantuml, etc.).
type DiagramBlock struct {
	Content    string // The diagram source code
	Hash       string // Short hash of content (first 8 chars of SHA256)
	SourceFile string // Absolute path to the .md file
	RelPath    string // Relative path from staging dir (for logging)
	BlockIndex int    // Index of block in file (0, 1, 2, ...)
	Filename   string // Generated SVG filename
	StartPos   int    // Start position in file (for replacement)
	EndPos     int    // End position in file (for replacement)
}

// DiagramCacheStatus represents the cache state for a diagram block.
type DiagramCacheStatus struct {
	Block     DiagramBlock
	Cached    bool   // true if SVG exists in cache
	CachePath string // absolute path to cached SVG (if exists)
}

// diagramIndexEntry maps a single diagram block to its rendered SVG in the builder output.
type diagramIndexEntry struct {
	SourceFile  string `json:"source_file"`
	BlockIndex  int    `json:"block_index"`
	ContentHash string `json:"content_hash"`
	SVGFilename string `json:"svg_filename"`
}

// diagramIndex is the manifest written by a diagram builder.
type diagramIndex struct {
	Entries []diagramIndexEntry `json:"entries"`
}

// DiagramConfig parameterizes the generic diagram processing functions
// for a specific diagram type (mermaid, plantuml, etc.).
type DiagramConfig struct {
	// Name is the diagram type name for log messages ("mermaid", "plantuml").
	Name string

	// BlockPattern is the regex to find diagram code blocks in markdown.
	// Must have a capture group for the diagram content.
	BlockPattern *regexp.Regexp

	// FilePrefix is the prefix for generated filenames (e.g., "mermaid", "plantuml").
	FilePrefix string

	// HashFn computes a short hash (8 chars) from diagram content.
	HashFn func(content string) string

	// PreHashFn optionally transforms content before hashing (e.g., strip size directives).
	// If nil, content is hashed as-is.
	PreHashFn func(content string) string

	// BuildOutputPath returns the builder output directory for this diagram type.
	BuildOutputPath func(workspaceRoot string) string

	// IndexFilename is the name of the builder index file (e.g., "mermaid-index.json").
	IndexFilename string

	// BuildImgTag generates the <img> tag for a rendered diagram.
	// Parameters: relPath to SVG, prefix text before block in file, block metadata.
	BuildImgTag func(relPath, filePrefix string, block DiagramBlock) string

	// CacheSubdir is the subdirectory name under assets/cache/ (e.g., "mermaid", "plantuml").
	CacheSubdir string
}

// ExtractBlocks scans markdown content for diagram code blocks matching the config's pattern.
// Returns all blocks with metadata for caching and rendering.
func ExtractBlocks(cfg DiagramConfig, content, absSourcePath, baseDir string) []DiagramBlock {
	blocks := []DiagramBlock{}

	relPath, relErr := filepath.Rel(baseDir, absSourcePath)
	if relErr != nil || relPath == "" {
		relPath = filepath.Base(absSourcePath)
	}

	basename := filepath.Base(absSourcePath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	matches := cfg.BlockPattern.FindAllStringSubmatchIndex(content, -1)

	for idx, match := range matches {
		if len(match) < 4 {
			continue
		}

		diagramContent := strings.TrimSpace(content[match[2]:match[3]])
		if diagramContent == "" {
			continue
		}

		// Apply pre-hash transform if configured (e.g., strip size directives)
		contentForHash := diagramContent
		if cfg.PreHashFn != nil {
			contentForHash = cfg.PreHashFn(diagramContent)
		}

		hash := cfg.HashFn(contentForHash)
		filename := fmt.Sprintf("%s_%s_%d_%s.svg", basename, cfg.FilePrefix, idx, hash)

		blocks = append(blocks, DiagramBlock{
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

// CheckDiagramCache checks which diagrams have pre-rendered SVGs from the builder output.
func CheckDiagramCache(cfg DiagramConfig, workspaceRoot string, blocks []DiagramBlock, debugf func(string, ...any)) ([]DiagramCacheStatus, error) {
	if debugf == nil {
		debugf = func(string, ...any) {}
	}

	builderOutputDir := cfg.BuildOutputPath(workspaceRoot)
	indexPath := filepath.Join(builderOutputDir, cfg.IndexFilename)

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		debugf("%s: no builder index at %s: %v", cfg.Name, indexPath, err)
		statuses := make([]DiagramCacheStatus, 0, len(blocks))
		for _, block := range blocks {
			statuses = append(statuses, DiagramCacheStatus{
				Block:  block,
				Cached: false,
			})
		}
		return statuses, nil
	}

	var idx diagramIndex
	if err := json.Unmarshal(indexData, &idx); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", cfg.IndexFilename, err)
	}

	hashToSVG := make(map[string]string, len(idx.Entries))
	for _, entry := range idx.Entries {
		hashToSVG[entry.ContentHash] = entry.SVGFilename
	}

	debugf("%s: loaded builder index with %d entries", cfg.Name, len(idx.Entries))

	statuses := make([]DiagramCacheStatus, 0, len(blocks))
	for _, block := range blocks {
		// Re-hash content using the config's hash function (with pre-hash transform)
		contentForHash := block.Content
		if cfg.PreHashFn != nil {
			contentForHash = cfg.PreHashFn(block.Content)
		}
		hash := cfg.HashFn(contentForHash)

		svgFilename, found := hashToSVG[hash]
		if !found {
			debugf("%s: builder index MISS block=%d hash=%s", cfg.Name, block.BlockIndex, hash)
			statuses = append(statuses, DiagramCacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		svgPath := filepath.Join(builderOutputDir, svgFilename)
		if _, err := os.Stat(svgPath); err != nil {
			debugf("%s: builder SVG missing block=%d file=%s", cfg.Name, block.BlockIndex, svgPath)
			statuses = append(statuses, DiagramCacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		debugf("%s: builder HIT block=%d hash=%s file=%s", cfg.Name, block.BlockIndex, hash, svgFilename)
		statuses = append(statuses, DiagramCacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: svgPath,
		})
	}

	return statuses, nil
}

// ReplaceBlocksWithImages replaces diagram code blocks with img references.
// extraPathPrefix is "" for PDF, "../" for site builds.
func ReplaceBlocksWithImages(
	cfg DiagramConfig,
	blocksByFile map[string][]DiagramBlock,
	statuses []DiagramCacheStatus,
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

			// Use config's BuildImgTag for type-specific img generation
			prefix := modified[:block.StartPos]
			imgTag := cfg.BuildImgTag(relPath, prefix, block)

			modified = modified[:block.StartPos] + imgTag + modified[block.EndPos:]
		}

		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}

		logf("      Replaced %d %s block(s) in %s", len(blocks), cfg.Name, blocks[0].RelPath)
	}

	return nil
}

// ScanForDiagrams scans all markdown files in staging directory for diagram blocks.
// Returns all blocks found (grouped by file) and their cache statuses.
func ScanForDiagrams(
	cfg DiagramConfig,
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	logf func(string, ...any),
	debugf func(string, ...any),
) (map[string][]DiagramBlock, []DiagramCacheStatus, error) {
	blocksByFile := make(map[string][]DiagramBlock)
	allBlocks := []DiagramBlock{}

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := ExtractBlocks(cfg, string(content), path, stagingDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
	}

	if len(allBlocks) == 0 {
		logf("    No %s diagrams found", cfg.Name)
		return blocksByFile, nil, nil
	}

	statuses, err := CheckDiagramCache(cfg, workspaceRoot, allBlocks, debugf)
	if err != nil {
		return nil, nil, fmt.Errorf("checking builder output: %w", err)
	}

	cacheDir := filepath.Join(stagingDir, "assets", "cache", cfg.CacheSubdir)
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
		return nil, nil, fmt.Errorf("%d %s diagram(s) not found in builder output (run %s build first)", misses, cfg.Name, cfg.Name)
	}

	return blocksByFile, statuses, nil
}

// defaultHashContent returns first 8 chars of SHA256 hash (no normalization).
func defaultHashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// defaultImgTag generates a simple img tag with 100% width.
func defaultImgTag(altText string) func(relPath, prefix string, block DiagramBlock) string {
	return func(relPath, _ string, _ DiagramBlock) string {
		return fmt.Sprintf(
			"<img src=\"%s\" alt=\"%s\" style=\"display:block; width:100%%; margin:0 auto;\">",
			relPath, altText,
		)
	}
}
