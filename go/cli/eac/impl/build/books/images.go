package books

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/go/core/paths"
)

// MaxImageWidthPDF is the maximum width for images in PDF output
// A4 at 150 DPI is about 1240px wide, so 1200px leaves some margin.
const MaxImageWidthPDF = 1200

// DrawioImage represents a discovered drawio.png image in the docs.
type DrawioImage struct {
	// SourceFile is the absolute path to the drawio.png file
	SourceFile string
	// RelPath is the relative path from docs directory
	RelPath string
	// Hash is the SHA256 hash of the source file content
	Hash string
}

// optimizeDrawioImages reads pre-rendered PNGs from the drawio builder output (Step 12)
//
// The drawio builder (drawio-render) has already rendered all .drawio.png files
// and placed them at out/build/docs/drawio/drawio/ preserving relative paths.
// This step:
// 1. Scans markdown files for *.drawio.png references
// 2. Looks up pre-rendered PNGs from builder output
// 3. Copies rendered PNGs into staging
// 4. Updates markdown references to point to the staged copies
// 5. Deletes original drawio.png files from staging (they bloat PDFs)
func (p *Preprocessor) optimizeDrawioImages() error {
	p.log("    Processing drawio images from builder output...")

	// Builder output directory: out/build/docs/drawio/drawio/
	builderOutputDir := paths.DrawioBuildOutputPath(p.workspaceRoot)

	// Track unique images found: absolute staging path -> DrawioImageRef
	type DrawioImageRef struct {
		AbsPath  string // absolute path in staging
		RelPath  string // relative path from docs/assets/ (matches builder output structure)
		Filename string // basename
	}
	imageRefs := make(map[string]DrawioImageRef)

	// Step 1: Scan markdown files for drawio.png references
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		refs := findDrawioReferences(string(content))
		for _, ref := range refs {
			mdDir := filepath.Dir(path)
			absPath := filepath.Clean(filepath.Join(mdDir, ref))

			if _, exists := imageRefs[absPath]; !exists {
				if _, err := os.Stat(absPath); err == nil {
					// Calculate relative path from staging assets dir
					// The builder output preserves the same relative structure as docs/assets/
					assetsDir := filepath.Join(p.stagingDir, "assets")
					relPath, relErr := filepath.Rel(assetsDir, absPath)
					if relErr != nil {
						p.warn("failed to compute relative path for %s: %v", ref, relErr)
						continue
					}
					imageRefs[absPath] = DrawioImageRef{
						AbsPath:  absPath,
						RelPath:  relPath,
						Filename: filepath.Base(ref),
					}
				}
			}
		}
	}

	if len(imageRefs) == 0 {
		p.log("    No drawio images found")
		return nil
	}

	// Step 2: Copy rendered PNGs from builder output to staging
	// Use staging/assets/rendered/drawio/ as the destination
	renderedDir := paths.RenderedAssetsPath(p.stagingDir, "drawio")
	if err := os.MkdirAll(renderedDir, 0o755); err != nil {
		return fmt.Errorf("creating rendered directory: %w", err)
	}

	copied := 0
	cachePathBySource := make(map[string]string)

	for _, ref := range imageRefs {
		// Look up pre-rendered PNG from builder output
		builderPath := filepath.Join(builderOutputDir, ref.RelPath)
		stagingRenderedPath := filepath.Join(renderedDir, ref.RelPath)

		if err := os.MkdirAll(filepath.Dir(stagingRenderedPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", ref.Filename, err)
		}

		if _, err := os.Stat(builderPath); err != nil {
			return fmt.Errorf("builder output not found for %s at %s (run drawio build first)", ref.RelPath, builderPath)
		}

		if err := copyFile(builderPath, stagingRenderedPath); err != nil {
			return fmt.Errorf("copying rendered drawio %s: %w", ref.Filename, err)
		}
		copied++

		cachePathBySource[ref.AbsPath] = stagingRenderedPath
	}

	// Step 3: Update markdown references to point to rendered images
	updated := 0
	for _, path := range p.fileIndex.GetMarkdownFiles() {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		original := string(content)
		modified := p.rewriteDrawioReferences(original, path, cachePathBySource)

		if modified != original {
			if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}
			updated++
		}
	}

	// Step 4: Delete original drawio.png files from staging
	deleted := 0
	for _, ref := range imageRefs {
		if err := os.Remove(ref.AbsPath); err != nil {
			if !os.IsNotExist(err) {
				p.warn("failed to delete %s: %v", ref.Filename, err)
			}
		} else {
			deleted++
		}
	}

	p.log("    Found %d drawio images: %d copied from builder output",
		len(imageRefs), copied)
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

