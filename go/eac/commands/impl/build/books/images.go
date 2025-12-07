package books

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nfnt/resize"
)

// MaxImageWidthPDF is the maximum width for images in PDF output
// A4 at 150 DPI is about 1240px wide, so 1200px leaves some margin
const MaxImageWidthPDF = 1200

// renderedDrawioDir is the output directory for optimized drawio images (relative to assets)
// Note: Using "rendered" (not ".rendered") to avoid MkDocs default exclusion of dot-directories
const renderedDrawioDir = "rendered/drawio"

// optimizeDrawioImages finds, optimizes, and stages drawio.png files for PDF (Step 6b)
//
// This step:
// 1. Scans markdown files for *.drawio.png references
// 2. Resizes large images to MaxImageWidthPDF while preserving aspect ratio
// 3. Stages optimized images in assets/rendered/drawio/
// 4. Updates markdown references to point to optimized versions
func (p *Preprocessor) optimizeDrawioImages() error {
	p.log("    Optimizing drawio images for PDF...")

	// Track unique images found and their source paths
	imageRefs := make(map[string]string) // relative path -> absolute source path

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
					imageRefs[ref] = absPath
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
	outputDir := filepath.Join(p.stagingDir, "assets", renderedDrawioDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Step 2: Optimize each image
	optimized := 0
	for relPath, absPath := range imageRefs {
		// Get just the filename for output
		filename := filepath.Base(relPath)
		outputPath := filepath.Join(outputDir, filename)

		if err := optimizeImage(absPath, outputPath, MaxImageWidthPDF); err != nil {
			p.log("    Warning: failed to optimize %s: %v", filename, err)
			// Copy original as fallback
			if err := copyFile(absPath, outputPath); err != nil {
				return fmt.Errorf("copying fallback image %s: %w", filename, err)
			}
		}
		optimized++
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

	p.log("    Optimized %d images, updated %d markdown files", optimized, updated)
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
		renderedDir := filepath.Join(stagingDir, "assets", renderedDrawioDir)

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
