// plantuml-render.go - Build handler for rendering PlantUML diagrams via the plantuml container.
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
	tool.GlobalBuildBridge().RegisterNativeHandler(&PlantUMLRenderHandler{})
}

// PlantUMLRenderHandler renders PlantUML diagrams from markdown sources via
// the plantuml container. Scans docs/**/*.md for ```plantuml blocks and
// docs/ for standalone *.puml files, renders SVGs, and writes them to
// out/build/docs/plantuml/plantuml/.
// Uses .cache/eac/plantuml/ for incremental builds.
type PlantUMLRenderHandler struct{}

func (h *PlantUMLRenderHandler) Name() string { return "plantuml-render" }

func (h *PlantUMLRenderHandler) Requirements() []string { return []string{"docker"} }

func (h *PlantUMLRenderHandler) IsContainer() bool { return true }

func (h *PlantUMLRenderHandler) IsHostInstalled() bool { return false }

func (h *PlantUMLRenderHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	if !IsDockerAvailable() {
		if IsDockerInDocker() {
			return fmt.Errorf("Docker socket not mounted")
		}
		return fmt.Errorf("Docker is not available (required for plantuml rendering)")
	}
	return nil
}

func (h *PlantUMLRenderHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{"plantuml/"}
}

func (h *PlantUMLRenderHandler) Build(
	module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, rawOpts any,
) int {
	opts, _ := rawOpts.(BuildOptions)
	Logln(logWriter, "=== Rendering plantuml diagrams via container ===")

	docsDir := filepath.Join(workspaceRoot, "docs")
	cacheDir := filepath.Join(paths.PlantUMLAccelCachePath(workspaceRoot), module.GetMoniker(), filepath.Base(outputDir))
	outputPlantUMLDir := filepath.Join(outputDir, "plantuml")

	// 1. Scan all markdown files for plantuml blocks
	allBlocks, blocksByFile, err := scanDocsForPlantUMLBlocks(docsDir)
	if err != nil {
		Logln(logWriter, "Error scanning for plantuml blocks: %v", err)
		return 1
	}

	// 2. Also scan for standalone .puml files
	standaloneBlocks, err := diagrams.FindStandalonePUMLFiles(docsDir)
	if err != nil {
		Logln(logWriter, "Error scanning for standalone .puml files: %v", err)
		return 1
	}

	allBlocks = append(allBlocks, standaloneBlocks...)

	if len(allBlocks) == 0 {
		Logln(logWriter, "No plantuml diagrams found")
		_ = os.MkdirAll(outputPlantUMLDir, 0o755)
		return 0
	}

	Logln(logWriter, "Found %d plantuml diagram(s) in %d file(s) + %d standalone file(s)",
		len(allBlocks)-len(standaloneBlocks), len(blocksByFile), len(standaloneBlocks))

	// 3. Convert to itemcache.Item slice and build lookup map
	blocksByKey := make(map[string]diagrams.PlantUMLBlock, len(allBlocks))
	items := make([]itemcache.Item, 0, len(allBlocks))
	for _, block := range allBlocks {
		hash := diagrams.HashPlantUMLContent(block.Content)
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

	// 4. Build function renders misses via plantuml container
	buildFunc := func(misses []itemcache.Item, dir string) (int, error) {
		return h.renderMisses(misses, blocksByKey, workspaceRoot, dir, logWriter)
	}

	// 5. Execute with cache
	cache := itemcache.New(cacheDir)
	result, err := cache.Execute(items, outputPlantUMLDir, buildFunc, opts.ForceRebuild)
	if err != nil {
		Logln(logWriter, "Error: %v", err)
		return 1
	}

	// 6. Write index manifest
	manifest := buildPlantUMLIndex(allBlocks, docsDir)
	manifestPath := filepath.Join(outputPlantUMLDir, "plantuml-index.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile(manifestPath, manifestData, 0o644)

	Logln(logWriter, "PlantUML: %d cached, %d rendered (%.1f%% hit rate)",
		result.CacheHits, result.Built, result.HitRate)
	Logln(logWriter, "Output: %s (%d SVGs, 1 index)", outputPlantUMLDir, result.TotalItems)

	return 0
}

// scanDocsForPlantUMLBlocks scans all markdown files under docsDir for plantuml blocks.
// Returns all blocks, grouped by file.
func scanDocsForPlantUMLBlocks(docsDir string) ([]diagrams.PlantUMLBlock, map[string][]diagrams.PlantUMLBlock, error) {
	var allBlocks []diagrams.PlantUMLBlock
	blocksByFile := make(map[string][]diagrams.PlantUMLBlock)

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

		blocks := diagrams.ExtractPlantUMLBlocks(string(content), path, docsDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)
		}
		return nil
	})

	return allBlocks, blocksByFile, err
}

// PlantUMLIndex maps source plantuml blocks to their rendered SVG filenames.
type PlantUMLIndex struct {
	Entries []PlantUMLIndexEntry `json:"entries"`
}

// PlantUMLIndexEntry maps a single plantuml block to its SVG.
type PlantUMLIndexEntry struct {
	SourceFile  string `json:"source_file"`  // Relative to docs/
	BlockIndex  int    `json:"block_index"`  // 0-based index of plantuml block in file
	ContentHash string `json:"content_hash"` // First 8 chars of SHA256 of plantuml code
	SVGFilename string `json:"svg_filename"` // Filename in the plantuml output directory
}

// buildPlantUMLIndex creates the index mapping source blocks to SVG files.
func buildPlantUMLIndex(blocks []diagrams.PlantUMLBlock, docsDir string) PlantUMLIndex {
	entries := make([]PlantUMLIndexEntry, 0, len(blocks))
	for _, block := range blocks {
		relPath, _ := filepath.Rel(docsDir, block.SourceFile)
		relPath = filepath.ToSlash(relPath)

		entries = append(entries, PlantUMLIndexEntry{
			SourceFile:  relPath,
			BlockIndex:  block.BlockIndex,
			ContentHash: block.Hash,
			SVGFilename: block.Filename,
		})
	}
	return PlantUMLIndex{Entries: entries}
}

// renderMisses renders cache-missed plantuml diagrams via the plantuml tool container.
// The plantuml container renders all .puml files found in its mounted /data directory.
func (h *PlantUMLRenderHandler) renderMisses(
	misses []itemcache.Item, blocksByKey map[string]diagrams.PlantUMLBlock,
	workspaceRoot, cacheDir string, logWriter io.Writer,
) (int, error) {
	if len(misses) == 0 {
		return 0, nil
	}

	// Create work directory for .puml source files
	workDir := filepath.Join(cacheDir, ".plantuml-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return 0, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Write .puml files to work directory
	// The plantuml container renders all .puml files in /data to SVG in-place.
	// We use the cache filename (minus .svg extension) as the .puml filename
	// so the output SVG has the correct name.
	for _, item := range misses {
		block := blocksByKey[item.Key]
		content := diagrams.EnsurePlantUMLWrapper(block.Content)

		// Name the .puml file so the container produces {cacheFilename}.svg
		pumlName := strings.TrimSuffix(item.CacheFilename, ".svg") + ".puml"
		pumlPath := filepath.Join(workDir, pumlName)
		if err := os.WriteFile(pumlPath, []byte(content), 0o644); err != nil {
			return 0, fmt.Errorf("writing temp file for %s: %w", block.Filename, err)
		}
	}

	// Execute plantuml container via tool bridge
	// The plantuml tool mounts {output} -> /data and renders all .puml -> .svg in-place
	bridge := tool.GlobalHandlerToolBridge()
	tc := &tool.ToolContext{
		WorkspaceRoot: workspaceRoot,
		OutputDir:     workDir,
		LogWriter:     logWriter,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "plantuml", tc)
	if execErr != nil {
		return 0, fmt.Errorf("plantuml tool failed: %w", execErr)
	}
	if exitCode != 0 {
		return 0, fmt.Errorf("plantuml tool exited with code %d", exitCode)
	}

	// Copy rendered SVGs from work directory to cache directory
	rendered := 0
	for _, item := range misses {
		svgName := strings.TrimSuffix(item.CacheFilename, ".svg") + ".svg"
		svgPath := filepath.Join(workDir, svgName)

		if _, statErr := os.Stat(svgPath); statErr != nil {
			continue
		}

		cachePath := filepath.Join(cacheDir, item.CacheFilename)
		if cpErr := CopyFile(svgPath, cachePath); cpErr != nil {
			return 0, fmt.Errorf("copying SVG to cache: %w", cpErr)
		}
		rendered++
	}

	return rendered, nil
}

var _ build.BuilderPort = (*PlantUMLRenderHandler)(nil)
