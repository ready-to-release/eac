// Command: get dependencies
// Short: Get direct dependencies of a module
// Usage: get dependencies <module>
//
//	--as-yaml: Output as YAML (default)
//	--as-json: Output as JSON
//	--format: Output format (list, space, json)
//
// Long:
// Long: Returns the direct dependencies of a module.
// Long:
// Long: Expected Output:
// Long:   - List of module monikers this module depends on
// Long:
// Long: Examples:
// Long:   get dependencies r2r-cli                    # YAML list
// Long:   get dependencies r2r-cli --format space     # Space-separated: "dep1 dep2 dep3"
// Long:   get dependencies r2r-cli --format list      # One per line
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/get/internal"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func init() {
	registry.Register(GetDependenciesModule)
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
