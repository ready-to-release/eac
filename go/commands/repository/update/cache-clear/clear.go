package cacheclear

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

type clearCacheCommand struct{}

var _ core.SimpleCommandPort = (*clearCacheCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&clearCacheCommand{},
	}
}

func (c *clearCacheCommand) Name() string { return "update cache-clear" }

func (c *clearCacheCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-cache-clear",
		Short:         "Clear incremental cache state files",
		Long:          "Removes cache files used by build, lint, test, and scan commands.\nThis forces a full rebuild/retest/relint on the next run.\n\nCache types:\n  state    - Incremental state (build/lint/test state.json, capacity semaphores)\n  asset    - Rendered assets (mermaid, drawio, structurizr caches)\n  work     - Ephemeral work directories (npm work dirs)\n  registry - Docker image cache (runs docker image prune)\n  layer    - Docker builder cache (runs docker builder prune)\n  all      - Everything\n\nThe state cache includes capacity semaphore files (.global-*-capacity.*)\nwhich coordinate parallel test/build execution. Clear these if tests hang.\n\nDefault (no --type): state + work (same as --skip-cache default)\n\nExamples:\n  cache-clear                     Clear state + work caches (default)\n  cache-clear --type=all          Clear all caches including Docker\n  cache-clear --type=state        Clear only state files\n  cache-clear --type=asset        Clear only asset caches\n  cache-clear --type=registry     Clear Docker image cache\n  cache-clear --type=layer        Clear Docker builder cache\n  cache-clear --type=local:state  Clear only local state (fine-grained)\n\nUse --dry-run to see what would be deleted without actually deleting.",
		Flags: []core.FlagSpec{
			{Name: "type", Type: "string", DefaultValue: "state+work", Usage: "Cache type to clear (state, asset, work, registry, layer, all, or level:type)"},
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show what would be deleted without deleting"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show each file being deleted"},
		},
	}
}

func (c *clearCacheCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ClearCache()
}

var log = logging.C()
// CacheSetResult tracks clearing results for a single cacheset (ClearDir).
type CacheSetResult struct {
	Description  string
	DeletedCount int
	DeletedBytes int64
	Items        []string // paths of deleted items
}

// CategoryResult tracks clearing results per cache type.
type CategoryResult struct {
	Type      cache.Type
	CacheSets []CacheSetResult
}

// ClearCache removes cache files based on the --type flag.
func ClearCache() int {
	args := os.Args[3:] // Skip program name, "update", and "cache-clear"
	dryRun := false
	verbose := false
	typeSpec := ""

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--dry-run":
			dryRun = true
		case args[i] == "-v", args[i] == "--verbose":
			verbose = true
		case args[i] == "-h", args[i] == "--help":
			printUsage()
			return 0
		case strings.HasPrefix(args[i], "--type="):
			typeSpec = strings.TrimPrefix(args[i], "--type=")
		case args[i] == "--type" && i+1 < len(args):
			typeSpec = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				log.Errorf("Unknown flag: %s", args[i])
				printUsage()
				return 1
			}
		}
	}

	// Parse the type spec
	specs, err := ParseTypeFlag(typeSpec)
	if err != nil {
		log.Errorf("Invalid cache type: %s", err)
		printUsage()
		return 1
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build and filter targets
	allDirs := GetAllClearDirs()
	targets := BuildTargets(allDirs, repoRoot)
	filtered := FilterTargets(targets, specs)

	if len(filtered) == 0 {
		fmt.Println("No cache directories match the specified type")
		return 0
	}

	// Format specs for display
	specStr := formatSpecs(specs)

	if dryRun {
		fmt.Printf("Dry run - would delete (type=%s):\n", specStr)
	} else {
		fmt.Printf("Clearing cache (type=%s)...\n", specStr)
	}

	// Clear the filtered targets with per-category output
	results := clearTargetsWithCategories(filtered, repoRoot, dryRun, verbose)

	// Print summary
	printCategorySummary(results, dryRun)

	return 0
}

// formatSpecs formats specs for display.
func formatSpecs(specs []cache.Spec) string {
	if len(specs) == 1 {
		return specs[0].String()
	}
	var parts []string
	for _, spec := range specs {
		parts = append(parts, spec.String())
	}
	return strings.Join(parts, ",")
}

// clearTargetsWithCategories clears targets in parallel and groups results by type.
func clearTargetsWithCategories(targets []CacheTarget, repoRoot string, dryRun, verbose bool) []CategoryResult {
	// Clear all targets in parallel
	type targetResult struct {
		count int
		bytes int64
		items []string
	}

	results := make([]targetResult, len(targets))
	var wg sync.WaitGroup

	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t CacheTarget) {
			defer wg.Done()
			switch t.Dir.Mode {
			case ClearContents:
				results[idx].count, results[idx].bytes, results[idx].items = clearDirectoryContentsWithDetails(t.FullPath, t.Dir.RelPath, dryRun, verbose)
			case ClearDocker:
				results[idx].count, results[idx].bytes, results[idx].items = clearDockerCache(t.Dir.Type, dryRun, verbose)
			case ClearSemaphore:
				results[idx].count, results[idx].bytes, results[idx].items = clearSemaphoreFiles(t.FullPath, repoRoot, dryRun, verbose)
			}
		}(i, target)
	}

	wg.Wait()

	// Group by type for display
	byType := make(map[cache.Type][]CacheSetResult)
	for i, target := range targets {
		r := results[i]
		if r.count > 0 || len(r.items) > 0 {
			byType[target.Dir.Type] = append(byType[target.Dir.Type], CacheSetResult{
				Description:  target.Dir.Description,
				DeletedCount: r.count,
				DeletedBytes: r.bytes,
				Items:        r.items,
			})
		}
	}

	var categoryResults []CategoryResult
	typeOrder := []cache.Type{cache.TypeState, cache.TypeWork, cache.TypeAsset, cache.TypeRegistry, cache.TypeLayer}
	for _, cacheType := range typeOrder {
		sets, ok := byType[cacheType]
		if !ok || len(sets) == 0 {
			continue
		}
		categoryResults = append(categoryResults, CategoryResult{Type: cacheType, CacheSets: sets})
	}

	return categoryResults
}

// printCategorySummary prints per-category results and total summary.
func printCategorySummary(results []CategoryResult, dryRun bool) {
	var totalCount int
	var totalBytes int64

	const maxDisplayItems = 5

	for _, r := range results {
		if len(r.CacheSets) == 0 {
			continue
		}
		fmt.Printf("\n  %s:\n", r.Type)
		for _, cs := range r.CacheSets {
			fmt.Printf("    %s:\n", cs.Description)
			displayed := len(cs.Items)
			if displayed > maxDisplayItems {
				displayed = maxDisplayItems
			}
			for _, item := range cs.Items[:displayed] {
				fmt.Printf("      %s\n", item)
			}
			if remaining := len(cs.Items) - displayed; remaining > 0 {
				fmt.Printf("      ... and %d more\n", remaining)
			}
			totalCount += cs.DeletedCount
			totalBytes += cs.DeletedBytes
		}
		fmt.Printf("  → %d %s, %s\n", categoryCount(r), pluralize("item", categoryCount(r)), formatBytes(categoryBytes(r)))
	}

	verb := "deleted"
	if dryRun {
		verb = "would delete"
	}
	fmt.Printf("\nSummary: %d items %s, %s freed\n", totalCount, verb, formatBytes(totalBytes))
}

func categoryCount(r CategoryResult) int {
	n := 0
	for _, cs := range r.CacheSets {
		n += cs.DeletedCount
	}
	return n
}

func categoryBytes(r CategoryResult) int64 {
	var b int64
	for _, cs := range r.CacheSets {
		b += cs.DeletedBytes
	}
	return b
}

// clearDirectoryContentsWithDetails deletes all entries in a directory, returning details.
// File sizes come from DirEntry metadata (no recursive walk). Directory entries
// report only their name — this avoids the expensive subtree traversal that was
// the primary performance bottleneck.
func clearDirectoryContentsWithDetails(fullPath, relPath string, dryRun, verbose bool) (int, int64, []string) {
	var deleted int
	var bytes int64
	var items []string

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if !os.IsNotExist(err) && verbose {
			items = append(items, fmt.Sprintf("[skip] %s (error: %v)", relPath, err))
		}
		return 0, 0, items
	}

	for _, entry := range entries {
		entryPath := filepath.Join(fullPath, entry.Name())
		entryRelPath := filepath.Join(relPath, entry.Name())

		// File size from DirEntry metadata (no recursive walk for directories)
		var entrySize int64
		if info, err := entry.Info(); err == nil && !info.IsDir() {
			entrySize = info.Size()
		}
		bytes += entrySize

		if entrySize > 0 {
			items = append(items, fmt.Sprintf("%s (%s)", entryRelPath, formatBytes(entrySize)))
		} else {
			items = append(items, entryRelPath)
		}

		if !dryRun {
			if err := fileutil.RemoveAllWithRetry(entryPath); err != nil {
				log.Errorf("Failed to delete %s: %v", entryRelPath, err)
				continue
			}
		}
		deleted++
	}

	return deleted, bytes, items
}

// formatBytes formats bytes in human-readable form.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// pluralize returns singular or plural form based on count.
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func printUsage() {
	fmt.Println("Usage: update cache-clear [--type=<spec>] [flags]")
	fmt.Println()
	fmt.Println("Cache types:")
	fmt.Println("  state        Clear incremental state (build/lint/test + capacity semaphores)")
	fmt.Println("  asset        Clear mermaid/drawio/structurizr caches")
	fmt.Println("  work         Clear npm work dirs, preprocessing state")
	fmt.Println("  registry     Clear Docker image cache (runs docker image prune)")
	fmt.Println("  layer        Clear Docker builder cache (runs docker builder prune)")
	fmt.Println("  all          Clear everything")
	fmt.Println("  local:state  Fine-grained: local state only")
	fmt.Println()
	fmt.Println("Default (no --type): state + work (same as --skip-cache default)")
	fmt.Println()
	fmt.Println("The state cache includes capacity semaphore files (.global-*-capacity.*)")
	fmt.Println("which coordinate parallel execution. Clear these if tests/builds hang.")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --type=<spec>  Cache type to clear (default: state,work)")
	fmt.Println("  --dry-run      Show what would be deleted without deleting")
	fmt.Println("  -v, --verbose  Show each file being deleted")
	fmt.Println("  -h, --help     Show this help")
}
