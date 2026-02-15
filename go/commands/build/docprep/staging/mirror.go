package staging

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
// Copies files that are new or changed, deletes files in dst that don't exist in src,
// and preserves modification times for accurate future comparisons.
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
			return os.MkdirAll(dstPath, 0o755)
		}

		// Mark as seen (not an orphan)
		delete(dstFiles, relPath)

		srcInfo, err := d.Info()
		if err != nil {
			return err
		}

		if !NeedsCopy(srcInfo, dstPath) {
			stats.Skipped++
			return nil
		}

		if err := CopyFilePreserveMtime(srcPath, dstPath, srcInfo); err != nil {
			return err
		}
		stats.Copied++

		return nil
	})

	if err != nil {
		return stats, err
	}

	// Delete orphaned files
	for orphan := range dstFiles {
		orphanPath := filepath.Join(dst, orphan)
		if err := os.Remove(orphanPath); err == nil {
			stats.Deleted++
		}
	}

	CleanEmptyDirs(dst)

	return stats, nil
}

// NeedsCopy checks if srcInfo differs from the file at dstPath.
// Returns true if dst doesn't exist or has different size/mtime.
func NeedsCopy(srcInfo os.FileInfo, dstPath string) bool {
	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		return true // dst doesn't exist
	}

	if srcInfo.Size() != dstInfo.Size() {
		return true
	}

	// Compare mtime (truncate to second precision for cross-platform compatibility)
	srcMtime := srcInfo.ModTime().Truncate(1e9)
	dstMtime := dstInfo.ModTime().Truncate(1e9)
	return !srcMtime.Equal(dstMtime)
}

// CopyFilePreserveMtime copies a file and preserves its modification time.
func CopyFilePreserveMtime(src, dst string, srcInfo os.FileInfo) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	dstFile.Close()

	mtime := srcInfo.ModTime()
	return os.Chtimes(dst, mtime, mtime)
}

// CopyFile copies a file from src to dst, creating directories as needed.
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CleanEmptyDirs removes empty directories from the tree (bottom-up).
func CleanEmptyDirs(root string) {
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
