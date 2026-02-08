// Command: update docs
// Short: Sync documentation assets and command references
// Long: Updates documentation by:
// Long:   - Rendering mermaid diagrams to cache
// Long:   - Optimizing drawio images to cache
// Long:   - Syncing command reference docs (creates missing, removes orphans)
// Long:
// Long: Use --area to run specific areas (for testing).
// Long: Use --dry-run to preview changes without applying them.
// Long:
// Long: Cache Pruning:
// Long:   Use --prune-cache to remove orphaned mermaid/drawio cache files.
// Flag.area: type=string, default="", usage=Area to update (mermaid, drawio, command-refs, all). Repeatable.
// Flag.dry-run: type=bool, default=false, usage=Preview changes without applying them
// Flag.force: type=bool, default=false, usage=Force re-render/re-optimize all assets even if cached
// Flag.verbose: type=bool, shorthand=v, default=false, usage=Show detailed progress
// Flag.prune-cache: type=bool, default=false, usage=Remove orphaned cache files (mermaid/drawio)
package docs

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

var log = logging.C()

func init() {
	registry.Register(UpdateDocs)
}

// UpdateDocs syncs documentation assets and command references.
func UpdateDocs() int {
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	opts, err := parseFlags(args)
	if err != nil {
		log.Errorf("%v", err)
		return 1
	}

	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Handle --prune-cache mode (separate from normal update)
	if opts.PruneCache {
		return runPruneCache(repoRoot, opts.Verbose, opts.DryRun)
	}

	logWriter := os.Stdout
	exitCode := 0

	// Run selected areas
	if opts.Areas.Includes(AreaMermaid) {
		result, err := runMermaidUpdate(repoRoot, opts, logWriter)
		if err != nil {
			log.Errorf("Mermaid error: %v", err)
			exitCode = 1
		} else if result.Failed > 0 {
			exitCode = 1
		}
		fmt.Println()
	}

	if opts.Areas.Includes(AreaDrawio) {
		result, err := runDrawioUpdate(repoRoot, opts, logWriter)
		if err != nil {
			log.Errorf("Drawio error: %v", err)
			exitCode = 1
		} else if result.Failed > 0 {
			exitCode = 1
		}
		fmt.Println()
	}

	if opts.Areas.Includes(AreaCommandRefs) {
		_, err := runCommandRefsUpdate(repoRoot, opts, logWriter)
		if err != nil {
			log.Errorf("Command refs error: %v", err)
			exitCode = 1
		}
		fmt.Println()
	}

	// Final summary
	if exitCode == 0 && !opts.DryRun {
		cacheDir := paths.DocsCachePath(repoRoot)
		fmt.Printf("Cache: %s\n", cacheDir)
	}

	return exitCode
}

// parseFlags parses command line flags into UpdateOptions.
func parseFlags(args []string) (UpdateOptions, error) {
	opts := UpdateOptions{
		Areas: NewAreaSet(nil), // Default to all areas
	}

	var areas []Area

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--dry-run":
			opts.DryRun = true
		case arg == "--force":
			opts.Force = true
		case arg == "-v" || arg == "--verbose":
			opts.Verbose = true
		case arg == "--prune-cache":
			opts.PruneCache = true
		case arg == "--area":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--area requires a value")
			}
			i++
			area, err := ParseArea(args[i])
			if err != nil {
				return opts, err
			}
			areas = append(areas, area)
		case strings.HasPrefix(arg, "--area="):
			value := strings.TrimPrefix(arg, "--area=")
			area, err := ParseArea(value)
			if err != nil {
				return opts, err
			}
			areas = append(areas, area)
		}
	}

	if len(areas) > 0 {
		opts.Areas = NewAreaSet(areas)
	}

	return opts, nil
}

// runPruneCache handles the --prune-cache mode.
func runPruneCache(repoRoot string, verbose, dryRun bool) int {
	fmt.Println("Analyzing cache for orphaned files...")

	result, err := PruneCache(repoRoot, verbose)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Println()
	fmt.Println("Mermaid cache:")
	fmt.Printf("  Files: %d, Active: %d, Orphans: %d (%s)\n",
		result.MermaidCacheFiles, result.MermaidActiveHashes,
		len(result.MermaidOrphans), formatBytes(result.MermaidBytesRecovered))

	fmt.Println()
	fmt.Println("Drawio cache:")
	fmt.Printf("  Files: %d, Active: %d, Orphans: %d (%s)\n",
		result.DrawioCacheFiles, result.DrawioActiveHashes,
		len(result.DrawioOrphans), formatBytes(result.DrawioBytesRecovered))

	totalOrphans := result.TotalOrphans()
	totalBytes := result.TotalBytesRecovered()

	if totalOrphans == 0 {
		fmt.Println()
		fmt.Println("No orphaned files found.")
		return 0
	}

	fmt.Println()
	fmt.Printf("Total: %d orphaned files (%s)\n", totalOrphans, formatBytes(totalBytes))

	if dryRun {
		fmt.Println()
		fmt.Println("[DRY RUN] Would delete:")
		if len(result.MermaidOrphans) > 0 {
			for _, f := range result.MermaidOrphans {
				fmt.Printf("  - mermaid/%s\n", f)
			}
		}
		if len(result.DrawioOrphans) > 0 {
			for _, f := range result.DrawioOrphans {
				fmt.Printf("  - drawio/%s\n", f)
			}
		}
		return 0
	}

	cacheDir := paths.DocsCachePath(repoRoot)
	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		log.Errorf("Error deleting orphans: %v", err)
		return 1
	}

	fmt.Println()
	fmt.Printf("Deleted %d files, recovered %s\n", deleted, formatBytes(totalBytes))
	return 0
}
