package cleanup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/build/docprep/staging"
)

// AssetExtensions are file extensions considered assets (images, PDFs, etc.)
var AssetExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".ico":  true,
	".pdf":  true,
}

// assetReferencePattern is the shared regex for matching asset references in markdown/HTML.
// Defined in staging.AssetReferencePattern to avoid import cycles.
var assetReferencePattern = staging.AssetReferencePattern

// CleanupUnreferencedAssets removes asset files from staging that are not
// referenced by any markdown. Runs after all preprocessing to ensure only
// necessary files are included in the final output.
func CleanupUnreferencedAssets(stagingDir string, logf func(string, ...any), warnf func(string, ...any)) error {
	// Step 1: Collect all asset references from markdown files
	referencedPaths := make(map[string]bool) // normalized absolute paths
	referencedNames := make(map[string]bool) // basenames (fallback matching)

	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		refs := assetReferencePattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range refs {
			if len(match) <= 1 {
				continue
			}
			ref := match[1]
			if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") ||
				strings.HasPrefix(ref, "#") || strings.HasPrefix(ref, "mailto:") {
				continue
			}

			ref = filepath.FromSlash(ref)
			mdDir := filepath.Dir(path)
			absPath := filepath.Clean(filepath.Join(mdDir, ref))

			referencedPaths[strings.ToLower(absPath)] = true
			referencedNames[strings.ToLower(filepath.Base(ref))] = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Step 2: Find and delete unreferenced asset files
	deleted := 0
	var deletedSize int64

	err = filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !AssetExtensions[ext] {
			return nil
		}

		baseName := strings.ToLower(filepath.Base(path))
		normalizedPath := strings.ToLower(path)
		if referencedPaths[normalizedPath] || referencedNames[baseName] {
			return nil
		}

		info, err := d.Info()
		if err == nil {
			deletedSize += info.Size()
		}

		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				warnf("failed to delete unreferenced asset %s: %v",
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
		err = filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() || path == stagingDir {
				return nil
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
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
		logf("    Deleted %d unreferenced assets (%.1f KB), pruned %d empty directories",
			deleted, float64(deletedSize)/1024, pruned)
	} else {
		logf("    No unreferenced assets or empty directories found")
	}

	return nil
}
