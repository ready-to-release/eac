// Command: show ghosts
// Short: Display discovered ghost entities in human-readable format
// Long: Shows ghosts grouped by module with summary statistics.
// Long: The report includes:
// Long:   - Summary statistics (total, files, directories, by module)
// Long:   - Ghosts grouped by owning module
// Long:   - Unowned ghosts section
// Long:
// Long: Ghost entities enable dark launching, L4 monitoring, and feature toggles.
// Long:
// Long: Filter Examples:
// Long:   show ghosts                      # All ghosts
// Long:   show ghosts --module core        # Ghosts in core module
// Long:   show ghosts --type directory     # Only ghost directories
// Long:
// Long: Expected Output:
// Long: Markdown-formatted report with summary table and ghosts by module.
// Flag.type: type=string, usage=Filter by type (file, directory)
// Flag.module: type=string, usage=Filter to ghosts in specific module
// Flag.unowned: type=bool, usage=Only show unowned ghosts
package show

import (
	"fmt"
	"os"
	"sort"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/ghost"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(ShowGhosts)
}

func ShowGhosts() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	filters := parseShowGhostFilterFlags(os.Args[1:])

	cache := repository.NewFileCache(workspaceRoot, true)
	moduleRegistry, _ := modules.LoadFromWorkspace(workspaceRoot)

	alias := "ghost"
	if cfg := config.Global(); cfg != nil && cfg.Repository != nil {
		if a := cfg.Repository.Repository.GhostTracking.Alias; a != "" {
			alias = a
		}
	}

	report, err := ghost.BuildReport(cache, moduleRegistry, alias)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Apply filters using shared ghost.Filter function
	report.Ghosts = ghost.Filter(report.Ghosts, filters)
	report.Summary = ghost.BuildSummary(report.Ghosts)

	if len(report.Ghosts) == 0 {
		fmt.Println("No ghost entities found.")
		fmt.Printf("\nConfiguration: prefix=%q\n", report.Config.Alias)
		return 0
	}

	fmt.Println("# Ghost Tracking Report")
	fmt.Println()

	printGhostSummary(report.Summary, report.Config)
	printGhostsByModule(report.Ghosts)
	printUnownedGhosts(report.Ghosts)

	return 0
}

// parseShowGhostFilterFlags extracts filter flags from command arguments.
func parseShowGhostFilterFlags(args []string) ghost.FilterOptions {
	var opts ghost.FilterOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				opts.Type = args[i+1]
				i++
			}
		case "--module":
			if i+1 < len(args) {
				opts.Module = args[i+1]
				i++
			}
		case "--unowned":
			opts.Unowned = true
		}
	}
	return opts
}

func printGhostSummary(summary ghost.GhostSummary, cfg ghost.GhostConfigSummary) {
	fmt.Println("## Summary")
	fmt.Println()

	tb := render.NewTableBuilder().
		WithHeaders("Metric", "Value").
		AddRow("Total Ghosts", fmt.Sprintf("%d", summary.Total)).
		AddRow("Files", fmt.Sprintf("%d", summary.Files)).
		AddRow("Directories", fmt.Sprintf("%d", summary.Directories)).
		AddRow("Modules with Ghosts", fmt.Sprintf("%d", len(summary.ByModule))).
		AddRow("Unowned", fmt.Sprintf("%d", summary.Unowned))

	fmt.Println(tb.Build())
	fmt.Println()

	fmt.Println("### Configuration")
	fmt.Printf("- Prefix: `%s`\n", cfg.Alias)
	fmt.Printf("- Patterns: %v\n", cfg.Patterns)
	fmt.Println()
}

func printGhostsByModule(ghosts []ghost.Ghost) {
	byModule := make(map[string][]ghost.Ghost)
	for _, g := range ghosts {
		if g.Module != "" {
			byModule[g.Module] = append(byModule[g.Module], g)
		}
	}

	if len(byModule) == 0 {
		return
	}

	fmt.Println("## Ghosts by Module")
	fmt.Println()

	var moduleNames []string
	for m := range byModule {
		moduleNames = append(moduleNames, m)
	}
	sort.Strings(moduleNames)

	for _, module := range moduleNames {
		gs := byModule[module]
		fmt.Printf("### %s (%d)\n\n", module, len(gs))

		tb := render.NewTableBuilder().WithHeaders("Path", "Type", "Ghost Name")
		for _, g := range gs {
			tb.AddRow(fmt.Sprintf("`%s`", g.Path), string(g.Type), g.GhostName)
		}

		fmt.Println(tb.Build())
		fmt.Println()
	}
}

func printUnownedGhosts(ghosts []ghost.Ghost) {
	var unowned []ghost.Ghost
	for _, g := range ghosts {
		if g.Module == "" {
			unowned = append(unowned, g)
		}
	}

	if len(unowned) == 0 {
		return
	}

	fmt.Println("## Unowned Ghosts")
	fmt.Println()
	fmt.Println("These ghosts are not associated with any module:")
	fmt.Println()

	tb := render.NewTableBuilder().WithHeaders("Path", "Type", "Ghost Name")
	for _, g := range unowned {
		tb.AddRow(fmt.Sprintf("`%s`", g.Path), string(g.Type), g.GhostName)
	}

	fmt.Println(tb.Build())
	fmt.Println()
}
