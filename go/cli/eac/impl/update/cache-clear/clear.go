// Command: update cache-clear
// Short: Clear incremental cache state files
// Long: Removes cache files used by build, lint, test, and scan commands.
// Long: This forces a full rebuild/retest/relint on the next run.
// Long:
// Long: Cache types:
// Long:   state    - Incremental state (build/lint/test state.json files)
// Long:   asset    - Rendered assets (mermaid, drawio, structurizr caches)
// Long:   work     - Ephemeral work directories (npm work dirs)
// Long:   registry - Docker image cache (runs docker image prune)
// Long:   layer    - Docker builder cache (runs docker builder prune)
// Long:   all      - Everything
// Long:
// Long: Default (no --type): state + work (same as --skip-cache default)
// Long:
// Long: Examples:
// Long:   cache-clear                     Clear state + work caches (default)
// Long:   cache-clear --type=all          Clear all caches including Docker
// Long:   cache-clear --type=state        Clear only state files
// Long:   cache-clear --type=asset        Clear only asset caches
// Long:   cache-clear --type=registry     Clear Docker image cache
// Long:   cache-clear --type=layer        Clear Docker builder cache
// Long:   cache-clear --type=local:state  Clear only local state (fine-grained)
// Long:
// Long: Use --dry-run to see what would be deleted without actually deleting.
// Flag.type: type=string, default=state+work, usage=Cache type to clear (state, asset, work, registry, layer, all, or level:type)
// Flag.dry-run: type=bool, default=false, usage=Show what would be deleted without deleting
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show each file being deleted
package cacheclear

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(ClearCache)
}

// CategoryResult tracks clearing results per cache type.
type CategoryResult struct {
	Type         cache.Type
	DeletedCount int
	DeletedBytes int64
	Items        []string // paths of deleted items
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

// clearTargetsWithCategories clears targets and groups results by category.
func clearTargetsWithCategories(targets []CacheTarget, repoRoot string, dryRun, verbose bool) []CategoryResult {
	// Group targets by type
	byType := make(map[cache.Type][]CacheTarget)
	for _, target := range targets {
		byType[target.Dir.Type] = append(byType[target.Dir.Type], target)
	}

	var results []CategoryResult

	// Process each type in order
	typeOrder := []cache.Type{cache.TypeState, cache.TypeWork, cache.TypeAsset, cache.TypeRegistry, cache.TypeLayer}
	for _, cacheType := range typeOrder {
		typeTargets, ok := byType[cacheType]
		if !ok || len(typeTargets) == 0 {
			continue
		}

		result := CategoryResult{Type: cacheType}

		for _, target := range typeTargets {
			switch target.Dir.Mode {
			case ClearStateFiles:
				count, bytes, items := clearStateFilesWithDetails(target.FullPath, repoRoot, dryRun, verbose)
				result.DeletedCount += count
				result.DeletedBytes += bytes
				result.Items = append(result.Items, items...)
			case ClearContents:
				count, bytes, items := clearDirectoryContentsWithDetails(target.FullPath, target.Dir.RelPath, dryRun, verbose)
				result.DeletedCount += count
				result.DeletedBytes += bytes
				result.Items = append(result.Items, items...)
			case ClearDocker:
				count, bytes, items := clearDockerCache(target.Dir.Type, dryRun, verbose)
				result.DeletedCount += count
				result.DeletedBytes += bytes
				result.Items = append(result.Items, items...)
			}
		}

		if result.DeletedCount > 0 || len(result.Items) > 0 {
			results = append(results, result)
		}
	}

	return results
}

// printCategorySummary prints per-category results and total summary.
func printCategorySummary(results []CategoryResult, dryRun bool) {
	var totalCount int
	var totalBytes int64

	for _, r := range results {
		if r.DeletedCount == 0 && len(r.Items) == 0 {
			continue
		}
		fmt.Printf("\n  %s:\n", r.Type)
		for _, item := range r.Items {
			fmt.Printf("    %s\n", item)
		}
		fmt.Printf("  → %d %s, %s\n", r.DeletedCount,
			pluralize("item", r.DeletedCount), formatBytes(r.DeletedBytes))
		totalCount += r.DeletedCount
		totalBytes += r.DeletedBytes
	}

	verb := "deleted"
	if dryRun {
		verb = "would delete"
	}
	fmt.Printf("\nSummary: %d items %s, %s freed\n", totalCount, verb, formatBytes(totalBytes))
}

// clearStateFilesWithDetails recursively finds and deletes all state.json files, returning details.
func clearStateFilesWithDetails(rootPath, repoRoot string, dryRun, verbose bool) (int, int64, []string) {
	var deleted int
	var bytes int64
	var items []string

	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "state.json" {
			relPath, _ := filepath.Rel(repoRoot, path)
			if relPath == "" {
				relPath = path
			}
			bytes += info.Size()
			items = append(items, fmt.Sprintf("%s (%s)", relPath, formatBytes(info.Size())))

			if !dryRun {
				if err := os.Remove(path); err != nil {
					log.Errorf("Failed to delete %s: %v", relPath, err)
					return nil
				}
			}
			deleted++
		}

		return nil
	})

	return deleted, bytes, items
}

// clearDirectoryContentsWithDetails deletes all entries in a directory, returning details.
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

		// Get size
		var entrySize int64
		info, err := entry.Info()
		if err == nil {
			if info.IsDir() {
				// Walk directory to get total size
				_ = filepath.Walk(entryPath, func(_ string, fi os.FileInfo, _ error) error {
					if fi != nil && !fi.IsDir() {
						entrySize += fi.Size()
					}
					return nil
				})
			} else {
				entrySize = info.Size()
			}
		}

		bytes += entrySize
		items = append(items, fmt.Sprintf("%s (%s)", entryRelPath, formatBytes(entrySize)))

		if !dryRun {
			if err := os.RemoveAll(entryPath); err != nil {
				log.Errorf("Failed to delete %s: %v", entryRelPath, err)
				continue
			}
		}
		deleted++
	}

	return deleted, bytes, items
}

// clearDockerCache clears Docker caches using docker commands.
func clearDockerCache(cacheType cache.Type, dryRun, verbose bool) (int, int64, []string) {
	var cmd string
	var args []string
	var description string

	switch cacheType {
	case cache.TypeRegistry:
		cmd = "docker"
		args = []string{"image", "prune", "-f"}
		description = "Docker image cache"
	case cache.TypeLayer:
		cmd = "docker"
		args = []string{"builder", "prune", "-f"}
		description = "Docker builder cache"
	default:
		return 0, 0, nil
	}

	if dryRun {
		return 1, 0, []string{fmt.Sprintf("Would run: %s %s (%s)", cmd, strings.Join(args, " "), description)}
	}

	// Execute docker command
	execCmd := exec.Command(cmd, args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return 0, 0, []string{fmt.Sprintf("Failed to run %s: %v", cmd, err)}
	}

	// Try to parse reclaimed space from output
	// Docker outputs: "Total reclaimed space: 1.234GB"
	var reclaimedBytes int64
	outputStr := string(output)
	if strings.Contains(outputStr, "reclaimed space:") {
		// Parse the output (this is best-effort)
		reclaimedBytes = parseDockerReclaimedSpace(outputStr)
	}

	return 1, reclaimedBytes, []string{fmt.Sprintf("%s %s (%s)", cmd, strings.Join(args, " "), description)}
}

// parseDockerReclaimedSpace tries to parse reclaimed space from docker prune output.
func parseDockerReclaimedSpace(output string) int64 {
	// Docker outputs: "Total reclaimed space: 1.234GB" or similar
	// This is best-effort parsing
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "reclaimed space:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				sizeStr := strings.TrimSpace(parts[len(parts)-1])
				return parseSizeString(sizeStr)
			}
		}
	}
	return 0
}

// parseSizeString parses a size string like "1.234GB" into bytes.
func parseSizeString(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0B" {
		return 0
	}

	var multiplier float64 = 1
	suffix := ""

	// Extract numeric part and suffix
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] >= '0' && s[i] <= '9' || s[i] == '.' {
			suffix = s[i+1:]
			s = s[:i+1]
			break
		}
	}

	switch strings.ToUpper(suffix) {
	case "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	var value float64
	fmt.Sscanf(s, "%f", &value)
	return int64(value * multiplier)
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
	fmt.Println("  state        Clear incremental state (build/lint/test)")
	fmt.Println("  asset        Clear mermaid/drawio/structurizr caches")
	fmt.Println("  work         Clear npm work dirs, preprocessing state")
	fmt.Println("  registry     Clear Docker image cache (runs docker image prune)")
	fmt.Println("  layer        Clear Docker builder cache (runs docker builder prune)")
	fmt.Println("  all          Clear everything")
	fmt.Println("  local:state  Fine-grained: local state only")
	fmt.Println()
	fmt.Println("Default (no --type): state + work (same as --skip-cache default)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --type=<spec>  Cache type to clear (default: state,work)")
	fmt.Println("  --dry-run      Show what would be deleted without deleting")
	fmt.Println("  -v, --verbose  Show each file being deleted")
	fmt.Println("  -h, --help     Show this help")
}
