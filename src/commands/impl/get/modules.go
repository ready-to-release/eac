// Command: get modules
// Description: Get all module contracts in the repository
// Flags:
//   --as-yaml: Output as YAML (default)
//   --as-json: Output as JSON
//   --as-toml: Output as TOML
//   --as-<name>: Output using custom renderer (e.g., --as-summary, --as-count)
// HasSideEffects: false
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/impl/get/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/contracts/reports"
	"github.com/ready-to-release/eac/src/core/repository"
)

func init() {
	registry.Register(GetModules)
}

func GetModules() int {
	// Get repository root
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to find repository root: %v\n", err)
		return 1
	}

	// Use the shared get command helper
	return get.ExecuteGetCommand(func() (interface{}, error) {
		report, err := reports.GetModuleContracts(workspaceRoot, "0.1.0")
		if err != nil {
			return nil, err
		}
		return report.Modules, nil
	})
}
