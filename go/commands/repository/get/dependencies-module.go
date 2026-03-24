// Usage: get dependencies <module>
package get

import (
	"context"
	"fmt"
	"os"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/commands/repository/get/internal"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

type getDependenciesModuleCommand struct{}

var _ core.SimpleCommandPort = (*getDependenciesModuleCommand)(nil)

func (c *getDependenciesModuleCommand) Name() string { return "get dependencies" }

func (c *getDependenciesModuleCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "get-dependencies",
		Short:         "Get direct dependencies of a module",
		Long: "Returns the direct dependencies of a module.",
		Notes: "Expected Output:\n  - List of module monikers this module depends on",
		Examples: []string{
			"eac get dependencies clie                 # YAML list",
			"eac get dependencies clie --format space  # Space-separated: \"dep1 dep2 dep3\"",
			"eac get dependencies clie --format list   # One per line",
		},
		Args: "<module>",
		Flags: []core.FlagSpec{
			{Name: "format", Type: "string", Usage: "Output format (list, space, json)"},
		},
	}
}

func (c *getDependenciesModuleCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return GetDependenciesModule()
}

// DependenciesOutput represents the output structure.
type DependenciesOutput struct {
	Module       string   `json:"module" yaml:"module"`
	Dependencies []string `json:"dependencies" yaml:"dependencies"`
}

func GetDependenciesModule() int {
	// Parse arguments
	module := ""
	format := ""

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		case !strings.HasPrefix(arg, "--") && module == "":
			module = arg
		}
	}

	if module == "" {
		fmt.Fprintln(os.Stderr, "Error: module moniker required")
		fmt.Fprintln(os.Stderr, "Usage: get dependencies <module> [--format space|list]")
		return 1
	}

	// Load module registry
	moduleRegistry, err := modules.LoadFromWorkspace("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load modules: %v\n", err)
		return 1
	}

	// Get module
	mod, exists := moduleRegistry.Get(module)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: module '%s' not found\n", module)
		return 1
	}

	deps := mod.DependsOn
	if deps == nil {
		deps = []string{}
	}

	// Handle special output formats
	switch format {
	case "space":
		fmt.Println(strings.Join(deps, " "))
		return 0
	case "list":
		for _, dep := range deps {
			fmt.Println(dep)
		}
		return 0
	}

	// Default: use standard get command helper for YAML/JSON output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return DependenciesOutput{
			Module:       module,
			Dependencies: deps,
		}, nil
	})
}
