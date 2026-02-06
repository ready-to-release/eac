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

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

func init() {
	h := &MermaidRenderHandler{}
	RegisterHandler(h)
	tool.GlobalBuildBridge().RegisterNativeHandler(h)
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

func (h *MermaidRenderHandler) ValidateModule(module interfaces.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for mermaid rendering)")
	}
	return nil
}

func (h *MermaidRenderHandler) ListArtifacts(module interfaces.ModuleContractPort, workspaceRoot string) []string {
	return []string{"mermaid/"}
}

func (h *MermaidRenderHandler) Build(
	module interfaces.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts BuildOptions,
) int {
	Logln(logWriter, "=== Rendering mermaid diagrams via container ===")

	docsDir := filepath.Join(workspaceRoot, "docs")
	cacheDir := paths.MermaidAccelCachePath(workspaceRoot)
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

	// 2. Ensure directories exist
	_ = os.MkdirAll(cacheDir, 0o755)
	_ = os.MkdirAll(outputMermaidDir, 0o755)

	// 3. Check cache for each block
	var statuses []mermaidBlockStatus
	var toRender []mermaidBlockStatus
	cacheHits := 0

	for _, block := range allBlocks {
		cleanContent := books.StripSizeDirective(block.Content)
		hash := books.HashContent(cleanContent)

		// Cache file: .cache/eac/mermaid/{identifier}_{idx}_{hash}.svg
		relPath, _ := filepath.Rel(docsDir, block.SourceFile)
		identifier := paths.SanitizeForCacheName(relPath)
		cacheFilename := fmt.Sprintf("%s_%d_%s.svg", identifier, block.BlockIndex, hash)
		cachePath := filepath.Join(cacheDir, cacheFilename)

		// Output path
		outputPath := filepath.Join(outputMermaidDir, block.Filename)

		bs := mermaidBlockStatus{
			Block:      block,
			CachePath:  cachePath,
			OutputPath: outputPath,
		}

		if _, statErr := os.Stat(cachePath); statErr == nil && !opts.ForceRebuild {
			bs.Cached = true
			cacheHits++
		} else {
			toRender = append(toRender, bs)
		}

		statuses = append(statuses, bs)
	}

	// 4. Render cache misses via mermaid-render tool
	if len(toRender) > 0 {
		Logln(logWriter, "Rendering %d diagram(s) (%d cached)...", len(toRender), cacheHits)

		rendered, renderErr := renderMermaidBatchForBuilder(toRender, workspaceRoot, cacheDir, logWriter)
		if renderErr != nil {
			Logln(logWriter, "Error rendering mermaid diagrams: %v", renderErr)
			return 1
		}
		Logln(logWriter, "Rendered %d diagram(s) successfully", rendered)
	} else {
		Logln(logWriter, "All %d diagrams cached", cacheHits)
	}

	// 5. Copy all from cache to output
	copied := 0
	for _, s := range statuses {
		if _, statErr := os.Stat(s.CachePath); statErr != nil {
			Logln(logWriter, "Error: expected cached SVG not found: %s", s.CachePath)
			return 1
		}
		if cpErr := CopyFile(s.CachePath, s.OutputPath); cpErr != nil {
			Logln(logWriter, "Error copying mermaid SVG: %v", cpErr)
			return 1
		}
		copied++
	}

	// 6. Write index manifest mapping source file + block index -> SVG filename
	manifest := buildMermaidIndex(allBlocks, docsDir)
	manifestPath := filepath.Join(outputMermaidDir, "mermaid-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	hitRate := 0.0
	if len(allBlocks) > 0 {
		hitRate = float64(cacheHits) / float64(len(allBlocks)) * 100
	}
	Logln(logWriter, "Mermaid: %d cached, %d rendered (%.1f%% hit rate)", cacheHits, len(toRender), hitRate)
	Logln(logWriter, "Output: %s (%d SVGs, 1 index)", outputMermaidDir, copied)

	return 0
}

// mermaidBlockStatus tracks cache state for a mermaid block during builder execution.
type mermaidBlockStatus struct {
	Block      books.MermaidBlock
	Cached     bool
	CachePath  string
	OutputPath string
}

// scanDocsForMermaidBlocks scans all markdown files under docsDir for mermaid blocks.
// Returns all blocks, grouped by file.
func scanDocsForMermaidBlocks(docsDir string) ([]books.MermaidBlock, map[string][]books.MermaidBlock, error) {
	var allBlocks []books.MermaidBlock
	blocksByFile := make(map[string][]books.MermaidBlock)

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

		blocks := books.ExtractMermaidBlocks(string(content), path, docsDir)
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
func buildMermaidIndex(blocks []books.MermaidBlock, docsDir string) MermaidIndex {
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

// renderMermaidBatchForBuilder renders cache-missed mermaid diagrams via the
// mermaid-render tool container. Returns number of successfully rendered diagrams.
func renderMermaidBatchForBuilder(toRender []mermaidBlockStatus, workspaceRoot, cacheDir string, logWriter io.Writer) (int, error) {
	if len(toRender) == 0 {
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

	diagrams := make([]mermaidSpec, 0, len(toRender))
	for i, status := range toRender {
		block := status.Block

		// Write diagram content to temp .mmd file
		mmdFile := filepath.Join(workDir, fmt.Sprintf("diagram_%d.mmd", i))
		if err := os.WriteFile(mmdFile, []byte(block.Content), 0o644); err != nil {
			return 0, fmt.Errorf("writing temp file for %s: %w", block.Filename, err)
		}

		// Container paths: workDir -> /work, cacheDir -> /staging
		containerInput := fmt.Sprintf("/work/diagram_%d.mmd", i)

		// Output to cacheDir as container-relative path
		relCache, _ := filepath.Rel(cacheDir, status.CachePath)
		containerOutput := "/staging/" + filepath.ToSlash(relCache)

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
	for _, status := range toRender {
		if _, statErr := os.Stat(status.CachePath); statErr == nil {
			rendered++
		}
	}

	return rendered, nil
}
