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
)

// MaxImageWidthPDF is the maximum width for images in PDF output
// A4 at 150 DPI is about 1240px wide, so 1200px leaves some margin.
const MaxImageWidthPDF = 1200

// drawioCacheDir is the cache directory for optimized drawio images (relative to assets)
// Uses the same cache directory as the persistent cache (docs/assets/cache/drawio).
const drawioCacheDir = "cache/drawio"

// DrawioImage represents a discovered drawio.png image in the docs.
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
// 5. Updates markdown references to point to optimized versions.
func (p *Preprocessor) optimizeDrawioImages() error {
	p.log("    Optimizing drawio images for PDF...")

	// Track unique images found: absolute path -> DrawioImageRef
	type DrawioImageRef struct {
		AbsPath  string
		Hash     string
		Filename string
	}
	imageRefs := make(map[string]DrawioImageRef) // absolute path -> ref

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
			// Resolve the absolute path from the markdown file location
			mdDir := filepath.Dir(path)
			absPath := filepath.Clean(filepath.Join(mdDir, ref))

			if _, exists := imageRefs[absPath]; !exists {
				if _, err := os.Stat(absPath); err == nil {
					// Hash the source file for cache lookup
					hash, hashErr := HashFileContent(absPath)
					if hashErr != nil {
						p.log("    Warning: failed to hash %s: %v", ref, hashErr)
						continue
					}
					imageRefs[absPath] = DrawioImageRef{
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

	// Cache directory: staging/assets/cache/drawio/ (copied from docs/assets/cache/)
	cacheDir := filepath.Join(p.stagingDir, "assets", drawioCacheDir)

	// Step 2: Check cache for each image
	// Cache files are at staging/assets/cache/drawio/{hash}.png (copied from docs/assets/cache/drawio/)
	cacheHits := 0
	cacheMisses := 0
	// Map from original path to cache path for markdown rewriting
	cachePathBySource := make(map[string]string)

	for _, ref := range imageRefs {
		// Check cache using hash
		cacheKey := DrawioCacheKey{
			SourceHash: ref.Hash,
			MaxWidth:   MaxImageWidthPDF,
		}
		// Get the hash filename from assetCache
		persistentPath, _ := p.assetCache.GetDrawio(cacheKey)
		hashFilename := filepath.Base(persistentPath)
		stagingCachePath := filepath.Join(cacheDir, hashFilename)

		// Check if file exists in staging (copied from docs/assets/cache/)
		if _, err := os.Stat(stagingCachePath); err == nil {
			cacheHits++
			cachePathBySource[ref.AbsPath] = stagingCachePath
			continue
		}

		// Cache miss - need to optimize the image
		cacheMisses++

		// Ensure cache directory exists for new renders
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			return fmt.Errorf("creating cache directory: %w", err)
		}

		if err := optimizeImage(ref.AbsPath, stagingCachePath, MaxImageWidthPDF); err != nil {
			p.log("    Warning: failed to optimize %s: %v", ref.Filename, err)
			// Copy original as fallback
			if err := copyFile(ref.AbsPath, stagingCachePath); err != nil {
				return fmt.Errorf("copying fallback image %s: %w", ref.Filename, err)
			}
		}

		// Store in persistent cache for future builds
		if err := p.assetCache.PutDrawio(stagingCachePath, cacheKey); err != nil {
			p.log("    Warning: failed to cache %s: %v", ref.Filename, err)
			// Non-fatal - continue
		}

		cachePathBySource[ref.AbsPath] = stagingCachePath
	}

	// Step 3: Update markdown references to point to cached images
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
		modified := p.rewriteDrawioReferences(original, path, cachePathBySource)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("updating markdown references: %w", err)
	}

	// Step 4: Delete original drawio.png files from staging
	// These are no longer referenced (markdown now points to rendered/) and would bloat the PDF
	deleted := 0
	for _, ref := range imageRefs {
		if err := os.Remove(ref.AbsPath); err != nil {
			if !os.IsNotExist(err) {
				p.log("    Warning: failed to delete %s: %v", ref.Filename, err)
			}
		} else {
			deleted++
		}
	}

	// Log summary with cache statistics
	totalImages := len(imageRefs)
	hitRate := 0.0
	if totalImages > 0 {
		hitRate = float64(cacheHits) / float64(totalImages) * 100
	}
	p.log("    Found %d drawio images: %d cached, %d optimized (%.1f%% hit rate)",
		totalImages, cacheHits, cacheMisses, hitRate)
	p.log("    Updated %d markdown files, deleted %d originals", updated, deleted)
	return nil
}

// drawioPattern matches references to *.drawio.png files in markdown and HTML
// Handles: ![](path.drawio.png), [](path.drawio.png), and <img src="path.drawio.png">.
var drawioPattern = regexp.MustCompile(`(?:\]\(|src=")([^)"]*\.drawio\.png)`)

// findDrawioReferences extracts all drawio.png paths from markdown content.
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

// fullDrawioImagePattern matches full markdown images with optional attr_list:
// ![alt](path.drawio.png) or ![alt](path.drawio.png){ width="100%" }.
var fullDrawioImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]*\.drawio\.png)\)(\s*\{[^}]*\})?`)

// htmlDrawioSrcPattern matches HTML img src: src="path.drawio.png".
var htmlDrawioSrcPattern = regexp.MustCompile(`src="([^"]*\.drawio\.png)"`)

// rewriteDrawioReferences updates drawio.png paths to point to cached images
// For site mode (non-PDF), converts markdown images to HTML <img> tags with path adjustment
// For PDF mode, preserves the original format.
func (p *Preprocessor) rewriteDrawioReferences(content, mdPath string, cachePathBySource map[string]string) string {
	mdDir := filepath.Dir(mdPath)

	// Helper to calculate new path from old path
	getNewPath := func(oldPath string) (string, bool) {
		absPath := filepath.Clean(filepath.Join(mdDir, oldPath))
		cachePath, found := cachePathBySource[absPath]
		if !found {
			return "", false
		}

		relPath, err := filepath.Rel(mdDir, cachePath)
		if err != nil {
			return "", false
		}

		newPath := filepath.ToSlash(relPath)

		// For site mode, add extra ../ to account for MkDocs converting file.md to file/index.html
		if !p.pdfMode {
			newPath = "../" + newPath
		}

		return newPath, true
	}

	// First pass: Convert full markdown images to HTML (site mode) or update paths (PDF mode)
	content = fullDrawioImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := fullDrawioImagePattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		alt := parts[1]
		oldPath := parts[2]
		attrList := ""
		if len(parts) >= 4 {
			attrList = parts[3] // Optional attr_list like { width="100%" }
		}

		newPath, ok := getNewPath(oldPath)
		if !ok {
			return match
		}

		if p.pdfMode {
			// PDF mode: keep markdown format with attr_list
			return fmt.Sprintf("![%s](%s)%s", alt, newPath, attrList)
		}

		// Site mode: convert to HTML <img> tag
		// Parse attr_list and convert to HTML attributes
		style := ""
		if strings.Contains(attrList, `width="100%"`) || strings.Contains(attrList, "width=100%") {
			style = ` style="max-width: 100%;"`
		}
		return fmt.Sprintf(`<img src="%s" alt="%s"%s>`, newPath, alt, style)
	})

	// Second pass: Update any existing HTML img src attributes
	content = htmlDrawioSrcPattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := htmlDrawioSrcPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		oldPath := parts[1]

		newPath, ok := getNewPath(oldPath)
		if !ok {
			return match
		}

		return fmt.Sprintf(`src="%s"`, newPath)
	})

	return content
}

// optimizeImage resizes a PNG image to maxWidth while preserving aspect ratio.
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
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
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
// Returns a list of DrawioImage with source path, relative path, and content hash.
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

// HashFileContent returns the SHA256 hash of a file's content.
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
// Returns an error if optimization fails.
func OptimizeSingleImage(srcPath, dstPath string, maxWidth int) error {
	return optimizeImage(srcPath, dstPath, uint(maxWidth))
}

// DrawioCacheStatus tracks whether a drawio image is cached.
type DrawioCacheStatus struct {
	Image     DrawioImage
	Cached    bool
	CachePath string
}

// UpdateDrawioCache ensures all drawio.png images in docs/ are cached.
// This is fast if images are already cached (just hash comparison).
// Returns the number of images optimized (0 if all cached).
func UpdateDrawioCache(workspaceRoot string, logWriter io.Writer) (int, error) {
	docsDir := filepath.Join(workspaceRoot, "docs")
	cache := NewAssetCache(workspaceRoot)

	// Find all drawio images
	images, err := FindDrawioImages(docsDir)
	if err != nil {
		return 0, fmt.Errorf("scanning for drawio images: %w", err)
	}

	if len(images) == 0 {
		return 0, nil
	}

	// Check cache and optimize any missing
	optimized := 0
	for _, img := range images {
		cacheKey := DrawioCacheKey{
			SourceHash: img.Hash,
			MaxWidth:   MaxImageWidthPDF,
		}

		cachePath, hit := cache.GetDrawio(cacheKey)
		if hit {
			continue // Already cached
		}

		// Optimize and cache
		if err := OptimizeSingleImage(img.SourceFile, cachePath, MaxImageWidthPDF); err != nil {
			fmt.Fprintf(logWriter, "    Warning: failed to optimize %s: %v\n", img.RelPath, err)
			continue
		}

		if err := cache.PutDrawio(cachePath, cacheKey); err != nil {
			fmt.Fprintf(logWriter, "    Warning: failed to cache %s: %v\n", img.RelPath, err)
		}

		optimized++
	}

	return optimized, nil
}
