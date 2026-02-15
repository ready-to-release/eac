package get

import (
	"context"
	"fmt"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/ghost"
	"github.com/ready-to-release/eac/go/core/repository"
)

type getGhostsCommand struct{}

var _ core.SimpleCommandPort = (*getGhostsCommand)(nil)

func (c *getGhostsCommand) Name() string { return "get ghosts" }

func (c *getGhostsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-ghosts",
		Short:         "Get discovered ghost entities in structured format",
		Long:          "Returns structured data about all ghost (dark launch) entities discovered\nin the repository. Ghosts are files/directories matching the configured\nnaming convention (default: ghost-* prefix).\n\nGhost entities enable:\n  - Dark launching: Code deployed but inactive\n  - L4 monitoring: Hidden observability probes\n  - Feature toggles: Without a full feature flag system\n\nFilter Examples:\n  get ghosts                      # All ghosts\n  get ghosts --type file          # Only ghost files\n  get ghosts --type directory     # Only ghost directories\n  get ghosts --module core        # Ghosts in core module\n  get ghosts --unowned            # Ghosts not owned by any module\n\nExpected Output:\nGhostReport with ghosts list, summary statistics, and effective config.",
		Flags: []core.FlagSpec{
			{Name: "as-yaml", Type: "bool", DefaultValue: "", Usage: "Output as YAML (default format)"},
			{Name: "as-json", Type: "bool", DefaultValue: "", Usage: "Output as JSON"},
			{Name: "as-toml", Type: "bool", DefaultValue: "", Usage: "Output as TOML"},
			{Name: "type", Type: "string", DefaultValue: "", Usage: "Filter by type (file, directory)"},
			{Name: "module", Type: "string", DefaultValue: "", Usage: "Filter to ghosts in specific module"},
			{Name: "unowned", Type: "bool", DefaultValue: "", Usage: "Only show unowned ghosts"},
		},
	}
}

func (c *getGhostsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetGhosts()
}

func GetGhosts() int {
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	filters := parseGhostFilterFlags(os.Args[1:])

	return internal.ExecuteGetCommand(func() (interface{}, error) {
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
			return nil, err
		}

		// Apply filters using shared ghost.Filter function
		report.Ghosts = ghost.Filter(report.Ghosts, filters)
		report.Summary = ghost.BuildSummary(report.Ghosts)

		return report, nil
	})
}

// parseGhostFilterFlags extracts filter flags from command arguments.
func parseGhostFilterFlags(args []string) ghost.FilterOptions {
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
