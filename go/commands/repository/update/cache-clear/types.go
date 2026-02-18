// Package cacheclear provides types and functions for cache clearing operations.
// It integrates with the 2D cache taxonomy (Level x Type) from go/core/cache.
package cacheclear

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ClearMode specifies how to clear a directory.
type ClearMode int

const (
	ClearContents  ClearMode = iota // Delete all contents in the directory
	ClearDocker                     // Execute docker prune command
	ClearSemaphore                  // Delete semaphore state files (capacity coordination)
)

// String returns a string representation of ClearMode.
func (m ClearMode) String() string {
	switch m {
	case ClearContents:
		return "contents"
	case ClearDocker:
		return "docker"
	case ClearSemaphore:
		return "semaphore"
	default:
		return "unknown"
	}
}

// ClearDir represents a directory that can be cleared.
type ClearDir struct {
	RelPath     string      // Relative path from repo root
	Description string      // Human-readable description
	Mode        ClearMode   // How to clear this directory
	Level       cache.Level // Cache level (local, remote)
	Type        cache.Type  // Cache type (state, asset, work, etc.)
}

// CacheTarget combines a ClearDir with its absolute path for clearing.
type CacheTarget struct {
	Dir      ClearDir // Directory definition
	FullPath string   // Absolute path to the directory
}

// Matches returns true if this target matches the given cache spec.
func (t CacheTarget) Matches(spec cache.Spec) bool {
	return spec.Matches(t.Dir.Level, t.Dir.Type)
}

// ClearResult holds the result of a cache clearing operation.
type ClearResult struct {
	DeletedCount int           // Number of items deleted
	DeletedBytes int64         // Total bytes deleted
	Targets      []CacheTarget // Targets that were processed
	DryRun       bool          // Whether this was a dry run
	Errors       []error       // Any errors encountered
}

// HasErrors returns true if any errors occurred during clearing.
func (r ClearResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// GetAllClearDirs returns all known cache directories.
func GetAllClearDirs() []ClearDir {
	return []ClearDir{
		// State: incremental UoW state (build/lint/test/scan state.json files)
		// All state.json files live under .cache/eac/incremental/.
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "incremental"),
			Description: "incremental state (build/lint/test/scan)",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		// State: build acceleration
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "build"),
			Description: "build acceleration cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		// State: UoW manifests in out/ directories (build/test/lint/scan results)
		// These contain input_hash, executed_at, and artifact records that drive
		// incremental change detection. Without clearing these, cache-clear has
		// no effect because the next run still sees valid manifests.
		{
			RelPath:     filepath.Join(paths.OutDir, paths.BuildDir),
			Description: "build output manifests",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, paths.TestDir),
			Description: "test output manifests",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, paths.LintDir),
			Description: "lint output manifests",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, paths.SecurityDir),
			Description: "scan output manifests",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "staging"),
			Description: "staging cache (preprocessed docs with renders)",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "pdf-screenshots"),
			Description: "PDF screenshot cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		// Semaphore files (capacity coordination - can cause test hangs if stale)
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "semaphores"), // Directory containing semaphore files
			Description: "capacity semaphore state",
			Mode:        ClearSemaphore,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},

		// Asset directories (ClearContents mode - delete everything)
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "mermaid"),
			Description: "mermaid acceleration cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "drawio"),
			Description: "drawio acceleration cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "structurizr"),
			Description: "structurizr acceleration cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},

		// Work directories (ClearContents mode - delete everything)
		{
			RelPath:     filepath.Join(paths.EACCacheRoot, "npm"),
			Description: "npm work directories and download cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeWork,
		},

		// Docker caches (ClearDocker mode - uses docker prune commands)
		{
			RelPath:     "", // Not a filesystem path
			Description: "Docker image cache",
			Mode:        ClearDocker,
			Level:       cache.LevelLocal,
			Type:        cache.TypeRegistry,
		},
		{
			RelPath:     "", // Not a filesystem path
			Description: "Docker builder cache",
			Mode:        ClearDocker,
			Level:       cache.LevelLocal,
			Type:        cache.TypeLayer,
		},
	}
}

// ParseTypeFlag parses the --type flag value into cache specs.
// Empty string defaults to DefaultSkipSpecs (state + work).
// Returns multiple specs to support the default case.
func ParseTypeFlag(s string) ([]cache.Spec, error) {
	if s == "" {
		// Default: same as --skip-cache (state + work)
		return cache.DefaultSkipSpecs(), nil
	}
	if s == "all" {
		return []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeAll}}, nil
	}
	return cache.ParseSpecs(s)
}

// FilterTargets filters targets based on cache specs.
// A target is included if it matches ANY of the given specs.
func FilterTargets(targets []CacheTarget, specs []cache.Spec) []CacheTarget {
	var filtered []CacheTarget
	for _, t := range targets {
		for _, spec := range specs {
			if t.Matches(spec) {
				filtered = append(filtered, t)
				break // Only add once
			}
		}
	}
	return filtered
}

// BuildTargets creates CacheTargets from ClearDirs and a repo root.
func BuildTargets(dirs []ClearDir, repoRoot string) []CacheTarget {
	targets := make([]CacheTarget, len(dirs))
	for i, d := range dirs {
		targets[i] = CacheTarget{
			Dir:      d,
			FullPath: filepath.Join(repoRoot, d.RelPath),
		}
	}
	return targets
}

// ClearTargets performs the actual clearing operation on targets.
// Note: Docker clearing (ClearDocker mode) is not handled here - it's handled
// separately in clear.go via clearDockerCache which requires running docker commands.
func ClearTargets(targets []CacheTarget, dryRun, verbose bool) ClearResult {
	result := ClearResult{
		DryRun:  dryRun,
		Targets: targets,
	}

	for _, target := range targets {
		switch target.Dir.Mode {
		case ClearContents:
			count, bytes := clearDirectoryContents(target.FullPath, dryRun)
			result.DeletedCount += count
			result.DeletedBytes += bytes
		case ClearDocker:
			// Docker clearing is handled separately in clear.go
		case ClearSemaphore:
			// Semaphore clearing is handled separately in clear.go
		}
	}

	return result
}

// clearDirectoryContents deletes all entries in a directory.
// File sizes come from DirEntry metadata (no recursive walk for directories).
func clearDirectoryContents(fullPath string, dryRun bool) (int, int64) {
	deleted := 0
	var bytes int64

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return 0, 0
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fullPath, entry.Name())

		// File size from DirEntry metadata (no recursive walk for directories)
		if info, err := entry.Info(); err == nil && !info.IsDir() {
			bytes += info.Size()
		}

		if dryRun {
			deleted++
		} else {
			if err := fileutil.RemoveAllWithRetry(entryPath); err == nil {
				deleted++
			}
		}
	}

	return deleted, bytes
}
