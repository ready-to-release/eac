package cacheclear

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
)

// semaphoreFiles lists the semaphore files to clear (relative to out directory).
var semaphoreFiles = []string{
	".global-capacity.json",
	".global-capacity.lock",
	".global-docker-capacity.json",
	".global-docker-capacity.lock",
}

// clearSemaphoreFiles deletes capacity semaphore state files.
// These files coordinate resource allocation across concurrent processes.
// Stale semaphore state can cause test hangs and should be cleared when debugging timeouts.
func clearSemaphoreFiles(semaphoreDir, repoRoot string, dryRun, verbose bool) (int, int64, []string) {
	var deleted int
	var bytes int64
	var items []string

	for _, filename := range semaphoreFiles {
		fullPath := filepath.Join(semaphoreDir, filename)
		relPath, _ := filepath.Rel(repoRoot, fullPath)
		if relPath == "" {
			relPath = filepath.Join(paths.EACCacheRoot, "semaphores", filename)
		}

		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			continue // File doesn't exist, nothing to clear
		}
		if err != nil {
			if verbose {
				items = append(items, fmt.Sprintf("[skip] %s (error: %v)", relPath, err))
			}
			continue
		}

		fileSize := info.Size()
		bytes += fileSize
		items = append(items, fmt.Sprintf("%s (%s)", relPath, formatBytes(fileSize)))

		if !dryRun {
			if err := os.Remove(fullPath); err != nil {
				log.Errorf("Failed to delete %s: %v", relPath, err)
				continue
			}
		}
		deleted++
	}

	return deleted, bytes, items
}
