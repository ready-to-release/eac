// Package config provides a central configuration loader for all EAC repository configs.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// containsAny checks if s contains any of the substrings.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// findRepositoryRoot finds the git repository root by walking up directories.
// This is a local implementation to avoid import cycles with the repository package.
func findRepositoryRoot(startPath string) (string, error) {
	// Check for explicit repository root override first
	// This takes precedence over R2R_DOCKER_MODE to allow test isolation
	if repoRoot := os.Getenv("R2R_REPO_ROOT"); repoRoot != "" {
		return filepath.Clean(repoRoot), nil
	}

	// Check for Docker R2R mode - repository is mounted at ContainerRepoRoot
	// Only applies when no explicit override is set
	if os.Getenv("R2R_DOCKER_MODE") == "true" {
		return paths.ContainerRepoRoot, nil
	}

	// Use current directory if no path provided
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up looking for .git directory
	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				return currentPath, nil
			}
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
		}
		currentPath = parentPath
	}
}

// LoadOrNil safely loads config, returning nil on failure instead of propagating errors.
// Useful for commands that need to degrade gracefully when config is unavailable (e.g., tests).
// Disables schema validation for performance since this is often called as a fallback.
// Returns nil if config loading fails OR if the loaded config is unusable (e.g., nil Repository).
func LoadOrNil(repoRoot string) *EACConfig {
	cfg, err := Load(LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: false,
	})
	if err != nil {
		return nil
	}
	// Verify the config is actually usable (Repository must be non-nil)
	// In test environments without contract files, Repository can be nil
	if cfg == nil || cfg.Repository == nil {
		return nil
	}
	return cfg
}

// GetLogsPath returns the logs path for a subsystem.
// Always returns "out/logs/<subsystem>" relative to repoRoot for consistency.
// This is used for debug logs and command diagnostic output.
// Note: LogsPathAbs returns command output paths (out/<command>), not logs.
func GetLogsPath(repoRoot, subsystem string) string {
	return filepath.Join(repoRoot, "out", "logs", subsystem)
}

// GetSpecsPath returns the specs path with graceful fallback to defaults.
// If config loading fails, returns "specs/<moduleName>" relative to repoRoot.
func GetSpecsPath(repoRoot, moduleName string) string {
	cfg := LoadOrNil(repoRoot)
	if cfg == nil {
		return filepath.Join(repoRoot, "specs", moduleName)
	}
	return filepath.Join(repoRoot, cfg.Repository.Paths.SpecsRoot, moduleName)
}

// GetTestOutputPath returns the test output path with graceful fallback to defaults.
// If config loading fails, returns "out/test" relative to repoRoot.
func GetTestOutputPath(repoRoot string) string {
	cfg := LoadOrNil(repoRoot)
	if cfg == nil {
		return filepath.Join(repoRoot, "out", "test")
	}
	return cfg.Repository.TestOutputDirAbs(repoRoot)
}

// GetTestModuleOutputPath returns the test module output path with graceful fallback to defaults.
// If config loading fails, returns "out/test/<moniker>" relative to repoRoot.
func GetTestModuleOutputPath(repoRoot, moniker string) string {
	cfg := LoadOrNil(repoRoot)
	if cfg == nil {
		return filepath.Join(repoRoot, "out", "test", moniker)
	}
	return cfg.Repository.TestModuleDirAbs(repoRoot, moniker)
}
