// Command: get ghosts
// Short: Get discovered ghost entities in structured format
// Long: Returns structured data about all ghost (dark launch) entities discovered
// Long: in the repository. Ghosts are files/directories matching the configured
// Long: naming convention (default: ghost-* prefix).
// Long:
// Long: Ghost entities enable:
// Long:   - Dark launching: Code deployed but inactive
// Long:   - L4 monitoring: Hidden observability probes
// Long:   - Feature toggles: Without a full feature flag system
// Long:
// Long: Filter Examples:
// Long:   get ghosts                      # All ghosts
// Long:   get ghosts --type file          # Only ghost files
// Long:   get ghosts --type directory     # Only ghost directories
// Long:   get ghosts --module core        # Ghosts in core module
// Long:   get ghosts --unowned            # Ghosts not owned by any module
// Long:
// Long: Expected Output:
// Long: GhostReport with ghosts list, summary statistics, and effective config.
// Flag.as-yaml: type=bool, usage=Output as YAML (default format)
// Flag.as-json: type=bool, usage=Output as JSON
// Flag.as-toml: type=bool, usage=Output as TOML
// Flag.type: type=string, usage=Filter by type (file, directory)
// Flag.module: type=string, usage=Filter to ghosts in specific module
// Flag.unowned: type=bool, usage=Only show unowned ghosts
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/ghost"
	"github.com/ready-to-release/eac/go/core/repository"
)

func init() {
	registry.Register(GetGhosts)
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
