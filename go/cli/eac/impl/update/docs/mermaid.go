package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	dockerutil "github.com/ready-to-release/eac/go/adapters/docker/util"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/books"
	"github.com/ready-to-release/eac/go/core/paths"
)

// MermaidResult holds the result of a mermaid update operation.
type MermaidResult struct {
	FilesScanned int
	DiagramsFound int
	FilesWithDiagrams int
	CacheHits    int
	CacheMisses  int
	Rendered     int
	Failed       int
}

// runMermaidUpdate handles the mermaid area update.
// It scans docs for mermaid diagrams and renders missing ones to cache.
func runMermaidUpdate(repoRoot string, opts UpdateOptions, logWriter io.Writer) (MermaidResult, error) {
	result := MermaidResult{}

	fmt.Fprintln(logWriter, "Updating mermaid cache...")

	// Create asset cache (nil = use cache normally)
	cache := books.NewAssetCache(repoRoot, nil)

	// Scan docs/ for markdown files
	docsDir := paths.DocsSourcePath(repoRoot)

	// Collect all mermaid blocks
	allBlocks := []books.MermaidBlock{}
	blocksByFile := make(map[string][]books.MermaidBlock)

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

		blocks := books.ExtractMermaidBlocks(string(content), path, docsDir)
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
	statuses := []books.CacheStatus{}

	for _, block := range allBlocks {
		cleanContent := books.StripSizeDirective(block.Content)
		cachePath, hit := cache.GetMermaid(books.MermaidCacheKey{
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

		statuses = append(statuses, books.CacheStatus{
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

	// Ensure mermaid Docker image exists
	fmt.Fprintln(logWriter, "Ensuring mermaid-tool Docker image...")
	if err := books.EnsureMermaidImage(repoRoot, logWriter); err != nil {
		return result, fmt.Errorf("ensuring mermaid image: %w", err)
	}

	// Render cache misses
	fmt.Fprintf(logWriter, "Rendering %d diagram(s)...\n", result.CacheMisses)

	for i := range statuses {
		status := &statuses[i]
		if status.Cached {
			continue
		}

		block := status.Block
		relPath, relErr := filepath.Rel(docsDir, block.SourceFile)
		if relErr != nil {
			relPath = block.SourceFile // Fallback to absolute
		}

		if opts.Verbose {
			fmt.Fprintf(logWriter, "  Rendering %s [%d]...\n", relPath, block.BlockIndex)
		}

		// Render the diagram
		err := books.RenderSingleDiagram(block, status.CachePath, repoRoot, logWriter)
		if err != nil {
			log.Errorf("  Failed to render %s [%d]: %v", relPath, block.BlockIndex, err)
			result.Failed++
			continue
		}

		// Store in persistent cache
		cleanContent := books.StripSizeDirective(block.Content)
		if err := cache.PutMermaid(status.CachePath, books.MermaidCacheKey{
			SourceFile: block.SourceFile,
			BlockIndex: block.BlockIndex,
			Code:       cleanContent,
		}); err != nil {
			log.Warnf("  Failed to cache %s [%d]: %v", relPath, block.BlockIndex, err)
			// Non-fatal - continue
		}

		result.Rendered++

		if opts.Verbose || result.Rendered%10 == 0 {
			fmt.Fprintf(logWriter, "  Progress: %d/%d diagrams rendered\n", result.Rendered, result.CacheMisses)
		}
	}

	if result.Failed > 0 {
		fmt.Fprintf(logWriter, "Mermaid: Completed with errors: %d rendered, %d failed\n", result.Rendered, result.Failed)
	} else if result.Rendered > 0 {
		fmt.Fprintf(logWriter, "Mermaid: Successfully rendered %d diagram(s)\n", result.Rendered)
	}

	return result, nil
}
