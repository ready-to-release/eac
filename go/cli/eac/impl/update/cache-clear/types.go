// Package cacheclear provides types and functions for cache clearing operations.
// It integrates with the 2D cache taxonomy (Level x Type) from go/core/cache.
package cacheclear

import (
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/paths"
)

// ClearMode specifies how to clear a directory.
type ClearMode int

const (
	ClearStateFiles ClearMode = iota // Only delete state.json files (recursive)
	ClearContents                    // Delete all contents in the directory
	ClearDocker                      // Execute docker prune command
	ClearSemaphore                   // Delete semaphore state files (capacity coordination)
)

// String returns a string representation of ClearMode.
func (m ClearMode) String() string {
	switch m {
	case ClearStateFiles:
		return "state-files"
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
		// State directories (ClearStateFiles mode - only delete state.json)
		{
			RelPath:     filepath.Join(paths.OutDir, paths.BuildDir),
			Description: "build state",
			Mode:        ClearStateFiles,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, paths.LintDir),
			Description: "lint state",
			Mode:        ClearStateFiles,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, paths.TestDir),
			Description: "test state",
			Mode:        ClearStateFiles,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, "cache", "build-state"),
			Description: "book build cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		{
			RelPath:     filepath.Join(paths.OutDir, "cache", "preprocess-state"),
			Description: "preprocessing cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},
		// Semaphore files (capacity coordination - can cause test hangs if stale)
		{
			RelPath:     paths.OutDir, // Directory containing semaphore files
			Description: "capacity semaphore state",
			Mode:        ClearSemaphore,
			Level:       cache.LevelLocal,
			Type:        cache.TypeState,
		},

		// Asset directories (ClearContents mode - delete everything)
		{
			RelPath:     filepath.Join("docs", "assets", "cache", "mermaid"),
			Description: "mermaid cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		{
			RelPath:     filepath.Join("docs", "assets", "cache", "drawio"),
			Description: "drawio cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},
		{
			RelPath:     filepath.Join("docs", "assets", "cache", "structurizr"),
			Description: "structurizr cache",
			Mode:        ClearContents,
			Level:       cache.LevelLocal,
			Type:        cache.TypeAsset,
		},

		// Work directories (ClearContents mode - delete everything)
		{
			RelPath:     filepath.Join(".cache", "npm", "work"),
			Description: "npm work directories",
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
		return cache.DefaultSkipSpecs, nil
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
		case ClearStateFiles:
			count, bytes := clearStateFilesInDir(target.FullPath, dryRun)
			result.DeletedCount += count
			result.DeletedBytes += bytes
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

// clearStateFilesInDir recursively finds and deletes all state.json files.
func clearStateFilesInDir(rootPath string, dryRun bool) (int, int64) {
	deleted := 0
	var bytes int64

	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "state.json" {
			bytes += info.Size()
			if dryRun {
				deleted++
			} else {
				if err := os.Remove(path); err == nil {
					deleted++
				}
			}
		}
		return nil
	})

	return deleted, bytes
}

// clearDirectoryContents deletes all entries in a directory.
func clearDirectoryContents(fullPath string, dryRun bool) (int, int64) {
	deleted := 0
	var bytes int64

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return 0, 0
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fullPath, entry.Name())

		// Get file info for size
		info, err := entry.Info()
		if err == nil {
			if info.IsDir() {
				// For directories, walk to get total size
				_ = filepath.Walk(entryPath, func(_ string, fi os.FileInfo, _ error) error {
					if fi != nil && !fi.IsDir() {
						bytes += fi.Size()
					}
					return nil
				})
			} else {
				bytes += info.Size()
			}
		}

		if dryRun {
			deleted++
		} else {
			if err := os.RemoveAll(entryPath); err == nil {
				deleted++
			}
		}
	}

	return deleted, bytes
}
