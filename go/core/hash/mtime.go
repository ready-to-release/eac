// Package hash provides deterministic file content hashing for change detection.
package hash

import (
	"os"
	"path/filepath"
)

// mtimesUnchanged checks if all files in the cached map still have the same mtime.
// Returns false if any file is missing, has different mtime, or stat fails.
func mtimesUnchanged(workspaceRoot string, cached map[string]int64) bool {
	for relPath, cachedNano := range cached {
		info, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(relPath)))
		if err != nil {
			return false
		}
		if info.ModTime().UnixNano() != cachedNano {
			return false
		}
	}
	return true
}

// collectMtimes stats each file, returns map[relativePath]int64 (UnixNano).
// Skips files that can't be stat'd (returns partial map).
func collectMtimes(workspaceRoot string, files []string) map[string]int64 {
	result := make(map[string]int64, len(files))
	for _, relPath := range files {
		info, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(relPath)))
		if err != nil {
			continue
		}
		result[relPath] = info.ModTime().UnixNano()
	}
	return result
}
