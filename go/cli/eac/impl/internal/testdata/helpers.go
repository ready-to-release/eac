// Package testdata provides shared test data preparation functions
// used by get-tests, show-tests, get-suite, and show-suite commands.
package testdata

import (
	"github.com/ready-to-release/eac/go/core/environments"
	"fmt"
	"os"
	"path/filepath"
)

// FindRepoRoot finds the repository root by looking for .git directory.
// In isolated test environments, respects CLIE_REPO_ROOT environment variable.
func FindRepoRoot(startPath string) (string, error) {
	// Check if running in isolated test environment
	// Trust this environment variable when set - it's explicitly configured by the test framework
	if repoRoot := os.Getenv(environments.EnvCLIERepoRoot); repoRoot != "" {
		return repoRoot, nil
	}

	// Normal mode: walk up looking for .git
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
