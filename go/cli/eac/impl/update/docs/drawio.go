package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/caching"
	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/diagrams"
	"github.com/ready-to-release/eac/go/core/paths"
)

// DrawioResult holds the result of a drawio update operation.
type DrawioResult struct {
	ImagesFound int
	CacheHits   int
	CacheMisses int
	Optimized   int
	Failed      int
}

// runDrawioUpdate handles the drawio area update.
// It scans docs for drawio.png images and optimizes missing ones to cache.
func runDrawioUpdate(repoRoot string, opts UpdateOptions, logWriter io.Writer) (DrawioResult, error) {
	result := DrawioResult{}

	fmt.Fprintln(logWriter, "Updating drawio image cache...")

	// Create asset cache (nil = use cache normally)
	cache := caching.NewAssetCache(repoRoot, nil, nil)

	// Scan docs/ for drawio images
	docsDir := paths.DocsSourcePath(repoRoot)

	// Find all drawio.png images in docs/
	drawioImages, err := diagrams.FindDrawioImages(docsDir)
	if err != nil {
		return result, fmt.Errorf("scanning for drawio images: %w", err)
	}

	result.ImagesFound = len(drawioImages)
	fmt.Fprintf(logWriter, "Found %d drawio.png image(s)\n", result.ImagesFound)

	if len(drawioImages) == 0 {
		fmt.Fprintln(logWriter, "No drawio images found.")
		return result, nil
	}

	// Ensure drawio cache directory exists
	cacheDir := paths.DocsCachePath(repoRoot)
	drawioCacheDir := filepath.Join(cacheDir, "drawio")
	if err := os.MkdirAll(drawioCacheDir, 0o755); err != nil {
		return result, fmt.Errorf("creating drawio cache directory: %w", err)
	}

	// Local cache status type (was books.DrawioCacheStatus, removed during builder migration)
	type drawioCacheStatus struct {
		Image     diagrams.DrawioImage
		Cached    bool
		CachePath string
	}

	// Check cache status for each image
	drawioStatuses := []drawioCacheStatus{}

	for _, img := range drawioImages {
		cachePath, hit := cache.GetDrawio(caching.DrawioCacheKey{
			SourcePath: img.SourceFile,
			SourceHash: img.Hash,
			MaxWidth:   diagrams.MaxImageWidthPDF,
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

		drawioStatuses = append(drawioStatuses, drawioCacheStatus{
			Image:     img,
			Cached:    hit,
			CachePath: cachePath,
		})
	}

	fmt.Fprintf(logWriter, "Drawio cache status: %d hits, %d misses (%.1f%% hit rate)\n",
		result.CacheHits, result.CacheMisses,
		float64(result.CacheHits)/float64(len(drawioImages))*100)

	if result.CacheMisses == 0 {
		fmt.Fprintln(logWriter, "All drawio images are cached. Nothing to optimize.")
		return result, nil
	}

	if opts.DryRun {
		fmt.Fprintln(logWriter, "\n[DRY RUN] Would optimize the following images:")
		for _, status := range drawioStatuses {
			if !status.Cached {
				fmt.Fprintf(logWriter, "  - %s\n", status.Image.RelPath)
			}
		}
		return result, nil
	}

	// Optimize cache misses
	fmt.Fprintf(logWriter, "Optimizing %d image(s)...\n", result.CacheMisses)

	for _, status := range drawioStatuses {
		if status.Cached {
			continue
		}

		img := status.Image

		if opts.Verbose {
			fmt.Fprintf(logWriter, "  Optimizing %s...\n", img.RelPath)
		}

		// Copy the image to cache (builder handles actual rendering/optimization)
		err := copyFileForCache(img.SourceFile, status.CachePath)
		if err != nil {
			log.Errorf("  Failed to optimize %s: %v", img.RelPath, err)
			result.Failed++
			continue
		}

		// Store in persistent cache
		if err := cache.PutDrawio(status.CachePath, caching.DrawioCacheKey{
			SourcePath: img.SourceFile,
			SourceHash: img.Hash,
			MaxWidth:   diagrams.MaxImageWidthPDF,
		}); err != nil {
			log.Warnf("  Failed to cache %s: %v", img.RelPath, err)
			// Non-fatal - continue
		}

		result.Optimized++

		if opts.Verbose || result.Optimized%10 == 0 {
			fmt.Fprintf(logWriter, "  Progress: %d/%d images optimized\n", result.Optimized, result.CacheMisses)
		}
	}

	if result.Failed > 0 {
		fmt.Fprintf(logWriter, "Drawio: Completed with errors: %d optimized, %d failed\n", result.Optimized, result.Failed)
	} else if result.Optimized > 0 {
		fmt.Fprintf(logWriter, "Drawio: Successfully optimized %d image(s)\n", result.Optimized)
	}

	return result, nil
}

// copyFileForCache copies a file from src to dst, creating directories as needed.
func copyFileForCache(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
