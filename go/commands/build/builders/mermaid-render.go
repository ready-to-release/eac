// mermaid-render.go - Build handler for rendering mermaid diagrams via the mermaid-oci container.
package builders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
	"github.com/ready-to-release/eac/go/commands/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/clibase/caching/itemcache"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	tool.GlobalBuildBridge().RegisterNativeHandler(&MermaidRenderHandler{})
}

// MermaidRenderHandler renders mermaid diagrams from markdown sources via
// the mermaid-oci container. Scans docs/**/*.md for ```mermaid blocks,
// renders SVGs, and writes them to out/build/docs/mermaid/mermaid/.
// Uses .cache/eac/mermaid/ for incremental builds.
type MermaidRenderHandler struct{}

func (h *MermaidRenderHandler) Name() string { return "mermaid-render" }

func (h *MermaidRenderHandler) Requirements() []string { return []string{"docker"} }

func (h *MermaidRenderHandler) IsContainer() bool { return true }

func (h *MermaidRenderHandler) IsHostInstalled() bool { return false }

func (h *MermaidRenderHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for mermaid rendering)")
	}
	return nil
}

func (h *MermaidRenderHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"mermaid/"}
}

func (h *MermaidRenderHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any,
) int {
	opts, _ := rawOpts.(BuildOptions)
	Logln(logWriter, "=== Rendering mermaid diagrams via container ===")

	docsDir := filepath.Join(workspaceRoot, "docs")
	cacheDir := filepath.Join(paths.MermaidAccelCachePath(workspaceRoot), module.GetMoniker(), filepath.Base(outputDir))
	outputMermaidDir := filepath.Join(outputDir, "mermaid")

	// 1. Scan all markdown files for mermaid blocks
	allBlocks, blocksByFile, err := scanDocsForMermaidBlocks(docsDir)
	if err != nil {
		Logln(logWriter, "Error scanning for mermaid blocks: %v", err)
		return 1
	}

	if len(allBlocks) == 0 {
		Logln(logWriter, "No mermaid diagrams found")
		_ = os.MkdirAll(outputMermaidDir, 0o755)
		return 0
	}

	Logln(logWriter, "Found %d mermaid diagram(s) in %d file(s)", len(allBlocks), len(blocksByFile))

	// 2. Convert to itemcache.Item slice and build lookup map
	blocksByKey := make(map[string]diagrams.MermaidBlock, len(allBlocks))
	items := make([]itemcache.Item, 0, len(allBlocks))
	for _, block := range allBlocks {
		cleanContent := diagrams.StripSizeDirective(block.Content)
		hash := diagrams.HashContent(cleanContent)
		relPath, _ := filepath.Rel(docsDir, block.SourceFile)
		identifier := paths.SanitizeForCacheName(relPath)
		cacheFilename := fmt.Sprintf("%s_%d_%s.svg", identifier, block.BlockIndex, hash)
		key := fmt.Sprintf("%s:%d", filepath.ToSlash(relPath), block.BlockIndex)

		item := itemcache.Item{
			Key:           key,
			ContentHash:   hash,
			CacheFilename: cacheFilename,
			OutputRelPath: block.Filename,
		}
		items = append(items, item)
		blocksByKey[key] = block
	}

	// 3. Build function renders misses via mermaid-render container
	buildFunc := func(misses []itemcache.Item, dir string) (int, error) {
		return h.renderMisses(misses, blocksByKey, workspaceRoot, dir, logWriter)
	}

	// 4. Execute with cache
	cache := itemcache.New(cacheDir)
	result, err := cache.Execute(items, outputMermaidDir, buildFunc, opts.ForceRebuild)
	if err != nil {
		Logln(logWriter, "Error: %v", err)
		return 1
	}

	// 5. Write index manifest
	manifest := buildMermaidIndex(allBlocks, docsDir)
	manifestPath := filepath.Join(outputMermaidDir, "mermaid-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	Logln(logWriter, "Mermaid: %d cached, %d rendered (%.1f%% hit rate)",
		result.CacheHits, result.Built, result.HitRate)
	Logln(logWriter, "Output: %s (%d SVGs, 1 index)", outputMermaidDir, result.TotalItems)

	return 0
}

// scanDocsForMermaidBlocks scans all markdown files under docsDir for mermaid blocks.
// Returns all blocks, grouped by file.
func scanDocsForMermaidBlocks(docsDir string) ([]diagrams.MermaidBlock, map[string][]diagrams.MermaidBlock, error) {
	var allBlocks []diagrams.MermaidBlock
	blocksByFile := make(map[string][]diagrams.MermaidBlock)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("reading %s: %w", path, readErr)
		}

		blocks := diagrams.ExtractMermaidBlocks(string(content), path, docsDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
		return nil
	})

	return allBlocks, blocksByFile, err
}

// MermaidIndex maps source mermaid blocks to their rendered SVG filenames.
type MermaidIndex struct {
	Entries []MermaidIndexEntry `json:"entries"`
}

// MermaidIndexEntry maps a single mermaid block to its SVG.
type MermaidIndexEntry struct {
	SourceFile  string `json:"source_file"`  // Relative to docs/ (e.g., "explanation/cd-model/overview.md")
	BlockIndex  int    `json:"block_index"`  // 0-based index of mermaid block in file
	ContentHash string `json:"content_hash"` // First 8 chars of SHA256 of mermaid code
	SVGFilename string `json:"svg_filename"` // Filename in the mermaid output directory
}

// buildMermaidIndex creates the index mapping source blocks to SVG files.
func buildMermaidIndex(blocks []diagrams.MermaidBlock, docsDir string) MermaidIndex {
	entries := make([]MermaidIndexEntry, 0, len(blocks))
	for _, block := range blocks {
		relPath, _ := filepath.Rel(docsDir, block.SourceFile)
		relPath = filepath.ToSlash(relPath)

		entries = append(entries, MermaidIndexEntry{
			SourceFile:  relPath,
			BlockIndex:  block.BlockIndex,
			ContentHash: block.Hash,
			SVGFilename: block.Filename,
		})
	}
	return MermaidIndex{Entries: entries}
}

// renderMisses renders cache-missed mermaid diagrams via the mermaid-render tool container.
func (h *MermaidRenderHandler) renderMisses(
	misses []itemcache.Item, blocksByKey map[string]diagrams.MermaidBlock,
	workspaceRoot, cacheDir string, logWriter io.Writer,
) (int, error) {
	if len(misses) == 0 {
		return 0, nil
	}

	// Create work directory for manifest and temp .mmd files
	workDir := filepath.Join(cacheDir, ".mermaid-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return 0, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Create diagram specs and write temp .mmd files
	type mermaidSpec struct {
		Input  string `json:"input"`
		Output string `json:"output"`
		Config string `json:"config"`
	}

	diagrams := make([]mermaidSpec, 0, len(misses))
	for i, item := range misses {
		block := blocksByKey[item.Key]

		// Write diagram content to temp .mmd file
		mmdFile := filepath.Join(workDir, fmt.Sprintf("diagram_%d.mmd", i))
		if err := os.WriteFile(mmdFile, []byte(block.Content), 0o644); err != nil {
			return 0, fmt.Errorf("writing temp file for %s: %w", block.Filename, err)
		}

		// Container paths: workDir -> /work, cacheDir -> /staging
		containerInput := fmt.Sprintf("/work/diagram_%d.mmd", i)
		containerOutput := "/staging/" + item.CacheFilename

		diagrams = append(diagrams, mermaidSpec{
			Input:  containerInput,
			Output: containerOutput,
			Config: "/etc/mermaid/mermaid-config.json",
		})
	}

	// Create manifest
	type mermaidManifest struct {
		Diagrams    []mermaidSpec `json:"diagrams"`
		Concurrency int           `json:"concurrency"`
		Theme       string        `json:"theme"`
	}
	manifest := mermaidManifest{
		Diagrams:    diagrams,
		Concurrency: 4,
		Theme:       "dark",
	}

	manifestPath := filepath.Join(workDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshaling manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return 0, fmt.Errorf("writing manifest: %w", err)
	}

	// Execute mermaid-render tool via bridge
	bridge := tool.GlobalHandlerToolBridge()
	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		StagingDir:    cacheDir,
		LogWriter:     logWriter,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "mermaid-render", tc)
	if execErr != nil {
		return 0, fmt.Errorf("mermaid-render tool failed: %w", execErr)
	}
	if exitCode != 0 {
		return 0, fmt.Errorf("mermaid-render tool exited with code %d", exitCode)
	}

	// Verify outputs were created
	rendered := 0
	for _, item := range misses {
		cachePath := filepath.Join(cacheDir, item.CacheFilename)
		if _, statErr := os.Stat(cachePath); statErr == nil {
			rendered++
		}
	}

	return rendered, nil
}

var _ build.BuilderPort = (*MermaidRenderHandler)(nil)
