// Command: get execution order
// Description: Get execution order for specific modules based on dependencies
// Usage: get execution order <moniker1> <moniker2> ...
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --no-deps: Don't expand to include transitive dependencies (only order the specified modules)
// Long:
// Long: Expected Output:
// Long: YAML ordered list of modules for execution, topologically sorted based on dependencies.
// Long: By default includes transitive dependencies. Use --no-deps to only order specified modules.
package get

import (
	"fmt"
	"os"

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
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Collect module monikers from command line args and parse flags
	var monikers []string
	includeDeps := true // Default: expand dependencies
	skipNext := false
	for i, arg := range os.Args {
		if i < 4 { // Skip program name, "get", "execution", "order"
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		// Parse --no-deps flag
		if arg == "--no-deps" {
			includeDeps = false
			continue
		}
		// Skip output format flags
		if arg == "--as-yaml" || arg == "--as-json" || arg == "--as-toml" {
			continue
		}
		// Skip flag values
		if i > 0 && (os.Args[i-1] == "--as-yaml" || os.Args[i-1] == "--as-json" || os.Args[i-1] == "--as-toml") {
			continue
		}
		monikers = append(monikers, arg)
	}

	// Use the shared get command helper
	return internal.ExecuteGetCommand(func() (interface{}, error) {
		plan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, includeDeps)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate execution order: %w", err)
		}
		return plan, nil
	})
}
