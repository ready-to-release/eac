// Command: update docs
// Short: Scan docs for mermaid diagrams and drawio images, update cache
// Long: Scans all files in docs/ for mermaid code blocks and drawio.png images,
// Long: renders any missing mermaid diagrams to SVG and optimizes drawio images,
// Long: updating the cache at docs/assets/cache/. Cached assets are tracked in git
// Long: for faster CI builds and consistent rendering.
// Long:
// Long: Expected Output:
// Long:   - SVG files in docs/assets/cache/ for mermaid diagrams
// Long:   - Optimized drawio images in docs/assets/cache/
// Long:   - Cache status summary showing hits/misses
// Flag.dry-run: type=bool, default=false, usage=Show what would be updated without actually updating
// Flag.force: type=bool, default=false, usage=Force re-render/re-optimize all assets even if cached
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress for each file
package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/commands/impl/build/buildutil"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(UpdateDocs)
}

// UpdateDocs scans docs/ for mermaid diagrams and drawio images, updates the cache
func UpdateDocs() int {
	// Parse flags
	args := os.Args[2:]
	dryRun := false
	force := false
	verbose := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--force":
			force = true
		case "-v", "--verbose":
			verbose = true
		}
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Create output writer
	var logWriter io.Writer = os.Stdout

	fmt.Println("Updating docs mermaid cache...")

	// Create asset cache
	cache := books.NewAssetCache(repoRoot)

	// Scan docs/ for markdown files
	docsDir := paths.DocsSourcePath(repoRoot)

	// Collect all mermaid blocks
	allBlocks := []books.MermaidBlock{}
	blocksByFile := make(map[string][]books.MermaidBlock)
	filesScanned := 0

	err = filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		filesScanned++

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		blocks := books.ExtractMermaidBlocks(string(content), path, docsDir)
		if len(blocks) > 0 {
			blocksByFile[path] = blocks
			allBlocks = append(allBlocks, blocks...)

			if verbose {
				relPath, _ := filepath.Rel(docsDir, path)
				fmt.Printf("  Found %d diagram(s) in %s\n", len(blocks), relPath)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning docs: %v\n", err)
		return 1
	}

	fmt.Printf("Scanned %d files, found %d mermaid diagrams in %d files\n",
		filesScanned, len(allBlocks), len(blocksByFile))

	// Track processing state
	mermaidFailed := false
	mermaidRendered := 0

	// Check cache status for each block
	cacheDir := paths.DocsCachePath(repoRoot)
	mermaidCacheDir := filepath.Join(cacheDir, "mermaid")

	// Ensure cache directory exists
	if err := os.MkdirAll(mermaidCacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating cache directory: %v\n", err)
		return 1
	}

	// Process mermaid diagrams if any found
	if len(allBlocks) == 0 {
		fmt.Println("No mermaid diagrams found.")
	} else {
		// Check cache and identify what needs rendering
		statuses := []books.CacheStatus{}
		cacheHits := 0
		cacheMisses := 0

		for _, block := range allBlocks {
			cleanContent := books.StripSizeDirective(block.Content)
			cachePath, hit := cache.GetMermaid(books.MermaidCacheKey{
				Code: cleanContent,
			})

			// In force mode, treat everything as a cache miss
			if force {
				hit = false
			}

			if hit {
				cacheHits++
			} else {
				cacheMisses++
			}

			statuses = append(statuses, books.CacheStatus{
				Block:     block,
				Cached:    hit,
				CachePath: cachePath,
			})
		}

		fmt.Printf("Mermaid cache status: %d hits, %d misses (%.1f%% hit rate)\n",
			cacheHits, cacheMisses,
			float64(cacheHits)/float64(len(allBlocks))*100)

		if cacheMisses == 0 {
			fmt.Println("All mermaid diagrams are cached. Nothing to render.")
		} else if dryRun {
			fmt.Println("\n[DRY RUN] Would render the following mermaid diagrams:")
			for _, status := range statuses {
				if !status.Cached {
					relPath, _ := filepath.Rel(docsDir, status.Block.SourceFile)
					firstLine := strings.Split(status.Block.Content, "\n")[0]
					if len(firstLine) > 50 {
						firstLine = firstLine[:50] + "..."
					}
					fmt.Printf("  - %s [%d]: %s\n", relPath, status.Block.BlockIndex, firstLine)
				}
			}
		} else {
			// Check Docker availability
			if !buildutil.IsDockerAvailable() {
				fmt.Fprintln(os.Stderr, "Error: Docker is not available but required for mermaid rendering")
				fmt.Fprintln(os.Stderr, "Ensure Docker is installed and the daemon is running")
				mermaidFailed = true
			} else {
				// Ensure mermaid Docker image exists
				fmt.Println("Ensuring mermaid-cli Docker image...")
				if err := books.EnsureMermaidImage(repoRoot, logWriter); err != nil {
					fmt.Fprintf(os.Stderr, "Error ensuring mermaid image: %v\n", err)
					mermaidFailed = true
				} else {
					// Render cache misses
					fmt.Printf("Rendering %d diagram(s)...\n", cacheMisses)
					failed := 0

					for _, status := range statuses {
						if status.Cached {
							continue
						}

						block := status.Block
						relPath, _ := filepath.Rel(docsDir, block.SourceFile)

						if verbose {
							fmt.Printf("  Rendering %s [%d]...\n", relPath, block.BlockIndex)
						}

						// Render the diagram
						err := books.RenderSingleDiagram(block, status.CachePath, repoRoot, logWriter)
						if err != nil {
							fmt.Fprintf(os.Stderr, "  ❌ Failed to render %s [%d]: %v\n",
								relPath, block.BlockIndex, err)
							failed++
							continue
						}

						// Store in persistent cache
						cleanContent := books.StripSizeDirective(block.Content)
						if err := cache.PutMermaid(status.CachePath, books.MermaidCacheKey{
							Code: cleanContent,
						}); err != nil {
							fmt.Fprintf(os.Stderr, "  ⚠️  Failed to cache %s [%d]: %v\n",
								relPath, block.BlockIndex, err)
							// Non-fatal - continue
						}

						mermaidRendered++

						if verbose || mermaidRendered%10 == 0 {
							fmt.Printf("  ✓ Progress: %d/%d diagrams rendered\n", mermaidRendered, cacheMisses)
						}
					}

					mermaidFailed = failed > 0
					if mermaidFailed {
						fmt.Printf("Mermaid: Completed with errors: %d rendered, %d failed\n", mermaidRendered, failed)
					} else if mermaidRendered > 0 {
						fmt.Printf("✓ Mermaid: Successfully rendered %d diagram(s)\n", mermaidRendered)
					}
				}
			}
		}
	}

	// ========================================================================
	// Phase 2: Drawio image optimization
	// ========================================================================

	fmt.Println()
	fmt.Println("Updating drawio image cache...")

	// Find all drawio.png images in docs/
	drawioImages, err := books.FindDrawioImages(docsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning for drawio images: %v\n", err)
		return 1
	}

	fmt.Printf("Found %d drawio.png image(s)\n", len(drawioImages))

	if len(drawioImages) == 0 {
		fmt.Println("No drawio images found.")
		if mermaidFailed {
			return 1
		}
		return 0
	}

	// Ensure drawio cache directory exists
	drawioCacheDir := filepath.Join(cacheDir, "drawio")
	if err := os.MkdirAll(drawioCacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating drawio cache directory: %v\n", err)
		return 1
	}

	// Check cache status for each image
	drawioStatuses := []books.DrawioCacheStatus{}
	drawioCacheHits := 0
	drawioCacheMisses := 0

	for _, img := range drawioImages {
		cachePath, hit := cache.GetDrawio(books.DrawioCacheKey{
			SourceHash: img.Hash,
			MaxWidth:   books.MaxImageWidthPDF,
		})

		// In force mode, treat everything as a cache miss
		if force {
			hit = false
		}

		if hit {
			drawioCacheHits++
		} else {
			drawioCacheMisses++
		}

		drawioStatuses = append(drawioStatuses, books.DrawioCacheStatus{
			Image:     img,
			Cached:    hit,
			CachePath: cachePath,
		})
	}

	fmt.Printf("Drawio cache status: %d hits, %d misses (%.1f%% hit rate)\n",
		drawioCacheHits, drawioCacheMisses,
		float64(drawioCacheHits)/float64(len(drawioImages))*100)

	if drawioCacheMisses == 0 {
		fmt.Println("All drawio images are cached. Nothing to optimize.")
		if mermaidFailed {
			return 1
		}
		return 0
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] Would optimize the following images:")
		for _, status := range drawioStatuses {
			if !status.Cached {
				fmt.Printf("  - %s\n", status.Image.RelPath)
			}
		}
		if mermaidFailed {
			return 1
		}
		return 0
	}

	// Optimize cache misses
	fmt.Printf("Optimizing %d image(s)...\n", drawioCacheMisses)
	drawioOptimized := 0
	drawioFailed := 0

	for _, status := range drawioStatuses {
		if status.Cached {
			continue
		}

		img := status.Image

		if verbose {
			fmt.Printf("  Optimizing %s...\n", img.RelPath)
		}

		// Optimize the image
		err := books.OptimizeSingleImage(img.SourceFile, status.CachePath, books.MaxImageWidthPDF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Failed to optimize %s: %v\n", img.RelPath, err)
			drawioFailed++
			continue
		}

		// Store in persistent cache
		if err := cache.PutDrawio(status.CachePath, books.DrawioCacheKey{
			SourceHash: img.Hash,
			MaxWidth:   books.MaxImageWidthPDF,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠️  Failed to cache %s: %v\n", img.RelPath, err)
			// Non-fatal - continue
		}

		drawioOptimized++

		if verbose || drawioOptimized%10 == 0 {
			fmt.Printf("  ✓ Progress: %d/%d images optimized\n", drawioOptimized, drawioCacheMisses)
		}
	}

	fmt.Println()
	if drawioFailed > 0 {
		fmt.Printf("Drawio: Completed with errors: %d optimized, %d failed\n", drawioOptimized, drawioFailed)
	} else if drawioOptimized > 0 {
		fmt.Printf("✓ Drawio: Successfully optimized %d image(s)\n", drawioOptimized)
	}

	// Final summary
	fmt.Println()
	fmt.Printf("Cache updated at: %s\n", cacheDir)

	if mermaidFailed || drawioFailed > 0 {
		return 1
	}
	return 0
}
