// Command: get execution order
// Description: Get execution order for specific modules based on dependencies
// Usage: get execution order <moniker1> <moniker2> ...
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
// HasSideEffects: false
package get

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/core/repository"
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

	// Collect module monikers from command line args
	var monikers []string
	skipNext := false
	for i, arg := range os.Args {
		if i < 3 { // Skip program name and command
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		// Skip flags
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
	return get.ExecuteGetCommand(func() (interface{}, error) {
		plan, err := repository.CalculateExecutionOrder(monikers, workspaceRoot, "0.1.0")
		if err != nil {
			return nil, fmt.Errorf("failed to calculate execution order: %w", err)
		}
		return plan, nil
	})
}
