package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/caching"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/tool"
)

// MermaidResult holds the result of a mermaid update operation.
type MermaidResult struct {
	FilesScanned      int
	DiagramsFound     int
	FilesWithDiagrams int
	CacheHits         int
	CacheMisses       int
	Rendered          int
	Failed            int
}

// runMermaidUpdate handles the mermaid area update.
// It scans docs for mermaid diagrams and renders missing ones to cache.
func runMermaidUpdate(repoRoot string, opts UpdateOptions, logWriter io.Writer) (MermaidResult, error) {
	result := MermaidResult{}

	fmt.Fprintln(logWriter, "Updating mermaid cache...")

	// Create asset cache (nil = use cache normally)
	cache := caching.NewAssetCache(repoRoot, nil, nil)

	// Scan docs/ for markdown files
	docsDir := paths.DocsSourcePath(repoRoot)

	// Collect all mermaid blocks
	allBlocks := []diagrams.MermaidBlock{}
	blocksByFile := make(map[string][]diagrams.MermaidBlock)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		result.FilesScanned++

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		blocks := diagrams.ExtractMermaidBlocks(string(content), path, docsDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)

			if opts.Verbose {
				relPath, relErr := filepath.Rel(docsDir, path)
				if relErr != nil {
					relPath = path // Fallback to absolute
				}
				fmt.Fprintf(logWriter, "  Found %d diagram(s) in %s\n", len(blocks), relPath)
			}
		}

		return nil
	})
	if err != nil {
		return result, fmt.Errorf("scanning docs: %w", err)
	}

	result.DiagramsFound = len(allBlocks)
	result.FilesWithDiagrams = len(blocksByFile)

	fmt.Fprintf(logWriter, "Scanned %d files, found %d mermaid diagrams in %d files\n",
		result.FilesScanned, result.DiagramsFound, result.FilesWithDiagrams)

	if len(allBlocks) == 0 {
		fmt.Fprintln(logWriter, "No mermaid diagrams found.")
		return result, nil
	}

	// Check cache status for each block
	cacheDir := paths.DocsCachePath(repoRoot)
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	// Ensure cache directory exists
	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		return result, fmt.Errorf("creating cache directory: %w", err)
	}

	// Check cache and identify what needs rendering
	statuses := []diagrams.CacheStatus{}

	for _, block := range allBlocks {
		cleanContent := diagrams.StripSizeDirective(block.Content)
		cachePath, hit := cache.GetMermaid(caching.MermaidCacheKey{
			SourceFile: block.SourceFile,
			BlockIndex: block.BlockIndex,
			Code:       cleanContent,
		})

		// In force mode, treat everything as a cache miss
		if opts.Force {
			hit = false
		}

		if hit {
			result.CacheHits++
		} else {
			result.CacheMisses++
		}

		statuses = append(statuses, diagrams.CacheStatus{
			Block:     block,
			Cached:    hit,
			CachePath: cachePath,
		})
	}

	fmt.Fprintf(logWriter, "Mermaid cache status: %d hits, %d misses (%.1f%% hit rate)\n",
		result.CacheHits, result.CacheMisses,
		float64(result.CacheHits)/float64(len(allBlocks))*100)

	if result.CacheMisses == 0 {
		fmt.Fprintln(logWriter, "All mermaid diagrams are cached. Nothing to render.")
		return result, nil
	}

	if opts.DryRun {
		fmt.Fprintln(logWriter, "\n[DRY RUN] Would render the following mermaid diagrams:")
		for i := range statuses {
			status := &statuses[i]
			if !status.Cached {
				relPath, relErr := filepath.Rel(docsDir, status.Block.SourceFile)
				if relErr != nil {
					relPath = status.Block.SourceFile // Fallback to absolute
				}
				firstLine := strings.Split(status.Block.Content, "\n")[0]
				if len(firstLine) > 50 {
					firstLine = firstLine[:50] + "..."
				}
				fmt.Fprintf(logWriter, "  - %s [%d]: %s\n", relPath, status.Block.BlockIndex, firstLine)
			}
		}
		return result, nil
	}

	// Check Docker availability
	if !dockerutil.IsDockerAvailable() {
		return result, fmt.Errorf("Docker is not available but required for mermaid rendering. Ensure Docker is installed and the daemon is running")
	}

	// Render cache misses using batch tool
	fmt.Fprintf(logWriter, "Rendering %d diagram(s) via mermaid-render tool...\n", result.CacheMisses)

	rendered, failed, err := renderMermaidBatch(repoRoot, statuses, cache, logWriter)
	result.Rendered = rendered
	result.Failed = failed

	if err != nil {
		return result, err
	}

	if result.Failed > 0 {
		fmt.Fprintf(logWriter, "Mermaid: Completed with errors: %d rendered, %d failed\n", result.Rendered, result.Failed)
	} else if result.Rendered > 0 {
		fmt.Fprintf(logWriter, "Mermaid: Successfully rendered %d diagram(s)\n", result.Rendered)
	}

	return result, nil
}

// renderMermaidBatch renders diagrams using the mermaid-render tool.
// Uses shared types from books package to avoid duplication.
func renderMermaidBatch(repoRoot string, statuses []diagrams.CacheStatus, cache *caching.AssetCache, logWriter io.Writer) (int, int, error) {
	// Filter for cache misses
	toRender := []diagrams.CacheStatus{}
	for i := range statuses {
		status := &statuses[i]
		if !status.Cached {
			toRender = append(toRender, *status)
		}
	}

	if len(toRender) == 0 {
		return 0, 0, nil
	}

	// Create work directory for manifest and temp files
	cacheDir := paths.DocsCachePath(repoRoot)
	workDir := filepath.Join(cacheDir, ".mermaid-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("creating work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Local types for mermaid rendering manifest (were in books package, removed during builder migration)
	type mermaidDiagramSpec struct {
		Input  string `json:"input"`
		Output string `json:"output"`
		Config string `json:"config"`
	}
	type mermaidManifest struct {
		Diagrams    []mermaidDiagramSpec `json:"diagrams"`
		Concurrency int                  `json:"concurrency"`
		Theme       string               `json:"theme"`
	}

	// Create diagram specs
	diagramSpecs := make([]mermaidDiagramSpec, 0, len(toRender))
	for i, status := range toRender {
		block := status.Block

		// Write diagram content to temp .mmd file
		mmdFile := filepath.Join(workDir, fmt.Sprintf("diagram_%d.mmd", i))
		if err := os.WriteFile(mmdFile, []byte(block.Content), 0o644); err != nil {
			return 0, 0, fmt.Errorf("writing temp file for %s: %w", block.Filename, err)
		}

		// Container paths: workDir maps to /work, cacheDir maps to /staging
		containerInput := fmt.Sprintf("/work/diagram_%d.mmd", i)

		// Output path relative to cacheDir
		relOutput, err := filepath.Rel(cacheDir, status.CachePath)
		if err != nil {
			return 0, 0, fmt.Errorf("calculating relative path for %s: %w", block.Filename, err)
		}
		containerOutput := "/staging/" + strings.ReplaceAll(relOutput, "\\", "/")

		diagramSpecs = append(diagramSpecs, mermaidDiagramSpec{
			Input:  containerInput,
			Output: containerOutput,
			Config: "/etc/mermaid/mermaid-config.json",
		})
	}

	// Create manifest
	manifest := mermaidManifest{
		Diagrams:    diagramSpecs,
		Concurrency: 4,
		Theme:       "dark",
	}

	// Write manifest file
	manifestPath := filepath.Join(workDir, "manifest.json")
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return 0, 0, fmt.Errorf("marshaling manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return 0, 0, fmt.Errorf("writing manifest: %w", err)
	}

	// Execute mermaid-render tool via bridge
	bridge := tool.GlobalHandlerToolBridge()
	tc := &tool.ToolContext{
		WorkspaceRoot: repoRoot,
		StagingDir:    cacheDir, // Use cache dir as staging for update command
		LogWriter:     logWriter,
	}

	exitCode, execErr := bridge.ExecuteTool(context.Background(), "mermaid-render", tc)
	if execErr != nil {
		return 0, len(toRender), fmt.Errorf("mermaid-render tool failed: %w", execErr)
	}
	if exitCode != 0 {
		return 0, len(toRender), fmt.Errorf("mermaid-render tool exited with code %d", exitCode)
	}

	// Verify outputs and update cache
	rendered := 0
	failed := 0
	for _, status := range toRender {
		if _, err := os.Stat(status.CachePath); err == nil {
			rendered++

			// Store in persistent cache
			block := status.Block
			cleanContent := diagrams.StripSizeDirective(block.Content)
			if err := cache.PutMermaid(status.CachePath, caching.MermaidCacheKey{
				SourceFile: block.SourceFile,
				BlockIndex: block.BlockIndex,
				Code:       cleanContent,
			}); err != nil {
				log.Warnf("Failed to cache %s: %v", block.Filename, err)
			}
		} else {
			failed++
		}
	}

	return rendered, failed, nil
}
