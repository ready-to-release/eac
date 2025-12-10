// Package testdata provides shared test data preparation functions
// used by get-tests, show-tests, get-suite, and show-suite commands.
package testdata

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRepoRoot finds the repository root by looking for .git directory
func FindRepoRoot(startPath string) (string, error) {
	current := startPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("not in a git repository")
		}
		current = parent
	}
}
