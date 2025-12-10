package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// assetExtensions are file extensions considered assets (images, PDFs, etc.)
var assetExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".ico":  true,
	".pdf":  true,
}

// assetReferencePattern matches asset references in markdown and HTML
// Captures paths from: ![](path), [](path), src="path", href="path"
var assetReferencePattern = regexp.MustCompile(`(?:\]\(|src="|href=")([^)"]+)`)

// cleanupUnreferencedAssets removes asset files from staging that are not referenced by any markdown
// This runs after all preprocessing to ensure only necessary files are included in the final output
func (p *Preprocessor) cleanupUnreferencedAssets() error {
	// Step 1: Collect all asset references from markdown files
	// We track both absolute paths and basenames for robust matching
	referencedPaths := make(map[string]bool) // normalized absolute paths
	referencedNames := make(map[string]bool) // basenames (fallback matching)

	err := filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories/files we can't access
			return nil
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files we can't read
			return nil
		}

		// Find all asset references
		refs := assetReferencePattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range refs {
			if len(match) > 1 {
				ref := match[1]
				// Skip external URLs and anchors
				if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") ||
					strings.HasPrefix(ref, "#") || strings.HasPrefix(ref, "mailto:") {
					continue
				}

				// Normalize path separators (markdown uses forward slashes)
				ref = filepath.FromSlash(ref)

				// Resolve to absolute path from markdown file location
				mdDir := filepath.Dir(path)
				absPath := filepath.Clean(filepath.Join(mdDir, ref))

				// Store normalized lowercase path for case-insensitive matching on Windows
				referencedPaths[strings.ToLower(absPath)] = true

				// Also store basename for fallback matching
				referencedNames[strings.ToLower(filepath.Base(ref))] = true
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Step 2: Find and delete unreferenced asset files
	deleted := 0
	var deletedSize int64

	err = filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories/files we can't access
			return nil
		}
		if d.IsDir() {
			return nil
		}

		// Check if this is an asset file
		ext := strings.ToLower(filepath.Ext(path))
		if !assetExtensions[ext] {
			return nil
		}

		// Check if it's referenced (by full path or basename)
		baseName := strings.ToLower(filepath.Base(path))
		normalizedPath := strings.ToLower(path)
		if referencedPaths[normalizedPath] || referencedNames[baseName] {
			return nil
		}

		// Get file size before deletion for reporting
		info, err := d.Info()
		if err == nil {
			deletedSize += info.Size()
		}

		// Delete unreferenced asset
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				p.log("    Warning: failed to delete unreferenced asset %s: %v",
					filepath.Base(path), err)
			}
		} else {
			deleted++
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Step 3: Prune empty directories (bottom-up)
	pruned := 0
	for {
		prunedThisPass := 0
		err = filepath.WalkDir(p.stagingDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// Skip directories we can't access
				return nil
			}
			if !d.IsDir() || path == p.stagingDir {
				return nil
			}

			// Check if directory is empty
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil // Skip on error
			}
			if len(entries) == 0 {
				if err := os.Remove(path); err == nil {
					prunedThisPass++
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		if prunedThisPass == 0 {
			break
		}
		pruned += prunedThisPass
	}

	if deleted > 0 || pruned > 0 {
		p.log("    Deleted %d unreferenced assets (%.1f KB), pruned %d empty directories",
			deleted, float64(deletedSize)/1024, pruned)
	} else {
		p.log("    No unreferenced assets or empty directories found")
	}

	return nil
}
