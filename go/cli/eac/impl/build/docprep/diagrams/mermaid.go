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

// MermaidSizePresets maps size names to CSS width values.
var MermaidSizePresets = map[string]string{
	"small":  "33%",
	"medium": "50%",
	"large":  "66%",
	"full":   "100%",
}

// mermaidBlockPattern matches mermaid code blocks with optional size directive.
// Captures: (1) size directive value, (2) mermaid content.
var mermaidBlockPattern = regexp.MustCompile("(?s)```mermaid\\s*\n%%\\{(?:size|width):([^}]+)\\}%%\\s*\n(.*?)```")

// mermaidBlockPlain matches plain mermaid blocks without size directive.
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// MermaidBlock represents a mermaid diagram found in markdown.
type MermaidBlock struct {
	Content    string // The mermaid diagram code
	Hash       string // SHA256 hash of content (first 8 chars for filename)
	SourceFile string // Absolute path to the .md file
	RelPath    string // Relative path from staging dir (for logging)
	BlockIndex int    // Index of block in file (0, 1, 2, ...)
	Filename   string // Generated SVG filename: {source}_mermaid_{idx}_{hash}.svg
	StartPos   int    // Start position in file (for replacement later)
	EndPos     int    // End position in file (for replacement later)
}

// CacheStatus represents the cache state for a mermaid block.
type CacheStatus struct {
	Block     MermaidBlock
	Cached    bool   // true if SVG exists in cache
	CachePath string // absolute path to cached SVG (if exists)
}

// mermaidIndexEntry maps a single mermaid block to its rendered SVG in the builder output.
type mermaidIndexEntry struct {
	SourceFile  string `json:"source_file"`
	BlockIndex  int    `json:"block_index"`
	ContentHash string `json:"content_hash"`
	SVGFilename string `json:"svg_filename"`
}

// mermaidIndex is the manifest written by the mermaid builder.
type mermaidIndex struct {
	Entries []mermaidIndexEntry `json:"entries"`
}

// ProcessMermaidSizing wraps mermaid blocks with size directives in container divs.
func ProcessMermaidSizing(fileIndex *staging.FileIndex, logf func(string, ...any)) error {
	logf("    Processing mermaid diagram sizing...")

	wrapped := 0
	mdFiles := fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := WrapMermaidBlocks(original)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			wrapped++
		}
	}

	logf("    Processed %d files, wrapped %d mermaid blocks with sizing", len(mdFiles), wrapped)
	return nil
}

// WrapMermaidBlocks finds mermaid blocks with size directives and wraps them.
func WrapMermaidBlocks(content string) string {
	result := mermaidBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := mermaidBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		sizeValue := strings.TrimSpace(submatches[1])
		mermaidContent := submatches[2]

		var wrapper strings.Builder
		wrapper.WriteString("<div class=\"mermaid-wrapper\" data-size=\"")
		wrapper.WriteString(strings.ToLower(sizeValue))
		wrapper.WriteString("\">\n\n")
		wrapper.WriteString("```mermaid\n")
		wrapper.WriteString(mermaidContent)
		wrapper.WriteString("```\n\n</div>")

		return wrapper.String()
	})

	return result
}

// ExtractMermaidBlocks scans a markdown file for mermaid code blocks.
// Returns all blocks with metadata for caching and rendering.
func ExtractMermaidBlocks(content, absSourcePath, stagingDir string) []MermaidBlock {
	blocks := []MermaidBlock{}

	relPath, relErr := filepath.Rel(stagingDir, absSourcePath)
	if relErr != nil || relPath == "" {
		relPath = filepath.Base(absSourcePath)
	}

	basename := filepath.Base(absSourcePath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	matches := mermaidBlockPlain.FindAllStringSubmatchIndex(content, -1)

	for idx, match := range matches {
		if len(match) < 4 {
			continue
		}

		diagramContent := strings.TrimSpace(content[match[2]:match[3]])
		if diagramContent == "" {
			continue
		}

		diagramForHash := StripSizeDirective(diagramContent)
		hash := HashContent(diagramForHash)
		filename := fmt.Sprintf("%s_mermaid_%d_%s.svg", basename, idx, hash)

		blocks = append(blocks, MermaidBlock{
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

// StripSizeDirective removes size directive lines from diagram content.
func StripSizeDirective(content string) string {
	lines := strings.Split(content, "\n")
	filtered := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%%{size:") || strings.HasPrefix(trimmed, "%%{width:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

// HashContent returns first 8 chars of SHA256 hash.
func HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// CheckMermaidCache checks which diagrams have pre-rendered SVGs from the builder output.
func CheckMermaidCache(workspaceRoot string, blocks []MermaidBlock, debugf func(string, ...any)) ([]CacheStatus, error) {
	if debugf == nil {
		debugf = func(string, ...any) {}
	}

	builderOutputDir := paths.MermaidBuildOutputPath(workspaceRoot)
	indexPath := filepath.Join(builderOutputDir, "mermaid-index.json")

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		debugf("mermaid: no builder index at %s: %v", indexPath, err)
		statuses := make([]CacheStatus, 0, len(blocks))
		for _, block := range blocks {
			statuses = append(statuses, CacheStatus{
				Block:  block,
				Cached: false,
			})
		}
		return statuses, nil
	}

	var idx mermaidIndex
	if err := json.Unmarshal(indexData, &idx); err != nil {
		return nil, fmt.Errorf("parsing mermaid-index.json: %w", err)
	}

	hashToSVG := make(map[string]string, len(idx.Entries))
	for _, entry := range idx.Entries {
		hashToSVG[entry.ContentHash] = entry.SVGFilename
	}

	debugf("mermaid: loaded builder index with %d entries", len(idx.Entries))

	statuses := make([]CacheStatus, 0, len(blocks))
	for _, block := range blocks {
		cleanContent := StripSizeDirective(block.Content)
		hash := HashContent(cleanContent)

		svgFilename, found := hashToSVG[hash]
		if !found {
			debugf("mermaid: builder index MISS block=%d hash=%s", block.BlockIndex, hash)
			statuses = append(statuses, CacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		svgPath := filepath.Join(builderOutputDir, svgFilename)
		if _, err := os.Stat(svgPath); err != nil {
			debugf("mermaid: builder SVG missing block=%d file=%s", block.BlockIndex, svgPath)
			statuses = append(statuses, CacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		debugf("mermaid: builder HIT block=%d hash=%s file=%s", block.BlockIndex, hash, svgFilename)
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: svgPath,
		})
	}

	return statuses, nil
}

// ReplaceMermaidBlocksWithImages replaces mermaid code blocks with img references.
// extraPathPrefix is "" for PDF, "../" for site builds (from OutputMode.ExtraPathPrefix()).
func ReplaceMermaidBlocksWithImages(
	blocksByFile map[string][]MermaidBlock,
	statuses []CacheStatus,
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

			imgWidth := "100%"
			prefix := modified[:block.StartPos]
			if wrapperIdx := strings.LastIndex(prefix, "data-size=\""); wrapperIdx >= 0 {
				rest := prefix[wrapperIdx+11:]
				if endIdx := strings.Index(rest, "\""); endIdx > 0 {
					sizeVal := rest[:endIdx]
					if w, ok := MermaidSizePresets[strings.ToLower(sizeVal)]; ok {
						imgWidth = w
					} else if strings.HasSuffix(sizeVal, "%") || strings.HasSuffix(sizeVal, "px") {
						imgWidth = sizeVal
					}
				}
			}

			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"Mermaid diagram\" style=\"display:block; width:%s; margin:0 auto;\">",
				relPath, imgWidth,
			)

			modified = modified[:block.StartPos] + imgTag + modified[block.EndPos:]
		}

		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}

		logf("      Replaced %d mermaid block(s) in %s", len(blocks), blocks[0].RelPath)
	}

	return nil
}

// ScanForMermaidDiagrams scans all markdown files in staging directory.
// Returns all mermaid blocks found (grouped by file) and their cache statuses.
func ScanForMermaidDiagrams(
	fileIndex *staging.FileIndex,
	stagingDir, workspaceRoot string,
	logf func(string, ...any),
	debugf func(string, ...any),
) (map[string][]MermaidBlock, []CacheStatus, error) {
	blocksByFile := make(map[string][]MermaidBlock)
	allBlocks := []MermaidBlock{}

	for _, path := range fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := ExtractMermaidBlocks(string(content), path, stagingDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
	}

	if len(allBlocks) == 0 {
		logf("    No mermaid diagrams found")
		return blocksByFile, nil, nil
	}

	statuses, err := CheckMermaidCache(workspaceRoot, allBlocks, debugf)
	if err != nil {
		return nil, nil, fmt.Errorf("checking builder output: %w", err)
	}

	cacheDir := filepath.Join(stagingDir, "assets", "cache", "mermaid")
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
		return nil, nil, fmt.Errorf("%d mermaid diagram(s) not found in builder output (run mermaid build first)", misses)
	}

	return blocksByFile, statuses, nil
}
