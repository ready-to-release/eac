package books

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/core/tool"
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

// MermaidManifest defines input for batch mermaid rendering.
type MermaidManifest struct {
	Diagrams    []MermaidDiagramSpec `json:"diagrams"`
	Concurrency int                  `json:"concurrency"`
	Theme       string               `json:"theme"`
}

// MermaidDiagramSpec defines a single diagram to render.
type MermaidDiagramSpec struct {
	Input  string `json:"input"`  // Path to .mmd file in work dir
	Output string `json:"output"` // Path for output .svg
	Config string `json:"config"` // Mermaid config file
}

// renderMermaidBatch renders multiple diagrams using the mermaid-render tool.
// This replaces the previous direct Docker execution with the tool bridge.
func (p *Preprocessor) renderMermaidBatch(toRender []CacheStatus) (int, error) {
	if len(toRender) == 0 {
		return 0, nil
	}

	// Create work directory for manifest and temp files
	workDir := filepath.Join(p.stagingDir, ".mermaid-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return 0, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up work directory

	// Create diagram specs and write temp .mmd files
	diagrams := make([]MermaidDiagramSpec, 0, len(toRender))
	for i, status := range toRender {
		block := status.Block

		// Write diagram content to temp .mmd file
		mmdFile := filepath.Join(workDir, fmt.Sprintf("diagram_%d.mmd", i))
		if err := os.WriteFile(mmdFile, []byte(block.Content), 0o644); err != nil {
			return 0, fmt.Errorf("writing temp file for %s: %w", block.Filename, err)
		}

		// Convert paths to container paths
		// workDir maps to /work in container
		// stagingDir maps to /staging in container
		containerInput := fmt.Sprintf("/work/diagram_%d.mmd", i)

		// Output path relative to staging, converted to container path
		relOutput, err := filepath.Rel(p.stagingDir, status.CachePath)
		if err != nil {
			return 0, fmt.Errorf("calculating relative path for %s: %w", block.Filename, err)
		}
		containerOutput := "/staging/" + strings.ReplaceAll(relOutput, "\\", "/")

		diagrams = append(diagrams, MermaidDiagramSpec{
			Input:  containerInput,
			Output: containerOutput,
			Config: "/etc/mermaid/mermaid-config.json",
		})
	}

	// Create manifest
	manifest := MermaidManifest{
		Diagrams:    diagrams,
		Concurrency: 4,
		Theme:       "dark",
	}

	// Write manifest file
	manifestPath := filepath.Join(workDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshaling manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return 0, fmt.Errorf("writing manifest: %w", err)
	}

	p.log("    Rendering %d diagram(s) via mermaid-render tool...", len(toRender))

	// Execute mermaid-render tool via bridge
	bridge := tool.GlobalHandlerToolBridge()
	tc := &tool.ToolContext{
		WorkspaceRoot: p.workspaceRoot,
		StagingDir:    p.stagingDir,
		LogWriter:     p.logWriter,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "mermaid-render", tc)
	if execErr != nil {
		return 0, fmt.Errorf("mermaid-render tool failed: %w", execErr)
	}
	if exitCode != 0 {
		return 0, fmt.Errorf("mermaid-render tool exited with code %d", exitCode)
	}

	// Verify outputs were created and update cache
	rendered := 0
	for _, status := range toRender {
		if _, err := os.Stat(status.CachePath); err == nil {
			rendered++

			// Save to persistent cache
			block := status.Block
			cleanContent := StripSizeDirective(block.Content)
			if err := p.assetCache.PutMermaid(status.CachePath, MermaidCacheKey{
				SourceFile: block.SourceFile,
				BlockIndex: block.BlockIndex,
				Code:       cleanContent,
			}); err != nil {
				p.log("      ⚠️  Failed to cache %s: %v", block.Filename, err)
			}
		}
	}

	return rendered, nil
}

// renderMermaidDiagrams renders multiple mermaid diagrams using the tool bridge.
// Only renders cache misses (cached diagrams are skipped).
// Returns number of diagrams rendered and any error.
func (p *Preprocessor) renderMermaidDiagrams(statuses []CacheStatus) (int, error) {
	// Check Docker availability first - fail fast if unavailable
	if !dockerutil.IsDockerAvailable() {
		errorMsg := "Docker is not available but required for mermaid diagram rendering"
		if dockerutil.IsDinD() {
			errorMsg += "\nRunning in container: mount Docker socket with -v /var/run/docker.sock:/var/run/docker.sock"
		} else {
			errorMsg += "\nEnsure Docker is installed and the daemon is running"
		}
		return 0, fmt.Errorf("%s", errorMsg)
	}

	// Filter for cache misses
	toRender := []CacheStatus{}
	for i := range statuses {
		status := &statuses[i]
		if !status.Cached {
			toRender = append(toRender, *status)
		}
	}

	if len(toRender) == 0 {
		p.log("    All diagrams cached, nothing to render")
		return 0, nil
	}

	// Use batch rendering via tool bridge
	rendered, err := p.renderMermaidBatch(toRender)
	if err != nil {
		return rendered, err
	}

	p.log("    ✓ Rendered %d diagram(s) successfully", rendered)
	return rendered, nil
}

// checkMermaidCache checks which diagrams are already cached
// Cache is at staging/assets/cache/mermaid/ (copied from docs/assets/cache/mermaid/)
// Returns all blocks with their cache status.
func (p *Preprocessor) checkMermaidCache(blocks []MermaidBlock) ([]CacheStatus, error) {
	// Cache directory: staging/assets/cache/mermaid/ (copied from docs/assets/cache/)
	cacheDir := filepath.Join(p.stagingDir, "assets", "cache", "mermaid")
	log.Debugf("cache: checkMermaidCache blocks=%d cacheDir=%s", len(blocks), cacheDir)

	statuses := make([]CacheStatus, 0, len(blocks))

	for _, block := range blocks {
		// Use stripped content for cache key to ensure consistency
		cleanContent := StripSizeDirective(block.Content)
		cacheKey := MermaidCacheKey{
			SourceFile: block.SourceFile,
			BlockIndex: block.BlockIndex,
			Code:       cleanContent,
		}

		// Get the hash for this diagram (assetCache computes the hash)
		persistentPath, _ := p.assetCache.GetMermaid(cacheKey)
		hashFilename := filepath.Base(persistentPath)
		stagingCachePath := filepath.Join(cacheDir, hashFilename)

		// Check if file exists in staging (copied from docs/assets/cache/)
		cached := false
		if _, err := os.Stat(stagingCachePath); err == nil {
			cached = true
			log.Debugf("cache: mermaid staging HIT block=%d file=%s", block.BlockIndex, stagingCachePath)
		} else {
			log.Debugf("cache: mermaid staging MISS block=%d file=%s", block.BlockIndex, stagingCachePath)
		}

		statuses = append(statuses, CacheStatus{
			Block:     block,
			Cached:    cached,
			CachePath: stagingCachePath,
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

// scanForMermaidDiagrams scans all markdown files in staging directory
// Returns all mermaid blocks found (grouped by file) and their cache statuses
// Now includes cache checking and statistics.
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

	// Step 2: Check cache for all blocks
	statuses, err := p.checkMermaidCache(allBlocks)
	if err != nil {
		return nil, nil, fmt.Errorf("checking cache: %w", err)
	}

	// Step 3: Analyze cache statistics
	cacheHits := 0
	cacheMisses := 0
	for i := range statuses {
		status := &statuses[i]
		if status.Cached {
			cacheHits++
		} else {
			cacheMisses++
		}
	}

	// Step 4: Render cache misses
	rendered := 0
	var renderErr error
	if cacheMisses > 0 {
		rendered, renderErr = p.renderMermaidDiagrams(statuses)
		// Note: renderErr may be non-nil if some renders failed,
		// but we continue to show statistics
	}

	// Step 5: Log summary
	totalDiagrams := len(allBlocks)
	cacheHitRate := 0.0
	if totalDiagrams > 0 {
		cacheHitRate = float64(cacheHits) / float64(totalDiagrams) * 100
	}

	p.log("    Found %d diagrams in %d files", totalDiagrams, len(blocksByFile))
	p.log("    Cache: %d hits, %d misses (%.1f%% hit rate)",
		cacheHits, cacheMisses, cacheHitRate)
	if rendered > 0 {
		p.log("    Rendered: %d diagrams", rendered)
	}

	// Step 6: Log detailed breakdown by file
	for _, blocks := range blocksByFile {
		relPath := blocks[0].RelPath
		fileHits := 0
		fileMisses := 0

		for _, block := range blocks {
			// Find this block's cache status
			for j := range statuses {
				status := &statuses[j]
				if status.Block.Filename == block.Filename {
					if status.Cached {
						fileHits++
					} else {
						fileMisses++
					}
					break
				}
			}
		}

		p.log("      %s: %d diagram(s) (%d cached, %d to render)",
			relPath, len(blocks), fileHits, fileMisses)

		// Show details for each diagram
		for _, block := range blocks {
			// Find cache status
			cached := false
			for _, status := range statuses {
				if status.Block.Filename == block.Filename {
					cached = status.Cached
					break
				}
			}

			cacheMarker := "❌"
			if cached {
				cacheMarker = "✓"
			}

			// Show first line of content
			firstLine := strings.Split(block.Content, "\n")[0]
			if len(firstLine) > 45 {
				firstLine = firstLine[:45] + "..."
			}

			p.log("        [%d] %s %s %s",
				block.BlockIndex, cacheMarker, firstLine, block.Filename)
		}
	}

	// Return rendering error if any diagrams failed
	// (but still return the blocksByFile and statuses so caller can see what was found)
	if renderErr != nil {
		return blocksByFile, statuses, fmt.Errorf("rendering: %w", renderErr)
	}

	return blocksByFile, statuses, nil
}
