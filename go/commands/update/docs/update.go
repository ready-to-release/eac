package docs

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateDocsCommand struct{}

var _ core.SimpleCommandPort = (*updateDocsCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateDocsCommand{},
	}
}

func (c *updateDocsCommand) Name() string { return "update docs" }

func (c *updateDocsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-docs",
		Short:         "Sync documentation assets and command references",
		Long:          "Updates documentation by:\n  - Rendering mermaid diagrams to cache\n  - Optimizing drawio images to cache\n  - Syncing command reference docs (creates missing, removes orphans)\n\nUse --area to run specific areas (for testing).\nUse --dry-run to preview changes without applying them.\n\nCache Pruning:\n  Use --prune-cache to remove orphaned mermaid/drawio cache files.",
		Flags: []core.FlagSpec{
			{Name: "area", Type: "string", Usage: "Area to update (mermaid, drawio, command-refs, all). Repeatable."},
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Preview changes without applying them"},
			{Name: "force", Type: "bool", DefaultValue: "false", Usage: "Force re-render/re-optimize all assets even if cached"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed progress"},
			{Name: "prune-cache", Type: "bool", DefaultValue: "false", Usage: "Remove orphaned cache files (mermaid/drawio)"},
		},
	}
}

func (c *updateDocsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateDocs()
}

var log = logging.C()
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
		cacheDir := paths.CacheRootPath(repoRoot)
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

	cacheDir := paths.CacheRootPath(repoRoot)
	deleted, err := DeleteOrphans(result, cacheDir)
	if err != nil {
		log.Errorf("Error deleting orphans: %v", err)
		return 1
	}

	fmt.Println()
	fmt.Printf("Deleted %d files, recovered %s\n", deleted, formatBytes(totalBytes))
	return 0
}
