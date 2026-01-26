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
// Long:
// Long: Cache Pruning:
// Long:   Use --prune-cache to identify and remove orphaned cache files that are no
// Long:   longer referenced by any markdown files. Use --dry-run to preview what
// Long:   would be deleted without actually removing files.
// Flag.dry-run: type=bool, default=false, usage=Show what would be changed without actually changing
// Flag.force: type=bool, default=false, usage=Force re-render/re-optimize all assets even if cached
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress for each file
// Flag.prune-cache: type=bool, default=false, usage=Identify and remove orphaned cache files
package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/build/books"
	"github.com/ready-to-release/eac/go/eac/commands/internal/dockerutil"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// commandFlags defines valid flags for the update docs command

var log = logging.C()

func init() {
	registry.Register(UpdateDocs)
}

// UpdateDocs scans docs/ for mermaid diagrams and drawio images, updates the cache.
func UpdateDocs() int {
	// Validate flags
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	dryRun := false
	force := false
	verbose := false
	pruneCache := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		case "--force":
			force = true
		case "-v", "--verbose":
			verbose = true
		case "--prune-cache":
			pruneCache = true
		}
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Handle --prune-cache mode
	if pruneCache {
		return runPruneCache(repoRoot, verbose, dryRun)
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
				relPath, relErr := filepath.Rel(docsDir, path)
				if relErr != nil {
					relPath = path // Fallback to absolute
				}
				fmt.Printf("  Found %d diagram(s) in %s\n", len(blocks), relPath)
			}
		}

		return nil
	})
	if err != nil {
		log.Errorf("Error scanning docs: %v", err)
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
	if err := os.MkdirAll(mermaidCacheDir, 0o755); err != nil {
		log.Errorf("Error creating cache directory: %v", err)
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
				SourceFile: block.SourceFile,
				BlockIndex: block.BlockIndex,
				Code:       cleanContent,
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
					fmt.Printf("  - %s [%d]: %s\n", relPath, status.Block.BlockIndex, firstLine)
				}
			}
		} else {
			// Check Docker availability
			if !dockerutil.IsDockerAvailable() {
				log.Errorf("Error: Docker is not available but required for mermaid rendering")
				log.Errorf("Ensure Docker is installed and the daemon is running")
				mermaidFailed = true
			} else {
				// Ensure mermaid Docker image exists
				fmt.Println("Ensuring mermaid-cli Docker image...")
				if err := books.EnsureMermaidImage(repoRoot, logWriter); err != nil {
					log.Errorf("Error ensuring mermaid image: %v", err)
					mermaidFailed = true
				} else {
					// Render cache misses
					fmt.Printf("Rendering %d diagram(s)...\n", cacheMisses)
					failed := 0

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

						if verbose {
							fmt.Printf("  Rendering %s [%d]...\n", relPath, block.BlockIndex)
						}

						// Render the diagram
						err := books.RenderSingleDiagram(block, status.CachePath, repoRoot, logWriter)
						if err != nil {
							log.Errorf("  ❌ Failed to render %s [%d]: %v",
								relPath, block.BlockIndex, err)
							failed++
							continue
						}

						// Store in persistent cache
						cleanContent := books.StripSizeDirective(block.Content)
						if err := cache.PutMermaid(status.CachePath, books.MermaidCacheKey{
							SourceFile: block.SourceFile,
							BlockIndex: block.BlockIndex,
							Code:       cleanContent,
						}); err != nil {
							log.Errorf("  ⚠️  Failed to cache %s [%d]: %v",
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
		log.Errorf("Error scanning for drawio images: %v", err)
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
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		log.Errorf("Error creating drawio cache directory: %v", err)
		return 1
	}

	// Check cache status for each image
	drawioStatuses := []books.DrawioCacheStatus{}
	drawioCacheHits := 0
	drawioCacheMisses := 0

	for _, img := range drawioImages {
		cachePath, hit := cache.GetDrawio(books.DrawioCacheKey{
			SourcePath: img.SourceFile,
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
			log.Errorf("  ❌ Failed to optimize %s: %v", img.RelPath, err)
			drawioFailed++
			continue
		}

		// Store in persistent cache
		if err := cache.PutDrawio(status.CachePath, books.DrawioCacheKey{
			SourcePath: img.SourceFile,
			SourceHash: img.Hash,
			MaxWidth:   books.MaxImageWidthPDF,
		}); err != nil {
			log.Errorf("  ⚠️  Failed to cache %s: %v", img.RelPath, err)
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

// runPruneCache handles the --prune-cache mode for the update docs command.
// It identifies orphaned cache files and deletes them (unless --dry-run is set).
func runPruneCache(repoRoot string, verbose, dryRun bool) int {
	fmt.Println("Analyzing cache for orphaned files...")

	result, err := PruneCache(repoRoot, verbose)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Print summary
	fmt.Println()
	fmt.Println("Mermaid cache:")
	fmt.Printf("  Cache files: %d\n", result.MermaidCacheFiles)
	fmt.Printf("  Active hashes: %d\n", result.MermaidActiveHashes)
	fmt.Printf("  Orphans (V2): %d (%s)\n",
		len(result.MermaidOrphans),
		formatBytes(result.MermaidBytesRecovered))
	fmt.Printf("  Legacy files: %d (%s)\n",
		len(result.LegacyMermaidFiles),
		formatBytes(result.LegacyMermaidBytes))

	fmt.Println()
	fmt.Println("Drawio cache:")
	fmt.Printf("  Cache files: %d\n", result.DrawioCacheFiles)
	fmt.Printf("  Active hashes: %d\n", result.DrawioActiveHashes)
	fmt.Printf("  Orphans: %d (%s)\n",
		len(result.DrawioOrphans),
		formatBytes(result.DrawioBytesRecovered))

	totalOrphans := result.TotalOrphans()
	totalBytes := result.TotalBytesRecovered()

	if totalOrphans == 0 {
		fmt.Println()
		fmt.Println("No orphaned files found.")
		return 0
	}

	fmt.Println()
	fmt.Printf("Total: %d orphaned files (%s)\n", totalOrphans, formatBytes(totalBytes))

	if dryRun {
		fmt.Println()
		fmt.Println("[DRY RUN] Would delete the above orphaned files")
		if verbose {
			fmt.Println()
			if len(result.MermaidOrphans) > 0 {
				fmt.Println("Orphaned V2 mermaid files:")
				for _, f := range result.MermaidOrphans {
					fmt.Printf("  mermaid/%s\n", f)
				}
			}
			if len(result.LegacyMermaidFiles) > 0 {
				fmt.Println("Legacy mermaid files:")
				for _, f := range result.LegacyMermaidFiles {
					fmt.Printf("  mermaid/%s\n", f)
				}
			}
			if len(result.DrawioOrphans) > 0 {
				fmt.Println("Orphaned drawio files:")
				for _, f := range result.DrawioOrphans {
					fmt.Printf("  drawio/%s\n", f)
				}
			}
		}
		return 0
	}

	// Actually delete
	cacheDir := paths.DocsCachePath(repoRoot)
	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		log.Errorf("Error deleting orphans: %v", err)
		return 1
	}

	fmt.Println()
	fmt.Printf("Deleted %d orphaned files, recovered %s\n",
		deleted, formatBytes(totalBytes))
	return 0
}
