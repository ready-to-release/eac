// Package config provides a central configuration loader for all EAC repository configs.
package config

import (
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
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
	return paths.TestOutputDir(repoRoot)
}

// GetTestModuleOutputPath returns the test module output path with graceful fallback to defaults.
// If config loading fails, returns "out/test/<moniker>" relative to repoRoot.
func GetTestModuleOutputPath(repoRoot, moniker string) string {
	cfg := LoadOrNil(repoRoot)
	if cfg == nil {
		return filepath.Join(repoRoot, "out", "test", moniker)
	}
	return paths.TestModuleDir(repoRoot, moniker)
}
