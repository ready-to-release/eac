package books

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// assetRefPattern matches asset references in markdown and HTML.
// Captures paths from: ![](path), [](path), src="path", href="path"
var assetRefPattern = regexp.MustCompile(`(?:\]\(|src="|href=")([^)"]+)`)

// ScanAssetReferences scans all markdown files in a directory tree and returns
// a set of all referenced asset paths (normalized relative to the directory).
// This is used for lazy asset copying - only copy assets that are actually referenced.
func ScanAssetReferences(dir string) (map[string]bool, error) {
	refs := make(map[string]bool)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files/dirs we can't access
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Find all references
		matches := assetRefPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) <= 1 {
				continue
			}
			ref := match[1]

			// Skip external URLs and non-asset references
			if !IsAssetPath(ref) {
				continue
			}

			// Normalize to relative path from directory root
			normalized := NormalizeAssetPath(ref, path, dir)
			if normalized != "" {
				refs[normalized] = true
			}
		}

		return nil
	})

	return refs, err
}

// IsAssetPath checks if a reference path points to an asset (in assets/ directory).
// Returns false for external URLs, anchors, mailto links, and non-asset paths.
func IsAssetPath(path string) bool {
	// Skip external URLs
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}

	// Skip anchors and mailto
	if strings.HasPrefix(path, "#") || strings.HasPrefix(path, "mailto:") {
		return false
	}

	// Normalize path separators
	normalized := filepath.ToSlash(path)

	// Check if path contains "assets/" as a directory component
	// This handles: assets/..., ./assets/..., ../assets/..., ../../assets/...
	// But NOT files named "assets.md" or similar
	if strings.Contains(normalized, "assets/") {
		return true
	}

	// Also check if path starts with "assets/" (not just "assets")
	if strings.HasPrefix(normalized, "assets/") {
		return true
	}

	return false
}

// NormalizeAssetPath converts a relative asset reference to a normalized path
// relative to the workspace directory.
// Example: "../assets/logo/icon.png" from "/workspace/docs/explanation/test.md"
// becomes "assets/logo/icon.png"
func NormalizeAssetPath(ref, mdFilePath, workspaceDir string) string {
	// Convert ref to OS path separators
	ref = filepath.FromSlash(ref)

	// Get the directory containing the markdown file
	mdDir := filepath.Dir(mdFilePath)

	// Resolve the reference relative to markdown location
	absPath := filepath.Clean(filepath.Join(mdDir, ref))

	// Make it relative to workspace directory
	relPath, err := filepath.Rel(workspaceDir, absPath)
	if err != nil {
		return ""
	}

	// Ensure we got a valid relative path (not escaping workspace)
	if strings.HasPrefix(relPath, "..") {
		return ""
	}

	return relPath
}

// IsAssetReferenced checks if a given asset file is in the set of referenced assets.
// The assetPath should be relative to the workspace root.
// The referencedAssets set contains normalized paths relative to the docs directory.
func IsAssetReferenced(assetPath string, referencedAssets map[string]bool) bool {
	if len(referencedAssets) == 0 {
		// No reference scan performed - allow all assets
		return true
	}

	// Normalize path for comparison
	normalized := filepath.ToSlash(assetPath)

	// Direct match
	if referencedAssets[assetPath] || referencedAssets[normalized] {
		return true
	}

	// Check basename match (fallback for edge cases)
	baseName := filepath.Base(assetPath)
	for ref := range referencedAssets {
		if filepath.Base(ref) == baseName {
			return true
		}
	}

	return false
}
