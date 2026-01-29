// Command: update cache-clear
// Short: Clear incremental cache state files
// Long: Removes all incremental cache state files used by build, lint, test, and scan commands.
// Long: This forces a full rebuild/retest/relint on the next run.
// Long:
// Long: Cache files cleared:
// Long:   - out/build/.build-state.json (build incremental state)
// Long:   - out/.lint-state.json (lint incremental state)
// Long:   - out/test/.test-state.json (test incremental state)
// Long:   - out/cache/build-state/*.json (book build cache)
// Long:   - out/cache/preprocess-state/*.json (preprocessing cache)
// Long:
// Long: Use --dry-run to see what would be deleted without actually deleting.
// Flag.dry-run: type=bool, default=false, usage=Show what would be deleted without deleting
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show each file being deleted
package cacheclear

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(ClearCache)
}

// cacheTarget defines a cache file or directory to clear.
type cacheTarget struct {
	path        string // Relative path from repo root
	description string
	isDir       bool // If true, delete entire directory contents
}

// ClearCache removes all incremental cache state files.
func ClearCache() int {
	args := os.Args[3:] // Skip program name, "update", and "cache-clear"
	dryRun := false
	verbose := false

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			printUsage()
			return 0
		default:
			if strings.HasPrefix(arg, "-") {
				log.Errorf("Unknown flag: %s", arg)
				printUsage()
				return 1
			}
		}
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Define all cache targets
	targets := []cacheTarget{
		{filepath.Join(paths.OutDir, paths.BuildDir, ".build-state.json"), "build incremental state", false},
		{filepath.Join(paths.OutDir, ".lint-state.json"), "lint incremental state", false},
		{filepath.Join(paths.OutDir, paths.TestDir, ".test-state.json"), "test incremental state", false},
		{filepath.Join(paths.OutDir, "cache", "build-state"), "book build cache", true},
		{filepath.Join(paths.OutDir, "cache", "preprocess-state"), "preprocessing cache", true},
	}

	if dryRun {
		fmt.Println("Dry run - would delete:")
	} else {
		fmt.Println("Clearing cache...")
	}

	deleted := 0
	for _, target := range targets {
		fullPath := filepath.Join(repoRoot, target.path)

		if target.isDir {
			// Delete directory contents
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				if !os.IsNotExist(err) && verbose {
					fmt.Printf("  [skip] %s (error: %v)\n", target.path, err)
				}
				continue
			}

			for _, entry := range entries {
				entryPath := filepath.Join(fullPath, entry.Name())
				relPath := filepath.Join(target.path, entry.Name())

				if dryRun {
					fmt.Printf("  %s (%s)\n", relPath, target.description)
					deleted++
				} else {
					if err := os.RemoveAll(entryPath); err != nil {
						log.Errorf("Failed to delete %s: %v", relPath, err)
					} else {
						if verbose {
							fmt.Printf("  deleted: %s\n", relPath)
						}
						deleted++
					}
				}
			}
		} else {
			// Delete single file
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if verbose {
					fmt.Printf("  [skip] %s (not found)\n", target.path)
				}
				continue
			}

			if dryRun {
				fmt.Printf("  %s (%s)\n", target.path, target.description)
				deleted++
			} else {
				if err := os.Remove(fullPath); err != nil {
					log.Errorf("Failed to delete %s: %v", target.path, err)
				} else {
					if verbose {
						fmt.Printf("  deleted: %s\n", target.path)
					}
					deleted++
				}
			}
		}
	}

	if dryRun {
		fmt.Printf("\nWould delete %d cache file(s)\n", deleted)
	} else {
		fmt.Printf("\n✓ Cleared %d cache file(s)\n", deleted)
	}

	return 0
}

func printUsage() {
	fmt.Println("Usage: update cache-clear [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --dry-run    Show what would be deleted without deleting")
	fmt.Println("  -v, --verbose  Show each file being deleted")
	fmt.Println("  -h, --help   Show this help")
}
