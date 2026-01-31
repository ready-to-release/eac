// Command: update cache-clear
// Short: Clear incremental cache state files
// Long: Removes all incremental cache state files used by build, lint, test, and scan commands.
// Long: This forces a full rebuild/retest/relint on the next run.
// Long:
// Long: Cache files cleared:
// Long:   - out/build/**/state.json (per-module build state)
// Long:   - out/lint/**/state.json (per-module lint state)
// Long:   - out/test/**/state.json (per-module test state)
// Long:   - out/cache/build-state/* (book build cache)
// Long:   - out/cache/preprocess-state/* (preprocessing cache)
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

	if dryRun {
		fmt.Println("Dry run - would delete:")
	} else {
		fmt.Println("Clearing cache...")
	}

	deleted := 0

	// Clear per-module state files (state.json) in build/lint/test directories
	contextDirs := []struct {
		dir         string
		description string
	}{
		{filepath.Join(paths.OutDir, paths.BuildDir), "build state"},
		{filepath.Join(paths.OutDir, paths.LintDir), "lint state"},
		{filepath.Join(paths.OutDir, paths.TestDir), "test state"},
	}

	for _, ctx := range contextDirs {
		fullPath := filepath.Join(repoRoot, ctx.dir)
		count := clearStateFiles(fullPath, ctx.dir, ctx.description, dryRun, verbose)
		deleted += count
	}

	// Clear book cache directories (delete entire contents)
	bookCacheDirs := []struct {
		dir         string
		description string
	}{
		{filepath.Join(paths.OutDir, "cache", "build-state"), "book build cache"},
		{filepath.Join(paths.OutDir, "cache", "preprocess-state"), "preprocessing cache"},
		{filepath.Join(".cache", "npm", "work"), "npm isolated work directories"},
	}

	for _, cache := range bookCacheDirs {
		fullPath := filepath.Join(repoRoot, cache.dir)
		count := clearDirectoryContents(fullPath, cache.dir, cache.description, dryRun, verbose)
		deleted += count
	}

	if dryRun {
		fmt.Printf("\nWould delete %d cache file(s)\n", deleted)
	} else {
		fmt.Printf("\n✓ Cleared %d cache file(s)\n", deleted)
	}

	return 0
}

// clearStateFiles recursively finds and deletes all state.json files in a directory.
func clearStateFiles(rootPath, relRoot, description string, dryRun, verbose bool) int {
	deleted := 0

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "state.json" {
			relPath, _ := filepath.Rel(filepath.Dir(rootPath), path)
			if relPath == "" {
				relPath = path
			}

			if dryRun {
				fmt.Printf("  %s (%s)\n", relPath, description)
				deleted++
			} else {
				if err := os.Remove(path); err != nil {
					log.Errorf("Failed to delete %s: %v", relPath, err)
				} else {
					if verbose {
						fmt.Printf("  deleted: %s\n", relPath)
					}
					deleted++
				}
			}
		}

		return nil
	})

	if err != nil && verbose {
		fmt.Printf("  [skip] %s (error: %v)\n", relRoot, err)
	}

	return deleted
}

// clearDirectoryContents deletes all entries in a directory.
func clearDirectoryContents(fullPath, relPath, description string, dryRun, verbose bool) int {
	deleted := 0

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if !os.IsNotExist(err) && verbose {
			fmt.Printf("  [skip] %s (error: %v)\n", relPath, err)
		}
		return 0
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fullPath, entry.Name())
		entryRelPath := filepath.Join(relPath, entry.Name())

		if dryRun {
			fmt.Printf("  %s (%s)\n", entryRelPath, description)
			deleted++
		} else {
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("Failed to delete %s: %v", entryRelPath, err)
			} else {
				if verbose {
					fmt.Printf("  deleted: %s\n", entryRelPath)
				}
				deleted++
			}
		}
	}

	return deleted
}

func printUsage() {
	fmt.Println("Usage: update cache-clear [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --dry-run    Show what would be deleted without deleting")
	fmt.Println("  -v, --verbose  Show each file being deleted")
	fmt.Println("  -h, --help   Show this help")
}
