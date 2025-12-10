package books

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nfnt/resize"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// MaxImageWidthPDF is the maximum width for images in PDF output
// A4 at 150 DPI is about 1240px wide, so 1200px leaves some margin
const MaxImageWidthPDF = 1200

// renderedDrawioDir is the output directory for optimized drawio images (relative to assets)
// Note: Using "rendered" (not ".rendered") to avoid MkDocs default exclusion of dot-directories
const renderedDrawioDir = "rendered/drawio"

// DrawioImage represents a discovered drawio.png image in the docs
type DrawioImage struct {
	// SourceFile is the absolute path to the drawio.png file
	SourceFile string
	// RelPath is the relative path from docs directory
	RelPath string
	// Hash is the SHA256 hash of the source file content
	Hash string
}

// optimizeDrawioImages finds, optimizes, and stages drawio.png files for PDF (Step 12)
//
// This step uses the persistent cache (docs/assets/cache/drawio/) to avoid
// re-optimizing images that haven't changed:
// 1. Scans markdown files for *.drawio.png references
// 2. Checks cache for each image (by content hash)
// 3. Uses cached version if available, otherwise optimizes and caches
// 4. Stages optimized images in assets/rendered/drawio/
// 5. Updates markdown references to point to optimized versions
func (p *Preprocessor) optimizeDrawioImages() error {
	p.log("    Optimizing drawio images for PDF...")

	// Track unique images found: relative path -> DrawioImageRef
	type DrawioImageRef struct {
		AbsPath  string
		Hash     string
		Filename string
	}
	imageRefs := make(map[string]DrawioImageRef) // relative path -> ref

	// Step 1: Scan markdown files for drawio.png references
	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Find all drawio.png references
		refs := findDrawioReferences(string(content))
		for _, ref := range refs {
			if _, exists := imageRefs[ref]; !exists {
				// Resolve the absolute path from the markdown file location
				mdDir := filepath.Dir(path)
				absPath := filepath.Join(mdDir, ref)
				if _, err := os.Stat(absPath); err == nil {
					// Hash the source file for cache lookup
					hash, hashErr := HashFileContent(absPath)
					if hashErr != nil {
						p.log("    Warning: failed to hash %s: %v", ref, hashErr)
						continue
					}
					imageRefs[ref] = DrawioImageRef{
						AbsPath:  absPath,
						Hash:     hash,
						Filename: filepath.Base(ref),
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scanning markdown files: %w", err)
	}

	if len(imageRefs) == 0 {
		p.log("    No drawio images found to optimize")
		return nil
	}

	// Create output directory (in staging/assets, not staging/docs/assets)
	outputDir := paths.RenderedAssetsPath(p.stagingDir, "drawio")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Step 2: Check cache and optimize/copy each image
	cacheHits := 0
	cacheMisses := 0
	for _, ref := range imageRefs {
		outputPath := filepath.Join(outputDir, ref.Filename)

		// Check persistent cache
		cacheKey := DrawioCacheKey{
			SourceHash: ref.Hash,
			MaxWidth:   MaxImageWidthPDF,
		}
		cachePath, cached := p.assetCache.GetDrawio(cacheKey)

		if cached {
			// Copy from cache to staging
			if err := copyFile(cachePath, outputPath); err != nil {
				p.log("    Warning: failed to copy cached %s: %v", ref.Filename, err)
				// Fall through to optimize
			} else {
				cacheHits++
				continue
			}
		}

		// Cache miss - optimize the image
		cacheMisses++
		if err := optimizeImage(ref.AbsPath, outputPath, MaxImageWidthPDF); err != nil {
			p.log("    Warning: failed to optimize %s: %v", ref.Filename, err)
			// Copy original as fallback
			if err := copyFile(ref.AbsPath, outputPath); err != nil {
				return fmt.Errorf("copying fallback image %s: %w", ref.Filename, err)
			}
		}

		// Store in persistent cache for future builds
		if err := p.assetCache.PutDrawio(outputPath, cacheKey); err != nil {
			p.log("    Warning: failed to cache %s: %v", ref.Filename, err)
			// Non-fatal - continue
		}
	}

	// Step 3: Update markdown references
	updated := 0
	err = filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		original := string(content)
		modified := rewriteDrawioReferences(original, path, p.stagingDir)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0644); err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("updating markdown references: %w", err)
	}

	// Log summary with cache statistics
	totalImages := len(imageRefs)
	hitRate := 0.0
	if totalImages > 0 {
		hitRate = float64(cacheHits) / float64(totalImages) * 100
	}
	p.log("    Found %d drawio images: %d cached, %d optimized (%.1f%% hit rate)",
		totalImages, cacheHits, cacheMisses, hitRate)
	p.log("    Updated %d markdown files", updated)
	return nil
}

// drawioPattern matches references to *.drawio.png files in markdown and HTML
// Handles: ![](path.drawio.png), [](path.drawio.png), and <img src="path.drawio.png">
var drawioPattern = regexp.MustCompile(`(?:\]\(|src=")([^)"]*\.drawio\.png)`)

// findDrawioReferences extracts all drawio.png paths from markdown content
func findDrawioReferences(content string) []string {
	matches := drawioPattern.FindAllStringSubmatch(content, -1)
	refs := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			ref := match[1]
			if !seen[ref] {
				seen[ref] = true
				refs = append(refs, ref)
			}
		}
	}
	return refs
}

// rewriteDrawioReferences updates drawio.png paths to point to rendered directory
func rewriteDrawioReferences(content string, mdPath string, stagingDir string) string {
	return drawioPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Determine if this is markdown ]( or HTML src="
		var prefix string
		var oldPath string

		if strings.Contains(match, "](") {
			// Markdown syntax: ](path.drawio.png)
			prefix = "]("
			pathStart := strings.Index(match, "](") + 2
			oldPath = match[pathStart:]
		} else {
			// HTML syntax: src="path.drawio.png"
			prefix = "src=\""
			pathStart := strings.Index(match, "src=\"") + 5
			oldPath = match[pathStart:]
		}

		// Get just the filename
		filename := filepath.Base(oldPath)

		// Calculate the relative path from this markdown file to the rendered directory
		mdDir := filepath.Dir(mdPath)
		renderedDir := paths.RenderedAssetsPath(stagingDir, "drawio")

		relPath, err := filepath.Rel(mdDir, renderedDir)
		if err != nil {
			return match // Keep original on error
		}

		// Build new path using forward slashes (markdown standard)
		newPath := filepath.ToSlash(filepath.Join(relPath, filename))

		return prefix + newPath
	})
}

// optimizeImage resizes a PNG image to maxWidth while preserving aspect ratio
func optimizeImage(srcPath, dstPath string, maxWidth uint) error {
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer srcFile.Close()

	// Decode PNG
	img, err := png.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("decoding PNG: %w", err)
	}

	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// Only resize if larger than maxWidth
	var resized image.Image
	if width > maxWidth {
		// Calculate new height preserving aspect ratio
		newHeight := uint(float64(height) * float64(maxWidth) / float64(width))
		resized = resize.Resize(maxWidth, newHeight, img, resize.Lanczos3)
	} else {
		resized = img
	}

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer dstFile.Close()

	// Encode with compression
	encoder := &png.Encoder{CompressionLevel: png.BestCompression}

	// Write to buffer first to check size
	var buf bytes.Buffer
	if err := encoder.Encode(&buf, resized); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}

	// Write buffer to file
	if _, err := dstFile.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

// ============================================================================
// Exported functions for update docs command
// ============================================================================

// FindDrawioImages scans the docs directory for all .drawio.png files
// Returns a list of DrawioImage with source path, relative path, and content hash
func FindDrawioImages(docsDir string) ([]DrawioImage, error) {
	var images []DrawioImage
	seen := make(map[string]bool)

	err := filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Check if this is a drawio.png file
		if !strings.HasSuffix(path, ".drawio.png") {
			return nil
		}

		// Skip duplicates (same absolute path)
		if seen[path] {
			return nil
		}
		seen[path] = true

		// Calculate hash of file content
		hash, err := HashFileContent(path)
		if err != nil {
			return fmt.Errorf("hashing %s: %w", path, err)
		}

		// Calculate relative path from docs directory
		relPath, err := filepath.Rel(docsDir, path)
		if err != nil {
			relPath = path
		}

		images = append(images, DrawioImage{
			SourceFile: path,
			RelPath:    relPath,
			Hash:       hash,
		})

		return nil
	})

	return images, err
}

// HashFileContent returns the SHA256 hash of a file's content
func HashFileContent(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// OptimizeSingleImage optimizes a single drawio.png image to the output path
// Returns an error if optimization fails
func OptimizeSingleImage(srcPath, dstPath string, maxWidth int) error {
	return optimizeImage(srcPath, dstPath, uint(maxWidth))
}

// DrawioCacheStatus tracks whether a drawio image is cached
type DrawioCacheStatus struct {
	Image     DrawioImage
	Cached    bool
	CachePath string
}
