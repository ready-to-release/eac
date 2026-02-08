package staging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AssetReferencePattern matches asset references in markdown and HTML.
// Captures paths from: ![](path), [](path), src="path", href="path".
//
// Shared by both staging/assetref.go (ScanAssetReferences) and
// cleanup/assets.go (CleanupUnreferencedAssets).
var AssetReferencePattern = regexp.MustCompile(`(?:\]\(|src="|href=")([^)"]+)`)

// ScanAssetReferences scans all markdown files in a directory tree and returns
// a set of all referenced asset paths (normalized relative to the directory).
func ScanAssetReferences(dir string) (map[string]bool, error) {
	refs := make(map[string]bool)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := AssetReferencePattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) <= 1 {
				continue
			}
			ref := match[1]

			if !IsAssetPath(ref) {
				continue
			}

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
func IsAssetPath(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}

	if strings.HasPrefix(path, "#") || strings.HasPrefix(path, "mailto:") {
		return false
	}

	normalized := filepath.ToSlash(path)

	return strings.Contains(normalized, "assets/")
}

// NormalizeAssetPath converts a relative asset reference to a normalized path
// relative to the workspace directory.
func NormalizeAssetPath(ref, mdFilePath, workspaceDir string) string {
	ref = filepath.FromSlash(ref)

	mdDir := filepath.Dir(mdFilePath)

	absPath := filepath.Clean(filepath.Join(mdDir, ref))

	relPath, err := filepath.Rel(workspaceDir, absPath)
	if err != nil {
		return ""
	}

	if strings.HasPrefix(relPath, "..") {
		return ""
	}

	return relPath
}

// ResolveSourceDir extracts the source directory from a glob pattern.
// For patterns like "docs/explanation/**", returns "docs/explanation".
// For patterns like "docs/**/*.md", returns "docs".
// Returns empty string if no directory can be extracted.
func ResolveSourceDir(pattern, workspaceRoot string) string {
	parts := strings.Split(pattern, "**")
	if len(parts) > 0 && parts[0] != "" {
		dir := strings.TrimSuffix(parts[0], "/")
		return filepath.Join(workspaceRoot, dir)
	}
	return ""
}

// IsAssetReferenced checks if a given asset file is in the set of referenced assets.
func IsAssetReferenced(assetPath string, referencedAssets map[string]bool) bool {
	if len(referencedAssets) == 0 {
		return true
	}

	normalized := filepath.ToSlash(assetPath)

	if referencedAssets[assetPath] || referencedAssets[normalized] {
		return true
	}

	baseName := filepath.Base(assetPath)
	for ref := range referencedAssets {
		if filepath.Base(ref) == baseName {
			return true
		}
	}

	return false
}
