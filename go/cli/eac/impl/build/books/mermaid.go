package books

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

// Size presets for mermaid diagrams.
var mermaidSizePresets = map[string]string{
	"small":  "33%",
	"medium": "50%",
	"large":  "66%",
	"full":   "100%",
}

// mermaidBlockPattern matches mermaid code blocks with optional size directive
// Captures: (1) size directive value, (2) mermaid content.
var mermaidBlockPattern = regexp.MustCompile("(?s)```mermaid\\s*\n%%\\{(?:size|width):([^}]+)\\}%%\\s*\n(.*?)```")

// mermaidBlockPlain matches plain mermaid blocks without size directive.
var mermaidBlockPlain = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```")

// processMermaidSizing wraps mermaid blocks with size directives in container divs
// This enables CSS-based sizing for both web (mermaid2) and PDF (mermaid-to-svg)
//
// Syntax in markdown:
//
//	```mermaid
//	%%{size:medium}%%
//	flowchart TD
//	    A --> B
//	```
//
// Or with explicit width:
//
//	```mermaid
//	%%{width:40%}%%
//	flowchart TD
//	    A --> B
//	```
//
// Size presets: small (33%), medium (50%), large (66%), full (100%).
func (p *Preprocessor) processMermaidSizing() error {
	p.log("    Processing mermaid diagram sizing...")

	wrapped := 0

	// Use file index for efficient iteration (avoids repeated WalkDir calls)
	mdFiles := p.fileIndex.GetMarkdownFiles()

	for _, path := range mdFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := wrapMermaidBlocks(original)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			wrapped++
		}
	}

	p.log("    Processed %d files, wrapped %d mermaid blocks with sizing", len(mdFiles), wrapped)
	return nil
}

// wrapMermaidBlocks finds mermaid blocks with size directives and wraps them.
func wrapMermaidBlocks(content string) string {
	// Process blocks with size directives
	result := mermaidBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract size value and content
		submatches := mermaidBlockPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		sizeValue := strings.TrimSpace(submatches[1])
		mermaidContent := submatches[2]

		// Resolve preset name to percentage
		width := sizeValue
		if preset, ok := mermaidSizePresets[strings.ToLower(sizeValue)]; ok {
			width = preset
		}

		// Ensure width ends with %
		if !strings.HasSuffix(width, "%") && !strings.HasSuffix(width, "px") {
			width += "%"
		}

		// Build wrapped block
		// Store size in data-size attribute for img tag to read
		// Don't set width on div - let img handle its own sizing
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

// MermaidBlock represents a mermaid diagram found in markdown
// Used for caching and pre-rendering during preprocessing
// Exported for use by update docs command.
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

// extractMermaidBlocks scans a markdown file for mermaid code blocks
// Returns all blocks with metadata for caching and rendering.
func extractMermaidBlocks(content, absSourcePath, stagingDir string) []MermaidBlock {
	blocks := []MermaidBlock{}

	// Get relative path for logging
	relPath, relErr := filepath.Rel(stagingDir, absSourcePath)
	if relErr != nil || relPath == "" {
		relPath = filepath.Base(absSourcePath)
	}

	// Get base filename for SVG naming
	basename := filepath.Base(absSourcePath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))

	// Find all mermaid blocks (both with and without size directives)
	// Use mermaidBlockPlain to match all ```mermaid...``` blocks
	matches := mermaidBlockPlain.FindAllStringSubmatchIndex(content, -1)

	for idx, match := range matches {
		// match is an index slice: [fullStart, fullEnd, group1Start, group1End]
		if len(match) < 4 {
			continue
		}

		// Extract the diagram content (group 1)
		diagramContent := strings.TrimSpace(content[match[2]:match[3]])

		// Skip empty blocks
		if diagramContent == "" {
			continue
		}

		// Remove size directives from content before hashing
		// This ensures the hash is based on actual diagram code, not formatting
		diagramForHash := StripSizeDirective(diagramContent)

		// Hash the content for cache key (8 chars like the plugin does)
		hash := HashContent(diagramForHash)

		// Generate filename: {basename}_mermaid_{idx}_{hash}.svg
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

// ExtractMermaidBlocks is the exported version for use by update docs command.
func ExtractMermaidBlocks(content, absSourcePath, baseDir string) []MermaidBlock {
	return extractMermaidBlocks(content, absSourcePath, baseDir)
}

// StripSizeDirective removes size directive lines from diagram content
// Example: %%{size:medium}%% is removed before hashing
// Exported for use by update docs command.
func StripSizeDirective(content string) string {
	// Remove lines starting with %%{size: or %%{width:
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

// HashContent returns first 8 chars of SHA256 hash
// This matches the naming convention used by the mermaid-to-svg plugin
// Exported for use by update docs command.
func HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)[:8]
}

// CacheStatus represents the cache state for a mermaid block
// Exported for use by update docs command.
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

// checkMermaidCache checks which diagrams have pre-rendered SVGs from the builder output.
// Reads mermaid-index.json from out/build/docs/mermaid/mermaid/ and matches by content hash.
// Returns all blocks with their cache status (CachePath points to builder output SVG).
func (p *Preprocessor) checkMermaidCache(blocks []MermaidBlock) ([]CacheStatus, error) {
	builderOutputDir := paths.MermaidBuildOutputPath(p.workspaceRoot)
	indexPath := filepath.Join(builderOutputDir, "mermaid-index.json")

	// Load the builder's index manifest
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		log.Debugf("mermaid: no builder index at %s: %v", indexPath, err)
		// No builder output - all blocks are cache misses
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

	// Build lookup: content_hash -> SVG filename
	hashToSVG := make(map[string]string, len(idx.Entries))
	for _, entry := range idx.Entries {
		hashToSVG[entry.ContentHash] = entry.SVGFilename
	}

	log.Debugf("mermaid: loaded builder index with %d entries", len(idx.Entries))

	statuses := make([]CacheStatus, 0, len(blocks))
	for _, block := range blocks {
		cleanContent := StripSizeDirective(block.Content)
		hash := HashContent(cleanContent)

		svgFilename, found := hashToSVG[hash]
		if !found {
			log.Debugf("mermaid: builder index MISS block=%d hash=%s", block.BlockIndex, hash)
			statuses = append(statuses, CacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		svgPath := filepath.Join(builderOutputDir, svgFilename)
		if _, err := os.Stat(svgPath); err != nil {
			log.Debugf("mermaid: builder SVG missing block=%d file=%s", block.BlockIndex, svgPath)
			statuses = append(statuses, CacheStatus{
				Block:  block,
				Cached: false,
			})
			continue
		}

		log.Debugf("mermaid: builder HIT block=%d hash=%s file=%s", block.BlockIndex, hash, svgFilename)
		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    true,
			CachePath: svgPath,
		})
	}

	return statuses, nil
}

// replaceMermaidBlocksWithImages replaces mermaid code blocks with img references
// This is done ONLY in staging directory, source markdown stays pure
// Uses cache statuses to get the correct SVG path for each block.
func (p *Preprocessor) replaceMermaidBlocksWithImages(blocksByFile map[string][]MermaidBlock, statuses []CacheStatus) error {
	// Build a map from block filename to cache path for quick lookup
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

		// Replace blocks in reverse order (to preserve positions)
		for i := len(blocks) - 1; i >= 0; i-- {
			block := blocks[i]

			// Get the SVG path from cache status
			svgAbsPath := cachePathByBlock[block.Filename]
			if svgAbsPath == "" {
				return fmt.Errorf("no cache path found for block %s", block.Filename)
			}

			// Calculate relative path from markdown file to SVG using link translator
			relPath, err := p.linkTranslator.CalculateRelativePath(filePath, svgAbsPath)
			if err != nil {
				return fmt.Errorf("calculating relative path for %s: %w", block.Filename, err)
			}

			// For site builds (non-PDF), MkDocs converts file.md to file/index.html,
			// adding an extra directory level. Prepend ../ to account for this.
			if !p.pdfMode {
				relPath = "../" + relPath
			}

			// Check if this block is wrapped in a mermaid-wrapper div with data-size
			// Look backwards from the block start to find the wrapper
			imgWidth := "100%"
			prefix := modified[:block.StartPos]
			if wrapperIdx := strings.LastIndex(prefix, "data-size=\""); wrapperIdx >= 0 {
				// Extract size value from data-size="value"
				rest := prefix[wrapperIdx+11:] // Skip data-size="
				if endIdx := strings.Index(rest, "\""); endIdx > 0 {
					sizeVal := rest[:endIdx]
					// Convert preset to percentage
					switch strings.ToLower(sizeVal) {
					case "small":
						imgWidth = "33%"
					case "medium":
						imgWidth = "50%"
					case "large":
						imgWidth = "66%"
					case "full":
						imgWidth = "100%"
					default:
						// Custom value
						if strings.HasSuffix(sizeVal, "%") || strings.HasSuffix(sizeVal, "px") {
							imgWidth = sizeVal
						}
					}
				}
			}

			// Build img tag with relative path to SVG
			// Apply width directly to img for reliable sizing
			imgTag := fmt.Sprintf(
				"<img src=\"%s\" alt=\"Mermaid diagram\" style=\"display:block; width:%s; margin:0 auto;\">",
				relPath, imgWidth,
			)

			// Replace mermaid block with img tag
			modified = modified[:block.StartPos] + imgTag + modified[block.EndPos:]
		}

		// Write back to staging file
		if err := os.WriteFile(filePath, []byte(modified), 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", filePath, err)
		}

		p.log("      ✓ Replaced %d mermaid block(s) in %s", len(blocks), blocks[0].RelPath)
	}

	return nil
}

// scanForMermaidDiagrams scans all markdown files in staging directory.
// Returns all mermaid blocks found (grouped by file) and their cache statuses.
// Reads pre-rendered SVGs from the mermaid builder output and copies them to staging.
func (p *Preprocessor) scanForMermaidDiagrams() (map[string][]MermaidBlock, []CacheStatus, error) {
	blocksByFile := make(map[string][]MermaidBlock)
	allBlocks := []MermaidBlock{}

	// Step 1: Extract all mermaid blocks using file index
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", path, err)
		}

		blocks := extractMermaidBlocks(string(content), path, p.stagingDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
	}

	if len(allBlocks) == 0 {
		p.log("    No mermaid diagrams found")
		return blocksByFile, nil, nil
	}

	// Step 2: Check builder output for pre-rendered SVGs
	statuses, err := p.checkMermaidCache(allBlocks)
	if err != nil {
		return nil, nil, fmt.Errorf("checking builder output: %w", err)
	}

	// Step 3: Copy builder SVGs to staging cache directory so replaceMermaidBlocksWithImages can reference them
	cacheDir := filepath.Join(p.stagingDir, "assets", "cache", "mermaid")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating staging cache directory: %w", err)
	}

	hits := 0
	misses := 0
	for i := range statuses {
		status := &statuses[i]
		if status.Cached {
			// Copy SVG from builder output to staging
			stagingPath := filepath.Join(cacheDir, status.Block.Filename)
			if err := copyFile(status.CachePath, stagingPath); err != nil {
				return nil, nil, fmt.Errorf("copying SVG to staging: %w", err)
			}
			// Update CachePath to the staging location
			status.CachePath = stagingPath
			hits++
		} else {
			misses++
		}
	}

	// Step 4: Log summary and fail if any diagrams are missing
	p.log("    Found %d diagrams in %d files", len(allBlocks), len(blocksByFile))
	p.log("    Builder output: %d available, %d missing", hits, misses)

	if misses > 0 {
		return nil, nil, fmt.Errorf("%d mermaid diagram(s) not found in builder output (run mermaid build first)", misses)
	}

	return blocksByFile, statuses, nil
}
