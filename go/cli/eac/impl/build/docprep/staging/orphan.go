package staging

import (
	"os"
	"path/filepath"
	"strings"
)

// CleanOrphanedStagedFiles removes files from staging that were not tracked during this build.
// Only removes files with source-type extensions (.md, .png, .jpg, etc.), not generated content.
func CleanOrphanedStagedFiles(stagingDir string, mapper FileMapper, logf func(string, ...any)) error {
	var removed int

	err := filepath.WalkDir(stagingDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Check if this file was tracked during this build's copy phase
		if _, tracked := mapper.GetSourcePath(path); tracked {
			return nil
		}

		// Only clean source file types, not generated content
		if !ShouldCleanOrphan(path) {
			return nil
		}

		if err := os.Remove(path); err != nil {
			logf("    Failed to remove orphan: %s (%v)", path, err)
			return nil
		}

		removed++
		return nil
	})

	if removed > 0 {
		logf("    Removed %d orphaned files from staging", removed)
	}

	CleanEmptyDirs(stagingDir)

	return err
}

// orphanCleanExts lists file extensions that should be cleaned as orphans.
// Note: .svg excluded since mermaid generates SVGs.
var orphanCleanExts = map[string]bool{
	".md":   true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
}

// ShouldCleanOrphan checks if a file type should be cleaned as an orphan.
// Only source file types are cleaned - generated content is preserved.
func ShouldCleanOrphan(path string) bool {
	return orphanCleanExts[strings.ToLower(filepath.Ext(path))]
}
