// Command: get execution-order
// Description: Get execution order for specific modules based on dependencies
// Usage: get execution-order <moniker1> <moniker2> ...
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --format: Output format (list = one per line, space = space-separated)
//   --skip-depm: Don't expand to include transitive module dependencies (only order the specified modules)
// Long:
// Long: Expected Output:
// Long: YAML ordered list of modules for execution, topologically sorted based on dependencies.
// Long: By default includes transitive dependencies. Use --skip-depm to only order specified modules.
// Long:
// Long: Example:
// Long:   get execution-order r2r-cli                     # YAML output with deps
// Long:   get execution-order r2r-cli --format list       # One module per line
// Long:   get execution-order r2r-cli --format space      # Space-separated
// Long:   for mod in $(get execution-order r2r-cli --format list); do ... done
package get

import (
	"fmt"
	"os"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/impl/get/internal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(GetExecutionOrder)
}

func GetExecutionOrder() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Collect module monikers from command line args and parse flags
	var monikers []string
	includeDeps := true // Default: expand dependencies
	format := ""        // list, space, or empty for default
	for i := 4; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--skip-depm":
			includeDeps = false
		case arg == "--format" && i+1 < len(os.Args):
			format = os.Args[i+1]
			i++
		case arg == "--as-yaml" || arg == "--as-json" || arg == "--as-toml":
			// Handled by internal.ExecuteGetCommand
		case !strings.HasPrefix(arg, "--"):
			monikers = append(monikers, arg)
		}
	}

	// Calculate execution order
	plan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, includeDeps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to calculate execution order: %v\n", err)
		return 1
	}

	// Handle special formats directly
	switch format {
	case "list":
		for _, mod := range plan.ExecutionOrder {
			fmt.Println(mod)
		}
		return 0
	case "space":
		fmt.Println(strings.Join(plan.ExecutionOrder, " "))
		return 0
	}

	// Use the shared get command helper for structured output
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		return plan, nil
	})
}
