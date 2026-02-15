package structurizr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	design "github.com/ready-to-release/eac/go/commands/repository/design/helper"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type updateStructurizrCommand struct{}

var _ core.SimpleCommandPort = (*updateStructurizrCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&updateStructurizrCommand{},
	}
}

func (c *updateStructurizrCommand) Name() string { return "update structurizr" }

func (c *updateStructurizrCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "update-structurizr",
		Short:         "Update Structurizr diagram cache from workspace.dsl files",
		Long:          "Exports all Structurizr views from workspace.dsl files to SVG format\nand stores them in the .cache/eac/structurizr/ acceleration cache.\n\nExpected Output:\n  - SVG files in .cache/eac/structurizr/\n  - Cache status summary showing hits/misses",
		Flags: []core.FlagSpec{
			{Name: "module", Shorthand: "m", Type: "string", Usage: "Export specific module only"},
			{Name: "force", Shorthand: "f", Type: "bool", DefaultValue: "false", Usage: "Force re-export even if cache is current"},
			{Name: "verbose", Shorthand: "v", Type: "bool", DefaultValue: "false", Usage: "Show detailed progress for each module"},
			{Name: "dry-run", Type: "bool", DefaultValue: "false", Usage: "Show what would be exported without actually exporting"},
		},
	}
}

func (c *updateStructurizrCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return UpdateStructurizr()
}

var log = logging.C()
// UpdateStructurizr exports Structurizr workspace views to SVG cache.
func UpdateStructurizr() int {
	// Validate flags
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	module := ""
	force := false
	verbose := false
	dryRun := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m", "--module":
			if i+1 < len(args) {
				module = args[i+1]
				i++
			}
		case "-f", "--force":
			force = true
		case "-v", "--verbose":
			verbose = true
		case "--dry-run":
			dryRun = true
		}
	}

	// Get repo root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Println("Updating Structurizr diagram cache...")

	// Create exporter
	exporter, err := design.NewExporter()
	if err != nil {
		log.Errorf("Error creating exporter: %v", err)
		return 1
	}

	// Check Docker availability
	if !dryRun && !exporter.IsDockerRunning() {
		log.Errorf("Error: Docker is not available but required for Structurizr export")
		log.Errorf("Ensure Docker is installed and the daemon is running")
		return 1
	}

	// Get list of modules to process
	var modulesToProcess []string
	if module != "" {
		// Single module specified
		modulesToProcess = []string{module}
	} else {
		// Get all modules with workspaces
		modules, err := design.ListModulesWithWorkspaces()
		if err != nil {
			log.Errorf("Error listing modules: %v", err)
			return 1
		}
		modulesToProcess = modules
	}

	if len(modulesToProcess) == 0 {
		fmt.Println("No modules with workspace.dsl files found.")
		return 0
	}

	fmt.Printf("Found %d module(s) with workspace.dsl files\n", len(modulesToProcess))

	// Ensure cache directory exists
	cacheDir := paths.StructurizrAccelCachePath(repoRoot)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		log.Errorf("Error creating cache directory: %v", err)
		return 1
	}

	// Check cache status for each module
	type moduleStatus struct {
		Name       string
		DSLHash    string
		CacheHit   bool
		CachedSVGs []string
	}

	statuses := make([]moduleStatus, 0, len(modulesToProcess))
	cacheHits := 0
	cacheMisses := 0

	for _, modName := range modulesToProcess {
		// Get DSL hash
		dslHash, err := design.GetModuleDSLHash(modName)
		if err != nil {
			if verbose {
				fmt.Printf("  Skipping %s: %v\n", modName, err)
			}
			continue
		}

		// Check for existing cached SVGs with matching hash
		cachedSVGs := findCachedSVGs(cacheDir, modName, dslHash)

		status := moduleStatus{
			Name:       modName,
			DSLHash:    dslHash,
			CacheHit:   len(cachedSVGs) > 0 && !force,
			CachedSVGs: cachedSVGs,
		}

		if status.CacheHit {
			cacheHits++
		} else {
			cacheMisses++
		}

		statuses = append(statuses, status)

		if verbose {
			if status.CacheHit {
				fmt.Printf("  %s: cached (%d SVGs, hash %s)\n", modName, len(cachedSVGs), dslHash)
			} else if force {
				fmt.Printf("  %s: force re-export (hash %s)\n", modName, dslHash)
			} else {
				fmt.Printf("  %s: needs export (hash %s)\n", modName, dslHash)
			}
		}
	}

	fmt.Printf("Cache status: %d hits, %d misses (%.1f%% hit rate)\n",
		cacheHits, cacheMisses,
		float64(cacheHits)/float64(len(statuses))*100)

	if cacheMisses == 0 && !force {
		fmt.Println("All Structurizr diagrams are cached. Nothing to export.")
		return 0
	}

	if dryRun {
		fmt.Println("\n[DRY RUN] Would export the following modules:")
		for _, status := range statuses {
			if !status.CacheHit {
				fmt.Printf("  - %s (hash %s)\n", status.Name, status.DSLHash)
			}
		}
		return 0
	}

	// Export modules that need updating
	fmt.Printf("Exporting %d module(s)...\n", cacheMisses)
	exported := 0
	failed := 0
	totalViews := 0

	for _, status := range statuses {
		if status.CacheHit {
			continue
		}

		if verbose {
			fmt.Printf("  Exporting %s...\n", status.Name)
		}

		// Export module
		result, err := exporter.ExportModule(status.Name)
		if err != nil {
			log.Errorf("  Error exporting %s: %v", status.Name, err)
			failed++
			continue
		}

		if result.Error != nil {
			log.Errorf("  Error exporting %s: %v", status.Name, result.Error)
			failed++
			continue
		}

		exported++
		totalViews += len(result.Views)

		if verbose {
			fmt.Printf("  Exported %d view(s) from %s in %v\n",
				len(result.Views), status.Name, result.ExecutionTime)
			for _, view := range result.Views {
				fmt.Printf("    - %s\n", filepath.Base(view.SVGPath))
			}
		}
	}

	fmt.Println()
	if failed > 0 {
		fmt.Printf("Completed with errors: %d module(s) exported (%d views), %d failed\n",
			exported, totalViews, failed)
		return 1
	}

	if exported > 0 {
		fmt.Printf("Successfully exported %d module(s) (%d views total)\n", exported, totalViews)
	}

	fmt.Printf("Cache location: %s\n", cacheDir)
	return 0
}

// findCachedSVGs returns all cached SVG files for a module with matching hash.
func findCachedSVGs(cacheDir, moduleName, dslHash string) []string {
	var matches []string

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return matches
	}

	// Pattern: {module}_{viewKey}_{hash}.svg
	prefix := moduleName + "_"
	suffix := "_" + dslHash + ".svg"

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
			matches = append(matches, filepath.Join(cacheDir, name))
		}
	}

	return matches
}
