package docsmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File names for manifest system
const (
	DescriptionsFileName   = "descriptions.yml"
	CacheFileName          = ".manifest-cache.json"
	LegacyManifestFileName = "manifest.json"
)

// ScanAssets discovers all documentation assets in the given directory.
// It returns a map of relative paths (with forward slashes) to DiscoveredAsset.
func ScanAssets(assetsDir string) (map[string]DiscoveredAsset, error) {
	assets := make(map[string]DiscoveredAsset)

	err := filepath.WalkDir(assetsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip hidden directories and special directories
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "cache" || name == "rendered" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches any of our patterns
		assetType := determineAssetType(d.Name())
		if assetType == "" {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Calculate relative path with forward slashes
		relPath, err := filepath.Rel(assetsDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		// Skip manifest files and descriptions
		if relPath == LegacyManifestFileName || relPath == DescriptionsFileName || relPath == CacheFileName {
			return nil
		}

		// Determine category from first directory component
		category := ""
		if idx := strings.Index(relPath, "/"); idx > 0 {
			category = relPath[:idx]
		}

		// Calculate content hash
		hash, err := hashFile(path)
		if err != nil {
			// Non-fatal: continue without hash
			hash = ""
		}

		assets[relPath] = DiscoveredAsset{
			RelPath:      relPath,
			AbsPath:      path,
			Category:     category,
			Type:         assetType,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			Hash:         hash,
		}

		return nil
	})

	return assets, err
}

// determineAssetType classifies an asset based on its filename.
// Returns empty string if file should not be tracked.
func determineAssetType(name string) string {
	nameLower := strings.ToLower(name)

	// Check for .drawio.png first (most specific)
	if strings.HasSuffix(nameLower, ".drawio.png") {
		return "drawio"
	}

	// Check for .drawio source files
	if strings.HasSuffix(nameLower, ".drawio") {
		return "drawio-source"
	}

	// Check for regular PNG images (not drawio)
	if strings.HasSuffix(nameLower, ".png") {
		return "image"
	}

	// Not a tracked file type
	return ""
}

// hashFile computes the SHA-256 hash of a file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
