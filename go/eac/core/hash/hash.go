// Package hash provides deterministic file content hashing for change detection.
// It consolidates hashing logic previously duplicated across buildstate, lintstate, and teststate.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
)

// Files computes a SHA-256 hash of file contents.
// It includes filenames in the hash to detect renames.
// Files are processed in sorted order for determinism.
// Returns an empty string for an empty file list.
func Files(workspaceRoot string, files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	h := sha256.New()

	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", file, err)
		}

		// Include filename in hash (detects renames)
		h.Write([]byte(file + "\n"))
		if _, err := io.Copy(h, f); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to read %s: %w", file, err)
		}
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// UncommittedState computes a hash representing uncommitted file state.
// Uses a short hash (16 chars) since this is an optimization hint, not identity.
// Handles deleted files by marking them in the hash.
// Returns an empty string for an empty file list.
func UncommittedState(workspaceRoot string, files []string) string {
	if len(files) == 0 {
		return ""
	}

	h := sha256.New()

	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	for _, file := range sorted {
		path := filepath.Join(workspaceRoot, file)
		f, err := os.Open(path)
		if err != nil {
			// File might be deleted - include that in hash
			h.Write([]byte(file + ":deleted\n"))
			continue
		}
		if _, err := io.Copy(h, f); err != nil {
			h.Write([]byte(file + ":error\n"))
		}
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil))[:16]
}

// ExpandGlobPatterns expands glob patterns against the filesystem.
// Returns relative paths with forward slashes.
// Results are deduplicated and sorted.
// Directories matching patterns are excluded - only files are returned.
func ExpandGlobPatterns(workspaceRoot string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		absPattern := pattern
		if !filepath.IsAbs(pattern) {
			absPattern = filepath.Join(workspaceRoot, pattern)
		}

		// Use doublestar for glob expansion to support ** patterns
		matches, err := doublestar.FilepathGlob(absPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			rel, err := filepath.Rel(workspaceRoot, match)
			if err != nil {
				continue
			}
			// Normalize to forward slashes for cross-platform consistency
			rel = filepath.ToSlash(rel)

			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				// Skip directories and files we can't stat
				continue
			}

			if !seen[rel] {
				seen[rel] = true
				result = append(result, rel)
			}
		}
	}

	sort.Strings(result)
	return result, nil
}
