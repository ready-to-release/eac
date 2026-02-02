package books

import (
	"io"
	"os"
	"path/filepath"
)

// MirrorStats tracks statistics from a mirror sync operation.
type MirrorStats struct {
	Copied  int // Files copied (new or changed)
	Skipped int // Files skipped (unchanged)
	Deleted int // Files deleted (orphans)
}

// MirrorSync synchronizes src directory to dst directory using mtime/size comparison.
// - Copies files that are new or changed (different size or mtime)
// - Deletes files in dst that don't exist in src (orphans)
// - Preserves modification times for accurate future comparisons
//
// This is similar to "robocopy /MIR" or "rsync -a --delete".
func MirrorSync(src, dst string) (MirrorStats, error) {
	var stats MirrorStats

	// Build map of existing dst files for orphan detection
	dstFiles := make(map[string]bool)
	_ = filepath.WalkDir(dst, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(dst, path)
		dstFiles[relPath] = true
		return nil
	})

	// Walk source and sync to dst
	err := filepath.WalkDir(src, func(srcPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, srcPath)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			// Ensure directory exists in dst
			return os.MkdirAll(dstPath, 0o755)
		}

		// Mark as seen (not an orphan)
		delete(dstFiles, relPath)

		// Check if file needs copying
		srcInfo, err := d.Info()
		if err != nil {
			return err
		}

		if !needsCopy(srcInfo, dstPath) {
			stats.Skipped++
			return nil
		}

		// Copy file
		if err := copyFilePreserveMtime(srcPath, dstPath, srcInfo); err != nil {
			return err
		}
		stats.Copied++

		return nil
	})

	if err != nil {
		return stats, err
	}

	// Delete orphaned files (exist in dst but not in src)
	for orphan := range dstFiles {
		orphanPath := filepath.Join(dst, orphan)
		if err := os.Remove(orphanPath); err == nil {
			stats.Deleted++
		}
	}

	// Clean up empty directories
	cleanEmptyDirs(dst)

	return stats, nil
}

// needsCopy checks if srcInfo differs from the file at dstPath.
// Returns true if dst doesn't exist or has different size/mtime.
func needsCopy(srcInfo os.FileInfo, dstPath string) bool {
	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		return true // dst doesn't exist
	}

	// Compare size first (fast)
	if srcInfo.Size() != dstInfo.Size() {
		return true
	}

	// Compare mtime (truncate to second precision for cross-platform compatibility)
	srcMtime := srcInfo.ModTime().Truncate(1e9)
	dstMtime := dstInfo.ModTime().Truncate(1e9)
	return !srcMtime.Equal(dstMtime)
}

// copyFilePreserveMtime copies a file and preserves its modification time.
func copyFilePreserveMtime(src, dst string, srcInfo os.FileInfo) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Open source
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Close before setting mtime
	dstFile.Close()

	// Preserve modification time
	mtime := srcInfo.ModTime()
	return os.Chtimes(dst, mtime, mtime)
}

// cleanEmptyDirs removes empty directories from the tree (bottom-up).
func cleanEmptyDirs(root string) {
	// Walk in reverse order (deepest first) by collecting then sorting
	var dirs []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})

	// Remove empty directories (deepest first = reverse order)
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err == nil && len(entries) == 0 {
			os.Remove(dirs[i])
		}
	}
}
