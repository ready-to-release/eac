package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/tool"
)

// downloadArtifacts downloads artifacts matching a pattern from a workflow run
// Returns the list of downloaded artifact names.
func downloadArtifacts(runID int, pattern, destDir, workspaceRoot string) ([]string, error) {
	// Use gh run download with pattern matching
	output, exitCode, err := tool.GlobalToolSystem().RunToolCombined(context.Background(), "gh", workspaceRoot, "run", "download",
		fmt.Sprintf("%d", runID),
		"--pattern", pattern,
		"--dir", destDir,
	)
	if err != nil {
		return nil, fmt.Errorf("gh execution failed: %w", err)
	}
	if exitCode != 0 {
		// Check if it's just "no artifacts found" (exit code 1 with specific message)
		outputStr := string(output)
		if strings.Contains(outputStr, "no artifact") || strings.Contains(outputStr, "no matching") {
			return nil, nil // No artifacts is not an error
		}
		return nil, fmt.Errorf("%s", strings.TrimSpace(outputStr))
	}

	// Parse output to get downloaded artifact names
	// gh run download outputs lines like "Downloading artifact-name..."
	var downloaded []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Downloading ") {
			name := strings.TrimSuffix(strings.TrimPrefix(line, "Downloading "), "...")
			if name != "" {
				downloaded = append(downloaded, name)
			}
		}
	}

	// If we got no output but no error, assume success
	if len(downloaded) == 0 && err == nil {
		// Try to detect what was downloaded by listing the directory
		entries, readErr := os.ReadDir(destDir)
		if readErr == nil {
			for _, e := range entries {
				if e.IsDir() && strings.Contains(e.Name(), strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "-")) {
					downloaded = append(downloaded, e.Name())
				}
			}
		}
	}

	// Flatten directory structure: gh run download creates artifact-name/module/
	// but evidence loader expects just module/ directly under destDir
	if err := flattenArtifactDirs(destDir); err != nil {
		return downloaded, fmt.Errorf("failed to flatten directories: %w", err)
	}

	return downloaded, nil
}

// flattenArtifactDirs flattens the artifact directory structure.
// gh run download creates: destDir/artifact-name/module/files
// We need: destDir/module/files.
func flattenArtifactDirs(destDir string) error {
	entries, err := os.ReadDir(destDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		artifactDir := filepath.Join(destDir, entry.Name())

		// Check if this looks like an artifact directory (test-results-* or scan-results-*)
		if !strings.HasPrefix(entry.Name(), "test-results-") && !strings.HasPrefix(entry.Name(), "scan-results-") {
			continue
		}

		// Move contents up one level
		subEntries, err := os.ReadDir(artifactDir)
		if err != nil {
			continue
		}

		for _, subEntry := range subEntries {
			srcPath := filepath.Join(artifactDir, subEntry.Name())
			dstPath := filepath.Join(destDir, subEntry.Name())

			// If destination exists, merge directories
			if _, err := os.Stat(dstPath); err == nil {
				if subEntry.IsDir() {
					// Merge directory contents
					if err := mergeDirectories(srcPath, dstPath); err != nil {
						return fmt.Errorf("failed to merge %s: %w", subEntry.Name(), err)
					}
					os.RemoveAll(srcPath)
				}
				// Skip files that already exist
				continue
			}

			// Move to destination
			if err := os.Rename(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move %s: %w", subEntry.Name(), err)
			}
		}

		// Remove empty artifact directory
		os.Remove(artifactDir)
	}

	return nil
}

// mergeDirectories recursively merges src into dst.
func mergeDirectories(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Create destination dir if needed
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			// Recurse
			if err := mergeDirectories(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file if it doesn't exist
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				if err := os.Rename(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
